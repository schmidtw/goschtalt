// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"encoding/json"
	"errors"
	"io/fs"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/psanford/memfs"
	"github.com/schmidtw/goschtalt"
	"github.com/schmidtw/goschtalt/pkg/decoder"
	"github.com/schmidtw/goschtalt/pkg/meta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtensions(t *testing.T) {
	assert := assert.New(t)

	want := []string{Extension}
	got := cliDecoder{}.Extensions()

	assert.Empty(cmp.Diff(want, got))
}

func TestDecode(t *testing.T) {
	unknown := errors.New("unknown")
	tests := []struct {
		description string
		fileStr     string
		file        instructions
		expected    meta.Object
		expectedErr error
	}{
		{
			description: "A small test.",
			file: instructions{
				Delimiter: ".",
				Entries: []kvp{
					{Key: "foo.bar.0", Value: "zero"},
					{Key: "foo.bar.1", Value: "one"},
					{Key: "a", Value: "one"},
					{Key: "b", Value: "two"},
				},
			},
			expected: meta.Object{
				Origins: []meta.Origin{{File: "filename.cli"}},
				Map: map[string]meta.Object{
					"foo": {
						Origins: []meta.Origin{{File: "filename.cli"}},
						Map: map[string]meta.Object{
							"bar": {
								Origins: []meta.Origin{{File: "filename.cli"}},
								Array: []meta.Object{
									{
										Origins: []meta.Origin{{File: "filename.cli"}},
										Value:   "zero",
									}, {
										Origins: []meta.Origin{{File: "filename.cli"}},
										Value:   "one",
									},
								},
							},
						},
					},
					"a": {
						Origins: []meta.Origin{{File: "filename.cli"}},
						Value:   "one",
					},
					"b": {
						Origins: []meta.Origin{{File: "filename.cli"}},
						Value:   "two",
					},
				},
			},
		}, {
			description: "A test of invalid json.",
			fileStr:     `{ invalid json }`,
			expectedErr: unknown,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			var d cliDecoder
			var got meta.Object
			var b []byte
			var err error

			if len(tc.fileStr) > 0 {
				b = []byte(tc.fileStr)
			} else {
				b, err = json.Marshal(tc.file)
				require.NoError(err)
			}
			ctx := decoder.Context{
				Filename:  "filename.cli",
				Delimiter: ".",
			}
			err = d.Decode(ctx, b, &got)

			if tc.expectedErr == nil {
				assert.NoError(err)
				assert.Empty(cmp.Diff(tc.expected, got, cmpopts.IgnoreUnexported(meta.Object{})))
				return
			}

			if tc.expectedErr == unknown {
				assert.Error(err)
				return
			}

			assert.ErrorIs(err, tc.expectedErr)
		})
	}
}

type testDecoder struct{}

func (d testDecoder) Extensions() []string {
	return []string{"json"}
}

// Decode decodes a byte arreay into the meta.Object tree.
func (d testDecoder) Decode(ctx decoder.Context, b []byte, m *meta.Object) error {
	var raw map[string]any

	if len(b) == 0 {
		*m = meta.Object{}
		return nil
	}

	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	*m = meta.ObjectFromRaw(raw)
	return nil
}

type expect struct {
	key      string
	val      string
	filename string
}

