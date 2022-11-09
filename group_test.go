// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	iofs "io/fs"
	"sort"
	"strings"
	"testing"

	"github.com/psanford/memfs"
	"github.com/schmidtw/goschtalt/pkg/decoder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWalk(t *testing.T) {
	tests := []struct {
		description string
		grp         group
		expected    []string
		expectedErr error
	}{
		{
			description: "Process one file.",
			grp: group{
				paths: []string{"nested/conf/1.json"},
			},
			expected: []string{
				`1.json`,
			},
		}, {
			description: "Process two files.",
			grp: group{
				paths: []string{
					"nested/conf/1.json",
					"nested/4.json",
				},
			},
			expected: []string{
				`1.json`,
				`4.json`,
			},
		},
		{
			description: "Process most files.",
			grp: group{
				paths:   []string{"nested"},
				recurse: true,
			},
			expected: []string{
				`1.json`,
				`2.json`,
				`3.json`,
				`4.json`,
				`ignore`,
			},
		}, {
			description: "Process some files.",
			grp: group{
				paths: []string{"nested"},
			},
			expected: []string{
				`3.json`,
				`4.json`,
			},
		}, {
			description: "Trailing slashes are not allowed.",
			grp: group{
				paths: []string{"nested/"},
			},
			expectedErr: iofs.ErrInvalid,
		}, {
			description: "Absolute addressing is not allowed.",
			grp: group{
				paths: []string{"/nested"},
			},
			expectedErr: iofs.ErrInvalid,
		}, {
			description: "No file or directory with this patth.",
			grp: group{
				paths: []string{"invalid"},
			},
			expectedErr: iofs.ErrNotExist,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fs := memfs.New()
			require.NoError(fs.MkdirAll("nested/conf", 0777))
			require.NoError(fs.WriteFile("nested/conf/1.json", []byte(`{"hello":"world"}`), 0755))
			require.NoError(fs.WriteFile("nested/conf/2.json", []byte(`{"water":"blue"}`), 0755))
			require.NoError(fs.WriteFile("nested/conf/ignore", []byte(`ignore this file`), 0755))
			require.NoError(fs.WriteFile("nested/3.json", []byte(`{"sky":"overcast"}`), 0755))
			require.NoError(fs.WriteFile("nested/4.json", []byte(`{"ground":"green"}`), 0755))
			require.NoError(fs.WriteFile("invalid.json", []byte(`{ground:green}`), 0755))
			tc.grp.fs = fs

			dr := newRegistry[decoder.Decoder]()
			require.NotNil(dr)
			dr.register(&testDecoder{extensions: []string{"json"}})

			got, err := tc.grp.toRecords(".", dr)

			if tc.expectedErr == nil {
				assert.NoError(err)
				require.NotNil(got)
				sort.SliceStable(got, func(i, j int) bool {
					return got[i].name < got[j].name
				})

				require.Equalf(len(tc.expected), len(got),
					"wanted { %s }\n   got { %s }\n",
					strings.Join(tc.expected, ", "),
					func() string {
						list := make([]string, 0, len(got))
						for _, g := range got {
							list = append(list, g.name)
						}
						return strings.Join(list, ", ")
					}())
				for i := range tc.expected {
					assert.Equalf(tc.expected[i], got[i].name,
						"wanted %s\n   got %s\n", tc.expected[i], got[i].name)
				}
				return
			}
			assert.ErrorIs(err, tc.expectedErr)
		})
	}
}
