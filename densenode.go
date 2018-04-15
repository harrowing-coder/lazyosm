package top_level

/*
This code contains lazy methods for parsing densenodes in different contexts.

For example one function returns a raw nodeid point map, another
reads a pbf struct (bytes) lazily and returns LazyPrimitiveBlock context struct.
*/

import (
	"fmt"
	m "github.com/murphy214/mercantile"
	"github.com/murphy214/pbf"
)

func c() {
	fmt.Println()
}

//
type Config struct {
	Granularity int // nanodegrees defualt = 0
	LatOffset   int
	LonOffset   int
}

type Node struct {
	Tags  map[string]string
	Point []float64
}

type DenseNode struct {
	NodeMap     map[int]*Node
	DenseInfo   int
	KeyValue    int
	BoundingBox m.Extrema
	Tags        []uint32
	Buf         *pbf.PBF
}

// returns a default config
func NewConfig() Config {
	return Config{Granularity: 100, LatOffset: 0, LonOffset: 0}
}

type tagUnpacker struct {
	stringTable []string
	keysVals    []int32
	index       int
}

// Make tags map from stringtable and array of IDs (used in DenseNodes encoding).
func (tu *tagUnpacker) next() map[string]string {
	tags := make(map[string]string)
	var key, val string
	for tu.index < len(tu.keysVals) {
		keyID := int(tu.keysVals[tu.index])
		tu.index++
		if keyID == 0 {
			break
		}

		valID := int(tu.keysVals[tu.index])
		tu.index++

		if len(tu.stringTable) > keyID {

			key = tu.stringTable[keyID]
		}
		if len(tu.stringTable) > valID {
			val = tu.stringTable[valID]
		}
		if (len(tu.stringTable) > keyID) && (len(tu.stringTable) > valID) {
			tags[key] = val
		}

	}
	return tags
}

// parses a dense node out of a node data structure
func (prim *PrimitiveBlock) NewDenseNode() *DenseNode {
	var tu *tagUnpacker

	densenode := &DenseNode{NodeMap: map[int]*Node{}}
	var idpbf, latpbf, longpbf *pbf.PBF
	//fmt.Println(idpbf,longpbf,latpbf,keypbf)
	key, val := prim.Buf.ReadKey()
	if key == 1 && val == 2 {
		// logic for getting the ids pbf here
		size := prim.Buf.ReadVarint()
		endpos := prim.Buf.Pos + size
		idpbf = pbf.NewPBF(prim.Buf.Pbf[prim.Buf.Pos:endpos])
		prim.Buf.Pos += size
		key, val = prim.Buf.ReadKey()
	}
	if key == 5 && val == 2 {
		// do some shit with dense info here
		size := prim.Buf.ReadVarint()
		densenode.DenseInfo = prim.Buf.Pos
		prim.Buf.Pos += size
		key, val = prim.Buf.ReadKey()
	}
	if key == 8 && val == 2 {
		size := prim.Buf.ReadVarint()
		endpos := prim.Buf.Pos + size
		latpbf = pbf.NewPBF(prim.Buf.Pbf[prim.Buf.Pos:endpos])
		prim.Buf.Pos += size
		key, val = prim.Buf.ReadKey()
	}
	if key == 9 && val == 2 {
		size := prim.Buf.ReadVarint()
		endpos := prim.Buf.Pos + size
		longpbf = pbf.NewPBF(prim.Buf.Pbf[prim.Buf.Pos:endpos])
		prim.Buf.Pos += size
		key, val = prim.Buf.ReadKey()
	}
	if key == 10 && val == 2 {
		densenode.KeyValue = prim.Buf.Pos
		tags := prim.Buf.ReadPackedInt32()
		tu = &tagUnpacker{prim.StringTable, tags, 0}
		key, val = prim.Buf.ReadKey()
	}

	// collecting the node map
	// this is the point of this after all
	var id, lat, long int
	west, south, east, north := 180.0, 90.0, -180.0, -90.0
	var pt []float64

	for i := 0; i < 8000 || prim.Buf.Pos < prim.Buf.Length; i++ {
		tags := tu.next()
		id = id + int(idpbf.ReadSVarint())
		lat = lat + int(latpbf.ReadSVarint())
		long = long + int(longpbf.ReadSVarint())
		// getting the point vlaue
		pt = []float64{
			(float64(prim.Config.LonOffset+(long*prim.Config.Granularity)) * 1e-9),
			(float64(prim.Config.LatOffset+(lat*prim.Config.Granularity)) * 1e-9),
		}

		// adding the node to the nodemap
		densenode.NodeMap[id] = &Node{Point: pt, Tags: tags}

		x, y := pt[0], pt[1]
		// can only be one condition
		// using else if reduces one comparison
		if x < west {
			west = x
		} else if x > east {
			east = x
		}

		if y < south {
			south = y
		} else if y > north {
			north = y
		}

		// currently ignoring the keys for now :P
	}

	bds := m.Extrema{N: north, S: south, E: east, W: west}
	densenode.BoundingBox = bds
	densenode.Buf = prim.Buf

	return densenode
}

