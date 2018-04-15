package top_level

/*
This structure manages a priority of queue of nodes and removes / adds nodemap to the total nodemap
accordingly. So you hand it a set of nodes to add to the nodemap, and it pushes out lowest in the priority
queue accordingly. This manages the bulk of i/o that we do.

*/

import "fmt"
import "sort"

type NodeMap struct {
	HitMap  map[int]int
	Limit   int
	Popped  int
	NodeMap map[int]map[int][]float64
}

// adds or updates a hit map
func (hits *NodeMap) AddUpdate(stringval int) bool {
	_, boolval := hits.HitMap[stringval]
	if boolval == true {
		hits.HitMap[stringval] = 0
		var popped int
		var maxlimit int
		for k, v := range hits.HitMap {
			if k != stringval {
				if v+1 != hits.Limit {
					hits.HitMap[k] = v + 1
				} else {
					hits.HitMap[k] = v

				}
			}
			if hits.HitMap[k] >= maxlimit {
				maxlimit = hits.HitMap[k]
				popped = k
			}
		}
		hits.Popped = popped
		return false
	} else {
		if len(hits.HitMap) == hits.Limit {
			delete(hits.HitMap, hits.Popped)
			delete(hits.NodeMap, hits.Popped)
		}
		var popped int
		var maxlimit int
		for k, v := range hits.HitMap {
			hits.HitMap[k] = v + 1
			if hits.HitMap[k] >= maxlimit {
				maxlimit = hits.HitMap[k]
				popped = k
			}
		}
		hits.Popped = popped
		hits.HitMap[stringval] = 0
		if len(hits.HitMap) == 1 {
			hits.Popped = stringval
		}
		return true
	}
}

func NewNodeMap(limit int) *NodeMap {
	return &NodeMap{HitMap: map[int]int{}, Limit: limit, NodeMap: map[int]map[int][]float64{}}
}

func (d *decoder) EmptyNodeMap() {
	d.NodeMap = NewNodeMap(d.Limit)
}

// adds values to the nodemap
func (d *decoder) AddUpdate(position int) {
	if d.NodeMap.AddUpdate(position) {
		val, boolval := d.DenseNodes[position]
		if boolval {
			d.NodeMap.NodeMap[position] = d.NewDenseNodeMap(val)
		} else {
			fmt.Println("Positon not available", position)
		}
	}
}

var debug = false

// gets the node for a given relationship
func (d *decoder) GetNode(id int) []float64 {
	if debug == true {
		id2 := d.IdMap.GetBlock(id)
		val, boolval := d.NodeMap.NodeMap[id2][id]
		if boolval {
			return val
		} else {
			fmt.Println(boolval, id2, id, "Not found")
		}
	} else {
		return d.NodeMap.NodeMap[d.IdMap.GetBlock(id)][id]
	}
	return []float64{}
}

type OutputStruct struct {
	Position int
	Map      map[int][]float64
	Priority int
}

// this adds a list of nodes to the node map
// the node map is pretty fault tolerant, if you hand it soemthing over the nodemap limit
// it will read them all in as it assumes there needed
// i.e. you have a limit of 2000 and you input 2121 node ids it will add all those to
// the map
func (d *decoder) AddUpdates(stringval []int) {
	hits := d.NodeMap

	// updating string val so that only new values aree within stringval
	newlist := []int{}
	dups := []int{}
	for _, val := range stringval {
		_, boolval := hits.HitMap[val]
		if !boolval {
			newlist = append(newlist, val)
		} else {
			dups = append(dups, val)
		}
	}
	stringval = newlist

	// now moving the dups to the top of the chain
	dupmap := map[int]string{}
	for pos, i := range dups {
		hits.HitMap[i] = pos
		dupmap[i] = ""
	}
	// now adding the size of dups to each non dup
	size := len(dupmap)
	for k, v := range hits.HitMap {
		_, boolval := dupmap[k]
		if !boolval {
			hits.HitMap[k] = v + size
		}
	}

	// corner cases for all the ways nodes can come in
	if len(dups)+len(stringval) > hits.Limit {
		for k := range hits.HitMap {
			_, boolval := dupmap[k]
			if !boolval {
				delete(hits.HitMap, k)
				delete(hits.NodeMap, k)
			}
		}

	} else if hits.Limit < len(stringval) {
		hits = &NodeMap{NodeMap: map[int]map[int][]float64{}, HitMap: map[int]int{}, Limit: hits.Limit}
		//fmt.Println("allowing to exceed limit to accomade new", len(stringval))
	} else if hits.Limit < len(hits.HitMap)+len(stringval) {
		number_to_remove := len(hits.HitMap) + len(stringval) - hits.Limit
		intlist := make([]int, len(hits.HitMap))
		intmap := map[int]int{}
		i := 0
		for k, v := range hits.HitMap {
			intmap[v] = k
			intlist[i] = v
			i++
		}

		sort.Ints(intlist)
		//fmt.Println(number_to_remove)
		// getting the ints to remove
		remove_ints := intlist[len(intlist)-number_to_remove:]

		// deleting the integer values
		for _, intval := range remove_ints {
			delete(hits.HitMap, intmap[intval])
			delete(hits.NodeMap, intmap[intval])
		}

		// updating the old integer values
		for k, v := range hits.HitMap {
			hits.HitMap[k] = v + number_to_remove
		}

	} else {
		number_add := len(stringval)

		for k, v := range hits.HitMap {
			hits.HitMap[k] = v + number_add
		}

	}

	// finally reading each added value concurrently
	c := make(chan OutputStruct)
	for i, intval := range stringval {
		lazy, boolval := d.DenseNodes[intval]
		go func(i int, lazy *LazyPrimitiveBlock, boolval bool, c chan OutputStruct) {
			if !boolval {
				c <- OutputStruct{}
			} else {
				c <- OutputStruct{
					Map:      d.NewDenseNodeMap(lazy),
					Position: lazy.Position,
					Priority: i,
				}
			}

		}(i, lazy, boolval, c)
	}

	for range stringval {
		output := <-c
		if output.Position != 0 {
			hits.NodeMap[output.Position] = output.Map
			hits.HitMap[output.Position] = output.Priority
		}
	}

	d.NodeMap = hits

}
