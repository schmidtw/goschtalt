// SPDX-FileCopyrightText: 2022-2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/goschtalt/goschtalt/pkg/meta"
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
	type withAltTags struct {
		Foo string `goschtalt:"flags"`
		Bob string
	}
	type withAll struct {
		Foo      string
		Duration time.Duration
		Time     time.Time
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
			description: "Convert from string to a time.Duration",
			input:       `{"Foo":"bar", "Delta": "1s"}`,
			opts:        []UnmarshalOption{adaptStringToDuration()},
			want:        withDuration{},
			expected: withDuration{
				Foo:   "bar",
				Delta: time.Second,
			},
		}, {
			description: "Fail to convert from string to a time.Duration",
			input:       `{"Foo":"bar", "Delta": "invalid"}`,
			opts: []UnmarshalOption{
				adaptStringToDuration(), // Have the same handler registered multiple
				adaptStringToDuration(), // times so there are multiple errors for
				adaptStringToDuration(), // the same field.
			},
			want:        withDuration{},
			expectedErr: unknownErr,
		}, {
			description: "Convert from string to a all supoorted types",
			input: `{
						"Foo":"bar",
						"Duration": "1s",
						"Time": "2022-12-30"
					}`,
			opts: []UnmarshalOption{
				adaptStringToDuration(),
				adaptStringToTime("2006-01-02"),
			},
			want: withAll{},
			expected: withAll{
				Foo:      "bar",
				Duration: time.Second,
				Time:     time.Date(2022, time.December, 30, 0, 0, 0, 0, time.UTC),
			},
		}, {
			description: "Verify the Strictness(EXACT) behavior succeeds with exact match.",
			input:       `{"Foo":"bar", "Delta": "1s"}`,
			opts:        []UnmarshalOption{Strictness(EXACT)},
			want:        simple{},
			expected: simple{
				Foo:   "bar",
				Delta: "1s",
			},
		}, {
			description: "Verify the Strictness(EXACT) behavior catches extra config.",
			input:       `{"Foo":"bar", "Delta": "1s", "extra": "value"}`,
			opts:        []UnmarshalOption{Strictness(EXACT)},
			want:        simple{},
			expectedErr: unknownErr,
		}, {
			description: "Verify the Strictness(EXACT) behavior catches missing config.",
			input:       `{"Foo":"bar"}`,
			opts:        []UnmarshalOption{Strictness(EXACT)},
			want:        simple{},
			expectedErr: unknownErr,
		}, {
			description: "Verify the Strictness(SUBSET) behavior succeeds with exact match.",
			input:       `{"Foo":"bar", "Delta": "1s"}`,
			opts:        []UnmarshalOption{Strictness(SUBSET)},
			want:        simple{},
			expected: simple{
				Foo:   "bar",
				Delta: "1s",
			},
		}, {
			description: "Verify the Strictness(SUBSET) behavior succeeds with subset.",
			input:       `{"Foo":"bar"}`,
			opts:        []UnmarshalOption{Strictness(SUBSET)},
			want:        simple{},
			expected: simple{
				Foo: "bar",
			},
		}, {
			description: "Verify the Strictness(SUBSET) behavior catches extra config.",
			input:       `{"Foo":"bar", "extra": "value"}`,
			opts:        []UnmarshalOption{Strictness(SUBSET)},
			want:        simple{},
			expectedErr: unknownErr,
		}, {
			description: "Verify the Strictness(COMPLETE) behavior succeeds with exact match.",
			input:       `{"Foo":"bar", "Delta": "1s"}`,
			opts:        []UnmarshalOption{Strictness(COMPLETE)},
			want:        simple{},
			expected: simple{
				Foo:   "bar",
				Delta: "1s",
			},
		}, {
			description: "Verify the Strictness(COMPLETE) behavior errors with subset.",
			input:       `{"Foo":"bar"}`,
			opts:        []UnmarshalOption{Strictness(COMPLETE)},
			want:        simple{},
			expectedErr: unknownErr,
		}, {
			description: "Verify the Strictness(COMPLETE) behavior ignores extra config.",
			input:       `{"Foo":"bar", "Delta": "1s", "extra": "value"}`,
			opts:        []UnmarshalOption{Strictness(COMPLETE)},
			want:        simple{},
			expected: simple{
				Foo:   "bar",
				Delta: "1s",
			},
		}, {
			description: "Verify the Strictness(NONE) behavior succeeds with exact match.",
			input:       `{"Foo":"bar", "Delta": "1s"}`,
			opts:        []UnmarshalOption{Strictness(NONE)},
			want:        simple{},
			expected: simple{
				Foo:   "bar",
				Delta: "1s",
			},
		}, {
			description: "Verify the Strictness(NONE) behavior succeeds with subset.",
			input:       `{"Foo":"bar"}`,
			opts:        []UnmarshalOption{Strictness(NONE)},
			want:        simple{},
			expected: simple{
				Foo: "bar",
			},
		}, {
			description: "Verify the Strictness(NONE) behavior ignores extra config.",
			input:       `{"Foo":"bar", "extra": "value"}`,
			opts:        []UnmarshalOption{Strictness(NONE)},
			want:        simple{},
			expected: simple{
				Foo: "bar",
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
			description: "Verify the required (default) behavior.",
			key:         "not_present",
			input:       `{"flags":"bar"}`,
			want:        simple{},
			expectedErr: meta.ErrNotFound,
		}, {
			description: "Verify the optional behavior.",
			key:         "not_present",
			input:       `{"flags":"bar"}`,
			opts:        []UnmarshalOption{Optional()},
			want:        simple{},
			expected:    simple{},
		}, {
			description: "Verify the optional behavior with a real failure.",
			key:         "flags.dog",
			input:       `{"flags":["bar"]}`,
			opts:        []UnmarshalOption{Optional()},
			want:        simple{},
			expectedErr: unknownErr,
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
				// Processed first, renames the string to flood
				Keymap(map[string]string{
					"Foo": "flood",
				}),
				// Processed second, renames flood to foo
				Keymap(map[string]string{
					"flood": "foo",
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
			description: "AdaptFromCfg",
			input:       `{"Foo":"2022-05-01"}`,
			key:         "Foo",
			want:        time.Time{},
			opts: []UnmarshalOption{
				AdaptFromCfg(func(f, t reflect.Value) (any, error) {
					if f.Kind() != reflect.String {
						return f.Interface(), nil
					}
					return time.Parse("2006-01-02", f.Interface().(string))
				}),
			},
			expected: time.Date(2022, time.May, 1, 0, 0, 0, 0, time.UTC),
		}, {
			description: "Verify the DefaultUnmarshalOptions() works.",
			input:       `{"Foo":"bar", "Delta": "bob"}`,
			defOpts:     []Option{DefaultUnmarshalOptions(Strictness(EXACT))},
			want:        simple{},
			expected: simple{
				Foo:   "bar",
				Delta: "bob",
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
