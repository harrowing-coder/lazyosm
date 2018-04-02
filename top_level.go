package top_level

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/murphy214/pbf"
	"io"
	"os"
	"sync"
	"time"
	//"github.com/paulmach/osm"
	osmpbf "./osmpbf"
	g "github.com/murphy214/geobuf"
	m "github.com/murphy214/mercantile"
)

const (
	MaxBlobHeaderSize = 64 * 1024
	MaxBlobSize       = 32 * 1024 * 1024
)

var (
	parseCapabilities = map[string]bool{
		"OsmSchema-V0.6":        true,
		"DenseNodes":            true,
		"HistoricalInformation": true,
	}
)

// osm block data types
const (
	osmHeaderType = "OSMHeader"
	osmDataType   = "OSMData"
)

// Header contains the contents of the header in the pbf file.
type Header struct {
	Bounds               *m.Extrema
	RequiredFeatures     []string
	OptionalFeatures     []string
	WritingProgram       string
	Source               string
	ReplicationTimestamp time.Time
	ReplicationSeqNum    uint64
	ReplicationBaseURL   string
}

// iPair is the group sent on the chan into the decoder
// goroutines that unzip and decode the pbf from the headerblock.
type iPair struct {
	Offset int64
	Blob   *osmpbf.Blob
	Err    error
}

// A Decoder reads and decodes OpenStreetMap PBF data from an input stream.
type decoder struct {
	Header     *Header
	r          io.Reader
	bytesRead  int64
	Count      int
	DenseNodes map[int]*LazyPrimitiveBlock
	Ways       map[int]*LazyPrimitiveBlock
	Relations  map[int]*LazyPrimitiveBlock
	Nodes      map[int]*LazyPrimitiveBlock
	IdMap      *IdMap
	NodeMap    *NodeMap
	Limit      int
	Geobuf     *g.Writer

	cancel func()
	wg     sync.WaitGroup

	// for data decoders
	inputs []chan<- iPair

	cOffset int64
	cIndex  int
	f       *os.File
}

// newDecoder returns a new decoder that reads from r.
func NewDecoder(f *os.File, limit int) *decoder {
	return &decoder{
		r:          f,
		f:          f,
		DenseNodes: map[int]*LazyPrimitiveBlock{},
		Ways:       map[int]*LazyPrimitiveBlock{},
		Relations:  map[int]*LazyPrimitiveBlock{},
		Nodes:      map[int]*LazyPrimitiveBlock{},
		NodeMap:    NewNodeMap(limit),
		IdMap:      NewIdMap(),
		Geobuf:     g.WriterFileNew("a.geobuf"),
		Limit:      limit,
	}
}

func (dec *decoder) Close() error {
	dec.cancel()
	dec.wg.Wait()
	return nil
}

func (dec *decoder) ReadDataPos(pos [2]int) []byte {
	buf := make([]byte, int(pos[1]-pos[0]))
	dec.f.ReadAt(buf, int64(pos[0]))

	blob := &osmpbf.Blob{}
	if err := proto.Unmarshal(buf, blob); err != nil {
		fmt.Println(err)
	}

	data, err := GetData(blob)
	if err != nil {
		fmt.Println(err)
	}

	return data
}

func (dec *decoder) ReadBlock(lazyprim LazyPrimitiveBlock) *osmpbf.PrimitiveBlock {
	primblock := &osmpbf.PrimitiveBlock{}
	err := proto.Unmarshal(dec.ReadDataPos(lazyprim.FilePos), primblock)
	if err != nil {
		fmt.Println(err)
	}
	return primblock
}

func ReadDecoder(f *os.File, limit int) *decoder {
	d := NewDecoder(f, limit)
	sizeBuf := make([]byte, 4)
	headerBuf := make([]byte, MaxBlobHeaderSize)
	blobBuf := make([]byte, MaxBlobSize)

	// read OSMHeader
	_, blob, _ := d.ReadFileBlock(sizeBuf, headerBuf, blobBuf)
	//_, blob, _ = d.ReadFileBlock(sizeBuf, headerBuf, blobBuf)
	//fmt.Println(headerblob)
	//headerblob2,_ := top_level.GetData(blob)
	header, err := DecodeOSMHeader(blob)
	if err != nil {
		fmt.Println(err)
	}
	d.Header = header

	boolval := true
	var oldsize int
	for boolval {
		d.Count++
		_, _, _ = d.ReadFileBlock(sizeBuf, headerBuf, blobBuf)
		size := len(d.DenseNodes) + len(d.Ways) + len(d.Relations)
		if size == oldsize {
			boolval = false
		}
		oldsize = size
	}

	return d
}

