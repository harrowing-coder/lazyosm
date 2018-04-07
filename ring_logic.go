package top_level

import (
	//"bytes"
	//"encoding/gob"
	//"fmt"
	//"github.com/paulmach/go.geojson"
	//"io/ioutil"
	//"math/rand"
	"sort"
)

func Reverse(s []int) []int {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
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
	if member[0] == member[len(member)-1] {
		return member[1:]
	} else {
		return member[1:]
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
	for len(membermap) > 2 && generation < 1000 {

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
							membermap[k] = append(member, Reverse(trymember)...)
							delete(membermap, ktry)

						} else if member[len(member)-1] == trymember[len(trymember)-1] {
							membermap[k] = append(member, Reverse(trymember)...)
							delete(membermap, ktry)

						} else if member[0] == trymember[len(trymember)-1] {
							membermap[k] = append(trymember, member...)
							delete(membermap, ktry)

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
		}
	}

	pos := 0
	for _, v := range membermap {
		totalmembers = append(totalmembers, v)
		pos++
	}

	return totalmembers

}
