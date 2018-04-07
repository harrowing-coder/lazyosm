package top_level

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/paulmach/go.geojson"
	"io/ioutil"
	"math"
)

type Poly [][]float64

func Round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}

func RoundPt(pt []float64) []float64 {
	return []float64{Round(pt[0], .5, 6), Round(pt[1], .5, 6)}
}

func (c Poly) Pip(p []float64) bool {
	// Cast ray from p.x towards the right
	intersections := 0
	for i := range c {
		curr := c[i]
		ii := i + 1
		if ii == len(c) {
			ii = 0
		}
		next := c[ii]

		// Is the point out of the edge's bounding box?
		// bottom vertex is inclusive (belongs to edge), top vertex is
		// exclusive (not part of edge) -- i.e. p lies "slightly above
		// the ray"
		bottom, top := curr, next
		if bottom[1] > top[1] {
			bottom, top = top, bottom
		}
		if p[1] < bottom[1] || p[1] > top[1] {
			continue
		}
		// Edge is from curr to next.

		if p[0] >= math.Max(curr[0], next[0]) ||
			next[1] == curr[1] {
			continue
		}

		// Find where the line intersects...
		xint := (p[1]-curr[1])*(next[0]-curr[0])/(next[1]-curr[1]) + curr[0]
		if curr[0] != next[0] && p[0] > xint {
			continue
		}

		intersections++
	}
	return intersections%2 != 0
}

func (poly Poly) Within(inner Poly) bool {
	boolval := true
	for _, pt := range inner {
		if !poly.Pip(pt) {
			boolval = false
			return boolval
		}
	}
	return boolval
}

