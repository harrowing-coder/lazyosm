package map_osm

import (
	"fmt"
	m "github.com/murphy214/mercantile"
	"github.com/paulmach/go.geojson"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"math"
	"strconv"
	///"strconv"

	"github.com/murphy214/vector-tile-go"
	"strings"
)

const mercatorPole = 20037508.34

func ConvertPoint(point []float64) []float64 {
	x := mercatorPole / 180.0 * point[0]

	y := math.Log(math.Tan((90.0+point[1])*math.Pi/360.0)) / math.Pi * mercatorPole
	y = math.Max(-mercatorPole, math.Min(y, mercatorPole))
	return []float64{x, y}
}

func reproject(coords [][]float64) [][]float64 {
	for i := range coords {
		coords[i] = ConvertPoint(coords[i])
	}
	return coords
}

func ConvertToMercator(geometry *geojson.Geometry) *geojson.Geometry {
	switch geometry.Type {
	case "Point":
		geometry.Point = ConvertPoint(geometry.Point)
	case "MultiPoint":
		geometry.MultiPoint = reproject(geometry.MultiPoint)
	case "LineString":
		geometry.LineString = reproject(geometry.LineString)
	case "MultiLineString":
		for i := range geometry.MultiLineString {
			geometry.MultiLineString[i] = reproject(geometry.MultiLineString[i])
		}
	case "Polygon":
		for i := range geometry.Polygon {
			geometry.Polygon[i] = reproject(geometry.Polygon[i])
		}
	case "MultiPolygon":
		for i := range geometry.MultiPolygon {
			for j := range geometry.MultiPolygon[i] {
				geometry.MultiPolygon[i][j] = reproject(geometry.MultiPolygon[i][j])
			}
		}
	}
	return geometry
}

type Areas struct {
	Area_Tags   []string
	Linear_Tags []string
	Map         map[string]string
}

type Generalized_Table struct {
	Source     string
	Sql_Filter string
	Tolerance  float64
}

// arg values
type Args struct {
	//Arg       string `yaml:",inline"`
	Arg       string
	ArgValues []string
	Args      map[string]interface{} `yaml:",inline"`
	Default   int
	Map       map[string]string
}

func (args *Args) Clean() {
	if args != nil {
		for k, v := range args.Args {
			args.Arg = k
			if k == "suffixes" {
				mymap := map[string]string{}
				for kk, vv := range v.(map[interface{}]interface{}) {
					mymap[kk.(string)] = vv.(string)
				}
				args.Map = mymap
			} else {
				for _, vv := range v.([]interface{}) {
					val, boolval := vv.(string)
					if boolval {
						args.ArgValues = append(args.ArgValues, val)
					}

				}
			}
		}
	}
}

type Column struct {
	Name string
	Key  string
	Type string
	Args *Args
}

type Filters struct {
	Type   string // require,reject,reject_regexp
	Field  string
	Filter []string
}
type Mapping struct {
	Mapping    map[string][]string `yaml:",inline"`
	MappingMap map[string]map[string]string
	Any        bool
	AnyKeys    map[string]string
}

// expands a mapping bool
func (mapping *Mapping) Expand() {
	mappingmap := map[string]map[string]string{}
	mapping.AnyKeys = map[string]string{}
	for k, v := range mapping.Mapping {
		var boolval bool
		mymap := map[string]string{}

		if len(v) == 1 {
			if v[0] == "__any__" {
				mapping.AnyKeys[k] = ""
				boolval = true
				mapping.Any = true
			}
		}
		if !boolval {
			for _, vv := range v {
				mymap[vv] = ""
			}
		}
		mappingmap[k] = mymap
	}
	mapping.MappingMap = mappingmap

}

func (mapping *Mapping) Map(tags map[string]string) (string, string, bool) {
	for k, v := range tags {
		_, boolval := mapping.AnyKeys[k]

		if mapping.Any && boolval {
			return k, v, boolval
		}

		val, boolval := mapping.MappingMap[k]
		if boolval {
			_, boolval2 := val[v]
			if boolval2 {
				return k, v, boolval2
			}
		}
	}
	return "", "", false
}

type Table struct {
	Columns  []*Column
	Mapping  Mapping
	Mappings map[string]map[string]Mapping
	Filters  Filters
	Type     string
	Args     *Args
}

func (table *Table) Clean() {
	table.Args.Clean()

	for pos, column := range table.Columns {
		column.Args.Clean()
		table.Columns[pos] = column
	}
}

type Generalized_Tables map[string]Generalized_Table

type TableMapping struct {
	Areas             Areas
	Generalized_Table Generalized_Table
	Tables            map[string]*Table
	SingleIdSpace     bool `yaml:"use_single_id_space"`
}

