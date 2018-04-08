package top_level

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/paulmach/go.geojson"
	"io/ioutil"
	"math/rand"
	"sort"
)

var colorkeys = []string{"#0030E5", "#0042E4", "#0053E4", "#0064E4", "#0075E4", "#0186E4", "#0198E3", "#01A8E3", "#01B9E3", "#01CAE3", "#02DBE3", "#02E2D9", "#02E2C8", "#02E2B7", "#02E2A6", "#03E295", "#03E184", "#03E174", "#03E163", "#03E152", "#04E142", "#04E031", "#04E021", "#04E010", "#09E004", "#19E005", "#2ADF05", "#3BDF05", "#4BDF05", "#5BDF05", "#6CDF06", "#7CDE06", "#8CDE06", "#9DDE06", "#ADDE06", "#BDDE07", "#CDDD07", "#DDDD07", "#DDCD07", "#DDBD07", "#DCAD08", "#DC9D08", "#DC8D08", "#DC7D08", "#DC6D08", "#DB5D09", "#DB4D09", "#DB3D09", "#DB2E09", "#DB1E09", "#DB0F0A"}
var sizecolorkeys = len(colorkeys)

func RandomColor() string {
	return colorkeys[rand.Intn(sizecolorkeys)]
}

func Reverse(s []int) []int {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	//fmt.Println(s)
	return s
}

func Satisify2(ring1 []int, ring2 []int) bool {
	_, lastid1 := ring1[0], ring1[len(ring1)-1]
	firstid2, _ := ring2[0], ring2[len(ring2)-1]
	return firstid2 == lastid1
}

