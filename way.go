package top_level

/*
This file implements methods for reading ways in different contexts.

Read lazy primitive way block and return all node positions map
Read lazy primitive way block with id inputs and return map[wayid][]nodids{} map
Read lazy primitive way block adn return way upper and lower bound

*/

import (
	"fmt"
	"github.com/murphy214/pbf"
)

func d() {
	fmt.Println()
}

// create primive block
func (d *decoder) CreatePrimitiveBlock(lazy *LazyPrimitiveBlock) *PrimitiveBlock {
	return &PrimitiveBlock{Buf: pbf.NewPBF(d.ReadDataPos(lazy.FilePos)), GroupIndex: lazy.BufPos, GroupType: 3}
}

// a lazy map reads an idmap and returns a map nodeid map
// so reads a way block lazily and returns all the node block positions taht exist within
// this way block
func (d *decoder) ReadWaysLazy(lazy *LazyPrimitiveBlock, idmap *IdMap) map[int]string {
	prim := d.CreatePrimitiveBlock(lazy)
	prim.Buf.Pos = prim.GroupIndex[0]
	mymap := map[int]string{}

	for prim.Buf.Pos < prim.GroupIndex[1] {
		prim.Buf.ReadKey()
		endpos2 := prim.Buf.Pos + prim.Buf.ReadVarint()

		key, val := prim.Buf.ReadKey()
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
				position := idmap.GetBlock(x)
				mymap[position] = ""
			}

			prim.Buf.Pos += size + 1
		}
		prim.Buf.Pos = endpos2
	}
	return mymap
}

// given a set of ids return the map[wayid][]int node id list map
// for the given ids input
func (d *decoder) ReadWaysLazyList(lazy *LazyPrimitiveBlock, ids []int) map[int][]int {
	idmap := map[int]string{}
	for _, i := range ids {
		idmap[i] = ""
	}

	prim := d.CreatePrimitiveBlock(lazy)
	prim.Buf.Pos = prim.GroupIndex[0]
	mymap := map[int][]int{}
	var boolval bool
	var id int
	for prim.Buf.Pos < prim.GroupIndex[1] {
		prim.Buf.ReadKey()
		endpos2 := prim.Buf.Pos + prim.Buf.ReadVarint()

		key, val := prim.Buf.ReadKey()
		// logic for handlign id
		if key == 1 && val == 0 {
			id = int(prim.Buf.ReadUInt64())
			_, boolval = idmap[id]
			key, val = prim.Buf.ReadKey()
		}
		// logic for handling tags
		if key == 2 {
			size := prim.Buf.ReadVarint()
			prim.Buf.Pos += size
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
			if boolval {
				var x int
				var xlist []int
				for prim.Buf.Pos < endpos {
					x += int(prim.Buf.ReadSVarint())
					xlist = append(xlist, x)
				}
				prim.Buf.Pos += size + 1
				mymap[id] = xlist
			} else {
				prim.Buf.Pos = endpos
			}

		}
		prim.Buf.Pos = endpos2
	}
	return mymap
}

// reads a lazyprimitive group and returns the wayid range
// for the given way block
func LazyWayRange(pbfval *pbf.PBF) (int, int) {
	var start, pos, id int
	for pbfval.Pos < pbfval.Length {
		pbfval.ReadKey()
		endpos2 := pbfval.Pos + pbfval.ReadVarint()

		key, val := pbfval.ReadKey()
		// logic for handlign id
		if key == 1 && val == 0 {
			id = int(pbfval.ReadUInt64())
			if pos == 0 {
				start = id
			}
			key, val = pbfval.ReadKey()
		}
		// logic for handling tags
		if key == 2 {
			//fmt.Println(feature)
			size := pbfval.ReadVarint()
			pbfval.Pos += size
			key, _ = pbfval.ReadKey()
		}
		// logic for handling features
		if key == 3 {
			size := pbfval.ReadVarint()
			pbfval.Pos += size
			key, _ = pbfval.ReadKey()
		}

		if key == 4 {
			size := pbfval.ReadVarint()
			pbfval.Pos += size
			key, _ = pbfval.ReadKey()
		}

		// logic for handling geometry
		if key == 8 {

			size := pbfval.ReadVarint()
			endpos := pbfval.Pos + size
			pbfval.Pos = endpos
		}
		pbfval.Pos = endpos2
		pos++
	}

	return start, id
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

// syncs multiple way blocks with an idmap
func (d *decoder) SyncWaysNodeMapMultiple(lazys []*LazyPrimitiveBlock, idmap *IdMap) {
	//s := time.Now()
	keymap := map[int]string{}
	c := make(chan map[int]string)
	current := 0
	for pos, lazy := range lazys {
		go func(lazy *LazyPrimitiveBlock) {
			c <- d.ReadWaysLazy(lazy, idmap)
		}(lazy)
		current++

		if pos%10 == 1 || len(lazys)-1 == pos {
			for i := 0; i < current; i++ {
				tempmap := <-c
				for k, v := range tempmap {
					keymap[k] = v
				}
			}
			current = 0
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
