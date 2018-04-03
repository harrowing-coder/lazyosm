package main

import (
	"./top_level"
	//"./top_level/osmpbf"
	"os"
	//"./top_level/osmpbf"
	//"./osm"
	//"./pbf"

	//"fmt"
	//"github.com/golang/protobuf/proto"
	"github.com/paulmach/go.geojson"
	//"io/ioutil"
	"sort"
	//"io/ioutil"
)

/*
func ReadHeaderBlob(bytevals []byte) *OSMPBF.PrimitiveBlock {

}
*/

func SortKeys(mymap map[int]*top_level.LazyPrimitiveBlock) []int {
	i := 0
	newlist := make([]int, len(mymap))
	for k := range mymap {
		newlist[i] = k
		i++
	}
	sort.Ints(newlist)
	return newlist
}

// func

func main() {
	f, _ := os.Open("wv.pbf")

	d := top_level.ReadDecoder(f, 1000)
	feat := geojson.NewPointFeature([]float64{-90.0, 40.0})
	feat.Properties = map[string]interface{}{"shit": "adfas"}
	d.Geobuf.WriteFeature(feat)
	//d.ProcessFile()
	/*
		count := 0
		fc := &geojson.FeatureCollection{}
		is := []*top_level.LazyPrimitiveBlock{}
		fmt.Println(len(d.Nodes), "shit")
		size := len(d.Ways)
		//sizedensenodes := len(d.DenseNodes)

		// processing dense nodes
		for pos, i := range d.DenseNodes {
			if i.TagsBool {
				is = append(is, i)
				if len(is) == 5 || pos == sizedensenodes-1 {
					d.ProcessMultipleDenseNode(is)
					is = []*top_level.LazyPrimitiveBlock{}
				}

			}
			count += 1
			fmt.Printf("\r[%d/%d] Dense Node Blocks Completed", count, sizedensenodes)
		}

		fmt.Println()
		count = 0
		waylist := SortKeys(d.Ways)
		pos := 0
		for _, key := range waylist {
			//tempmap := d.ReadWaysLazyRelations(i, d.IdMap))
			i := d.Ways[key]
			is = append(is, i)
			if len(is) == 3 || pos == size-1 {
				d.SyncWaysNodeMapMultiple(is, d.IdMap)
				d.ProcessMultiple(is)
				is = []*top_level.LazyPrimitiveBlock{}
			}

			count += 1
			pos += 1
			fmt.Printf("\r[%d/%d] Dense Node Blocks Completed", count, size)
		}
	*/

	relationlist := SortKeys(d.Relations)
	for _, key := range relationlist {
		//fmt.Println(d.Relations[key])

		primblock := d.Relations[key]
		//relations := d.ReadBlock(*primblock).Primitivegroup[0].Relations

		relmap := d.ReadRelationsLazy(primblock)

		totalmap := map[int]string{}
		for k := range relmap {
			val, boolval := d.Ways[k]
			if boolval {
				tempmap := d.ReadWaysLazy(val, d.IdMap)
				for k := range tempmap {
					totalmap[k] = ""
				}
			}
		}

		stringval := make([]*top_level.LazyPrimitiveBlock, len(totalmap))
		i := 0
		for k := range totalmap {
			stringval[i] = d.DenseNodes[k]
			i++
		}

		d.SyncWaysNodeMapMultiple(stringval, d.IdMap)

		pb := d.ReadBlock(*primblock)
		relations := pb.Primitivegroup[0].Relations
		waymap := map[int][]int{}
		for _, way := range relations {
			refs := way.Memids
			oldref := refs[0]
			pos := 1
			newrefs := make([]int, len(refs))
			newrefs[0] = int(refs[0])
			for _, ref := range refs[1:] {
				ref = ref + oldref
				newrefs[pos] = int(ref)
				pos++
				oldref = ref
			}

			for _, i := range newrefs {
				waymap[d.WayIdMap.GetBlock(i)] = append(waymap[d.WayIdMap.GetBlock(i)], i)
			}

		}
		totalwaynodemap := map[int][]int{}
		for k, v := range waymap {
			val, boolval := d.Ways[k]
			if boolval {
				tempwaynodemap := d.ReadWaysLazyList(val, v)
				for k, v := range tempwaynodemap {
					totalwaynodemap[k] = v
				}
			}
			//fmt.Println(k, len(v))
		}

		totalidmap := map[int]string{}
		totallist := make([][][]int, len(relations))

		for position, way := range relations {
			refs := way.Memids
			oldref := refs[0]
			pos := 1
			newrefs := make([]int, len(refs))
			newrefs[0] = int(refs[0])
			for _, ref := range refs[1:] {
				ref = ref + oldref
				newrefs[pos] = int(ref)
				pos++
				oldref = ref
			}

			goodmembers := [][]int{}
			mapmembers := [][]int{}
			for _, i := range newrefs {
				val, boolval := totalwaynodemap[i]
				if boolval {
					if val[0] == val[len(val)-1] {
						goodmembers = append(goodmembers, val)
					} else {
						mapmembers = append(mapmembers, val)
					}

					for _, i := range val {
						totalidmap[d.IdMap.GetBlock(i)] = ""
					}

					//newlist2 = append(newlist2, val)
				}
			}
			//fmt.Println(len(newlist2))

			mapmembers = top_level.Connect_Members(mapmembers)
			mapmembers = append(mapmembers, goodmembers...)

			totallist[position] = mapmembers

		}

		stringval2 := make([]int, len(totalidmap))
		newpos := 0
		for k := range totalidmap {
			stringval2[newpos] = k
			newpos++
		}
		d.AddUpdates(stringval2)

		for pos, way := range relations {
			polygon := [][][]float64{}
			for _, nodelist := range totallist[pos] {
				floatlist := make([][]float64, len(nodelist))
				for positional, i := range nodelist {
					floatlist[positional] = d.GetNode(i)
				}
				polygon = append(polygon, floatlist)
			}
			mymap := map[string]interface{}{}
			for i := range way.Keys {
				keypos, valpos := way.Keys[i], way.Vals[i]
				mymap[pb.Stringtable.S[keypos]] = pb.Stringtable.S[valpos]
			}

			if len(polygon) > 0 {
				feature := geojson.NewPolygonFeature(polygon)
				feature.Properties = mymap
				d.Geobuf.WriteFeature(feature)
			}

			//fmt.Println(len(polygon), way.Id)

		}

		//fmt.Println(waymap)
		//fmt.Println(relations)

	}

	//d.ProcessFile()
}
