package top_level

func (d *decoder) ReadRelationsLazy(lazy *LazyPrimitiveBlock, idmap *IdMap) map[int]int {
	prim := d.CreatePrimitiveBlock(lazy)
	prim.Buf.Pos = prim.GroupIndex[0]
	mymap := map[int]int{}

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
			prim.Buf.Pos = endpos
			key, _ = prim.Buf.ReadKey()
		}

		if key == 9 {
			size := prim.Buf.ReadVarint()
			endpos := prim.Buf.Pos + size
			var x int
			for prim.Buf.Pos < endpos {
				x += int(prim.Buf.ReadSVarint())
				//way.Refs = append(way.Refs, x)
				mymap[x] = 0
			}

		}

		prim.Buf.Pos = endpos2
	}
	return mymap
}
