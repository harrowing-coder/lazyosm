package top_level

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

type Areas struct {
	Area_Tags   []string
	Linear_Tags []string
}

type Generalized_Table struct {
	Source     string
	Sql_Filter string
	Tolerance  float64
}

type Args struct {
	Arg     string
	Args    []string
	Default int
	Map     map[string]string
}

type Column struct {
	Name string
	Key  string
	Type string
	Args Args
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
}

// expands a mapping bool
func (mapping *Mapping) Expand() {
	mappingmap := map[string]map[string]string{}
	for k, v := range mapping.Mapping {

		mymap := map[string]string{}
		for _, vv := range v {
			mymap[vv] = ""
		}
		mappingmap[k] = mymap
	}
	mapping.MappingMap = mappingmap

}

func (mapping *Mapping) Map(tags map[string]string) (string, string, bool) {
	for k, v := range tags {
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

type Mappings struct {
	Mappings map[string]*Mapping
}

type Table struct {
	Columns  []Column
	Mapping  Mapping
	Mappings Mappings
	Filters  Filters
	Type     string
}

type Generalized_Tables map[string]Generalized_Table

type TableMapping struct {
	Areas             Areas
	Generalized_Table Generalized_Table
	Tables            map[string]*Table
}

func (tablemapping *TableMapping) Expand() {
	for k, v := range tablemapping.Tables {
		if len(v.Mappings.Mappings) > 0 {
			for k, mapping := range v.Mappings.Mappings {
				mapping.Expand()
				v.Mappings.Mappings[k] = mapping
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

func (table *Table) MakeProperties(tags map[string]string, mappingkey, mappingvalue string) map[string]interface{} {
	mymap := map[string]interface{}{}
	for _, column := range table.Columns {
		switch column.Type {
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
			val := tags[column.Key]
			var myval int
			var donebool bool
			for pos, checkval := range column.Args.Args {
				if checkval == val && !donebool {
					myval = pos + 1
					donebool = true
				}
			}
			mymap[column.Name] = myval
		case "mapping_key":
			mymap[column.Name] = mappingkey
		case "mapping_value":
			mymap[column.Name] = mappingvalue
		case "wayzorder":
			val, boolval := tags[mappingvalue]
			var myval int
			var donebool bool
			for pos, checkval := range column.Args.Args {
				if checkval == val && !donebool {
					myval = pos + 1
					donebool = true
				}
			}
			fmt.Println(column)
			levelOffset := len(column.Args.Args)
			mylayer, boolval := tags["layer"]
			intval, boolval := defaultRanks[mylayer]

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
func (tablemap *TableMapping) TableMap(tags map[string]string, geomtype string) (string, map[string]interface{}, bool) {
	for k, v := range tablemap.Tables {
		mappingkey, mappingval, boolval := v.Mapping.Map(tags)
		if boolval {
			properties := v.MakeProperties(tags, mappingkey, mappingval)
			return k, properties, true
		}
		for k, vv := range v.Mappings.Mappings {
			mappingkey, mappingval, boolval := vv.Map(tags)
			if boolval {
				properties := v.MakeProperties(tags, mappingkey, mappingval)
				return k, properties, true
			}
		}
	}
	return "", map[string]interface{}{}, false
}
