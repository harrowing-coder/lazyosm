package osm

/*
This code holds the methods for processing a file block types
ways and nodes, and provides methods to do these operations concurrently.a

Helper and utility functions are also thrown in here as well.

*/

import (
	"github.com/murphy214/lazyosm/osmpbf"
	//"fmt"
	//g "github.com/murphy214/geobuf"
	"github.com/paulmach/go.geojson"
	"sync"
	//"io/ioutil"
	"fmt"
	"sort"
)

// processes a specific way block
func (d *decoder) ProcessBlockWay(lazy *LazyPrimitiveBlock) {
	block := d.ReadBlock(*lazy)
	var wg sync.WaitGroup
	if len(block.Primitivegroup) > 0 {
		for _, way := range block.Primitivegroup[0].Ways {
			wg.Add(1)
			go func(way *osmpbf.Way) {
				// getting keys
				mymap := map[string]interface{}{}

				for i := range way.Keys {
					keypos, valpos := way.Keys[i], way.Vals[i]
					mymap[block.Stringtable.S[keypos]] = block.Stringtable.S[valpos]
				}
				mymap["osm_id"] = int(way.GetId())

				refs := way.Refs
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
				//_, boolval := d.RelationMap[int(way.Id)]
				//if !boolval {

				line := make([][]float64, len(newrefs))

				for pos, i := range newrefs {
					line[pos] = d.GetNode(i)
				}

				closedbool := false
				// checking if closed way
				if newrefs[0] == newrefs[len(newrefs)-1] {
					closedbool = true
				}

				var feature *geojson.Feature
				//_,boundarybool := mymap[`boundary`]

				if (closedbool == true) && mymap["area"] != "no" {
					feature = geojson.NewPolygonFeature([][][]float64{line})
					feature.Properties = mymap
				} else {
					feature = geojson.NewLineStringFeature(line)
					feature.Properties = mymap
				}

				if d.WriteBool {
					//fmt.Println(line)
					d.Geobuf.WriteFeature(feature)
				}

				//}
				wg.Done()
			}(way)

		}
		wg.Wait()

	}
}

// proces multiple ways
func (d *decoder) ProcessMultipleWays(lazys []*LazyPrimitiveBlock) {
	var wg sync.WaitGroup
	for _, lazy := range lazys {
		wg.Add(1)
		go func(lazy *LazyPrimitiveBlock) {
			d.ProcessBlockWay(lazy)
			wg.Done()
		}(lazy)
	}
	wg.Wait()
}

// Make tags map from stringtable and two parallel arrays of IDs.
func extractTags(stringTable []string, keyIDs, valueIDs []uint32) map[string]string {
	tags := make(map[string]string, len(keyIDs))
	for index, keyID := range keyIDs {
		key := stringTable[keyID]
		val := stringTable[valueIDs[index]]
		tags[key] = val
	}
	return tags
}

// takes a lazy primitive block and process the points out of it
func (d *decoder) ProcessDenseNode(lazy *LazyPrimitiveBlock) {
	pb := d.ReadBlock(*lazy)
	dn := pb.Primitivegroup[0].Dense

	st := pb.GetStringtable().GetS()
	granularity := int64(pb.GetGranularity())
	latOffset := pb.GetLatOffset()
	lonOffset := pb.GetLonOffset()
	//dateGranularity := int64(pb.GetDateGranularity())
	ids := dn.GetId()
	lats := dn.GetLat()
	lons := dn.GetLon()
	//di := dn.GetDenseinfo()

	tu := tagUnpacker{st, dn.GetKeysVals(), 0}
	var id, lat, lon int64
	for index := range ids {
		id = ids[index] + id
		lat = lats[index] + lat
		lon = lons[index] + lon
		latitude := 1e-9 * float64((latOffset + (granularity * lat)))
		longitude := 1e-9 * float64((lonOffset + (granularity * lon)))
		tags := tu.next()
		//info := extractDenseInfo(st, &state, di, index, dateGranularity)
		if len(tags) != 0 {
			//id, latitude, longitude, tags
			mymap := map[string]interface{}{"osm_id": id}
			for k, v := range tags {
				mymap[k] = v
			}

			feature := geojson.NewPointFeature([]float64{longitude, latitude})
			feature.Properties = mymap
			if d.WriteBool {
				d.Geobuf.WriteFeature(feature)
			}
		}
	}
}