func (tablemapping *TableMapping) Expand() {
	for k, v := range tablemapping.Tables {
		v.Clean()
		if len(v.Mappings) > 0 {
			for k, mappings := range v.Mappings {
				if len(mappings) > 0 {
					for kk, mapping := range mappings {
						mapping.Expand()
						v.Mappings[k][kk] = mapping
					}
				}
				//v.Mappings[k] = mapping
				//fmt.Println(v)
			}
		} else {
			//fmt.Println(v.Mapping)
			vv := &v.Mapping
			vv.Expand()
			//fmt.Println(vv)
			v.Mapping = *vv
			tablemapping.Tables[k] = v
			//fmt.Println(v.Mapping)

		}
	}
	mymap := map[string]string{}
	for _, v := range tablemapping.Areas.Area_Tags {
		mymap[v] = "polygon"
	}

	for _, v := range tablemapping.Areas.Linear_Tags {
		mymap[v] = "linestring"
	}
	tablemapping.Areas.Map = mymap
}

var defaultRanks = map[string]int{
	"minor":          3,
	"road":           3,
	"unclassified":   3,
	"residential":    3,
	"tertiary_link":  3,
	"tertiary":       4,
	"secondary_link": 3,
	"secondary":      5,
	"primary_link":   3,
	"primary":        6,
	"trunk_link":     3,
	"trunk":          8,
	"motorway_link":  3,
	"motorway":       9,
}

func ReadYamlMapping(filename string) *TableMapping {
	bytevals, _ := ioutil.ReadFile("test_mapping.yml")
	var t TableMapping
	err := yaml.Unmarshal(bytevals, &t)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	myt := &t
	//fmt.Printf("%+v\n", myt.Tables["buildings"])
	myt.Expand()
	return myt
}

// signed area frunction
func SignedArea(ring [][]float64) float64 {
	sum := 0.0
	i := 0
	lenn := len(ring)
	j := lenn - 1
	var p1, p2 []float64

	for i < lenn {
		if i != 0 {
			j = i - 1
		}
		p1 = ring[i]
		p2 = ring[j]
		sum += (p2[0] - p1[0]) * (p1[1] + p2[1])
		i++
	}
	return sum
}

func GetArea(geometry *geojson.Geometry) float64 {
	var area float64
	switch geometry.Type {
	case "Polygon":
		for pos, i := range geometry.Polygon {
			if pos == 0 {
				area += SignedArea(i)
			} else {
				area -= SignedArea(i)
			}
		}

		return area
	case "MultiPolygon":
		for _, polygon := range geometry.MultiPolygon {
			var temparea float64
			for pos, i := range polygon {
				if pos == 0 {
					temparea += SignedArea(i)
				} else {
					temparea -= SignedArea(i)
				}
			}
			area += temparea
		}
		return area
	}
	return area

}

func WebmercArea(geometry *geojson.Geometry) interface{} {

	area := GetArea(geometry)

	bb := vt.Get_BoundingBox(geometry)

	bds := m.Extrema{W: bb[0], S: bb[1], E: bb[2], N: bb[3]}

	midY := bds.S + (bds.N-bds.S)/2

	pole := 6378137 * math.Pi // 20037508.342789244
	midLat := 2*math.Atan(math.Exp((midY/pole)*math.Pi)) - math.Pi/2

	area = area * math.Pow(math.Cos(midLat), 2)
	area = math.Abs(area)
	return area
}

