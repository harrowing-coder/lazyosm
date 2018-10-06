package osm

/*
The following code contains all the implementations needed for reading primitive blocks.
Each of these functions is sort of a itermediary between data types but you can see all of
them ahve sort of the same implementation.

The LazyPrimitiveBlock is effectively are mapping struct and is very important for the whole implementation.a
It gives us the context we need to add id ranges to the idmap, and read back the relevant bytes after we've passed it.a


*/

import (
	"fmt"
	"github.com/murphy214/pbf"
)

func a() {
	fmt.Println()
}

type PrimitiveBlock struct {
	StringTable []string
	//GroupPBF *pbf.PBF // a set of pbfs associated with each group
	GroupIndex [2]int // the index position of a set of features from the og file
	GroupType  int
	Buf        *pbf.PBF
	Config     Config
}

func NewPrimitiveBlock(pbfval *pbf.PBF) *PrimitiveBlock {
	primblock := &PrimitiveBlock{}
	//var endpos int
	key, val := pbfval.ReadKey()
	if key == 1 && val == 2 {
		size := pbfval.ReadVarint()
		endpos := pbfval.Pos + size
		for pbfval.Pos < endpos {
			_, _ = pbfval.ReadKey()
			primblock.StringTable = append(primblock.StringTable, pbfval.ReadString())
		}

		pbfval.Pos = endpos
		key, val = pbfval.ReadKey()
	}
	if key == 2 && val == 2 {
		// iterating through each pbf group
		endpos := pbfval.Pos + pbfval.ReadVarint()
		grouptype, _ := pbfval.ReadKey()
		if grouptype == 2 {
			pbfval.ReadVarint()
		} else if grouptype == 3 {
			pbfval.Pos -= 1
		}

		primblock.GroupIndex = [2]int{pbfval.Pos, endpos}

		primblock.GroupType = int(grouptype)
		pbfval.Pos = endpos
		key, val = pbfval.ReadKey()
	}
	if key == 100 {
		primblock.Config = NewConfig()
	}

	primblock.Buf = pbfval

	return primblock
}

type LazyPrimitiveBlock struct {
	Type     string // the type of block
	IdRange  [2]int // the id range of a dense node (if applicable)
	FilePos  [2]int // file position
	BufPos   [2]int // the positon of the block
	Position int    // position in which the block occurs within the file
	TagsBool bool
}

func ReadLazyPrimitiveBlock(pbfval *pbf.PBF) LazyPrimitiveBlock {
	var lazyblock LazyPrimitiveBlock
	//var endpos int
	key, val := pbfval.ReadKey()
	if key == 1 && val == 2 {
		size := pbfval.ReadVarint()
		endpos := pbfval.Pos + size
		pbfval.Pos = endpos
		key, val = pbfval.ReadKey()
	}
	if key == 2 && val == 2 {
		// iterating through each pbf group
		endpos := pbfval.Pos + pbfval.ReadVarint()
		grouptype, _ := pbfval.ReadKey()
		if grouptype == 1 {
			lazyblock.Type = "Nodes"
			pbfval.Pos -= 1
		} else if grouptype == 2 {
			pbfval.ReadVarint()
			lazyblock.Type = "DenseNodes"
		} else if grouptype == 3 {
			lazyblock.Type = "Ways"
			pbfval.Pos -= 1
		} else if grouptype == 4 {
			lazyblock.Type = "Relations"
			pbfval.Pos -= 1
		} else if grouptype == 5 {
			lazyblock.Type = "Changesets"
			pbfval.Pos -= 1
		}
		lazyblock.BufPos = [2]int{pbfval.Pos, endpos}
		if lazyblock.Type == "DenseNodes" {
			start, end, boolval := LazyDenseNode(pbfval)
			lazyblock.IdRange = [2]int{start, end}
			lazyblock.TagsBool = boolval
		} else if lazyblock.Type == "Ways" {
			start, end := LazyWayRange(pbfval)
			lazyblock.IdRange = [2]int{start, end}
		}

	}

	return lazyblock
}

func NewPrimitiveBlockLazy(pbfval *pbf.PBF) *PrimitiveBlock {
	primblock := &PrimitiveBlock{}
	//var endpos int

	key, val := pbfval.ReadKey()
	if key == 1 && val == 2 {

		size := pbfval.ReadVarint()
		endpos := pbfval.Pos + size
		pbfval.Pos = endpos
		//primblock.StringTable = pbfval.ReadPackedString()
		key, val = pbfval.ReadKey()
	}
	if key == 2 && val == 2 {
		// iterating through each pbf group
		endpos := pbfval.Pos + pbfval.ReadVarint()
		grouptype, _ := pbfval.ReadKey()
		if grouptype == 2 {
			pbfval.ReadVarint()
		} else if grouptype == 3 {
			pbfval.Pos -= 1
		}
		//fmt.Println(pbfval.Pos, pbfval.Length)

		primblock.GroupIndex = [2]int{pbfval.Pos, endpos}

		primblock.GroupType = int(grouptype)
		pbfval.Pos = endpos
		key, val = pbfval.ReadKey()
	}
	if key == 100 {
		primblock.Config = NewConfig()
	}

	primblock.Buf = pbfval

	return primblock
}
