package top_level

import (
	//"bytes"
	//"encoding/gob"
	"fmt"
	//"./osmpbf"
	"github.com/paulmach/go.geojson"
	//"io/ioutil"
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
		if bottom[1] >= top[1] {
			bottom, top = top, bottom
		}
		if p[1] <= bottom[1] || p[1] >= top[1] {
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

func (d *decoder) ProcessRelation(key int) {
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
	for ipos, way := range relations {
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
		totalidmap := map[int]string{}
		for _, i := range newrefs {

			vals, boolval := totalwaynodemap[i]
			if boolval {
				for _, nodeid := range vals {
					totalidmap[d.IdMap.GetBlock(nodeid)] = ""
				}
			}

		}

		stringval2 := make([]int, len(totalidmap))
		newpos := 0
		for k := range totalidmap {
			stringval2[newpos] = k
			newpos++
		}
		if len(stringval2) > 0 {
			d.AddUpdates(stringval2)
		}
		fmt.Println(len(stringval), ipos)
		// getting string table
		st := pb.GetStringtable().GetS()

		// creating nrew refs roles and way ids

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
				if role == "inner" {
					inners = append(inners, nodes)
				} else if role == "outer" {
					outers2 = append(outers2, nodes)
				}
			}
		}

		// dealing with roles and getting nodes
		inners = Connect(inners)
		outers2 = Connect(outers2)
		innermap := map[int][][]float64{}
		outers := [][][]float64{}
		for pos, inner := range inners {
			ring := make([][]float64, len(inner))
			for pos, node := range inner {
				ring[pos] = RoundPt(d.GetNode(node))
			}
			innermap[pos] = ring
		}

		// changing the outer from nodes to float
		for _, outer := range outers2 {
			ring := make([][]float64, len(outer))
			for pos, node := range outer {
				ring[pos] = RoundPt(d.GetNode(node))
			}
			outers = append(outers, ring)
		}

		// collecting each raw polygon
		//  non determining how to handle each outer ring and how to manipluate it
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

		//multipolygon := geojson.NewMultiPolygonFeature(polygons...)
		//multipolygon.Properties = mymap
		if len(polygons) > 0 && mymap["type"] == "multipolygon" {
			if len(polygons) == 1 {
				featpolygon := geojson.NewPolygonFeature(polygons[0])
				featpolygon.Properties = mymap
				d.Geobuf.WriteFeature(featpolygon)
			} else {
				featpolygon := geojson.NewMultiPolygonFeature(polygons...)
				featpolygon.Properties = mymap
				d.Geobuf.WriteFeature(featpolygon)
			}
		}
	}
}

// this method is designed to hackily handle large relations
func (d *decoder) ProcessRelations() {
	relationlist := SortKeys(d.Relations)
	//boolval5 := false

	// reading through each relation block
	sizerelation := len(relationlist)
	for i, key := range relationlist {
		d.ProcessRelation(key)
		fmt.Printf("\r[%d/%d] Processing Relations", i, sizerelation)
	}
}
