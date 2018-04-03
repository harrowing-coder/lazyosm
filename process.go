package top_level

import (
	"./osmpbf"
	//"fmt"
	//g "github.com/murphy214/geobuf"
	"github.com/paulmach/go.geojson"
	"sync"
	//"io/ioutil"
)

func (d *decoder) ProcessBlock(lazy *LazyPrimitiveBlock) {
	//fmt.Println(i)
	block := d.ReadBlock(*lazy)
	//fmt.Println("here")
	//fc := &geojson.FeatureCollection{}
	//feature := &geojson.Feature{Properties: map[string]interface{}{}}
	var wg sync.WaitGroup
	//tempwriter := g.WriterBufNew()

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

				line := make([][]float64, len(newrefs))

				for pos, i := range newrefs {
					line[pos] = d.GetNode(i)
				}
				feature := geojson.NewLineStringFeature(line)
				feature.Properties = mymap
				//feature.Geometry.Type = "LineString"
				//features = append(features, feature)
				//fmt.Println(feature)
				d.Geobuf.WriteFeature(feature)
				//count += 1

				//make(map[string]interface{}, len(keys))
				wg.Done()
			}(way)

		}
		wg.Wait()

	}
	//bytevals, _ := fc.MarshalJSON()
	//ioutil.WriteFile("a.geojson", bytevals, 0677)
}

// proces multiple
func (d *decoder) ProcessMultiple(lazys []*LazyPrimitiveBlock) {
	var wg sync.WaitGroup
	for _, lazy := range lazys {
		wg.Add(1)
		go func(lazy *LazyPrimitiveBlock) {
			d.ProcessBlock(lazy)
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
			mymap := map[string]interface{}{"id": id}
			for k, v := range tags {
				mymap[k] = v
			}

			feature := geojson.NewPointFeature([]float64{longitude, latitude})
			feature.Properties = mymap
			d.Geobuf.WriteFeature(feature)
		}
		//dec.q = append(dec.q, &Node{id, latitude, longitude, tags, info})
	}
}

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
