// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	//pp "github.com/k0kubun/pp/v3"
	"github.com/psanford/memfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadAll(t *testing.T) {
	tests := []struct {
		description string
		files       map[string]string
		options     []Option
		expected    annotatedMap
		expectedErr error
	}{
		{
			description: "Merge with no conflicts.",
			files: map[string]string{
				"1.json": `{"foo":"bar"}`,
				"2.json": `{"cats": {"sally":[12], "bats": [ "brown", "black" ] }}`,
			},
			expected: annotatedMap{
				files: []string{"1.json", "2.json"},
				m: map[string]any{
					"foo": annotatedValue{
						files: []string{"1.json"},
						value: "bar",
					},
					"cats": annotatedMap{
						files: []string{"2.json"},
						m: map[string]any{
							"sally": annotatedArray{
								files: []string{"2.json"},
								array: []any{
									annotatedValue{
										files: []string{"2.json"},
										value: int64(12),
									},
								},
							},
							"bats": annotatedArray{
								files: []string{"2.json"},
								array: []any{
									annotatedValue{
										files: []string{"2.json"},
										value: "brown",
									},
									annotatedValue{
										files: []string{"2.json"},
										value: "black",
									},
								},
							},
						},
					},
				},
			},
		}, {
			description: "Merge with a map conflict.",
			files: map[string]string{
				"1.json": `{"foo":{"bar": "one"}}`,
				"2.json": `{"foo":{"cat": "two"}}`,
			},
			expected: annotatedMap{
				files: []string{"1.json", "2.json"},
				m: map[string]any{
					"foo": annotatedMap{
						files: []string{"1.json", "2.json"},
						m: map[string]any{
							"bar": annotatedValue{
								files: []string{"1.json"},
								value: "one",
							},
							"cat": annotatedValue{
								files: []string{"2.json"},
								value: "two",
							},
						},
					},
				},
			},
		}, {
			description: "Merge with a deeper map conflict.",
			files: map[string]string{
				"1.json": `{"foo":{"bar": {"one": "red"}}}`,
				"2.json": `{"foo":{"bar": {"two": "blue"}}}`,
			},
			expected: annotatedMap{
				files: []string{"1.json", "2.json"},
				m: map[string]any{
					"foo": annotatedMap{
						files: []string{"1.json", "2.json"},
						m: map[string]any{
							"bar": annotatedMap{
								files: []string{"1.json", "2.json"},
								m: map[string]any{
									"one": annotatedValue{
										files: []string{"1.json"},
										value: "red",
									},
									"two": annotatedValue{
										files: []string{"2.json"},
										value: "blue",
									},
								},
							},
						},
					},
				},
			},
		}, {
			description: "Merge with a deeper map and leaf conflict - default rules.",
			files: map[string]string{
				"1.json": `{"foo":{"bar": "one"}}`,
				"2.json": `{"foo":{"bar": "two"}}`,
			},
			expected: annotatedMap{
				files: []string{"1.json", "2.json"},
				m: map[string]any{
					"foo": annotatedMap{
						files: []string{"1.json", "2.json"},
						m: map[string]any{
							"bar": annotatedValue{
								files: []string{"2.json"},
								value: "two",
							},
						},
					},
				},
			},
		}, {
			description: "Merge with a deeper map and leaf conflict - keep older on conflict.",
			files: map[string]string{
				"1.json": `{"foo":{"bar": "one"}}`,
				"2.json": `{"foo":{"bar": "two"}}`,
			},
			options: []Option{MergeStrategy(Value, Existing)},
			expected: annotatedMap{
				files: []string{"1.json", "2.json"},
				m: map[string]any{
					"foo": annotatedMap{
						files: []string{"1.json", "2.json"},
						m: map[string]any{
							"bar": annotatedValue{
								files: []string{"1.json"},
								value: "one",
							},
						},
					},
				},
			},
		}, {
			description: "Merge with a deeper map and leaf conflict - fail on conflict.",
			files: map[string]string{
				"1.json": `{"foo":{"bar": "one"}}`,
				"2.json": `{"foo":{"bar": "two"}}`,
			},
			options:     []Option{MergeStrategy(Value, Fail)},
			expectedErr: ErrConflict,
		}, {
			description: "Merge with a deeper map and array conflict - default rules.",
			files: map[string]string{
				"1.json": `{"foo":{"bar": ["one", "two"]}}`,
				"2.json": `{"foo":{"bar": ["three", "four"]}}`,
			},
			expected: annotatedMap{
				files: []string{"1.json", "2.json"},
				m: map[string]any{
					"foo": annotatedMap{
						files: []string{"1.json", "2.json"},
						m: map[string]any{
							"bar": annotatedArray{
								files: []string{"1.json", "2.json"},
								array: []any{
									annotatedValue{
										files: []string{"1.json"},
										value: "one",
									},
									annotatedValue{
										files: []string{"1.json"},
										value: "two",
									},
									annotatedValue{
										files: []string{"2.json"},
										value: "three",
									},
									annotatedValue{
										files: []string{"2.json"},
										value: "four",
									},
								},
							},
						},
					},
				},
			},
		}, {
			description: "Merge with a deeper map and leaf conflict - prepend new.",
			files: map[string]string{
				"1.json": `{"foo":{"bar": ["one", "two"]}}`,
				"2.json": `{"foo":{"bar": ["three", "four"]}}`,
			},
			options: []Option{MergeStrategy(Array, Prepend)},
			expected: annotatedMap{
				files: []string{"1.json", "2.json"},
				m: map[string]any{
					"foo": annotatedMap{
						files: []string{"1.json", "2.json"},
						m: map[string]any{
							"bar": annotatedArray{
								files: []string{"1.json", "2.json"},
								array: []any{
									annotatedValue{
										files: []string{"2.json"},
										value: "three",
									},
									annotatedValue{
										files: []string{"2.json"},
										value: "four",
									},
									annotatedValue{
										files: []string{"1.json"},
										value: "one",
									},
									annotatedValue{
										files: []string{"1.json"},
										value: "two",
									},
								},
							},
						},
					},
				},
			},
		}, {
			description: "Merge with a deeper map and leaf conflict - keep old.",
			files: map[string]string{
				"1.json": `{"foo":{"bar": ["one", "two"]}}`,
				"2.json": `{"foo":{"bar": ["three", "four"]}}`,
			},
			options: []Option{MergeStrategy(Array, Existing)},
			expected: annotatedMap{
				files: []string{"1.json", "2.json"},
				m: map[string]any{
					"foo": annotatedMap{
						files: []string{"1.json", "2.json"},
						m: map[string]any{
							"bar": annotatedArray{
								files: []string{"1.json"},
								array: []any{
									annotatedValue{
										files: []string{"1.json"},
										value: "one",
									},
									annotatedValue{
										files: []string{"1.json"},
										value: "two",
									},
								},
							},
						},
					},
				},
			},
		}, {
			description: "Merge with a deeper map and leaf conflict - keep new.",
			files: map[string]string{
				"1.json": `{"foo":{"bar": ["one", "two"]}}`,
				"2.json": `{"foo":{"bar": ["three", "four"]}}`,
			},
			options: []Option{MergeStrategy(Array, Latest)},
			expected: annotatedMap{
				files: []string{"1.json", "2.json"},
				m: map[string]any{
					"foo": annotatedMap{
						files: []string{"1.json", "2.json"},
						m: map[string]any{
							"bar": annotatedArray{
								files: []string{"2.json"},
								array: []any{
									annotatedValue{
										files: []string{"2.json"},
										value: "three",
									},
									annotatedValue{
										files: []string{"2.json"},
										value: "four",
									},
								},
							},
						},
					},
				},
			},
		}, {
			description: "Merge with a deeper map and leaf conflict - fail on conflict.",
			files: map[string]string{
				"1.json": `{"foo":{"bar": ["one", "two"]}}`,
				"2.json": `{"foo":{"bar": ["three", "four"]}}`,
			},
			options:     []Option{MergeStrategy(Array, Fail)},
			expectedErr: ErrConflict,
		}, {
			description: "Merge with a deeper map type conflict - default rules.",
			files: map[string]string{
				"1.json": `{"foo":{"bar": ["one", "two"]}}`,
				"2.json": `{"foo":{"bar": "oops"}}`,
			},
			expectedErr: ErrConflict,
		}, {
			description: "Merge with a deeper map and leaf conflict - keep newer.",
			files: map[string]string{
				"1.json": `{"foo":{"bar": ["one", "two"]}}`,
				"2.json": `{"foo":{"bar": "oops"}}`,
			},
			options: []Option{MergeStrategy(Map, Latest)},
			expected: annotatedMap{
				files: []string{"1.json", "2.json"},
				m: map[string]any{
					"foo": annotatedMap{
						files: []string{"1.json", "2.json"},
						m: map[string]any{
							"bar": annotatedValue{
								files: []string{"2.json"},
								value: "oops",
							},
						},
					},
				},
			},
		}, {
			description: "Merge with a deeper map and leaf conflict - keep older.",
			files: map[string]string{
				"1.json": `{"foo":{"bar": ["one", "two"]}}`,
				"2.json": `{"foo":{"bar": "oops"}}`,
			},
			options: []Option{MergeStrategy(Map, Existing)},
			expected: annotatedMap{
				files: []string{"1.json", "2.json"},
				m: map[string]any{
					"foo": annotatedMap{
						files: []string{"1.json"},
						m: map[string]any{
							"bar": annotatedArray{
								files: []string{"1.json"},
								array: []any{
									annotatedValue{
										files: []string{"1.json"},
										value: "one",
									},
									annotatedValue{
										files: []string{"1.json"},
										value: "two",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fs := memfs.New()
			require.NotNil(fs)
			for k, v := range tc.files {
				require.NoError(fs.WriteFile(k, []byte(v), 0755))
			}

			group := Group{
				Paths: []string{"."},
				FS:    fs,
			}

			c, err := New(FileGroup(group))
			require.NotNil(c)
			require.NoError(err)
			err = c.With(tc.options...)
			require.NoError(err)

			err = c.Compile()
			if tc.expectedErr == nil {
				assert.NoError(err)
				//Handy to keep around if you need to debug a failure.
				/*
					if !reflect.DeepEqual(tc.expected, c.annotated) {
						fmt.Printf("\n%s\n", tc.description)
						fmt.Println("\nWanted\n-------------------------------------------")
						pp.Print(tc.expected)
						fmt.Println("\nGot\n-------------------------------------------")
						pp.Print(c.annotated)
					}
				*/
				assert.True(reflect.DeepEqual(tc.expected, c.annotated))
				return
			}
			assert.ErrorIs(err, tc.expectedErr)
		})
	}
}

func TestFetch(t *testing.T) {
	tests := []struct {
		description     string
		file            string
		key             string
		options         []Option
		forceStringType bool
		expected        any
		expectedErr     error
		expectedErrText string
	}{
		{
			description:     "Fetch with a matching type.",
			file:            `{"foo":"bar"}`,
			key:             "foo",
			expected:        "bar",
			forceStringType: true,
		}, {
			description:     "Fetch a nested value through an array.",
			file:            `{"foo":[{"car":"map"},{"dog": "cat"}]}`,
			key:             "foo.1.dog",
			forceStringType: true,
			expected:        "cat",
		}, {
			description:     "Fetch a more deeply nested value through an array.",
			file:            `{"foo":[{"car":{"type":"ev"}},{"dog": "cat"}]}`,
			key:             "foo.0.car.type",
			forceStringType: true,
			expected:        "ev",
		}, {
			description:     "Fetch through nested arrays.",
			file:            `{"foo":[["dog", "cat"],["fish", "shark"]]}`,
			key:             "foo.0.1",
			forceStringType: true,
			expected:        "cat",
		}, {
			description: "Fetch an array.",
			file:        `{"foo":[["dog", "cat"],["fish", "shark"]]}`,
			key:         "foo.0",
			expected:    []any{"dog", "cat"},
		}, {
			description: "Fetch everything.",
			file:        `{"foo":[["dog", "cat"],["fish", "shark"]]}`,
			key:         "",
			expected: map[string]any{
				"foo": []any{
					[]any{"dog", "cat"},
					[]any{"fish", "shark"},
				},
			},
		}, {
			description: "Fetch an integer from a string using a mapper.",
			file:        `{"foo":"2s"}`,
			key:         "foo",
			expected:    int(123),
			options: []Option{
				func() Option {
					var typ int
					return CustomMapper(typ, func(i any) (any, error) {
						return 123, nil
					})
				}(),
			},
		}, {
			description: "Fetch with a non-numeric value as an array index to check the error.",
			file:        `{"foo":[123]}`,
			key:         "foo.pig",
			expectedErr: strconv.ErrSyntax,
		}, {
			description:     "Fetch with a non-matching type.",
			file:            `{"foo":123}`,
			key:             "foo",
			forceStringType: true,
			expectedErr:     ErrTypeMismatch,
			// Normally checking error text is a bad idea, but since this will
			// be hard to debug for the user, I think it's worth it in this case.
			expectedErrText: "type mismatch: expected type 'string' does not match type found 'int64'",
		}, {
			description:     "Fetch an array out of bounds to check the error.",
			file:            `{"foo":[123]}`,
			key:             "foo.1",
			forceStringType: true,
			expectedErr:     ErrArrayIndexOutOfBounds,
			// Normally checking error text is a bad idea, but since this will
			// be hard to debug for the user, I think it's worth it in this case.
			expectedErrText: "with 'foo.1' array index is out of bounds len(array) is 1",
		}, {
			description:     "Fetch a missing value to check the error.",
			file:            `{"foo":[{"car":"map"},{"dog": "cat"}]}`,
			key:             "foo.1.rat",
			forceStringType: true,
			expectedErr:     ErrNotFound,
			// Normally checking error text is a bad idea, but since this will
			// be hard to debug for the user, I think it's worth it in this case.
			expectedErrText: "with 'foo.1.rat' not found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fs := memfs.New()
			require.NotNil(fs)
			require.NoError(fs.WriteFile("file.json", []byte(tc.file), 0755))

			group := Group{
				Paths: []string{"."},
				FS:    fs,
			}

			c, err := New(FileGroup(group))
			require.NotNil(c)
			require.NoError(err)
			err = c.With(tc.options...)
			require.NoError(err)

			err = c.Compile()
			require.NoError(err)

			if tc.forceStringType {
				var got string
				got, err = Fetch(c, tc.key, got)
				if tc.expectedErr == nil {
					assert.NoError(err)
					assert.Equal(tc.expected, got)
					return
				}
				assert.ErrorIs(err, tc.expectedErr)
				if len(tc.expectedErrText) > 0 {
					assert.Equal(tc.expectedErrText, fmt.Sprintf("%s", err))
				}
				return
			}

			got, err := Fetch(c, tc.key, tc.expected)
			if tc.expectedErr == nil {
				assert.NoError(err)
				assert.Equal(tc.expected, got)
				return
			}
			assert.ErrorIs(err, tc.expectedErr)
			if len(tc.expectedErrText) > 0 {
				assert.Equal(tc.expectedErrText, fmt.Sprintf("%s", err))
			}
		})
	}
}

func TestDedupedAppend(t *testing.T) {
	tests := []struct {
		description string
		list        []string
		add         []string
		expected    []string
	}{
		{
			description: "Simple test.",
			list:        []string{"car", "bar"},
			add:         []string{"foo", "bar", "bar"},
			expected:    []string{"car", "bar", "foo"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			got := dedupedAppend(tc.list, tc.add...)
			assert.True(reflect.DeepEqual(tc.expected, got))
		})
	}
}
