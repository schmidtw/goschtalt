// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"sort"

	"github.com/schmidtw/goschtalt/internal/natsort"
)

// FileSortOrderCustom provides a way to specify how you want the files sorted
// prior to their merge.  This function provides a way to provide a completely
// custom sorting algorithm.
func FileSortOrderCustom(less func(a, b string) bool) Option {
	return func(c *Config) error {
		c.sorter = func(a []fileObject) {
			sort.SliceStable(a, func(i, j int) bool {
				return less(a[i].File, a[j].File)
			})
		}
		return nil
	}
}

// FileSortOrderLexical provides a built in sorter based on lexical order.
func FileSortOrderLexical() Option {
	return FileSortOrderCustom(func(a, b string) bool {
		return a < b
	})
}

// FileSortOrderNatural provides a built in sorter based on natural order.
// More information about natural sort order: https://en.wikipedia.org/wiki/Natural_sort_order
//
// Notes:
//
//   - Don't use floating point numbers.  They are treated like 2 integers separated
//     by the '.' rune.
//   - Any leading 0 values are dropped from the number.
//
// Example sort order for reference:
//
//	01_foo.yml
//	2_foo.yml
//	98_foo.yml
//	99 dogs.yml
//	99_Abc.yml
//	99_cli.yml
//	99_mine.yml
//	100_alpha.yml
func FileSortOrderNatural() Option {
	return FileSortOrderCustom(func(a, b string) bool {
		return natsort.Compare(a, b)
	})
}
