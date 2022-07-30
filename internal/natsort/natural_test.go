// SPDX-FileCopyrightText: 2015 Vincent Batoufflet and Marc Falzon
// SPDX-FileCopyrightText: 2022 Mark Karpelès
// SPDX-License-Identifier: BSD-3-Clause
//
// This file originated from https://github.com/facette/natsort/pull/2/files

package natsort

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

type stringSlice []string

func (s stringSlice) Len() int {
	return len(s)
}

func (s stringSlice) Less(a, b int) bool {
	return Compare(s[a], s[b])
}

func (s stringSlice) Swap(a, b int) {
	s[a], s[b] = s[b], s[a]
}

// Sort sorts a list of strings in a natural order
func Sort(l []string) {
	sort.Sort(stringSlice(l))
}

var testList = []string{
	"1000X Radonius Maximus",
	"10X Radonius",
	"200X Radonius",
	"20X Radonius",
	"20X Radonius Prime",
	"30X Radonius",
	"40X Radonius",
	"Allegia 50 Clasteron",
	"Allegia 500 Clasteron",
	"Allegia 50B Clasteron",
	"Allegia 51 Clasteron",
	"Allegia 6R Clasteron",
	"Alpha 100",
	"Alpha 2",
	"Alpha 200",
	"Alpha 2A",
	"Alpha 2A-8000",
	"Alpha 2A-900",
	"Callisto Morphamax",
	"Callisto Morphamax 500",
	"Callisto Morphamax 5000",
	"Callisto Morphamax 600",
	"Callisto Morphamax 6000 SE",
	"Callisto Morphamax 6000 SE2",
	"Callisto Morphamax 700",
	"Callisto Morphamax 7000",
	"Xiph Xlater 10000",
	"Xiph Xlater 2000",
	"Xiph Xlater 300",
	"Xiph Xlater 40",
	"Xiph Xlater 5",
	"Xiph Xlater 50",
	"Xiph Xlater 500",
	"Xiph Xlater 5000",
	"Xiph Xlater 58",
}

func Test_Sort1(t *testing.T) {
	testListSortedOK := []string{
		"10X Radonius",
		"20X Radonius",
		"20X Radonius Prime",
		"30X Radonius",
		"40X Radonius",
		"200X Radonius",
		"1000X Radonius Maximus",
		"Allegia 6R Clasteron",
		"Allegia 50 Clasteron",
		"Allegia 50B Clasteron",
		"Allegia 51 Clasteron",
		"Allegia 500 Clasteron",
		"Alpha 2",
		"Alpha 2A",
		"Alpha 2A-900",
		"Alpha 2A-8000",
		"Alpha 100",
		"Alpha 200",
		"Callisto Morphamax",
		"Callisto Morphamax 500",
		"Callisto Morphamax 600",
		"Callisto Morphamax 700",
		"Callisto Morphamax 5000",
		"Callisto Morphamax 6000 SE",
		"Callisto Morphamax 6000 SE2",
		"Callisto Morphamax 7000",
		"Xiph Xlater 5",
		"Xiph Xlater 40",
		"Xiph Xlater 50",
		"Xiph Xlater 58",
		"Xiph Xlater 300",
		"Xiph Xlater 500",
		"Xiph Xlater 2000",
		"Xiph Xlater 5000",
		"Xiph Xlater 10000",
	}
	testListSorted := testList[:]
	Sort(testListSorted)

	if !reflect.DeepEqual(testListSortedOK, testListSorted) {
		t.Fatalf(`ERROR: sorted list different from expected results:
	Expected results:
%v

	Got:
%v`, strings.Join(testListSortedOK, "\n"), strings.Join(testListSorted, "\n"))
	}
}

func Test_Sort2(t *testing.T) {
	testList := []string{
		"z1.doc",
		"z10.doc",
		"z100.doc",
		"z101.doc",
		"z102.doc",
		"z11.doc",
		"z12.doc",
		"z13.doc",
		"z14.doc",
		"z15.doc",
		"z16.doc",
		"z17.doc",
		"z18.doc",
		"z19.doc",
		"z2.doc",
		"z20.doc",
		"z3.doc",
		"z4.doc",
		"z5.doc",
		"z6.doc",
		"z7.doc",
		"z8.doc",
		"z9.doc",
	}

	testListSortedOK := []string{
		"z1.doc",
		"z2.doc",
		"z3.doc",
		"z4.doc",
		"z5.doc",
		"z6.doc",
		"z7.doc",
		"z8.doc",
		"z9.doc",
		"z10.doc",
		"z11.doc",
		"z12.doc",
		"z13.doc",
		"z14.doc",
		"z15.doc",
		"z16.doc",
		"z17.doc",
		"z18.doc",
		"z19.doc",
		"z20.doc",
		"z100.doc",
		"z101.doc",
		"z102.doc",
	}

	testListSorted := testList[:]
	Sort(testListSorted)

	if !reflect.DeepEqual(testListSortedOK, testListSorted) {
		t.Fatalf(`ERROR: sorted list different from expected results:
	Expected results:
%v

	Got:
%v`, strings.Join(testListSortedOK, "\n"), strings.Join(testListSorted, "\n"))
	}
}

func BenchmarkSort1(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Sort(testList)
	}
}
