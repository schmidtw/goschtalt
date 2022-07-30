// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	iofs "io/fs"
	"reflect"
	"sort"
	"testing"

	"github.com/psanford/memfs"
	"github.com/schmidtw/goschtalt/internal/encoding"
	"github.com/schmidtw/goschtalt/internal/encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type groupTestByFile []raw

func (a groupTestByFile) Len() int           { return len(a) }
func (a groupTestByFile) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a groupTestByFile) Less(i, j int) bool { return a[i].file < a[j].file }

func makeTestFs(t *testing.T) iofs.FS {
	require := require.New(t)
	fs := memfs.New()
	require.NoError(fs.MkdirAll("nested/conf", 0777))
	require.NoError(fs.WriteFile("nested/conf/1.json", []byte(`{"hello":"world"}`), 0755))
	require.NoError(fs.WriteFile("nested/conf/2.json", []byte(`{"water":"blue"}`), 0755))
	require.NoError(fs.WriteFile("nested/conf/ignore", []byte(`ignore this file`), 0755))
	require.NoError(fs.WriteFile("nested/3.json", []byte(`{"sky":"overcast"}`), 0755))
	require.NoError(fs.WriteFile("nested/4.json", []byte(`{"ground":"green"}`), 0755))
	require.NoError(fs.WriteFile("invalid.json", []byte(`{ground:green}`), 0755))
	return fs
}

func TestWalk(t *testing.T) {
	tests := []struct {
		description string
		opts        []encoding.Option
		group       Group
		expected    []raw
		expectedErr error
	}{
		{
			description: "Process one file.",
			opts:        []encoding.Option{encoding.WithCodec(json.Codec{})},
			group: Group{
				Paths: []string{"nested/conf/1.json"},
			},
			expected: []raw{
				{
					file: "1.json",
					values: &map[string]any{
						"hello": "world",
					},
				},
			},
		}, {
			description: "Process two files.",
			opts:        []encoding.Option{encoding.WithCodec(json.Codec{})},
			group: Group{
				Paths: []string{
					"nested/conf/1.json",
					"nested/4.json",
				},
			},
			expected: []raw{
				{
					file: "1.json",
					values: &map[string]any{
						"hello": "world",
					},
				}, {
					file: "4.json",
					values: &map[string]any{
						"ground": "green",
					},
				},
			},
		}, {
			description: "Process most files.",
			opts:        []encoding.Option{encoding.WithCodec(json.Codec{})},
			group: Group{
				Paths:   []string{"nested"},
				Recurse: true,
			},
			expected: []raw{
				{
					file: "1.json",
					values: &map[string]any{
						"hello": "world",
					},
				}, {
					file: "2.json",
					values: &map[string]any{
						"water": "blue",
					},
				}, {
					file: "3.json",
					values: &map[string]any{
						"sky": "overcast",
					},
				}, {
					file: "4.json",
					values: &map[string]any{
						"ground": "green",
					},
				},
			},
		}, {
			description: "Process some files.",
			opts:        []encoding.Option{encoding.WithCodec(json.Codec{})},
			group: Group{
				Paths: []string{"nested"},
			},
			expected: []raw{
				{
					file: "3.json",
					values: &map[string]any{
						"sky": "overcast",
					},
				}, {
					file: "4.json",
					values: &map[string]any{
						"ground": "green",
					},
				},
			},
		}, {
			description: "Process all files and fail.",
			opts:        []encoding.Option{encoding.WithCodec(json.Codec{})},
			group: Group{
				Paths:   []string{"."},
				Recurse: true,
			},
			expectedErr: encoding.ErrDecoding,
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

			tc.group.FS = makeTestFs(t)
			r, err := encoding.NewRegistry(tc.opts...)
			require.NotNil(r)
			require.NoError(err)

			got, err := tc.group.walk(r)
			if tc.expectedErr == nil {
				assert.NoError(err)
				require.NotNil(got)
				sort.Sort(groupTestByFile(got))
				assert.True(reflect.DeepEqual(tc.expected, got))
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
			assert.True(reflect.DeepEqual(tc.expected, got))
		})
	}
}
