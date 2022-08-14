// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/schmidtw/goschtalt/pkg/meta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yml "gopkg.in/yaml.v3"
)

func TestExtensions(t *testing.T) {
	assert := assert.New(t)

	var e Encoder
	got := e.Extensions()

	assert.Empty(cmp.Diff([]string{"yaml", "yml"}, got))
}

func TestEncodeExtended(t *testing.T) {
	tests := []struct {
		description      string
		in               string
		expected         string
		expectedExtended string
	}{
		{
			description:      "A test of empty.",
			in:               `---`,
			expected:         "null\n",
			expectedExtended: "null\n",
		},
		{
			description: "A simple test.",
			in: `---
candy: bar
cats:
  - madd
  - tabby
other:
  things:
    red: balloons
    green: [ grass, ground, water ]
  trending: now`,
			expected: `candy: bar
cats:
    - madd
    - tabby
other:
    things:
        green:
            - grass
            - ground
            - water
        red: balloons
    trending: now
`,
			expectedExtended: `candy: bar # file.yml:2[3]
cats: # file.yml:1[???]
    - madd # file.yml:4[9]
    - tabby # file.yml:5[12]
other: # file.yml:1[???]
    things: # file.yml:6[15]
        green: # file.yml:7[18]
            - grass # file.yml:9[24]
            - ground # file.yml:10[27]
            - water # file.yml:11[30]
        red: balloons # file.yml:12[33]
    trending: now # file.yml:13[36]
`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			var e Encoder
			var in any
			require.NoError(yml.Unmarshal([]byte(tc.in), &in))
			objs := meta.ObjectFromRaw(in)
			origin := meta.Origin{
				File: "file.yml",
				Line: 1,
			}
			objs = addOrigin(objs, &origin)

			got, err := e.EncodeExtended(objs)
			assert.NoError(err)
			assert.Empty(cmp.Diff(tc.expectedExtended, string(got)))

			got, err = e.Encode(in)
			assert.NoError(err)
			assert.Empty(cmp.Diff(tc.expected, string(got)))
		})
	}
}

// Test Utilities //////////////////////////////////////////////////////////////

func addOrigin(obj meta.Object, origin *meta.Origin) meta.Object {
	obj.Origins = append(obj.Origins, *origin)
	origin.Line++   // Not accurate, but interesting.
	origin.Col += 3 // Not accurate, but interesting.
	if origin.Col > 80 {
		origin.Col = 1
	}

	switch obj.Type {
	case meta.Array:
		array := make([]meta.Object, len(obj.Array))
		for i, val := range obj.Array {
			array[i] = addOrigin(val, origin)
		}
		obj.Array = array
	case meta.Map:
		// Ensure ordering so we can compare later.
		var keys []string
		for key := range obj.Map {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		m := make(map[string]meta.Object)

		for _, key := range keys {
			m[key] = addOrigin(obj.Map[key], origin)
		}
		obj.Map = m
	}

	return obj
}
