// SPDX-FileCopyrightText: 2022-2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/goschtalt/goschtalt/pkg/meta"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshal(t *testing.T) {
	var zeroOpt UnmarshalOption
	unknownErr := fmt.Errorf("unknown error")
	testErr := fmt.Errorf("test error")
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
			input:       `{"Foo":"bar", "Delta": "1s"}`,
			opts:        []UnmarshalOption{},
			want:        simple{},
			expected: simple{
				Foo:   "bar",
				Delta: "1s",
			},
		}, {
			description: "A simple tree showing the duration doesn't decode.",
			input:       `{"Foo":"bar", "Delta": "1s"}`,
			opts:        []UnmarshalOption{},
			want:        withDuration{},
			expectedErr: unknownErr,
		}, {
			description: "A simple tree showing the duration doesn't decode, with zero option.",
			input:       `{"Foo":"bar", "Delta": "1s"}`,
			opts:        []UnmarshalOption{zeroOpt},
			want:        withDuration{},
			expectedErr: unknownErr,
		}, {
			description: "A simple tree with the DecodeHook() behavior works with duration hook.",
			input:       `{"Foo":"bar", "Delta": "1s"}`,
			opts: []UnmarshalOption{
				DecodeHook(
					mapstructure.ComposeDecodeHookFunc(
						mapstructure.StringToTimeDurationHookFunc()))},
			want: withDuration{},
			expected: withDuration{
				Foo:   "bar",
				Delta: time.Second,
			},
		}, {
			description: "Verify the ErrorUnused() behavior succeeds.",
			input:       `{"Foo":"bar", "Delta": "1s"}`,
			opts:        []UnmarshalOption{ErrorUnused(true)},
			want:        simple{},
			expected: simple{
				Foo:   "bar",
				Delta: "1s",
			},
		}, {
			description: "Verify the ErrorUnused behavior fails.",
			input:       `{"Foo":"bar", "Delta": "1s", "extra": "arg"}`,
			opts:        []UnmarshalOption{ErrorUnused(true)},
			want:        simple{},
			expectedErr: unknownErr,
		}, {
			description: "Verify the ErrorUnset() behavior succeeds.",
			input:       `{"Foo":"bar", "Delta": "1s"}`,
			opts:        []UnmarshalOption{ErrorUnset(true)},
			want:        simple{},
			expected: simple{
				Foo:   "bar",
				Delta: "1s",
			},
		}, {
			description: "Verify the ErrorUnset() behavior fails.",
			input:       `{"Foo":"bar", "extra": "arg"}`,
			opts:        []UnmarshalOption{ErrorUnset(true)},
			want:        simple{},
			expectedErr: unknownErr,
		}, {
			description: "Verify the WeaklyTypedInput() behavior succeeds.",
			input:       `{"Foo":"bar", "Bool": "T"}`,
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
			description: "Verify the optional behavior.",
			key:         "not_present",
			input:       `{"flags":"bar"}`,
			want:        simple{},
			expected:    simple{},
		}, {
			description: "Verify the WithValidator(fn) behavior works.",
			input:       `{"Foo":"bar"}`,
			opts:        []UnmarshalOption{WithValidator(func(any) error { return nil })},
			want:        simple{},
			expected: simple{
				Foo: "bar",
			},
		}, {
			description: "Verify the WithValidator(nil) behavior works.",
			input:       `{"Foo":"bar"}`,
			opts: []UnmarshalOption{
				WithValidator(func(any) error { return unknownErr }),
				WithValidator(nil),
			},
			want: simple{},
			expected: simple{
				Foo: "bar",
			},
		}, {
			description: "Convert from camelCase to PascalCase",
			input:       `{"foo":"bar"}`,
			opts: []UnmarshalOption{
				Keymap(map[string]string{
					"Foo": "foo",
				}),
			},
			want: simple{},
			expected: simple{
				Foo: "bar",
			},
		}, {
			description: "Verify a parameter can be ignored",
			input:       `{"foo":"bar"}`,
			opts: []UnmarshalOption{
				Keymap(map[string]string{
					"Foo": "-",
				}),
			},
			want: simple{},
			expected: simple{
				Foo: "",
			},
		}, {
			description: "Perform transforms via two different Keymap() calls.",
			input:       `{"foo":"bar", "tree": "tree val"}`,
			opts: []UnmarshalOption{
				Keymap(map[string]string{
					"Foo": "food",
				}),
				Keymap(map[string]string{
					"Foo": "foo",
				}),
				Keymap(map[string]string{
					"Delta": "tree",
				}),
			},
			want: simple{},
			expected: simple{
				Foo:   "bar",
				Delta: "tree val",
			},
		}, {
			description: "Verify the KeymapFn() works",
			input:       `{"foo":"bar"}`,
			opts: []UnmarshalOption{
				KeymapFn(func(s string) string {
					return strings.ToLower(s)
				}),
			},
			want: simple{},
			expected: simple{
				Foo: "bar",
			},
		}, {
			description: "Verify the WithValidator(fn) failure mode works.",
			input:       `{"Foo":"bar"}`,
			opts: []UnmarshalOption{
				WithValidator(func(any) error { return unknownErr }),
			},
			want:        simple{},
			expectedErr: unknownErr,
		}, {
			description: "Verify handling an error option.",
			input:       `{"Foo":"bar"}`,
			opts: []UnmarshalOption{
				WithError(testErr),
			},
			want:        simple{},
			expectedErr: testErr,
		}, {
			description: "A struct that wasn't compiled.",
			input:       `{"Foo":"bar", "Delta": "1s"}`,
			notCompiled: true,
			opts:        []UnmarshalOption{},
			want:        simple{},
			expectedErr: ErrNotCompiled,
		}, {
			description: "A nil result value.",
			input:       `{"Foo":"bar", "Delta": "1s"}`,
			nilWanted:   true,
			opts:        []UnmarshalOption{},
			expectedErr: unknownErr,
		}, {
			description: "Make sure that indexing an array works",
			key:         "Foo.0",
			input:       `{"Foo":["one", "two"]}`,
			opts:        []UnmarshalOption{Required()},
			want:        "",
			expected:    "one",
		}, {
			description: "Make sure that indexing an array is a number or error",
			key:         "Foo.Bar",
			input:       `{"Foo":[{"Foo":"one"}, "two"]}`,
			expectedErr: unknownErr,
		}, {
			description: "Verify the AddDefaultUnmarshalOption() works.",
			input:       `{"Foo":"bar", "Bool": "T"}`,
			defOpts:     []Option{DefaultUnmarshalOptions(WeaklyTypedInput(true))},
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

			now := time.Time{}
			if !tc.notCompiled {
				now = time.Now()
			}

			c := Config{
				tree:       tree,
				compiledAt: now,
				opts: options{
					keyDelimiter: ".",
				},
			}

			for _, opt := range tc.defOpts {
				require.NoError(opt.apply(&c.opts))
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
			if !errors.Is(unknownErr, tc.expectedErr) {
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
								"Foo": {
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
