// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"sort"

	"github.com/schmidtw/goschtalt/internal/natsort"
)

// SortOrder mode options
const (
	Lexical = iota + 1 // Sorts the files in lexical order
	Natural            // Sorts the files in natural order
	Custom             // Allows client to specify their own algorithm
)

// Sorter is the sorting function used to prioritize the configuration files.
type Sorter func(a, b string) bool

// SortOrder provides a way to specify how you want the files (and ultimately
// the configuration values) sorted. There are 2 ordering schemes built in
// (Lexical and Natural) as well as the option for you to specify your own, using
// the mode Custom as well as providing a sorter function.
func SortOrder(mode int, sorter ...Sorter) Option {
	var fn Sorter
	switch mode {
	case Lexical:
		fn = func(a, b string) bool {
			return a < b
		}
	case Natural:
		fn = natsort.Compare
	case Custom:
		if len(sorter) > 0 {
			fn = sorter[0]
		}
	default:
		fn = nil
	}

	return func(c *Config) error {
		if fn == nil {
			return ErrInvalidOption
		}
		c.annotatedSorter = func(a []annotatedMap) {
			sort.SliceStable(a, func(i, j int) bool {
				return fn(a[i].files[0], a[j].files[0])
			})
		}
		return nil
	}
}
