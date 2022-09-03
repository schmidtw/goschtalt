// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/psanford/memfs"
	"github.com/schmidtw/goschtalt/pkg/meta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			opts:        []Option{KeyCaseLower(), AndCompile()},
		}, {
			description: "A case with an empty option.",
			opts:        []Option{zeroOpt},
		}, {
			description: "An error case where duplicate decoders are added.",
			opts:        []Option{DecoderRegister(&testDecoder{extensions: []string{"json", "json"}})},
			expectedErr: ErrDuplicateFound,
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
	required.NoError(fs2.WriteFile("b/90.json", []byte(`{"madd": "cat", "blue": "bird"}`), 0755))

	fs3 := memfs.New()
	required.NoError(fs3.MkdirAll("b", 0777))
	required.NoError(fs3.WriteFile("b/90.json", []byte(`{"Hello((fail))": "cat", "blue": "bird"}`), 0755))

	fs4 := memfs.New()
	required.NoError(fs4.MkdirAll("b", 0777))
	required.NoError(fs4.WriteFile("b/90.json", []byte(`I'm not valid json!`), 0755))

	type st1 struct {
		Hello string
		Blue  string
		Madd  string
	}

	tests := []struct {
		description string
		opts        []Option
		want        any
		expect      any
		files       []string
		expectedErr error
	}{
		{
			description: "A normal case with options.",
			opts: []Option{
				FileGroup(Group{
					FS:      fs1,
					Paths:   []string{"."},
					Recurse: true,
				}),
				FileGroup(Group{
					FS:      fs2,
					Paths:   []string{"."},
					Recurse: true,
				}),
				DecoderRegister(&testDecoder{extensions: []string{"json"}}),
			},
			want: st1{},
			expect: st1{
				Hello: "Mr. Blue Sky",
				Blue:  "bird",
				Madd:  "cat",
			},
			files: []string{"1.json", "2.json", "3.json", "90.json"},
		}, {
			description: "An empty case.",
			opts: []Option{
				DecoderRegister(&testDecoder{extensions: []string{"json"}}),
			},
			want:   st1{},
			expect: st1{},
		}, {
			description: "A merge failure case.",
			opts: []Option{
				FileGroup(Group{
					FS:      fs1,
					Paths:   []string{"."},
					Recurse: true,
				}),
				FileGroup(Group{
					FS:      fs3,
					Paths:   []string{"."},
					Recurse: true,
				}),
				DecoderRegister(&testDecoder{extensions: []string{"json"}}),
			},
			want:        st1{},
			expect:      st1{},
			expectedErr: meta.ErrConflict,
		}, {
			description: "A decode failure case.",
			opts: []Option{
				FileGroup(Group{
					FS:      fs1,
					Paths:   []string{"."},
					Recurse: true,
				}),
				FileGroup(Group{
					FS:      fs4,
					Paths:   []string{"."},
					Recurse: true,
				}),
				DecoderRegister(&testDecoder{extensions: []string{"json"}}),
			},
			want:        st1{},
			expect:      st1{},
			expectedErr: unknownErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cfg, err := New(tc.opts...)
			require.NoError(err)

			err = cfg.Compile()

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
			if tc.expectedErr != unknownErr {
				assert.ErrorIs(err, tc.expectedErr)
			}

			// check the file order is correct
			got, err := cfg.ShowOrder()
			assert.ErrorIs(err, ErrNotCompiled)
			assert.Empty(got)
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

			cfg, err := New(DecoderRegister(&testDecoder{extensions: []string{"json"}}))
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
			opts:        []Option{DecoderRegister(&testDecoder{extensions: []string{"json"}})},
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