func (dec *decoder) ReadFileBlock(sizeBuf, headerBuf, blobBuf []byte) (*osmpbf.BlobHeader, *osmpbf.Blob, error) {
	blobHeaderSize, err := dec.ReadBlobHeaderSize(sizeBuf)
	if err != nil {
		return nil, nil, err
	}
	headerBuf = headerBuf[:blobHeaderSize]
	blobHeader, err := dec.ReadBlobHeader(headerBuf)
	if err != nil {
		return nil, nil, err
	}

	blobBuf = blobBuf[:blobHeader.GetDatasize()]
	blob, err := dec.ReadBlob(blobHeader, blobBuf)
	if err != nil {
		return nil, nil, err
	}

	dec.bytesRead += 4 + int64(blobHeaderSize)
	index := [2]int{int(dec.bytesRead), int(dec.bytesRead) + int(blobHeader.GetDatasize())}
	//dec.DataIndexs = append(dec.DataIndexs, index)
	primblock := ReadLazyPrimitiveBlock(pbf.NewPBF(dec.ReadDataPos(index)))
	primblock.Position = dec.Count
	primblock.FilePos = index
	switch primblock.Type {
	case "DenseNodes":
		dec.DenseNodes[primblock.Position] = &primblock
		dec.IdMap.AddBlock(&primblock)
	case "Ways":
		dec.Ways[primblock.Position] = &primblock
	case "Relations":
		dec.Relations[primblock.Position] = &primblock
	case "Nodes":
		dec.Nodes[primblock.Position] = &primblock
	}

	dec.bytesRead += int64(blobHeader.GetDatasize())
	return blobHeader, blob, nil
}

func (dec *decoder) ReadBlobHeaderSize(buf []byte) (uint32, error) {
	if _, err := io.ReadFull(dec.r, buf); err != nil {
		return 0, err
	}

	size := binary.BigEndian.Uint32(buf)
	if size >= MaxBlobHeaderSize {
		return 0, errors.New("BlobHeader size >= 64Kb")
	}
	return size, nil
}

func (dec *decoder) ReadBlobHeader(buf []byte) (*osmpbf.BlobHeader, error) {
	if _, err := io.ReadFull(dec.r, buf); err != nil {
		return nil, err
	}

	blobHeader := &osmpbf.BlobHeader{}
	if err := proto.Unmarshal(buf, blobHeader); err != nil {
		return nil, err
	}

	if blobHeader.GetDatasize() >= MaxBlobSize {
		return nil, errors.New("Blob size >= 32Mb")
	}
	return blobHeader, nil
}

func (dec *decoder) ReadBlob(blobHeader *osmpbf.BlobHeader, buf []byte) (*osmpbf.Blob, error) {
	if _, err := io.ReadFull(dec.r, buf); err != nil {
		return nil, err
	}

	blob := &osmpbf.Blob{}
	if err := proto.Unmarshal(buf, blob); err != nil {
		return nil, err
	}
	return blob, nil
}

func GetData(blob *osmpbf.Blob) ([]byte, error) {
	switch {
	case blob.Raw != nil:
		return blob.GetRaw(), nil

	case blob.ZlibData != nil:
		r, err := zlib.NewReader(bytes.NewReader(blob.GetZlibData()))
		if err != nil {
			return nil, err
		}

		// using the bytes.Buffer allows for the preallocation of the necessary space.
		buf := bytes.NewBuffer(make([]byte, 0, blob.GetRawSize()+bytes.MinRead))
		if _, err = buf.ReadFrom(r); err != nil {
			return nil, err
		}

		if buf.Len() != int(blob.GetRawSize()) {
			return nil, fmt.Errorf("raw blob data size %d but expected %d", buf.Len(), blob.GetRawSize())
		}

		return buf.Bytes(), nil
	default:
		return nil, errors.New("unknown blob data")
	}
}

func DecodeOSMHeader(blob *osmpbf.Blob) (*Header, error) {
	data, err := GetData(blob)
	if err != nil {
		return nil, err
	}

	headerBlock := &osmpbf.HeaderBlock{}
	if err := proto.Unmarshal(data, headerBlock); err != nil {
		return nil, err
	}

	// Check we have the parse capabilities
	requiredFeatures := headerBlock.GetRequiredFeatures()
	for _, feature := range requiredFeatures {
		if !parseCapabilities[feature] {
			return nil, fmt.Errorf("parser does not have %s capability", feature)
		}
	}

	// read the header
	header := &Header{
		RequiredFeatures:   headerBlock.GetRequiredFeatures(),
		OptionalFeatures:   headerBlock.GetOptionalFeatures(),
		WritingProgram:     headerBlock.GetWritingprogram(),
		Source:             headerBlock.GetSource(),
		ReplicationBaseURL: headerBlock.GetOsmosisReplicationBaseUrl(),
		ReplicationSeqNum:  uint64(headerBlock.GetOsmosisReplicationSequenceNumber()),
	}

	// convert timestamp epoch seconds to golang time structure if it exists
	if headerBlock.OsmosisReplicationTimestamp != 0 {
		header.ReplicationTimestamp = time.Unix(headerBlock.OsmosisReplicationTimestamp, 0).UTC()
	}
	// read bounding box if it exists
	if headerBlock.Bbox != nil {
		// Units are always in nanodegree and do not obey granularity rules. See osmformat.proto
		header.Bounds = &m.Extrema{
			W: 1e-9 * float64(headerBlock.Bbox.Left),
			E: 1e-9 * float64(headerBlock.Bbox.Right),
			S: 1e-9 * float64(headerBlock.Bbox.Bottom),
			N: 1e-9 * float64(headerBlock.Bbox.Top),
		}
	}

	return header, nil
}