// processes multiple dense nodes
func (d *decoder) ProcessMultipleDenseNode(is []*LazyPrimitiveBlock) {
	var wg sync.WaitGroup
	for _, lazy := range is {
		wg.Add(1)
		go func(lazy *LazyPrimitiveBlock) {
			d.ProcessDenseNode(lazy)
			wg.Done()
		}(lazy)
	}
	wg.Wait()
}

// sorts the keys of a map
func SortKeys(mymap map[int]*LazyPrimitiveBlock) []int {
	i := 0
	newlist := make([]int, len(mymap))
	for k := range mymap {
		newlist[i] = k
		i++
	}
	sort.Ints(newlist)
	return newlist
}

// this reads ways from a decoder
func (d *decoder) ReadWays() {
	size := len(d.Ways)
	waylist := SortKeys(d.Ways)
	pos := 0
	is := []*LazyPrimitiveBlock{}

	for _, key := range waylist {
		i := d.Ways[key]
		is = append(is, i)
		if len(is) == 5 || pos == size-1 {
			d.SyncWaysNodeMapMultiple(is, d.IdMap)
			d.ProcessMultipleWays(is)
			is = []*LazyPrimitiveBlock{}
		}
		pos += 1
		fmt.Printf("\r[%d/%d] Way Blocks Completed", pos, size)
	}
}

// processes ways
func (d *decoder) ProcessWays2() {
	is := []*LazyPrimitiveBlock{}
	count := 0
	count = 0
	waylist := SortKeys(d.Ways)
	size := len(waylist)
	pos := 0
	totalidmap := map[int]string{}
	for _, key := range waylist {
		i := d.Ways[key]
		is = append(is, i)

		tempidmap := d.ReadWaysLazy(i, d.IdMap)
		for k, v := range tempidmap {
			totalidmap[k] = v
		}

		if len(totalidmap) > d.Limit || pos == size-1 || len(is) > 5 {
			//d.SyncWaysNodeMapMultiple(is, d.IdMap)
			keylist := make([]int, len(totalidmap))
			i := 0
			for k := range totalidmap {
				keylist[i] = k
				i++
			}
			d.AddUpdates(keylist)
			d.ProcessMultipleWays(is)
			is = []*LazyPrimitiveBlock{}
			totalidmap = map[int]string{}
		}

		count += 1
		pos += 1
		fmt.Printf("\r[%d/%d] Way Blocks Completed. Memory Throughput: %dmb", count, size, d.TotalMemory/1000000)
	}
	fmt.Println()
}

// this method processes the ways in a file
func (d *decoder) ProcessWays() {
	is := []*LazyPrimitiveBlock{}
	count := 0
	count = 0
	waylist := SortKeys(d.Ways)
	size := len(waylist)
	reads := d.AssembleWays()
	for _, read := range reads {

		keylist := make([]int, len(read.Nodes))
		//fmt.Println(read.Nodes)
		//fmt.Println(len(keylist))
		i := 0
		for k := range read.Nodes {
			keylist[i] = k
			i++
		}

		d.AddUpdates(keylist)
		for pos, i := range read.Ways {
			is = append(is, d.Ways[i])
			if len(is) == 50 || pos == len(read.Ways)-1 {
				//d.SyncWaysNodeMapMultiple(is, d.IdMap)
				d.ProcessMultipleWays(is)
				is = []*LazyPrimitiveBlock{}
			}
			count += 1
			fmt.Printf("\r[%d/%d] Way Blocks Completed. Memory Throughput: %dmb", count, size, d.TotalMemory/1000000)
		}

	}
	fmt.Println()
}

