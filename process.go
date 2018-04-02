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
