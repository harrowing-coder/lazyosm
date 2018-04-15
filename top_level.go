package top_level

/*
This code implements the top_level decoder data structure. Much of this code was repurposed from:
https://github.com/qedus/osmpbf

However the code is much different from the original implementation. The code was basically
used as a template to see how to traversed through a pbf file as osm pbfs arent technically a valid proto file.


*/

import (
	osmpbf "./osmpbf"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/gogo/protobuf/proto"
	g "github.com/murphy214/geobuf"
	m "github.com/murphy214/mercantile"
	"github.com/murphy214/pbf"
	"io"
	"os"
	"sync"
	"time"
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
	Header      *Header
	r           io.Reader
	bytesRead   int64
	Count       int
	DenseNodes  map[int]*LazyPrimitiveBlock // data structure for holding lazy dense nodes
	Ways        map[int]*LazyPrimitiveBlock // data structure for holding lazy ways
	Relations   map[int]*LazyPrimitiveBlock // data structure for holding lazy relations
	Nodes       map[int]*LazyPrimitiveBlock // data structure for holding nodes
	IdMap       *IdMap                      // the id map for nodes (see idmap.go)
	WayIdMap    *IdMap                      // the id map for ways (see idmap.go)
	NodeMap     *NodeMap                    // the nodemap for nodes
	RelationMap map[int]string              // the map for indicating whether a way is used in a relation
	Limit       int                         // the limit of how many nodes or ways can be in a map at once
	Geobuf      *g.Writer                   // the output writer that currently exists
	WriteBool   bool
	TotalMemory int // the total memory throughput
	cancel      func()
	wg          sync.WaitGroup

	// for data decoders
	inputs []chan<- iPair

	m sync.Mutex

	cOffset int64
	cIndex  int
	f       *os.File
}

// newDecoder returns a new decoder that reads from r.
func NewDecoder(f *os.File, limit int) *decoder {
	return &decoder{
		r:           f,
		f:           f,
		DenseNodes:  map[int]*LazyPrimitiveBlock{},
		Ways:        map[int]*LazyPrimitiveBlock{},
		Relations:   map[int]*LazyPrimitiveBlock{},
		Nodes:       map[int]*LazyPrimitiveBlock{},
		NodeMap:     NewNodeMap(limit),
		IdMap:       NewIdMap(),
		WayIdMap:    NewIdMap(),
		RelationMap: map[int]string{},
		Geobuf:      g.WriterFileNew("a.geobuf"),
		Limit:       limit,
		WriteBool:   true,
	}
}

func (dec *decoder) Close() error {
	dec.cancel()
	dec.wg.Wait()
	return nil
}

// reads the data at a given positon and decompresses it
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
	dec.TotalMemory += len(data)
	return data
}

// reads the lazy primitive block and returns the true blue
// osm primitive block structure
func (dec *decoder) ReadBlock(lazyprim LazyPrimitiveBlock) *osmpbf.PrimitiveBlock {
	primblock := &osmpbf.PrimitiveBlock{}
	err := proto.Unmarshal(dec.ReadDataPos(lazyprim.FilePos), primblock)
	if err != nil {
		fmt.Println(err)
	}
	return primblock
}

// reads and maps the decoder struct
// limit is the limit of node blocks open at one time.
func ReadDecoder(f *os.File, limit int) *decoder {
	d := NewDecoder(f, limit)
	sizeBuf := make([]byte, 4)
	headerBuf := make([]byte, MaxBlobHeaderSize)
	blobBuf := make([]byte, MaxBlobSize)

	// read OSMHeader
	_, blob, _, index := d.ReadFileBlock(sizeBuf, headerBuf, blobBuf)
	header, err := DecodeOSMHeader(blob)
	if err != nil {
		fmt.Println(err)
	}
	d.Header = header
	fi, _ := f.Stat()
	filesize := int(fi.Size()) / 1000000

	boolval := true
	var oldsize int64
	c := make(chan *LazyPrimitiveBlock)
	increment := 0
	for boolval {
		d.Count++
		_, blob, _, index = d.ReadFileBlock(sizeBuf, headerBuf, blobBuf)
		count := d.Count
		go func(blob *osmpbf.Blob, index [2]int, count int, c chan *LazyPrimitiveBlock) {
			if blob != nil {
				bytevals, err := GetData(blob)
				if err != nil {
					fmt.Println(err)
				}
				primblock := ReadLazyPrimitiveBlock(pbf.NewPBF(bytevals))
				primblock.Position = count
				primblock.FilePos = index
				c <- &primblock

			} else {
				c <- &LazyPrimitiveBlock{}
			}

		}(blob, index, count, c)

		increment++

		// collecting go functions if 1000 have been started or were at the end
		if increment == 1000 || d.bytesRead == oldsize {
			for myc := 0; myc < increment; myc++ {
				primblock := <-c
				switch primblock.Type {
				case "DenseNodes":
					d.DenseNodes[primblock.Position] = primblock
					d.IdMap.AddBlock(primblock)
				case "Ways":
					d.Ways[primblock.Position] = primblock
					d.WayIdMap.AddBlock(primblock)
				case "Relations":
					d.Relations[primblock.Position] = primblock
				case "Nodes":
					d.Nodes[primblock.Position] = primblock
				}
			}
			increment = 0
		}
		if d.bytesRead == oldsize {
			boolval = false
		}
		oldsize = d.bytesRead
		fmt.Printf("\r[%dmb/%dmb] concurrent preliminary read with %d fileblocks total", d.bytesRead/1000000, filesize, d.Count)
	}

	return d
}

func (dec *decoder) ReadFileBlock(sizeBuf, headerBuf, blobBuf []byte) (*osmpbf.BlobHeader, *osmpbf.Blob, error, [2]int) {
	blobHeaderSize, err := dec.ReadBlobHeaderSize(sizeBuf)
	if err != nil {
		return nil, nil, err, [2]int{0, 0}
	}
	headerBuf = headerBuf[:blobHeaderSize]
	blobHeader, err := dec.ReadBlobHeader(headerBuf)
	if err != nil {
		return nil, nil, err, [2]int{0, 0}
	}

	blobBuf = blobBuf[:blobHeader.GetDatasize()]
	blob, err := dec.ReadBlob(blobHeader, blobBuf)
	if err != nil {
		return nil, nil, err, [2]int{0, 0}
	}

	dec.bytesRead += 4 + int64(blobHeaderSize)
	index := [2]int{int(dec.bytesRead), int(dec.bytesRead) + int(blobHeader.GetDatasize())}

	dec.bytesRead += int64(blobHeader.GetDatasize())

	return blobHeader, blob, nil, index
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
