// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/schmidtw/goschtalt/pkg/meta"
	"github.com/stretchr/testify/assert"
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
		in               meta.Object
		expected         string
		expectedExtended string
	}{
		{
			description:      "A test of empty.",
			expected:         "null\n",
			expectedExtended: "null\n",
		},
		{
			description: "A simple test.",
			// Input vector in yaml:
			//candy: bar
			//cats:
			//    - madd
			//    - tabby
			//other:
			//    things:
			//        red: balloons
			//        green:
			//            - grass
			//            - ground
			//            - water
			//    trending: now
			in: meta.Object{
				Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 1, Col: 1}},
				Type:    meta.Map,
				Map: map[string]meta.Object{
					"candy": meta.Object{
						Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 1, Col: 8}},
						Type:    meta.Value,
						Value:   "bar",
					},
					"cats": meta.Object{
						Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 2, Col: 1}},
						Type:    meta.Array,
						Array: []meta.Object{
							meta.Object{
								Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 3, Col: 7}},
								Type:    meta.Value,
								Value:   "madd",
							},
							meta.Object{
								Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 4, Col: 7}},
								Type:    meta.Value,
								Value:   "tabby",
							},
						},
					},
					"other": meta.Object{
						Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 5, Col: 1}},
						Type:    meta.Map,
						Map: map[string]meta.Object{
							"things": meta.Object{
								Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 6, Col: 5}},
								Type:    meta.Map,
								Map: map[string]meta.Object{
									"red": meta.Object{
										Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 7, Col: 14}},
										Type:    meta.Value,
										Value:   "balloons",
									},
									"green": meta.Object{
										Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 8, Col: 9}},
										Type:    meta.Array,
										Array: []meta.Object{
											meta.Object{
												Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 9, Col: 15}},
												Type:    meta.Value,
												Value:   "grass",
											},
											meta.Object{
												Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 10, Col: 15}},
												Type:    meta.Value,
												Value:   "ground",
											},
											meta.Object{
												Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 11, Col: 15}},
												Type:    meta.Value,
												Value:   "water",
											},
										},
									},
								},
							},
							"trending": meta.Object{
								Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 12, Col: 15}},
								Type:    meta.Value,
								Value:   "now",
							},
						},
					},
				},
			},
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
			expectedExtended: `candy: bar # file.yml:1[8]
cats: # file.yml:2[1]
    - madd # file.yml:3[7]
    - tabby # file.yml:4[7]
other: # file.yml:5[1]
    things: # file.yml:6[5]
        green: # file.yml:8[9]
            - grass # file.yml:9[15]
            - ground # file.yml:10[15]
            - water # file.yml:11[15]
        red: balloons # file.yml:7[14]
    trending: now # file.yml:12[15]
`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			//require := require.New(t)

			var e Encoder
			got, err := e.EncodeExtended(tc.in)
			assert.NoError(err)
			assert.Empty(cmp.Diff(tc.expectedExtended, string(got)))

			raw := tc.in.ToRaw()

			got, err = e.Encode(raw)
			assert.NoError(err)
			assert.Empty(cmp.Diff(tc.expected, string(got)))
		})
	}
}