// this method processes dense nodes in a file
func (d *decoder) ProcessDenseNodes() {
	d.EmptyNodeMap()
	is := []*LazyPrimitiveBlock{}
	sizedensenodes := len(d.DenseNodes)
	count := 0
	// processing dense nodes (points)
	for pos, i := range d.DenseNodes {
		if i.TagsBool {
			is = append(is, i)
			if len(is) == 10 || pos == sizedensenodes-1 {
				d.ProcessMultipleDenseNode(is)
				is = []*LazyPrimitiveBlock{}
			}
		}
		count += 1
		fmt.Printf("\r[%d/%d] Dense Node Blocks Completed", count, sizedensenodes)
	}
	fmt.Println()

}

type OutputWay struct {
	ID  int
	Map map[int]string
}

type WayRow struct {
	Ways    []*LazyPrimitiveBlock
	KeyList []int
}

// assembles the shortest path structure
// which returns a data structure
// that will most effecientily traverse way file blocks.
func (d *decoder) AssembleWays() []ReadWay {
	d.EmptyNodeMap()
	waylist := SortKeys(d.Ways)
	c := make(chan OutputWay)
	waylistsize := len(waylist)
	count := 0
	totalmap := map[int]map[int]string{}
	for pos, i := range waylist {
		way := d.Ways[i]

		go func(way *LazyPrimitiveBlock, c chan OutputWay) {
			c <- OutputWay{Map: d.ReadWaysLazy(way, d.IdMap), ID: way.Position}
		}(way, c)
		count += 1
		fmt.Printf("\r[%d/%d] Reading Way IDs", pos, waylistsize-1)
		if count == 50 || waylistsize-1 == pos {
			for myc := 0; myc < count; myc++ {
				output := <-c
				totalmap[output.ID] = output.Map
			}
			count = 0
		}
	}
	fmt.Println()

	return MakeReads(totalmap, d.Limit)
}

// processes the osm pbf file
func (d *decoder) ProcessFile() {
	// processing relations
	d.ProcessRelations()

	// procesing ways
	d.ProcessWays()

	// procesing dense nodes
	d.ProcessDenseNodes()
}

// makes color key dense nodes
func (d *decoder) MakeColorDenseNodes() {
	pos := 0
	for _, lazy := range d.DenseNodes {
		colorkey := colorkeys[pos]
		pb := d.ReadBlock(*lazy)
		dn := pb.Primitivegroup[0].Dense

		st := pb.GetStringtable().GetS()
		granularity := int64(pb.GetGranularity())
		latOffset := pb.GetLatOffset()
		lonOffset := pb.GetLonOffset()
		//dateGranularity := int64(pb.GetDateGranularity())
		ids := dn.GetId()
		lats := dn.GetLat()
		lons := dn.GetLon()
		//di := dn.GetDenseinfo()

		tu := tagUnpacker{st, dn.GetKeysVals(), 0}
		var id, lat, lon int64
		for index := range ids {
			id = ids[index] + id
			lat = lats[index] + lat
			lon = lons[index] + lon
			latitude := 1e-9 * float64((latOffset + (granularity * lat)))
			longitude := 1e-9 * float64((lonOffset + (granularity * lon)))
			tags := tu.next()
			//info := extractDenseInfo(st, &state, di, index, dateGranularity)
			//id, latitude, longitude, tags
			mymap := map[string]interface{}{"id": id}
			for k, v := range tags {
				mymap[k] = v
			}

			mymap[`COLORKEY`] = colorkey
			mymap[`POSITION`] = lazy.Position

			feature := geojson.NewPointFeature([]float64{longitude, latitude})
			feature.Properties = mymap
			if d.WriteBool {
				d.Geobuf.WriteFeature(feature)
			}

		}
		pos++
		if pos == sizecolorkeys {
			pos = 0
		}
	}
}