func Reverse(s []int) []int {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func Collision(ring1 []int, ring2 []int) ([]int, bool, bool) {
	firstid1, lastid1 := ring1[0], ring1[len(ring1)-1]
	firstid2, lastid2 := ring2[0], ring2[len(ring2)-1]
	total := []int{}
	boolval := false
	if firstid1 == firstid2 {
		total = append(ring1, Reverse(ring2)...)
		boolval = true
	} else if firstid1 == lastid2 {
		total = append(ring1, ring2...)
		//boolval = true
	} else if lastid1 == lastid2 {
		total = append(Reverse(ring1), ring2...)
		boolval = true
	} else if lastid1 == firstid2 {
		total = append(ring2, ring1...)
		boolval = true
	}
	if len(total) == 0 {
		return []int{}, false, false
	}

	return total, boolval, total[0] != total[len(total)-1]
}

func Satisfy(member []int) bool {
	return member[0] == member[len(member)-1]
}

func Connect(members [][]int) [][]int {

	membermap := map[int][]int{}
	totalmembers := [][]int{}
	for pos, member := range members {
		if Satisfy(member) {
			totalmembers = append(totalmembers, member)
		} else {
			membermap[pos] = member
		}
	}

	for k, member := range membermap {
		for _, trymember := range membermap {
			newmember, mergebool, satisfy := Collision(member, trymember)
			if mergebool {
				if satisfy {
					totalmembers = append(totalmembers, newmember)
					delete(membermap, k)
				} else {
					membermap[k] = newmember
					//delete(membermap, newk)
				}

			}

		}
	}

	return totalmembers

}

// this method is designed to hackily handle large relations
func (d *decoder) ProcessRelation() {
	relationlist := SortKeys(d.Relations)
	boolval5 := false
	fc := &geojson.FeatureCollection{}
	for _, key := range relationlist[:1] {
		//fmt.Println(d.Relations[key])

		primblock := d.Relations[key]
		//relations := d.ReadBlock(*primblock).Primitivegroup[0].Relations

		relmap := d.ReadRelationsLazy(primblock)

		// lazily leading all the ways we need
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

		// lazily reading all the values to sync
		stringval := make([]*LazyPrimitiveBlock, len(totalmap))
		i := 0
		for k := range totalmap {
			stringval[i] = d.DenseNodes[k]
			i++
		}
		d.SyncWaysNodeMapMultiple(stringval, d.IdMap)

		// reading the primitive relation block
		pb := d.ReadBlock(*primblock)
		relations := pb.Primitivegroup[0].Relations
		waymap := map[int][]int{}
		// building the way map relation table
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

		// creating toal way nodemap
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

		// creating id map to update all needed nodes
		totalidmap := map[int]string{}
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

				vals, boolval := totalwaynodemap[i]
				if boolval {
					for _, nodeid := range vals {
						totalidmap[d.IdMap.GetBlock(nodeid)] = ""
					}
				}

			}
		}

		stringval2 := make([]int, len(totalidmap))
		newpos := 0
		for k := range totalidmap {
			stringval2[newpos] = k
			newpos++
		}
		d.AddUpdates(stringval2)

		//totalidmap := map[int]string{}
		//totallist := make([][][]int, len(relations))
		st := pb.GetStringtable().GetS()

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

			roles := make([]string, len(way.RolesSid))
			for pos, ri := range way.RolesSid {
				roles[pos] = st[int(ri)]
			}

			// processing each role / way relation
			//innermap := map[int][][]float64{}
			//outers := [][][]float64{}

			inners := [][]int{}
			outers2 := [][]int{}

			for i := range newrefs {
				role, wayid := roles[i], newrefs[i]
				// getting nodemap if possible
				nodes, boolval := totalwaynodemap[wayid]

				if boolval {
					/*
						ring := make([][]float64, len(nodes))
						for pos, node := range nodes {
							ring[pos] = RoundPt(d.GetNode(node))
						}
					*/
					/*
						firstpt, lastpt := ring[0], ring[len(ring)-1]
						if !(firstpt[0] == lastpt[0] && firstpt[1] == lastpt[1]) {
							ring = append(ring, firstpt)
						}
					*/
					//ring[len(ring)-1] = ring[0]

					if role == "inner" {
						inners = append(inners, nodes)
					} else if role == "outer" {
						outers2 = append(outers2, nodes)
					}
				}
			}

			//inners = Connect(inners)
			//outers2 = Connect(outers2)
			innermap := map[int][][]float64{}
			outers := [][][]float64{}
			for pos, inner := range inners {
				ring := make([][]float64, len(inner))
				for pos, node := range inner {
					ring[pos] = RoundPt(d.GetNode(node))
				}
				innermap[pos] = ring
			}
			for _, outer := range outers2 {
				ring := make([][]float64, len(outer))
				for pos, node := range outer {
					ring[pos] = RoundPt(d.GetNode(node))
				}
				outers = append(outers, ring)
			}

			// non determining how to handle each outer ring and how to manipluate it
			polygons := [][][][]float64{}
			for _, outer := range outers {
				newpolygon := [][][]float64{outer}
				for id, inner := range innermap {
					boolval := Poly(outer).Within(Poly(inner))
					if boolval {
						newpolygon = append(newpolygon, inner)
						delete(innermap, id)
					}
				}
				polygons = append(polygons, newpolygon)
			}

			// unpacking tags
			mymap := map[string]interface{}{}
			for i := range way.Keys {
				keypos, valpos := way.Keys[i], way.Vals[i]
				mymap[st[keypos]] = st[valpos]
			}
			_, boolval := mymap["name"]
			if boolval {

				if mymap["name"] == "Jefferson National Forest" && len(outers2) > 3 && boolval5 == false && mymap["type"] == "multipolygon" {
					var network bytes.Buffer        // Stand-in for a network connection
					enc := gob.NewEncoder(&network) // Will write to network.
					// Encode (send) the value.
					err := enc.Encode(d.NodeMap.NodeMap)
					if err != nil {
						fmt.Println(err)
					}
					ioutil.WriteFile("a.gob", network.Bytes(), 0677)
					fmt.Println(network.Len())
					fmt.Printf("%#v\n", inners)
					fmt.Printf("%#v\n", outers2)
					//fmt.Printf("%#v\n", d.NodeMap.NodeMap)
					boolval5 = true
				}
			}

			//multipolygon := geojson.NewMultiPolygonFeature(polygons...)
			//multipolygon.Properties = mymap
			if len(polygons) > 0 && mymap["type"] == "multipolygon" {

				for _, polygon := range polygons {
					featpolygon := geojson.NewPolygonFeature(polygon)
					featpolygon.Properties = mymap
					fc.Features = append(fc.Features, featpolygon)
				}
			}

			/*
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
			*/

			//fmt.Println(len(newlist2))

			//mapmembers = Connect_Members(mapmembers)
			//mapmembers = append(mapmembers, goodmembers...)

			//totallist[position] = mapmembers

		}
		/*
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
		*/
	}

	s, _ := fc.MarshalJSON()
	ioutil.WriteFile("a.geojson", s, 0677)
}
