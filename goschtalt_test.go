// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/goschtalt/goschtalt/pkg/meta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errOpt = errors.New("option error")

func TestNew(t *testing.T) {
	var zeroOpt Option
	tests := []struct {
		description string
		opts        []Option
		expectedErr error
	}{
		{
			description: "A normal case with no options.",
		}, {
			description: "A normal case with options.",
			opts:        []Option{AlterKeyCase(strings.ToLower), AutoCompile()},
		}, {
			description: "A case with an empty option.",
			opts:        []Option{zeroOpt},
		}, {
			description: "An error case.",
			opts:        []Option{WithError(errOpt)},
			expectedErr: errOpt,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			cfg, err := New(tc.opts...)

			if tc.expectedErr == nil {
				assert.NoError(err)
				assert.NotNil(cfg)
				return
			}

			assert.Error(err)
			assert.ErrorIs(err, tc.expectedErr)
		})
	}
}

func TestCompile(t *testing.T) {
	unknownErr := fmt.Errorf("unknown err")

	fs1 := fstest.MapFS{
		"a/1.json": &fstest.MapFile{
			Data: []byte(`{"Hello":"World"}`),
			Mode: 0755,
		},
		"2.json": &fstest.MapFile{
			Data: []byte(`{"Blue":"sky"}`),
			Mode: 0755,
		},
		"3.json": &fstest.MapFile{
			Data: []byte(`{"hello":"Mr. Blue Sky"}`),
			Mode: 0755,
		},
	}

	fs2 := fstest.MapFS{
		"b/90.json": &fstest.MapFile{
			Data: []byte(`{"madd": "cat", "blue": "${thing}"}`),
			Mode: 0755,
		},
	}

	fs3 := fstest.MapFS{
		"b/90.json": &fstest.MapFile{
			Data: []byte(`{"Hello((fail))": "cat", "blue": "bird"}`),
			Mode: 0755,
		},
	}

	fs4 := fstest.MapFS{
		"b/90.json": &fstest.MapFile{
			Data: []byte(`I'm not valid json!`),
			Mode: 0755,
		},
	}

	mapper1 := func(m string) string {
		switch m {
		case "thing":
			return "|bird|"
		}

		return ""
	}

	mapper2 := func(m string) string {
		switch m {
		case "bird":
			return "jay"
		}

		return ""
	}

	// Causes infinite loop
	mapper3 := func(m string) string {
		switch m {
		case "thing":
			return ".${bird}"
		case "bird":
			return ".${jay}"
		case "jay":
			return ".${bird}"
		}

		return ""
	}

	type st1 struct {
		Hello string
		Blue  string
		Madd  string
	}

	tests := []struct {
		description   string
		compileOption bool
		opts          []Option
		want          any
		expect        any
		files         []string
		expectedErr   error
	}{
		{
			description: "A normal case with options.",
			opts: []Option{
				AlterKeyCase(strings.ToLower),
				AddTree(fs1, "."),
				AddTree(fs2, "."),
				WithDecoder(&testDecoder{extensions: []string{"json"}}),
				AutoCompile(),
			},
			want: st1{},
			expect: st1{
				Hello: "Mr. Blue Sky",
				Blue:  "${thing}",
				Madd:  "cat",
			},
			files: []string{"1.json", "2.json", "3.json", "90.json"},
		}, {
			description: "A normal case with an encoded buffer.",
			opts: []Option{
				AlterKeyCase(strings.ToLower),
				AddBuffer("1.json", []byte(`{"Hello": "Mr. Blue Sky"}`)),
				AddBuffer("2.json", []byte(`{"Blue": "${thing}"}`)),
				AddBuffer("3.json", []byte(`{"Madd": "cat"}`)),
				WithDecoder(&testDecoder{extensions: []string{"json"}}),
				AutoCompile(),
			},
			want: st1{},
			expect: st1{
				Hello: "Mr. Blue Sky",
				Blue:  "${thing}",
				Madd:  "cat",
			},
			files: []string{"1.json", "2.json", "3.json"},
		}, {
			description: "A normal case with an encoded buffer function.",
			opts: []Option{
				AlterKeyCase(strings.ToLower),
				AddBuffer("3.json", []byte(`{"Madd": "cat"}`)),
				AddBuffer("2.json", []byte(`{"Blue": "${thing}"}`)),
				AddBufferFn("1.json", func(_ string, _ UnmarshalFunc) ([]byte, error) {
					return []byte(`{"Hello": "Mr. Blue Sky"}`), nil
				}),
				WithDecoder(&testDecoder{extensions: []string{"json"}}),
				AutoCompile(),
			},
			want: st1{},
			expect: st1{
				Hello: "Mr. Blue Sky",
				Blue:  "${thing}",
				Madd:  "cat",
			},
			files: []string{"1.json", "2.json", "3.json"},
		}, {
			description: "A case with an encoded buffer function that looks up something from the tree.",
			opts: []Option{
				AlterKeyCase(strings.ToLower),
				AddBuffer("1.json", []byte(`{"Madd": "cat"}`)),
				AddBufferFn("2.json", func(_ string, un UnmarshalFunc) ([]byte, error) {
					var s string
					_ = un("madd", &s)
					return []byte(fmt.Sprintf(`{"blue": "%s"}`, s)), nil
				}),
				AddBufferFn("3.json", func(_ string, un UnmarshalFunc) ([]byte, error) {
					var s string
					_ = un("blue", &s)
					return []byte(fmt.Sprintf(`{"hello": "%s"}`, s)), nil
				}),
				WithDecoder(&testDecoder{extensions: []string{"json"}}),
				AutoCompile(),
			},
			want: st1{},
			expect: st1{
				Hello: "cat",
				Blue:  "cat",
				Madd:  "cat",
			},
			files: []string{"1.json", "2.json", "3.json"},
		}, {
			description:   "A case with an encoded buffer that is invalid",
			compileOption: true,
			opts: []Option{
				AlterKeyCase(strings.ToLower),
				WithDecoder(&testDecoder{extensions: []string{"json"}}),
				AutoCompile(),
				AddBuffer("1.json", []byte(`invalid`)),
			},
			expectedErr: unknownErr,
		}, {
			description:   "An encoded buffer can't be decoded.",
			compileOption: true,
			opts: []Option{
				AlterKeyCase(strings.ToLower),
				AutoCompile(),
				AddBuffer("1.json", []byte(`invalid`)),
			},
			expectedErr: unknownErr,
		}, {
			description:   "A case with an encoded buffer fn that returns an error",
			compileOption: true,
			opts: []Option{
				AlterKeyCase(strings.ToLower),
				WithDecoder(&testDecoder{extensions: []string{"json"}}),
				AutoCompile(),
				AddBufferFn("3.json", func(_ string, _ UnmarshalFunc) ([]byte, error) {
					return nil, unknownErr
				}),
			},
			expectedErr: unknownErr,
		}, {
			description:   "A case with an encoded buffer fn that returns an invalidly formatted buffer",
			compileOption: true,
			opts: []Option{
				AlterKeyCase(strings.ToLower),
				WithDecoder(&testDecoder{extensions: []string{"json"}}),
				AutoCompile(),
				AddBufferFn("3.json", func(_ string, _ UnmarshalFunc) ([]byte, error) {
					return []byte(`invalid`), nil
				}),
			},
			expectedErr: unknownErr,
		}, {
			description: "A normal case with options including expansion.",
			opts: []Option{
				AlterKeyCase(strings.ToLower),
				AddTree(fs1, "."),
				AddTree(fs2, "."),
				WithDecoder(&testDecoder{extensions: []string{"json"}}),
				Expand(mapper1),
				Expand(mapper2, WithDelimiters("|", "|")),
			},
			want: st1{},
			expect: st1{
				Hello: "Mr. Blue Sky",
				Blue:  "jay",
				Madd:  "cat",
			},
			files: []string{"1.json", "2.json", "3.json", "90.json"},
		}, {
			description: "A normal case with values.",
			opts: []Option{
				AlterKeyCase(strings.ToLower),
				AddValue("record1", "", st1{
					Hello: "Mr. Blue Sky",
					Blue:  "jay",
					Madd:  "cat",
				}),
			},
			want: st1{},
			expect: st1{
				Hello: "Mr. Blue Sky",
				Blue:  "jay",
				Madd:  "cat",
			},
			files: []string{"record1"},
		}, {
			description: "An empty case.",
			opts: []Option{
				AlterKeyCase(strings.ToLower),
				WithDecoder(&testDecoder{extensions: []string{"json"}}),
			},
			want:   st1{},
			expect: st1{},
		}, {
			description: "An empty set of files.",
			opts: []Option{
				AlterKeyCase(strings.ToLower),
				AddFiles(fs1),
				WithDecoder(&testDecoder{extensions: []string{"json"}}),
			},
			want:   st1{},
			expect: st1{},
		}, {
			description: "A merge failure case.",
			opts: []Option{
				AlterKeyCase(strings.ToLower),
				AddTree(fs1, "."),
				AddTree(fs3, "."),
				WithDecoder(&testDecoder{extensions: []string{"json"}}),
			},
			want:        st1{},
			expect:      st1{},
			expectedErr: meta.ErrConflict,
		}, {
			description: "A decode failure case.",
			opts: []Option{
				AlterKeyCase(strings.ToLower),
				AddTree(fs1, "."),
				AddTree(fs4, "."),
				WithDecoder(&testDecoder{extensions: []string{"json"}}),
			},
			want:        st1{},
			expect:      st1{},
			expectedErr: unknownErr,
		}, {
			description: "A recursion case where a failure results",
			opts: []Option{
				AlterKeyCase(strings.ToLower),
				AddTree(fs1, "."),
				AddTree(fs2, "."),
				WithDecoder(&testDecoder{extensions: []string{"json"}}),
				Expand(mapper3),
				AutoCompile(),
			},
			compileOption: true,
			want:          st1{},
			expect:        st1{},
			expectedErr:   unknownErr,
		}, {
			description: "A case where the value decoder errors",
			opts: []Option{
				AlterKeyCase(strings.ToLower),
				AddValue("record1", "", st1{
					Hello: "Mr. Blue Sky",
					Blue:  "jay",
					Madd:  "cat",
				}, testSetResult(5)), // the result must be a pointer
			},
			want:        st1{},
			expect:      st1{},
			expectedErr: unknownErr,
		}, {
			description: "A case where the value doesn't have a record name.",
			opts: []Option{
				AlterKeyCase(strings.ToLower),
				AutoCompile(),
				AddValue("", "", st1{
					Hello: "Mr. Blue Sky",
					Blue:  "jay",
					Madd:  "cat",
				}),
			},
			compileOption: true,
			want:          st1{},
			expect:        st1{},
			expectedErr:   ErrInvalidInput,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cfg, err := New(tc.opts...)

			if !tc.compileOption {
				require.NoError(err)
				err = cfg.Compile()
			}

			if cfg != nil {
				fmt.Println(cfg.Explain())
			}

			if tc.expectedErr == nil {
				assert.NoError(err)
				require.NotNil(cfg)
				err = cfg.Unmarshal("", &tc.want)
				require.NoError(err)

				assert.Empty(cmp.Diff(tc.expect, tc.want))

				// check the file order
				got, err := cfg.ShowOrder()
				require.NoError(err)

				assert.Empty(cmp.Diff(tc.files, got))
				return
			}

			assert.Error(err)
			if !errors.Is(unknownErr, tc.expectedErr) {
				assert.ErrorIs(err, tc.expectedErr)
			}

			if !tc.compileOption {
				// check the file order is correct
				got, err := cfg.ShowOrder()
				assert.ErrorIs(err, ErrNotCompiled)
				assert.Empty(got)
			}
		})
	}
}