func Collision(ring1 []int, ring2 []int) ([]int, bool, bool) {
	firstid1, lastid1 := ring1[0], ring1[len(ring1)-1]
	firstid2, lastid2 := ring2[0], ring2[len(ring2)-1]
	total := []int{}
	boolval := false
	if firstid1 == firstid2 {

		total = append(ring1, Reverse(ring2)...)
		//total = Reverse(total)
		boolval = true

	} else if firstid1 == lastid2 {
		total = append(ring2, ring1...)
		// /total = Reverse(total)
		boolval = true
	} else if lastid1 == lastid2 {
		total = append(ring1, Reverse(ring2)...)
		boolval = true
	} else if lastid1 == firstid2 {
		total = append(ring1, ring2...)
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

/*
func Collision2(ring1 []int, ring2 []int) ([]int, bool, bool) {
	total := []int{}
	boolval := false
	if Satisify2(ring1, ring2) {

		lastpt := ring1[len(ring1)-1]
		if lastpt == ring2[0] {
			total = append(ring1, ring2...)
			boolval = true
			fmt.Println("error 1")
		} else if lastpt == ring2[len(ring2)-1] {
			fmt.Print("error 2")
			total = append(ring1, Reverse(ring2)...)
			boolval = true
		} else {
			if ring1[0] == ring2[0] {
				fmt.Println("error 3")
				total = append(ring1, Reverse(ring2)...)
				boolval = true
			} else if ring1[0] == ring2[len(ring2)-1] {
				total = append(ring1, Reverse(ring2)...)
				boolval = true
				fmt.Println("error4")
			} else {
				fmt.Println(ring1[0], ring1[len(ring1)-1], ring2[0], ring2[len(ring2)-1])
			}

		}
	}
	if len(total) == 0 {
		return []int{}, false, false
	} else {
		if total[0] == total[len(total)-1] {
			fmt.Println(total, ring2, total[0], total[len(total)-1])
		}

	}
	return total, boolval, total[0] != total[len(total)-1]

}
*/

func SortedMap(mymap map[int][]int) []int {
	newlist := make([]int, len(mymap))
	pos := 0
	for k := range mymap {
		newlist[pos] = k
		pos++
	}
	sort.Ints(newlist)
	return Reverse(newlist)
}
func cleanse(member []int) []int {
	if len(member) > 0 {
		if member[0] == member[len(member)-1] {
			return member[:len(member)-1]
		} else {
			return member
		}
	} else {
		return member
	}
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
	generation := 0
	for len(membermap) > 2 && generation < 10 {

		for _, k := range SortedMap(membermap) {
			member, boolval1 := membermap[k]
			boolval := true
			if boolval1 {
				lastpt := member[len(member)-1]
				for _, ktry := range SortedMap(membermap) {
					trymember, boolval2 := membermap[ktry]
					if boolval2 {
						if k != ktry && boolval == true {
							if lastpt == trymember[0] {
								if len(membermap) == 2 {
									//membermap[k] = append(member, trymember...)

								} else {
									membermap[k] = append(member, trymember...)

								}

								//membermap[k] = append(member, trymember...)
								delete(membermap, ktry)
								boolval = true
								//fmt.Println(len(membermap[k]), len(membermap))
							}
						}
					}
				}
			}
			generation += 1
		}
	}

	generation = 0
	for len(membermap) > 2 && generation < 100 {

		for _, k := range SortedMap(membermap) {
			member, boolval1 := membermap[k]
			if boolval1 {
				//lastpt := member[len(member)-1]
				boolval := false
				for _, ktry := range SortedMap(membermap) {
					trymember, boolval2 := membermap[ktry]
					if boolval2 {
						if k != ktry && boolval == false {
							if len(membermap) == 2 {
								if member[len(member)-1] != trymember[0] {
									membermap[k] = append(member, Reverse(trymember)...)
								} else {
									membermap[k] = append(member, trymember...)

								}
								delete(membermap, ktry)
							}

							if member[0] == trymember[0] {
								//membermap[ktry] = append(trymember, member...)
								membermap[ktry] = Reverse(trymember)
								//delete(membermap, k)

							} else if member[len(member)-1] == trymember[len(trymember)-1] {
								membermap[ktry] = Reverse(trymember)
								//delete(membermap, ktry)

							} else if member[0] == trymember[len(trymember)-1] {
								//membermap[ktry] = append(trymember, member...)
								membermap[ktry] = Reverse(trymember)
								//delete(membermap, k)

							} else if member[len(member)-1] == trymember[0] {
								membermap[k] = append(member, trymember...)

								delete(membermap, ktry)

							} else {

								//member, trymember = cleanse(member), cleanse(trymember)
								//membermap[k] = Reverse(member)
								//membermap[ktry] = Reverse(trymember)

								//fmt.Println(member[0], member[len(member)-1], trymember[0], trymember[len(trymember)-1])
							}
						}

					}
				}
				generation += 1
			}
		}

	}

	// final clean up if applicable
	if len(membermap) == 2 {
		var member, trymember []int
		var pos int
		var k, ktry int
		for kk, v := range membermap {
			if pos == 0 {
				pos = 1
				member = v
				k = kk
			} else if pos == 1 {
				trymember = v
				ktry = kk
			}
		}

		if member[len(member)-1] != trymember[0] {
			membermap[k] = append(member, Reverse(trymember)...)
		} else {
			membermap[k] = append(member, trymember...)

		}
		delete(membermap, ktry)
	}
	pos := 0
	for _, v := range membermap {
		totalmembers = append(totalmembers, v)
		pos++
	}

	return totalmembers

}

func ConvertNodes(nodes []int, nodemap map[int][]float64) [][]float64 {
	ring := make([][]float64, len(nodes))
	for pos, node := range nodes {
		ring[pos] = nodemap[node]
	}
	return ring
}

type TestStruct struct {
	Outers  [][]int
	Inners  [][]int
	NodeMap map[int][]float64
}

func (test *TestStruct) MakeOuters() *geojson.FeatureCollection {
	fc := &geojson.FeatureCollection{}
	for _, outer := range test.Outers {
		outerring := make([][]float64, len(outer))
		for pos, node := range outer {
			outerring[pos] = test.NodeMap[node]
		}
		feature := geojson.NewLineStringFeature(outerring)
		feature.Properties = map[string]interface{}{"COLORKEY": RandomColor()}
		fc.Features = append(fc.Features, feature)
	}
	return fc
}

// creates a polygon feature from a given test case
func (test *TestStruct) MakePolygon() *geojson.Feature {
	test.Outers = Connect(test.Outers)
	test.Inners = Connect(test.Inners)

	innermap := map[int][][]float64{}
	for pos, ring := range test.Inners {
		innermap[pos] = ConvertNodes(ring, test.NodeMap)
	}
	// collecting each raw polygon
	//  non determining how to handle each outer ring and how to manipluate it
	polygons := [][][][]float64{}
	for _, outerint := range test.Outers {
		outer := ConvertNodes(outerint, test.NodeMap)
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
	var featpolygon *geojson.Feature
	if len(polygons) == 1 {
		featpolygon = geojson.NewPolygonFeature(polygons[0])
	} else {
		featpolygon = geojson.NewMultiPolygonFeature(polygons...)

	}
	featpolygon.Properties = map[string]interface{}{"COLORKEY": RandomColor()}
	return featpolygon
}

func ReadTestCaseGob(nodefilename, nodemapfilename string) TestStruct {

	nodebytes, err := ioutil.ReadFile(nodefilename)
	if err != nil {
		fmt.Println(err)
	}

	nodemapbytes, err := ioutil.ReadFile(nodemapfilename)
	if err != nil {
		fmt.Println(err)
	}

	network := bytes.NewBuffer(nodebytes)
	dec := gob.NewDecoder(network)
	var v [][][]int
	err = dec.Decode(&v)
	if err != nil {
		fmt.Println(err)
	}
	var outers, inners [][]int
	if len(v) == 2 {
		outers, inners = v[0], v[1]
	}

	network = bytes.NewBuffer(nodemapbytes)
	dec = gob.NewDecoder(network)
	var vv map[int][]float64
	err = dec.Decode(&vv)
	if err != nil {
		fmt.Println(err)
	}

	return TestStruct{NodeMap: vv, Outers: outers, Inners: inners}
}
