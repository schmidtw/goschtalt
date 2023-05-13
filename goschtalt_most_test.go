// SPDX-FileCopyrightText: 2022-2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

//go:build !windows && !android

package goschtalt

import (
	"errors"
	"fmt"
	"testing"
	"testing/fstest"

	"github.com/goschtalt/goschtalt/pkg/debug"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStdCfgLayoutNotWin(t *testing.T) {
	assert := assert.New(t)
	got := StdCfgLayout("name")

	// Only make sure the other things are called.  Other tests ensure the
	// functionality works.
	assert.NotNil(got)
}

func TestPopulate(t *testing.T) {
	tests := []struct {
		description string
		home        string
		homeSet     bool
	}{
		{
			description: "no HOME set",
		}, {
			home:    "example",
			homeSet: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			var s stdLocations

			t.Setenv("HOME", tc.home)

			s.Populate("foo")

			assert.NotNil(s.local)
			assert.NotNil(s.root)
			if tc.homeSet {
				assert.NotNil(s.home)
				assert.NotNil(s.homeTree)
			} else {
				assert.Nil(s.home)
				assert.Nil(s.homeTree)
			}
			assert.NotNil(s.etc)
			assert.NotNil(s.etcTree)
		})
	}
}

func TestCompileNotWin(t *testing.T) {
	unknownErr := fmt.Errorf("unknown err")
	remappings := debug.Collect{}

	none := fstest.MapFS{}

	local := fstest.MapFS{
		"1.json": &fstest.MapFile{
			Data: []byte(`{"Status": "local - wanted"}`),
			Mode: 0755,
		},
		"dir/2.json": &fstest.MapFile{
			Data: []byte(`{"Other": "local - wanted"}`),
			Mode: 0755,
		},
	}

	home := fstest.MapFS{
		"example.json": &fstest.MapFile{
			Data: []byte(`{"Status": "home - wanted"}`),
			Mode: 0755,
		},
	}

	homeTree := fstest.MapFS{
		"example.json": &fstest.MapFile{
			Data: []byte(`{"Status": "home - wanted"}`),
			Mode: 0755,
		},
		"conf.d/2.json": &fstest.MapFile{
			Data: []byte(`{"Other": "home - wanted"}`),
			Mode: 0755,
		},
	}
	etc := fstest.MapFS{
		"example.json": &fstest.MapFile{
			Data: []byte(`{"Status": "etc - wanted"}`),
			Mode: 0755,
		},
	}

	etcTree := fstest.MapFS{
		"example.json": &fstest.MapFile{
			Data: []byte(`{"Status": "etc - wanted"}`),
			Mode: 0755,
		},
		"conf.d/2.json": &fstest.MapFile{
			Data: []byte(`{"Other": "etc - wanted"}`),
			Mode: 0755,
		},
	}

	never := fstest.MapFS{
		"example.json": &fstest.MapFile{
			Data: []byte(`{"Status": "never wanted"}`),
			Mode: 0755,
		},
	}

	type st struct {
		Status string
		Other  string
	}

	tests := []struct {
		description    string
		compileOption  bool
		opts           []Option
		want           any
		key            string
		expect         any
		files          []string
		expectedRemaps map[string]string
		expectedErr    error
		compare        func(assert *assert.Assertions, a, b any) bool
	}{
		{
			description: "local - one file in the list",
			opts: []Option{
				WithDecoder(&testDecoder{extensions: []string{"json"}}),
				nonWinStdCfgLayout("example", []string{"./1.json"}, stdLocations{
					local:    local,
					root:     local,
					home:     home,
					homeTree: homeTree,
					etc:      etc,
					etcTree:  etcTree,
				}),
				AddFile(never, "example.json"),
			},
			want: st{},
			expect: st{
				Status: "local - wanted",
			},
			files: []string{"1.json"},
		}, {
			description: "local - one dir in the list",
			opts: []Option{
				WithDecoder(&testDecoder{extensions: []string{"json"}}),
				nonWinStdCfgLayout("example", []string{"./dir"}, stdLocations{
					local:    local,
					root:     local,
					home:     home,
					homeTree: homeTree,
					etc:      etc,
					etcTree:  etcTree,
				}),
				AddFile(never, "example.json"),
			},
			want: st{},
			expect: st{
				Other: "local - wanted",
			},
			files: []string{"2.json"},
		}, {
			description: "local - a file and dir in the list",
			opts: []Option{
				WithDecoder(&testDecoder{extensions: []string{"json"}}),
				nonWinStdCfgLayout("example", []string{"./dir", "./1.json"}, stdLocations{
					local:    local,
					root:     local,
					home:     home,
					homeTree: homeTree,
					etc:      etc,
					etcTree:  etcTree,
				}),
				AddFile(never, "example.json"),
			},
			want: st{},
			expect: st{
				Status: "local - wanted",
				Other:  "local - wanted",
			},
			files: []string{"1.json", "2.json"},
		}, {
			description: "home - a file in the home",
			opts: []Option{
				WithDecoder(&testDecoder{extensions: []string{"json"}}),
				nonWinStdCfgLayout("example", nil, stdLocations{
					local:    local,
					root:     local,
					home:     home,
					homeTree: homeTree,
					etc:      etc,
					etcTree:  etcTree,
				}),
				AddFile(never, "example.json"),
			},
			want: st{},
			expect: st{
				Status: "home - wanted",
			},
			files: []string{"example.json"},
		}, {
			description: "home - one file in the dir",
			opts: []Option{
				WithDecoder(&testDecoder{extensions: []string{"json"}}),
				nonWinStdCfgLayout("example", nil, stdLocations{
					local:    local,
					root:     local,
					home:     none,
					homeTree: homeTree,
					etc:      etc,
					etcTree:  etcTree,
				}),
				AddFile(never, "example.json"),
			},
			want: st{},
			expect: st{
				Other: "home - wanted",
			},
			files: []string{"2.json"},
		}, {
			description: "etc - a file in the home",
			opts: []Option{
				WithDecoder(&testDecoder{extensions: []string{"json"}}),
				nonWinStdCfgLayout("example", nil, stdLocations{
					local:    local,
					root:     local,
					home:     none,
					homeTree: none,
					etc:      etc,
					etcTree:  etcTree,
				}),
				AddFile(never, "example.json"),
			},
			want: st{},
			expect: st{
				Status: "etc - wanted",
			},
			files: []string{"example.json"},
		}, {
			description: "etc - one file in the dir",
			opts: []Option{
				WithDecoder(&testDecoder{extensions: []string{"json"}}),
				nonWinStdCfgLayout("example", nil, stdLocations{
					local:    local,
					root:     local,
					home:     none,
					homeTree: none,
					etc:      none,
					etcTree:  etcTree,
				}),
				AddFile(never, "example.json"),
			},
			want: st{},
			expect: st{
				Other: "etc - wanted",
			},
			files: []string{"2.json"},
		}, {
			description:   "invalid appname",
			compileOption: true,
			opts: []Option{
				AutoCompile(),
				WithDecoder(&testDecoder{extensions: []string{"json"}}),
				nonWinStdCfgLayout("", nil, stdLocations{}),
				AddFile(never, "example.json"),
			},
			want:        st{},
			expect:      st{},
			expectedErr: ErrInvalidInput,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			t.Setenv("thing", "ocean")

			remappings.Mapping = make(map[string]string)

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
				err = cfg.Unmarshal(tc.key, &tc.want)
				require.NoError(err)

				if tc.compare != nil {
					assert.True(tc.compare(assert, tc.expect, tc.want))
				} else {
					assert.Equal(tc.expect, tc.want)
				}

				// check the file order
				got, err := cfg.ShowOrder()
				require.NoError(err)
				assert.Equal(tc.files, got)

				assert.NotEmpty(tell)

				if tc.expectedRemaps == nil {
					assert.Empty(remappings.Mapping)
				} else {
					assert.Equal(tc.expectedRemaps, remappings.Mapping)
				}
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