// parses a dense node out of a node data structure
// returns the sparse id point map
func (d *decoder) NewDenseNodeMap(lazy *LazyPrimitiveBlock) map[int][]float64 {
	prim := NewPrimitiveBlockLazy(pbf.NewPBF(d.ReadDataPos(lazy.FilePos)))
	prim.Buf.Pos = prim.GroupIndex[0]

	var idpbf, latpbf, longpbf *pbf.PBF
	key, val := prim.Buf.ReadKey()
	if key == 1 && val == 2 {
		// logic for getting the ids pbf here
		size := prim.Buf.ReadVarint()
		endpos := prim.Buf.Pos + size
		idpbf = pbf.NewPBF(prim.Buf.Pbf[prim.Buf.Pos:endpos])
		prim.Buf.Pos += size
		key, val = prim.Buf.ReadKey()
	}
	if key == 5 && val == 2 {
		// do some shit with dense info here
		size := prim.Buf.ReadVarint()
		//densenode.DenseInfo = prim.Buf.Pos
		prim.Buf.Pos += size
		key, val = prim.Buf.ReadKey()
	}
	if key == 8 && val == 2 {
		size := prim.Buf.ReadVarint()
		endpos := prim.Buf.Pos + size
		latpbf = pbf.NewPBF(prim.Buf.Pbf[prim.Buf.Pos:endpos])
		prim.Buf.Pos += size
		key, val = prim.Buf.ReadKey()
	}
	if key == 9 && val == 2 {
		size := prim.Buf.ReadVarint()
		endpos := prim.Buf.Pos + size
		longpbf = pbf.NewPBF(prim.Buf.Pbf[prim.Buf.Pos:endpos])
		prim.Buf.Pos += size
		key, val = prim.Buf.ReadKey()
	}
	if key == 10 && val == 2 {
		prim.Buf.ReadPackedInt32()
		key, val = prim.Buf.ReadKey()
	}

	var id, lat, long int
	var pt []float64
	nodemap := map[int][]float64{}
	//ii := 0
	for i := 0; i < 8000 && idpbf.Pos < idpbf.Length; i++ {
		id = id + int(idpbf.ReadSVarint())
		lat = lat + int(latpbf.ReadSVarint())
		long = long + int(longpbf.ReadSVarint())
		pt = []float64{
			(float64(prim.Config.LonOffset+(long*prim.Config.Granularity)) * 1e-9),
			(float64(prim.Config.LatOffset+(lat*prim.Config.Granularity)) * 1e-9),
		}

		// adding the node to the nodemap
		nodemap[id] = pt

		// currently ignoring the keys for now :P
	}
	return nodemap
}

// parses a dense node out of a node data structure
// from a pbf returns the minimal structure needed for a lazy dense node
// starting and ending positions as well a tag bool
func LazyDenseNode(pbfval *pbf.PBF) (int, int, bool) {
	var idpbf *pbf.PBF
	key, val := pbfval.ReadKey()
	var startid, endid int

	if key == 1 && val == 2 {
		// logic for getting the ids pbf here
		size := pbfval.ReadVarint()
		endpos := pbfval.Pos + size
		idpbf = pbf.NewPBF(pbfval.Pbf[pbfval.Pos:endpos])
		id := 0
		for i := 0; i < 8000 && idpbf.Pos < idpbf.Length; i++ {
			id = id + int(idpbf.ReadSVarint())
			if i == 0 {
				startid = id
			}
		}
		endid = id
		pbfval.Pos = endpos
		key, val = pbfval.ReadKey()
	}
	if key == 5 && val == 2 {
		// do some shit with dense info here
		size := pbfval.ReadVarint()
		//densenode.DenseInfo = pbfval.Pos
		pbfval.Pos += size
		key, val = pbfval.ReadKey()
	}
	if key == 8 && val == 2 {
		size := pbfval.ReadVarint()
		pbfval.Pos += size
		key, val = pbfval.ReadKey()
	}
	if key == 9 && val == 2 {
		size := pbfval.ReadVarint()
		pbfval.Pos += size
		key, val = pbfval.ReadKey()
	}
	if key == 10 && val == 2 {
		startpos := pbfval.Pos
		pbfval.ReadPackedInt32()
		sizevals := pbfval.Pos - startpos
		return startid, endid, 8002 != sizevals
	}

	return 0, 0, false
}