func TestOrderList(t *testing.T) {
	tests := []struct {
		description string
		in          []string
		expect      []string
	}{
		{
			description: "An empty list",
		}, {
			description: "A simple list",
			in: []string{
				"9.json",
				"3.txt",
				"1.json",
				"2.json",
			},
			expect: []string{
				"1.json",
				"2.json",
				"9.json",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cfg, err := New(WithDecoder(&testDecoder{extensions: []string{"json"}}))
			require.NotNil(cfg)
			require.NoError(err)

			got := cfg.OrderList(tc.in)

			assert.Empty(cmp.Diff(tc.expect, got))
		})
	}
}

func TestExtensions(t *testing.T) {
	tests := []struct {
		description string
		opts        []Option
		expect      []string
	}{
		{
			description: "An empty list",
		}, {
			description: "A simple list",
			opts:        []Option{WithDecoder(&testDecoder{extensions: []string{"json"}})},
			expect:      []string{"json"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cfg, err := New(tc.opts...)
			require.NotNil(cfg)
			require.NoError(err)

			got := cfg.Extensions()

			assert.Empty(cmp.Diff(tc.expect, got))
		})
	}
}

func TestHash(t *testing.T) {
	tests := []struct {
		description string
		opts        []Option
		expect      uint64
	}{
		{
			description: "An empty list",
			expect:      0xd199e2449a8d3676,
		}, {
			description: "A simple list",
			opts: []Option{
				AddValue("rec", "", map[string]string{"hello": "world"}),
				AutoCompile(),
			},
			expect: 0x66c6ba5f017f3756,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cfg, err := New(tc.opts...)
			require.NotNil(cfg)
			require.NoError(err)

			got := cfg.Hash()

			assert.Equal(tc.expect, got)
		})
	}
}

func TestCompiledAt(t *testing.T) {
	tests := []struct {
		description string
		opts        []Option
		timeZero    bool
	}{
		{
			description: "An empty list",
			timeZero:    true,
		}, {
			description: "A simple list",
			opts: []Option{
				AutoCompile(),
			},
			timeZero: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cfg, err := New(tc.opts...)
			require.NotNil(cfg)
			require.NoError(err)

			got := cfg.CompiledAt()

			if tc.timeZero {
				assert.Equal(time.Time{}, got)
			} else {
				assert.NotEqual(time.Time{}, got)
			}
		})
	}
}
