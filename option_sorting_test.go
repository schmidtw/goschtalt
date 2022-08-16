// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/schmidtw/goschtalt/pkg/meta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testStringsToObjects(s []string) (list []meta.Object) {
	for _, val := range s {
		o := meta.Object{
			Origins: []meta.Origin{
				{
					File: val,
				},
			},
		}

		list = append(list, o)
	}
	return list
}

func TestFileSortOrderLexical(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	var c Config

	require.NoError(c.With(FileSortOrderLexical()))
	require.NotNil(c.sorter)

	list := testStringsToObjects([]string{"zeta", "alpha", "bravo"})
	goal := testStringsToObjects([]string{"alpha", "bravo", "zeta"})
	c.sorter(list)
	assert.Empty(cmp.Diff(goal, list, cmpopts.IgnoreUnexported(meta.Object{})))

	list = testStringsToObjects([]string{"19beta", "19alpha", "4tango", "1alpha", "7alpha"})
	goal = testStringsToObjects([]string{"19alpha", "19beta", "1alpha", "4tango", "7alpha"})
	c.sorter(list)
	assert.Empty(cmp.Diff(goal, list, cmpopts.IgnoreUnexported(meta.Object{})))
}

func TestFileSortOrderNatural(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	var c Config

	require.NoError(c.With(FileSortOrderNatural()))
	require.NotNil(c.sorter)

	list := testStringsToObjects([]string{"zeta", "alpha", "bravo"})
	goal := testStringsToObjects([]string{"alpha", "bravo", "zeta"})
	c.sorter(list)
	assert.Empty(cmp.Diff(goal, list, cmpopts.IgnoreUnexported(meta.Object{})))

	list = testStringsToObjects([]string{"19beta", "19alpha", "4tango", "1alpha", "7alpha"})
	goal = testStringsToObjects([]string{"1alpha", "4tango", "7alpha", "19alpha", "19beta"})
	c.sorter(list)
	assert.Empty(cmp.Diff(goal, list, cmpopts.IgnoreUnexported(meta.Object{})))
}
