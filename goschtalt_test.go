// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/psanford/memfs"
	"github.com/schmidtw/goschtalt/pkg/meta"
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

	required := require.New(t)
	fs1 := memfs.New()
	required.NoError(fs1.MkdirAll("a", 0777))
	required.NoError(fs1.WriteFile("a/1.json", []byte(`{"Hello": "World"}`), 0755))
	required.NoError(fs1.WriteFile("2.json", []byte(`{"Blue": "sky"}`), 0755))
	required.NoError(fs1.WriteFile("3.json", []byte(`{"hello": "Mr. Blue Sky"}`), 0755))

	fs2 := memfs.New()
	required.NoError(fs2.MkdirAll("b", 0777))
	required.NoError(fs2.WriteFile("b/90.json", []byte(`{"madd": "cat", "blue": "${thing}"}`), 0755))

	fs3 := memfs.New()
	required.NoError(fs3.MkdirAll("b", 0777))
	required.NoError(fs3.WriteFile("b/90.json", []byte(`{"Hello((fail))": "cat", "blue": "bird"}`), 0755))

	fs4 := memfs.New()
	required.NoError(fs4.MkdirAll("b", 0777))
	required.NoError(fs4.WriteFile("b/90.json", []byte(`I'm not valid json!`), 0755))

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

			fmt.Println(cfg.Explain())

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
