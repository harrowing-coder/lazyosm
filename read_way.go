package top_level

import (
	"fmt"
	"github.com/murphy214/pbf"
	"github.com/paulmach/go.geojson"
)

func d() {
	fmt.Println()
}

type Way struct {
	Id   int
	Tags map[string]interface{}
	Info int
	Refs []int
}

func DeepCopy(a *geojson.Feature) *geojson.Feature {
	mymap := map[string]interface{}{}
	ehmap := a.Properties
	for k, v := range ehmap {
		mymap[k] = v
	}
	geometry := &geojson.Geometry{}
	*geometry = *a.Geometry
	aa := &geojson.Feature{Properties: mymap, Geometry: geometry, ID: a.ID}
	return aa
}

func (block *PrimitiveBlock) WriteWays(totalmap map[int]*Node) []*geojson.Feature {
	feats := []*geojson.Feature{}
	block.Buf.Pos = block.GroupIndex[0]
	for block.Buf.Pos < block.GroupIndex[1] {
		block.Buf.ReadKey()
		endpos := block.Buf.Pos + block.Buf.ReadVarint()

		//start,end := block.Buf.Pos,block.GroupIndex[1]

		way := block.ReadWay()
		line := make([][]float64, len(way.Refs))
		for pos, ref := range way.Refs {
			line[pos] = totalmap[ref].Point
		}
		block.Buf.Pos = endpos
		feat := geojson.NewFeature(geojson.NewLineStringGeometry(line))
		feat.Properties = way.Tags
		feat.ID = way.Id
		feat2 := DeepCopy(feat)
		feats = append(feats, feat2)
	}
	return feats
}

// create primive block
func (d *decoder) CreatePrimitiveBlock(lazy *LazyPrimitiveBlock) *PrimitiveBlock {
	return &PrimitiveBlock{Buf: pbf.NewPBF(d.ReadDataPos(lazy.FilePos)), GroupIndex: lazy.BufPos, GroupType: 3}
}

func (prim *PrimitiveBlock) ReadWay() *Way {
	key, val := prim.Buf.ReadKey()
	way := &Way{}
	var keys, values []uint32
	// logic for handlign id
	if key == 1 && val == 0 {
		way.Id = int(int64(prim.Buf.ReadUInt64()))
		key, val = prim.Buf.ReadKey()
	}
	// logic for handling tags
	if key == 2 {
		//fmt.Println(feature)
		keys = prim.Buf.ReadPackedUInt32()
		key, _ = prim.Buf.ReadKey()
	}
	// logic for handling features
	if key == 3 {
		values = prim.Buf.ReadPackedUInt32()
		key, _ = prim.Buf.ReadKey()
	}

	way.Tags = make(map[string]interface{})

	for i, keyx := range keys {
		if len(prim.StringTable) > int(keys[i]) && len(prim.StringTable) > int(values[i]) && i < len(values) {
			value := prim.StringTable[values[i]]
			keyval := prim.StringTable[int(keyx)]
			way.Tags[keyval] = value
		}

	}

	if key == 4 {
		size := prim.Buf.ReadVarint()
		way.Info = prim.Buf.Pos
		prim.Buf.Pos += size
		key, _ = prim.Buf.ReadKey()
	}

	// logic for handling geometry
	if key == 8 {

		size := prim.Buf.ReadVarint()
		endpos := prim.Buf.Pos + size
		var x int
		for prim.Buf.Pos < endpos {
			x += int(prim.Buf.ReadSVarint())
			way.Refs = append(way.Refs, x)
		}

		prim.Buf.Pos += size + 1
	}
	return way
}

func (d *decoder) ReadWaysLazy(lazy *LazyPrimitiveBlock, idmap *IdMap) map[int]string {
	prim := d.CreatePrimitiveBlock(lazy)
	prim.Buf.Pos = prim.GroupIndex[0]
	mymap := map[int]string{}

	for prim.Buf.Pos < prim.GroupIndex[1] {
		prim.Buf.ReadKey()
		endpos2 := prim.Buf.Pos + prim.Buf.ReadVarint()

		key, val := prim.Buf.ReadKey()
		//var keys, values []uint32
		// logic for handlign id
		if key == 1 && val == 0 {
			prim.Buf.ReadUInt64()
			key, val = prim.Buf.ReadKey()
		}
		// logic for handling tags
		if key == 2 {
			//fmt.Println(feature)
			size := prim.Buf.ReadVarint()
			prim.Buf.Pos += size
			//keys = prim.Buf.ReadPackedUInt32()
			key, _ = prim.Buf.ReadKey()
		}
		// logic for handling features
		if key == 3 {
			size := prim.Buf.ReadVarint()
			prim.Buf.Pos += size
			key, _ = prim.Buf.ReadKey()
		}

		if key == 4 {
			size := prim.Buf.ReadVarint()
			prim.Buf.Pos += size
			key, _ = prim.Buf.ReadKey()
		}

		// logic for handling geometry
		if key == 8 {

			size := prim.Buf.ReadVarint()
			endpos := prim.Buf.Pos + size
			var x int
			for prim.Buf.Pos < endpos {
				x += int(prim.Buf.ReadSVarint())
				//way.Refs = append(way.Refs, x)
				position := idmap.GetBlock(x)
				mymap[position] = ""
			}

			prim.Buf.Pos += size + 1
		}
		prim.Buf.Pos = endpos2
	}
	return mymap
}

// syncs the nodemap against a give way block and flushes old
// node maps out of memory if needed
func (d *decoder) SyncWaysNodeMap(lazy *LazyPrimitiveBlock, idmap *IdMap) {
	keymap := d.ReadWaysLazy(lazy, idmap)
	keylist := make([]int, len(keymap))
	i := 0
	for k := range keymap {
		keylist[i] = k
		i++
	}
	d.AddUpdates(keylist)
}

// syncs the nodemap against a give way block and flushes old
// node maps out of memory if needed
func (d *decoder) SyncWaysNodeMapMultiple(lazys []*LazyPrimitiveBlock, idmap *IdMap) {
	keymap := map[int]string{}
	for _, lazy := range lazys {
		tempkeymap := d.ReadWaysLazy(lazy, idmap)
		for k, v := range tempkeymap {
			keymap[k] = v
		}
	}
	keylist := make([]int, len(keymap))
	i := 0
	for k := range keymap {
		keylist[i] = k
		i++
	}

	d.AddUpdates(keylist)
}
