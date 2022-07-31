// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"reflect"
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
			options: []Option{WithMergeStrategy(Value, Existing)},
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
			options:     []Option{WithMergeStrategy(Value, Fail)},
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
			options: []Option{WithMergeStrategy(Array, Prepend)},
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
			options: []Option{WithMergeStrategy(Array, Existing)},
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
			options: []Option{WithMergeStrategy(Array, Latest)},
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
			options:     []Option{WithMergeStrategy(Array, Fail)},
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
			options: []Option{WithMergeStrategy(Map, Latest)},
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
			options: []Option{WithMergeStrategy(Map, Existing)},
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

			g, err := New(WithFileGroup(group))
			require.NotNil(g)
			require.NoError(err)
			err = g.Options(tc.options...)
			require.NoError(err)

			err = g.ReadInConfig()
			if tc.expectedErr == nil {
				assert.NoError(err)
				//Handy to keep around if you need to debug a failure.
				/*
					if !reflect.DeepEqual(tc.expected, g.annotated) {
						fmt.Printf("\n%s\n", tc.description)
						fmt.Println("\nWanted\n-------------------------------------------")
						pp.Print(tc.expected)
						fmt.Println("\nGot\n-------------------------------------------")
						pp.Print(g.annotated)
					}
				*/
				assert.True(reflect.DeepEqual(tc.expected, g.annotated))
				return
			}
			assert.ErrorIs(err, tc.expectedErr)
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
