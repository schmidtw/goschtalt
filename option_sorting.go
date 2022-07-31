// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"sort"

	"github.com/schmidtw/goschtalt/internal/natsort"
)

const (
	Lexical = iota + 1
	Natural
	Custom
)

// Sorter is the sorting function used to prioritize the configuration files.
type Sorter func(a, b string) bool

// WithSortOrder provides a way to specify how you want the files (and ultimately
// the configuration values) sorted. There are 2 ordering schemes built in
// (Lexical and Natural) as well as the option for you to specify your own, using
// the mode Custom as well as providing a sorter func.
//
// TODO clean this up.
//
// strings provided are the base filenames.  No directory information is provided.
// For the file 'etc/foo/bar.json' the string given to the sorter will be 'bar.json'.
//
// SortByLexical provides a simple lexical based sorter for the files where the
// configuration values originate.  This order determines which configuration
// values are adopted first and last.
//
// SortByNatural provides a simple lexical based sorter for the files where the
// configuration values originate.  This order determines which configuration
// values are adopted first and last.
func WithSortOrder(mode int, sorter ...Sorter) Option {
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

	return func(g *Goschtalt) error {
		if fn == nil {
			return ErrInvalidOption
		}
		g.rawSorter = func(r []raw) {
			sort.SliceStable(r, func(i, j int) bool {
				return fn(r[i].file, r[j].file)
			})
		}
		return nil
	}
}