func (tablemapping *TableMapping) MakeProperties(table *Table, tags map[string]string, geometry *geojson.Geometry, mappingkey, mappingvalue string) map[string]interface{} {

	mymap := map[string]interface{}{}
	for _, column := range table.Columns {
		switch column.Type {

		case "area":
			mymap[column.Name] = GetArea(geometry)
		case "webmerc_area":
			mymap[column.Name] = WebmercArea(geometry)
		case "integer":
			myval := tags[column.Key]

			intval, err := strconv.ParseInt(myval, 10, 64)
			if err == nil {
				mymap[column.Name] = int(intval)
			} else {
				mymap[column.Name] = 0
			}

		case "bool":
			myval, boolval := tags[column.Key]
			if !boolval {
				myval = ""
			}
			if myval == "true" || myval == "1" || myval == "yes" {
				mymap[column.Name] = true
			} else {
				mymap[column.Name] = false
			}

		case "boolint":
			myval, boolval := tags[column.Key]
			if !boolval {
				myval = ""
			}
			if myval == "true" || myval == "1" || myval == "yes" {
				mymap[column.Name] = 1
			} else {
				mymap[column.Name] = 0
			}

		case "string":
			myval, boolval := tags[column.Key]
			if !boolval {
				myval = ""
			}
			mymap[column.Name] = myval
		case "direction":
			val := tags[column.Key]
			var myval int
			if val == "1" || val == "yes" || val == "true" {
				myval = 1
			} else if val == "-1" {
				myval = -1
			} else {
				myval = 0
			}
			mymap[column.Name] = myval

		case "enumerate":
			val := mappingvalue
			var myval int
			var donebool bool
			if len(column.Args.ArgValues) == 0 {
				for pos, checkval := range table.Args.ArgValues {
					if checkval == val && !donebool {
						myval = pos + 1
						donebool = true
					}
				}
			} else {
				for pos, checkval := range column.Args.ArgValues {
					if checkval == val && !donebool {
						myval = pos + 1
						donebool = true
					}
				}
			}
			mymap[column.Name] = myval
		case "mapping_key":
			mymap[column.Name] = mappingkey
		case "mapping_value":
			mymap[column.Name] = mappingvalue
		case "wayzorder":
			//val, boolval := tags[mappingvalue]
			var myval int
			var donebool bool
			myvalues := column.Args.ArgValues
			//fmt.Println(table.Args)

			if table.Args != nil {
				myvalues = table.Args.ArgValues
				fmt.Println(myvalues)
			}

			for pos, checkval := range myvalues {
				if checkval == mappingvalue && !donebool {
					myval = pos + 1
					donebool = true
				}
			}
			//fmt.Println(column.Args, mappingvalue)
			levelOffset := len(myvalues)
			intval, boolval := defaultRanks[mappingvalue]

			z := levelOffset*intval + myval

			bridge, boolval := tags["bridge"]
			if !boolval {
				bridge = ""
			}
			if bridge == "yes" {
				z += levelOffset
			}

			bridge, boolval = tags["tunnel"]
			if !boolval {
				bridge = ""
			}
			if bridge == "yes" {
				z += levelOffset
			}
			mymap[column.Name] = z
		case "string_suffixreplace":
			myval, boolval := tags[column.Key]
			if !boolval {
				myval = ""
			}
			//fmt.Println(myval, boolval)

			//FIX THIS SHIT
			for k, v := range column.Args.Map {
				if strings.Contains(myval, k) {
					myval = strings.Replace(myval, k, v, -1)
				}
			}

			mymap[column.Name] = myval

		}
	}
	return mymap
}

// given a set of tags iterates through each tag set to determine whether it can be mapped
// type can be point, linestring, polygon, geometry, relation
func (tablemap *TableMapping) TableMap(tags map[string]string, geometry *geojson.Geometry) (map[string]*geojson.Feature, bool) {
	// getting geometry
	var geomtype string
	switch geometry.Type {
	case "Point", "MultiPoint":
		geomtype = "point"
	case "LineString", "MultiLineString":
		geomtype = "linestring"
	case "Polygon", "MultiPolygon":
		geomtype = "polygon"
	}

	// filtering out disagreeing area tags
	for k, v := range tablemap.Areas.Map {
		_, boolval := tags[k]
		if boolval {
			//fmt.Println(k, v, geomtype, tags)
			//fmt.Println(tags)
			/*
				var area string
				if v == "linestring" {
					area = "no"
				} else if v == "polygon" {
					area = "yes"
				}
			*/
			//areatest := tags["area"]
			geomtype = v

			//if area == "yes" && areatest == "no" {
			//	return map[string]*geojson.Feature{}, false
			//} else {
			//geomtype = v
			//}

			//totalbool = true

		}
	}

	// mapping against each table
	var notempty bool
	outputmap := map[string]*geojson.Feature{}
	for k, v := range tablemap.Tables {
		if !notempty {
			if geomtype == v.Type {
				mappingkey, mappingval, boolval := v.Mapping.Map(tags)
				if boolval {
					properties := tablemap.MakeProperties(v, tags, geometry, mappingkey, mappingval)
					newfeat := geojson.NewFeature(geometry)
					newfeat.Properties = properties
					outputmap[k] = newfeat
					//notempty = true
				}
				if !notempty {
					for _, vv := range v.Mappings {

						//fmt.Println(tablemapname)
						for _, vvv := range vv {
							mappingkey, mappingval, boolval := vvv.Map(tags)
							if boolval {
								properties := tablemap.MakeProperties(v, tags, geometry, mappingkey, mappingval)
								newfeat := geojson.NewFeature(geometry)
								newfeat.Properties = properties
								outputmap[k] = newfeat
								//notempty = true
							}
						}
					}
				}
			}
		}
	}

	return outputmap, len(outputmap) > 0
}
