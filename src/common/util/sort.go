package util

import (
	"sort"
)

type Pair1 struct {
	Key   int
	Value int
}

type Pair2 struct {
	Key   string
	Value int
}

func SortMap1(org map[int]int) PairList1 {

	pl := make(PairList1, len(org))

	i := 0
	for k, v := range org {
		pl[i] = Pair1{k, v}
		i++
	}

	//sort.Sort(sort.Reverse(pl))
	sort.Sort(pl)

	return pl
}

func ReverseSortMap1(org map[int]int) PairList1 {

	pl := make(PairList1, len(org))

	i := 0
	for k, v := range org {
		pl[i] = Pair1{k, v}
		i++
	}

	sort.Sort(sort.Reverse(pl))
	//sort.Sort(pl)

	return pl
}

func SortMap2(org map[string]int) PairList2 {

	pl := make(PairList2, len(org))

	i := 0
	for k, v := range org {
		pl[i] = Pair2{k, v}
		i++
	}

	//sort.Sort(sort.Reverse(pl))
	sort.Sort(pl)
	return pl
}

func ReverseSortMap2(org map[string]int) PairList2 {

	pl := make(PairList2, len(org))

	i := 0
	for k, v := range org {
		pl[i] = Pair2{k, v}
		i++
	}

	sort.Sort(sort.Reverse(pl))
	//sort.Sort(pl)
	return pl
}

type PairList1 []Pair1
type PairList2 []Pair2

func (p PairList1) Len() int           { return len(p) }
func (p PairList1) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList1) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (p PairList2) Len() int           { return len(p) }
func (p PairList2) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList2) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