func TestEndToEnd(t *testing.T) {
	unknown := errors.New("unknown")

	tests := []struct {
		description string
		args        []string
		expected    []expect
		notExpected []expect
		expectedErr error
	}{
		{
			description: "A simple test.",
			args: []string{
				"-d", "/etc",
				"-f", "local.json",
				"--kvp", "cli", "yes",
			},
			expected: []expect{
				{key: "foo", val: "etc", filename: "test.json"},
				{key: "bar", val: "local", filename: "local.json"},
				{key: "cli", val: "yes", filename: "cli.cli"},
			},
			notExpected: []expect{
				{key: "car", val: "red", filename: "other.json"},
				{key: "food", val: "nuts", filename: "test2.json"},
			},
		}, {
			description: "Include multiple files.",
			args: []string{
				"-d", "/etc",
				"-f", "local.json",
				"-f", "other.json",
				"--kvp", "cli", "yes",
			},
			expected: []expect{
				{key: "foo", val: "etc", filename: "test.json"},
				{key: "bar", val: "local", filename: "local.json"},
				{key: "cli", val: "yes", filename: "cli.cli"},
				{key: "car", val: "red", filename: "other.json"},
			},
			notExpected: []expect{
				{key: "food", val: "nuts", filename: "test2.json"},
			},
		}, {
			description: "A simple test with recursion.",
			args: []string{
				"-r", "/etc",
				"-f", "local.json",
				"--kvp", "cli", "yes",
			},
			expected: []expect{
				{key: "foo", val: "etc", filename: "test.json"},
				{key: "bar", val: "local", filename: "local.json"},
				{key: "cli", val: "yes", filename: "cli.cli"},
				{key: "food", val: "nuts", filename: "test2.json"},
			},
			notExpected: []expect{
				{key: "car", val: "red", filename: "other.json"},
			},
		}, {
			description: "A simple test - long names.",
			args: []string{
				"--dir", "/etc",
				"--file", "local.json",
				"--kvp", "cli", "yes",
			},
			expected: []expect{
				{key: "foo", val: "etc", filename: "test.json"},
				{key: "bar", val: "local", filename: "local.json"},
				{key: "cli", val: "yes", filename: "cli.cli"},
			},
			notExpected: []expect{
				{key: "car", val: "red", filename: "other.json"},
				{key: "food", val: "nuts", filename: "test2.json"},
			},
		}, {
			description: "A simple test with recursion - long names.",
			args: []string{
				"--recurse", "/etc",
				"--file", "local.json",
				"--kvp", "cli", "yes",
			},
			expected: []expect{
				{key: "foo", val: "etc", filename: "test.json"},
				{key: "bar", val: "local", filename: "local.json"},
				{key: "cli", val: "yes", filename: "cli.cli"},
				{key: "food", val: "nuts", filename: "test2.json"},
			},
			notExpected: []expect{
				{key: "car", val: "red", filename: "other.json"},
			},
		}, {
			description: "Missing the argument for -f",
			args: []string{
				"-r", "/etc",
				"-f",
				"--kvp", "cli", "yes",
			},
			expectedErr: ErrInput,
		}, {
			description: "Missing the argument for -f at the end",
			args: []string{
				"-r", "/etc",
				"--kvp", "cli", "yes",
				"-f",
			},
			expectedErr: ErrInput,
		}, {
			description: "Missing the argument for -d",
			args: []string{
				"-r", "/etc",
				"-d",
				"--kvp", "cli", "yes",
			},
			expectedErr: ErrInput,
		}, {
			description: "Missing the argument for -d at the end",
			args: []string{
				"-r", "/etc",
				"--kvp", "cli", "yes",
				"-d",
			},
			expectedErr: ErrInput,
		}, {
			description: "Missing the argument for -r",
			args: []string{
				"-r", "/etc",
				"-r",
				"--kvp", "cli", "yes",
			},
			expectedErr: ErrInput,
		}, {
			description: "Missing the argument for -r at the end",
			args: []string{
				"-r", "/etc",
				"--kvp", "cli", "yes",
				"-r",
			},
			expectedErr: ErrInput,
		}, {
			description: "Missing the argument for --kvp at the end",
			args: []string{
				"-r", "/etc",
				"--kvp", "cli",
			},
			expectedErr: ErrInput,
		}, {
			description: "Invalid argument",
			args: []string{
				"-rats", "/etc",
			},
			expectedErr: ErrInput,
		}, {
			description: "Missing file.",
			args: []string{
				"-f", "missing.json",
			},
			expectedErr: unknown,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			dirfs := func(s string) fs.FS {
				switch s {
				case "/etc":
					rv := memfs.New()
					require.NoError(rv.MkdirAll("foo", 0777))
					require.NoError(rv.WriteFile("foo/test2.json", []byte(`{"food":"nuts"}`), 0755))
					require.NoError(rv.WriteFile("test.json", []byte(`{"foo":"etc"}`), 0755))
					return rv
				case ".":
					rv := memfs.New()
					require.NoError(rv.WriteFile("local.json", []byte(`{"bar":"local"}`), 0755))
					require.NoError(rv.WriteFile("other.json", []byte(`{"car":"red"}`), 0755))
					return rv
				}

				require.Fail("invalid request")
				return nil
			}

			allOpts := append(Options("cli", ".", tc.args, dirfs), goschtalt.DecoderRegister(testDecoder{}))
			c, err := goschtalt.New(allOpts...)
			if err == nil {
				err = c.Compile()
			}

			if tc.expectedErr == nil {
				assert.NoError(err)
				for _, val := range tc.expected {
					got, origin, err := c.FetchWithOrigin(val.key)
					assert.Equal(val.val, got)
					assert.NoError(err)
					if len(origin) == 1 {
						assert.Equal(val.filename, origin[0].File)
					}
				}
				for _, val := range tc.notExpected {
					got, origin, err := c.FetchWithOrigin(val.key)
					assert.Nil(got)
					assert.Nil(origin)
					assert.ErrorIs(err, meta.ErrNotFound)
				}
				return
			}

			if tc.expectedErr == unknown {
				assert.Error(err)
				return
			}

			assert.ErrorIs(err, tc.expectedErr)
		})
	}
}
