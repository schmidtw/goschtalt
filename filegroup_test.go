// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"errors"
	iofs "io/fs"
	"sort"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/schmidtw/goschtalt/pkg/decoder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWalk(t *testing.T) {
	tests := []struct {
		description string
		grp         filegroup
		expected    []string
		expectedErr error
	}{
		{
			description: "Process one file.",
			grp: filegroup{
				paths: []string{"nested/conf/1.json"},
			},
			expected: []string{
				`1.json`,
			},
		}, {
			description: "Process two files.",
			grp: filegroup{
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
			grp: filegroup{
				paths:   []string{"nested"},
				recurse: true,
			},
			expected: []string{
				`1.json`,
				`2.json`,
				`3.json`,
				`4.json`,
			},
		}, {
			description: "Process some files.",
			grp: filegroup{
				paths: []string{"nested"},
			},
			expected: []string{
				`3.json`,
				`4.json`,
			},
		}, {
			description: "Trailing slashes are not allowed.",
			grp: filegroup{
				paths: []string{"nested/"},
			},
			expectedErr: iofs.ErrInvalid,
		}, {
			description: "Absolute addressing is not allowed.",
			grp: filegroup{
				paths: []string{"/nested"},
			},
			expectedErr: iofs.ErrInvalid,
		}, {
			description: "No file or directory with this patth.",
			grp: filegroup{
				paths: []string{"invalid"},
			},
			expectedErr: iofs.ErrNotExist,
		}, {
			description: "Ensure file is decoded.",
			grp: filegroup{
				paths:     []string{"ignore.txt"},
				exactFile: true,
			},
			expectedErr: ErrCodecNotFound,
		}, {
			description: "Ensure file is present.",
			grp: filegroup{
				paths:     []string{"fake.json"},
				exactFile: true,
			},
			expectedErr: iofs.ErrInvalid,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fs := fstest.MapFS{
				"nested/conf/1.json": &fstest.MapFile{
					Data: []byte(`{"hello":"world"}`),
					Mode: 0755,
				},
				"nested/conf/2.json": &fstest.MapFile{
					Data: []byte(`{"water":"blue"}`),
					Mode: 0755,
				},
				"nested/conf/ignore": &fstest.MapFile{
					Data: []byte(`ignore this file`),
					Mode: 0755,
				},
				"nested/3.json": &fstest.MapFile{
					Data: []byte(`{"sky":"overcast"}`),
					Mode: 0755,
				},
				"nested/4.json": &fstest.MapFile{
					Data: []byte(`{"ground":"green"}`),
					Mode: 0755,
				},
				"fake.json/ignored": &fstest.MapFile{
					Data: []byte(`ignore this file`),
					Mode: 0755,
				},
				"invalid.json": &fstest.MapFile{
					Data: []byte(`{ground:green}`),
					Mode: 0755,
				},
				"ignore.txt": &fstest.MapFile{
					Data: []byte(`ignore this file`),
					Mode: 0755,
				},
			}
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

// TODO: It would be nice to test random file system failures.
func TestToRecord(t *testing.T) {
	unknown := errors.New("unknown")
	tests := []struct {
		description string
		file        string
		expectedNil bool
		expectedErr error
	}{
		{
			description: "Test with an invalid file.",
			file:        "missing",
			expectedErr: unknown,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			g := filegroup{
				fs: fstest.MapFS{
					"1.json": &fstest.MapFile{
						Data: []byte(`{"hello":"world"}`),
						Mode: 0755,
					},
				},
			}

			dr := newRegistry[decoder.Decoder]()
			require.NotNil(dr)
			dr.register(&testDecoder{extensions: []string{"json"}})

			got, err := g.toRecord(tc.file, ".", dr)

			if tc.expectedErr == nil {
				if tc.expectedNil {
					assert.Nil(got)
				} else {
					assert.NotNil(got)
				}
				return
			}

			assert.Nil(got)
			if errors.Is(unknown, tc.expectedErr) {
				assert.Error(err)
				return
			}

			assert.ErrorIs(err, tc.expectedErr)
		})
	}
}
