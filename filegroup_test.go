// SPDX-FileCopyrightText: 2022-2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"errors"
	iofs "io/fs"
	"path"
	"sort"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/goschtalt/goschtalt/pkg/decoder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Implement a fake file system with files that fail in specific ways to test
// the different failure modes possible.
type fakeFS struct{}

func (fakeFS) Open(name string) (iofs.File, error) {
	if name == "stat-fails.json" {
		return &statFailsFile{}, nil
	}
	if name == "read-fails.json" {
		return &readFailsFile{}, nil
	}
	return nil, nil
}

type statFailsFile struct{}

func (statFailsFile) Stat() (iofs.FileInfo, error) { return nil, iofs.ErrPermission }
func (statFailsFile) Read([]byte) (int, error)     { return 0, nil }
func (statFailsFile) Close() error                 { return nil }

type readFailsFile struct{}

func (readFailsFile) Stat() (iofs.FileInfo, error) { return &readFailsFileInfo{}, nil }
func (readFailsFile) Read([]byte) (int, error)     { return 0, iofs.ErrPermission }
func (readFailsFile) Close() error                 { return nil }

type readFailsFileInfo struct{}

func (readFailsFileInfo) Name() string        { return "read-fails.json" }
func (readFailsFileInfo) Size() int64         { return 1234 }
func (readFailsFileInfo) Mode() iofs.FileMode { return iofs.ModeCharDevice }
func (readFailsFileInfo) ModTime() time.Time  { return time.Time{} }
func (readFailsFileInfo) IsDir() bool         { return false }
func (readFailsFileInfo) Sys() any            { return nil }

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
			description: "Trailing slashes are stripped.",
			grp: filegroup{
				paths: []string{"nested/"},
			},
			expected: []string{
				`3.json`,
				`4.json`,
			},
		}, {
			description: "Non-clean path is cleaned.",
			grp: filegroup{
				paths: []string{"./a/../nested"},
			},
			expected: []string{
				`3.json`,
				`4.json`,
			},
		}, {
			description: "Invalid file glob.",
			grp: filegroup{
				paths: []string{"invalid["},
			},
			expectedErr: path.ErrBadPattern,
		}, {
			description: "Absolute path is not allowed.",
			grp: filegroup{
				paths: []string{"/nested"},
			},
			expectedErr: iofs.ErrInvalid,
		}, {
			description: "A relative path can't be outside the fs.",
			grp: filegroup{
				paths: []string{"../nested"},
			},
			expectedErr: iofs.ErrInvalid,
		}, {
			description: "No file or directory with this path.",
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
		}, {
			description: "Ensure Stat() failures are handled.",
			grp: filegroup{
				fs:        &fakeFS{},
				paths:     []string{"stat-fails.json"},
				exactFile: true,
			},
			expectedErr: iofs.ErrNotExist,
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
		grp         filegroup
		expectedNil bool
		expectedErr error
	}{
		{
			description: "Test with an invalid file.",
			file:        "missing",
			grp: filegroup{
				fs: fstest.MapFS{
					"1.json": &fstest.MapFile{
						Data: []byte(`{"hello":"world"}`),
						Mode: 0755,
					},
				},
			},
			expectedErr: unknown,
		}, {
			description: "Test with an failing stat call.",
			file:        "stat-fails.json",
			grp: filegroup{
				fs: &fakeFS{},
			},
			expectedErr: iofs.ErrPermission,
		}, {
			description: "Ensure ReadAll() failures are handled.",
			file:        "read-fails.json",
			grp: filegroup{
				fs:        &fakeFS{},
				exactFile: true,
			},
			expectedErr: iofs.ErrPermission,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			dr := newRegistry[decoder.Decoder]()
			require.NotNil(dr)
			dr.register(&testDecoder{extensions: []string{"json"}})

			got, err := tc.grp.toRecord(tc.file, ".", dr)

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

func TestEnumeratePath(t *testing.T) {
	unknown := errors.New("unknown")
	tests := []struct {
		description string
		file        string
		grp         filegroup
		expectedNil bool
		expectedErr error
	}{
		{
			description: "Ensure Stat() failures are handled & skipped over.",
			file:        "stat-fails.json",
			grp: filegroup{
				fs:        &fakeFS{},
				paths:     []string{"stat-fails.json"},
				exactFile: true,
			},
			expectedNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			got, err := tc.grp.enumeratePath(tc.file)

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