// make colorkey ways
func (d *decoder) MakeColorWays() {
	newlist := []int{}
	for _, i := range d.DenseNodes {
		newlist = append(newlist, i.Position)
	}
	d.AddUpdates(newlist)

	pos := 0
	for _, lazy := range d.Ways {
		colorkey := colorkeys[pos]
		block := d.ReadBlock(*lazy)
		if len(block.Primitivegroup) > 0 {
			for _, way := range block.Primitivegroup[0].Ways {
				// getting keys
				mymap := map[string]interface{}{}
				for i := range way.Keys {
					keypos, valpos := way.Keys[i], way.Vals[i]
					mymap[block.Stringtable.S[keypos]] = block.Stringtable.S[valpos]
				}
				refs := way.Refs
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
				mymap["COLORKEY"] = colorkey
				mymap["POSITION"] = lazy.Position
				line := make([][]float64, len(newrefs))

				for pos, i := range newrefs {
					line[pos] = d.GetNode(i)
				}

				feature := geojson.NewLineStringFeature(line)
				feature.Properties = mymap
				d.Geobuf.WriteFeature(feature)
			}
		}
		pos++
		if pos == sizecolorkeys {
			pos = 0
		}
	}
}

// interpolates a colorkey depending on its position in the densenode list
func Interpolate(pos, limit int) string {
	val := float64(pos) / float64(limit) * float64(len(colorkeys)-1)
	return colorkeys[int(Round(val, .5, 0))]
}

//
func (d *decoder) MakeColorDenseNodeLines(limit int) []*geojson.Feature {
	feats := []*geojson.Feature{}

	for pos, lazy := range d.DenseNodes {
		if pos < limit {

			pb := d.ReadBlock(*lazy)
			dn := pb.Primitivegroup[0].Dense

			granularity := int64(pb.GetGranularity())
			latOffset := pb.GetLatOffset()
			lonOffset := pb.GetLonOffset()
			//dateGranularity := int64(pb.GetDateGranularity())
			ids := dn.GetId()
			lats := dn.GetLat()
			lons := dn.GetLon()
			//di := dn.GetDenseinfo()

			var id, lat, lon int64
			newlist := [][]float64{}
			for index := range ids {
				id = ids[index] + id
				lat = lats[index] + lat
				lon = lons[index] + lon
				latitude := 1e-9 * float64((latOffset + (granularity * lat)))
				longitude := 1e-9 * float64((lonOffset + (granularity * lon)))
				newlist = append(newlist, []float64{longitude, latitude})
				//info := extractDenseInfo(st, &state, di, index, dateGranularity)
				//id, latitude, longitude, tags

			}
			line := geojson.NewMultiPointFeature(newlist...)
			mymap := map[string]interface{}{}

			mymap[`COLORKEY`] = Interpolate(pos, limit)
			mymap[`FILEPOS`] = lazy.FilePos
			line.Properties = mymap
			feats = append(feats, line)
		}
	}
	return feats
}

//
func (d *decoder) MakeColorDenseNodeLines2(startpos, limit int) []*geojson.Feature {
	feats := []*geojson.Feature{}

	newlist := make([]int, len(d.DenseNodes))
	i := 0
	for k := range d.DenseNodes {
		newlist[i] = k
		i++
	}
	sort.Ints(newlist)
	for pos, key := range newlist[startpos:] {
		lazy := d.DenseNodes[key]
		if pos < limit {

			pb := d.ReadBlock(*lazy)
			dn := pb.Primitivegroup[0].Dense

			granularity := int64(pb.GetGranularity())
			latOffset := pb.GetLatOffset()
			lonOffset := pb.GetLonOffset()
			//dateGranularity := int64(pb.GetDateGranularity())
			ids := dn.GetId()
			lats := dn.GetLat()
			lons := dn.GetLon()
			//di := dn.GetDenseinfo()

			var id, lat, lon int64
			newlist := [][]float64{}
			for index := range ids {
				id = ids[index] + id
				lat = lats[index] + lat
				lon = lons[index] + lon
				latitude := 1e-9 * float64((latOffset + (granularity * lat)))
				longitude := 1e-9 * float64((lonOffset + (granularity * lon)))
				newlist = append(newlist, []float64{longitude, latitude})
				//info := extractDenseInfo(st, &state, di, index, dateGranularity)
				//id, latitude, longitude, tags

			}
			line := geojson.NewMultiPointFeature(newlist...)
			mymap := map[string]interface{}{}

			mymap[`COLORKEY`] = Interpolate(pos, limit)
			mymap[`FILEPOS`] = lazy.FilePos
			line.Properties = mymap
			feats = append(feats, line)
		}
	}
	return feats
}
