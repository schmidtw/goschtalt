// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"errors"
	"fmt"
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
			opts:        []Option{AutoCompile()},
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
	testErr := fmt.Errorf("test err")

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
			Data: []byte(`{"Hello":"Mr. Blue Sky"}`),
			Mode: 0755,
		},
	}

	fs2 := fstest.MapFS{
		"b/90.json": &fstest.MapFile{
			Data: []byte(`{"Madd": "cat", "Blue": "${thing}"}`),
			Mode: 0755,
		},
	}

	fs3 := fstest.MapFS{
		"b/90.json": &fstest.MapFile{
			Data: []byte(`{"Hello((fail))": "cat", "Blue": "bird"}`),
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

	type st2 struct {
		Hello string
		Blue  string
		Madd  string
		Func  func()
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
				AddBuffer("1.json", []byte(`{"Madd": "cat"}`)),
				AddBufferFn("2.json", func(_ string, un UnmarshalFunc) ([]byte, error) {
					var s string
					_ = un("Madd", &s)
					return []byte(fmt.Sprintf(`{"Blue": "%s"}`, s)), nil
				}),
				AddBufferFn("3.json", func(_ string, un UnmarshalFunc) ([]byte, error) {
					var s string
					_ = un("Blue", &s)
					return []byte(fmt.Sprintf(`{"Hello": "%s"}`, s)), nil
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
				WithDecoder(&testDecoder{extensions: []string{"json"}}),
				AutoCompile(),
				AddBuffer("1.json", []byte(`invalid`)),
			},
			expectedErr: unknownErr,
		}, {
			description:   "An encoded buffer can't be decoded.",
			compileOption: true,
			opts: []Option{
				AutoCompile(),
				AddBuffer("1.json", []byte(`invalid`)),
			},
			expectedErr: unknownErr,
		}, {
			description:   "An encoded buffer with an option that causes a failure.",
			compileOption: true,
			opts: []Option{
				WithDecoder(&testDecoder{extensions: []string{"json"}}),
				AutoCompile(),
				AddBuffer("1.json", []byte(`{}`), WithError(testErr)),
			},
			expectedErr: testErr,
		}, {
			description:   "A case with an encoded buffer fn that returns an error",
			compileOption: true,
			opts: []Option{
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
				AddValue("record1", Root, st1{
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
			description: "A normal case with non-serializable values being errors.",
			opts: []Option{
				AddValue("record1", Root,
					st1{
						Hello: "Mr. Blue Sky",
						Blue:  "jay",
						Madd:  "cat",
					},
					FailOnNonSerializable()),
			},
			want: st1{},
			expect: st1{
				Hello: "Mr. Blue Sky",
				Blue:  "jay",
				Madd:  "cat",
			},
			files: []string{"record1"},
		}, {
			description: "A normal case with non-serializable values being dropped.",
			opts: []Option{
				AddValue("record1", Root,
					st2{
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
				WithDecoder(&testDecoder{extensions: []string{"json"}}),
			},
			want:   st1{},
			expect: st1{},
		}, {
			description: "An empty set of files.",
			opts: []Option{
				AddFiles(fs1),
				WithDecoder(&testDecoder{extensions: []string{"json"}}),
			},
			want:   st1{},
			expect: st1{},
		}, {
			description: "A merge failure case.",
			opts: []Option{
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
				AddValue("record1", Root, st1{
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
				AutoCompile(),
				AddValue("", Root, st1{
					Hello: "Mr. Blue Sky",
					Blue:  "jay",
					Madd:  "cat",
				}),
			},
			compileOption: true,
			want:          st1{},
			expect:        st1{},
			expectedErr:   ErrInvalidInput,
		}, {
			description: "A case where the value function returns an error.",
			opts: []Option{
				AutoCompile(),
				AddValueFn("record", Root,
					func(string, UnmarshalFunc) (any, error) {
						return nil, testErr
					},
				),
			},
			compileOption: true,
			want:          st1{},
			expect:        st1{},
			expectedErr:   testErr,
		}, {
			description: "A case where the decode hook is invalid.",
			opts: []Option{
				AddValue("record", Root, st1{
					Hello: "Mr. Blue Sky",
					Blue:  "jay",
					Madd:  "cat",
				}, DecodeHook(func() {})),
			},
			want:        st1{},
			expect:      st1{},
			expectedErr: unknownErr,
		}, {
			description: "A case where the an option is/becomes an error.",
			opts: []Option{
				AddValue("record", Root, st1{
					Hello: "Mr. Blue Sky",
					Blue:  "jay",
					Madd:  "cat",
				},
					// Act like everything is fine the first time through, but then
					// fail the 2nd time to trigger a failure during the compile
					// code.
					testSetError([]error{nil, testErr}),
				),
			},
			want:        st1{},
			expect:      st1{},
			expectedErr: testErr,
		}, {
			description: "A case with non-serializable values producing an error.",
			opts: []Option{
				AddValue("record1", Root,
					st2{
						Hello: "Mr. Blue Sky",
						Blue:  "jay",
						Madd:  "cat",
						Func:  func() {},
					},
					FailOnNonSerializable(),
				),
			},
			want:        st1{},
			expect:      st1{},
			expectedErr: meta.ErrNonSerializable,
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

			var tell string
			if cfg != nil {
				tell = cfg.Explain()
			}

			if tc.expectedErr == nil {
				assert.NoError(err)
				require.NotNil(cfg)
				err = cfg.Unmarshal(Root, &tc.want)
				require.NoError(err)

				assert.Empty(cmp.Diff(tc.expect, tc.want))

				// check the file order
				got, err := cfg.ShowOrder()
				require.NoError(err)
				assert.Empty(cmp.Diff(tc.files, got))

				assert.NotEmpty(tell)
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
				assert.NotEmpty(tell)
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
				AddValue("rec", Root, map[string]string{"hello": "world"}),
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
