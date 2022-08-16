// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	iofs "io/fs"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/psanford/memfs"
	"github.com/schmidtw/goschtalt/pkg/meta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWalk(t *testing.T) {
	tests := []struct {
		description string
		group       Group
		expected    []string
		expectedErr error
	}{
		{
			description: "Process one file.",
			group: Group{
				Paths: []string{"nested/conf/1.json"},
			},
			expected: []string{
				`1.json`, `{"hello":"world"}`,
			},
		}, {
			description: "Process two files.",
			group: Group{
				Paths: []string{
					"nested/conf/1.json",
					"nested/4.json",
				},
			},
			expected: []string{
				`1.json`, `{"hello":"world"}`,
				`4.json`, `{"ground":"green"}`,
			},
		},
		{
			description: "Process most files.",
			group: Group{
				Paths:   []string{"nested"},
				Recurse: true,
			},
			expected: []string{
				`1.json`, `{"hello":"world"}`,
				`2.json`, `{"water":"blue"}`,
				`3.json`, `{"sky":"overcast"}`,
				`4.json`, `{"ground":"green"}`,
			},
		}, {
			description: "Process some files.",
			group: Group{
				Paths: []string{"nested"},
			},
			expected: []string{
				`3.json`, `{"sky":"overcast"}`,
				`4.json`, `{"ground":"green"}`,
			},
		}, {
			description: "Process all files and fail.",
			group: Group{
				Paths:   []string{"."},
				Recurse: true,
			},
			expectedErr: ErrDecoding,
		}, {
			description: "Trailing slashes are not allowed.",
			group: Group{
				Paths: []string{"nested/"},
			},
			expectedErr: iofs.ErrInvalid,
		}, {
			description: "Absolute addressing is not allowed.",
			group: Group{
				Paths: []string{"/nested"},
			},
			expectedErr: iofs.ErrInvalid,
		}, {
			description: "No file or directory with this patth.",
			group: Group{
				Paths: []string{"invalid"},
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
			tc.group.FS = fs

			dr := newDecoderRegistry()
			require.NotNil(dr)
			err := dr.register(&testDecoder{extensions: []string{"json"}})
			require.NoError(err)

			got, err := tc.group.walk(dr)
			if tc.expectedErr == nil {
				assert.NoError(err)
				require.NotNil(got)
				sort.SliceStable(got, func(i, j int) bool {
					return got[i].Origins[0].File < got[j].Origins[0].File
				})

				var expected []meta.Object

				for i := 0; i < len(tc.expected); i += 2 {
					file := tc.expected[i]
					data := tc.expected[i+1]
					tree := decode(file, data)
					expected = append(expected, tree)
				}
				assert.Empty(cmp.Diff(expected, got, cmpopts.IgnoreUnexported(meta.Object{})))
				return
			}
			assert.ErrorIs(err, tc.expectedErr)
		})
	}
}

func TestMatchExts(t *testing.T) {
	tests := []struct {
		description string
		exts        []string
		files       []string
		expected    []string
	}{
		{
			description: "Simple match",
			exts:        []string{"json", "yaml", "yml"},
			files: []string{
				"dir/file.json",
				"file.JSON",
				"other.yml",
				"a.tricky.file.json.that.really.isnt",
			},
			expected: []string{
				"dir/file.json",
				"file.JSON",
				"other.yml",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			got := matchExts(tc.exts, tc.files)
			assert.Empty(cmp.Diff(tc.expected, got))
		})
	}
}
