// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/mitchellh/mapstructure"
	"github.com/schmidtw/goschtalt/pkg/meta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshal(t *testing.T) {
	var zeroOpt UnmarshalOption
	unknownErr := fmt.Errorf("unknown error")
	type simple struct {
		Foo   string
		Delta string
	}
	type withDuration struct {
		Foo   string
		Delta time.Duration
	}
	type withBool struct {
		Foo  string
		Bool bool
	}
	type withAltTags struct {
		Foo string `goschtalt:"flags"`
		Bob string
	}
	/* Currently not needed due to what appears to be a broken upstream feature.
	type withSomeTags struct {
		Foo string `mapstructure:"foo"`
		Bob string //`mapstructure:"-"`
	}
	*/
	tests := []struct {
		description string
		key         string
		input       string
		want        any
		defOpts     []Option
		opts        []UnmarshalOption
		notCompiled bool
		nilWanted   bool
		expected    any
		expectedErr error
	}{
		{
			description: "A simple tree.",
			input:       `{"foo":"bar", "delta": "1s"}`,
			opts:        []UnmarshalOption{},
			want:        simple{},
			expected: simple{
				Foo:   "bar",
				Delta: "1s",
			},
		}, {
			description: "A simple tree showing the duration doesn't decode.",
			input:       `{"foo":"bar", "delta": "1s"}`,
			opts:        []UnmarshalOption{},
			want:        withDuration{},
			expectedErr: unknownErr,
		}, {
			description: "A simple tree showing the duration doesn't decode, with zero option.",
			input:       `{"foo":"bar", "delta": "1s"}`,
			opts:        []UnmarshalOption{zeroOpt},
			want:        withDuration{},
			expectedErr: unknownErr,
		}, {
			description: "A simple tree with the DecodeHook() behavior works with duration hook.",
			input:       `{"foo":"bar", "delta": "1s"}`,
			opts: []UnmarshalOption{DecodeHook(
				mapstructure.ComposeDecodeHookFunc(
					mapstructure.StringToTimeDurationHookFunc()))},
			want: withDuration{},
			expected: withDuration{
				Foo:   "bar",
				Delta: time.Second,
			},
		}, {
			description: "Verify the ErrorUnused() behavior succeeds.",
			input:       `{"foo":"bar", "delta": "1s"}`,
			opts:        []UnmarshalOption{ErrorUnused(true)},
			want:        simple{},
			expected: simple{
				Foo:   "bar",
				Delta: "1s",
			},
		}, {
			description: "Verify the ErrorUnused behavior fails.",
			input:       `{"foo":"bar", "delta": "1s", "extra": "arg"}`,
			opts:        []UnmarshalOption{ErrorUnused(true)},
			want:        simple{},
			expectedErr: unknownErr,
		}, {
			description: "Verify the ErrorUnset() behavior succeeds.",
			input:       `{"foo":"bar", "delta": "1s"}`,
			opts:        []UnmarshalOption{ErrorUnset(true)},
			want:        simple{},
			expected: simple{
				Foo:   "bar",
				Delta: "1s",
			},
		}, {
			description: "Verify the ErrorUnset() behavior fails.",
			input:       `{"foo":"bar", "extra": "arg"}`,
			opts:        []UnmarshalOption{ErrorUnset(true)},
			want:        simple{},
			expectedErr: unknownErr,
		}, {
			description: "Verify the WeaklyTypedInput() behavior succeeds.",
			input:       `{"foo":"bar", "bool": "T"}`,
			opts:        []UnmarshalOption{WeaklyTypedInput(true)},
			want:        withBool{},
			expected: withBool{
				Foo:  "bar",
				Bool: true,
			},
		}, {
			description: "Verify the TagName() behavior succeeds.",
			input:       `{"flags":"bar"}`,
			opts:        []UnmarshalOption{TagName("goschtalt")},
			want:        withAltTags{},
			expected: withAltTags{
				Foo: "bar",
			},
		}, {
			description: "Verify the Optional() behavior.",
			key:         "not_present",
			input:       `{"flags":"bar"}`,
			opts:        []UnmarshalOption{Optional(true)},
			want:        simple{},
			expected:    simple{},
		}, {
			description: "Verify the MatchName() behavior succeeds.",
			input:       `{"flags":"bar"}`,
			opts: []UnmarshalOption{MatchName(func(key, fieldName string) bool {
				return key == "flags" && strings.ToLower(fieldName) == "foo"
			})},
			want: simple{},
			expected: simple{
				Foo: "bar",
			},
			/*  This test is broken since the IgnoreUntaggedFields field appears
			    broken in mapstructure.
			}, {
				description: "Verify the IgnoreUntaggedFields() behavior fails.",
				input:       `{"foo":"bar", "bob": "dog"}`,
				opts:        []UnmarshalOption{IgnoreUntaggedFields(true)},
				want:        withSomeTags{},
				expected: withSomeTags{
					Foo: "bar",
				},
			*/
		}, {
			description: "A struct that wasn't compiled.",
			input:       `{"foo":"bar", "delta": "1s"}`,
			notCompiled: true,
			opts:        []UnmarshalOption{},
			want:        simple{},
			expectedErr: ErrNotCompiled,
		}, {
			description: "A nil result value.",
			input:       `{"foo":"bar", "delta": "1s"}`,
			nilWanted:   true,
			opts:        []UnmarshalOption{},
			expectedErr: unknownErr,
		}, {
			description: "Verify the DefaultUnmarshalOption() works.",
			input:       `{"foo":"bar", "bool": "T"}`,
			defOpts:     []Option{DefaultUnmarshalOption(WeaklyTypedInput(true))},
			want:        withBool{},
			expected: withBool{
				Foo:  "bar",
				Bool: true,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			tree, err := decode("file", tc.input).ResolveCommands()
			require.NoError(err)

			c := Config{
				tree:            tree,
				hasBeenCompiled: !tc.notCompiled,
				keySwizzler:     strings.ToLower,
				keyDelimiter:    ".",
			}

			for _, opt := range tc.defOpts {
				require.NoError(opt(&c))
			}

			if tc.nilWanted {
				err = c.Unmarshal(tc.key, nil, tc.opts...)
			} else {
				err = c.Unmarshal(tc.key, &tc.want, tc.opts...)
			}

			if tc.expectedErr == nil {
				assert.NoError(err)
				assert.Empty(cmp.Diff(tc.expected, tc.want))
				return
			}

			assert.Error(err)
			if tc.expectedErr != unknownErr {
				assert.ErrorIs(err, tc.expectedErr)
			}
		})
	}
}

func TestUnmarshalFn(t *testing.T) {
	type sub struct {
		Foo string
	}

	tests := []struct {
		description string
		key         string
		opts        []UnmarshalOption
		skipCompile bool
		want        sub
		expectedErr bool
	}{
		{
			description: "An empty struct",
			key:         "test",
			skipCompile: true,
			expectedErr: true,
		}, {
			description: "An valid struct",
			key:         "test",
			want: sub{
				Foo: "bar",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			g, err := New()
			require.NoError(err)
			require.NotNil(g)
			if !tc.skipCompile {
				err = g.Compile()
				require.NoError(err)
				g.tree = meta.Object{
					Map: map[string]meta.Object{
						"test": {
							Map: map[string]meta.Object{
								"foo": {
									Value: "bar",
								},
							},
						},
					},
				}
			}

			fn := UnmarshalFn[sub](tc.key, tc.opts...)
			require.NotNil(fn)

			got, err := fn(g)

			if tc.expectedErr == false {
				assert.NoError(err)
				assert.Equal(tc.want.Foo, got.Foo)
				return
			}

			assert.NotNil(err)
		})
	}
}
