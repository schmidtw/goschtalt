// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package meta

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func decode(s string) Object {
	var data any
	err := json.Unmarshal([]byte(s), &data)
	if err != nil {
		panic(err)
	}
	return ObjectFromRaw(data)
}

func TestOrigin_String(t *testing.T) {
	tests := []struct {
		description string
		origin      Origin
		expected    string
	}{
		{
			description: "Output an empty origin.",
			expected:    "unknown:???[???]",
		}, {
			description: "Output an filled out origin.",
			origin: Origin{
				File: "magic.file",
				Line: 42,
				Col:  88,
			},
			expected: "magic.file:42[88]",
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			got := tc.origin.String()

			assert.Equal(tc.expected, got)
		})
	}
}

func TestFetch(t *testing.T) {
	unknownErr := fmt.Errorf("unknown error")
	tests := []struct {
		description string
		in          string
		asks        []string
		expected    Object
		expectedErr error
	}{
		{
			description: "Simple map lookup.",
			in:          `{"foo":"something"}`,
			asks:        []string{"foo"},
			expected: Object{
				Origins: []Origin{},
				Type:    Value,
				Value:   "something",
			},
		}, {
			description: "Simple map then array lookup.",
			in:          `{"foo":["something", "else"]}`,
			asks:        []string{"foo", "1"},
			expected: Object{
				Origins: []Origin{},
				Type:    Value,
				Value:   "else",
			},
		}, {
			description: "A bit more nested.",
			in:          `{"foo":["something", {"else": "entirely"}]}`,
			asks:        []string{"foo", "1", "else"},
			expected: Object{
				Origins: []Origin{},
				Type:    Value,
				Value:   "entirely",
			},
		}, {
			description: "Not found.",
			in:          `{"foo":["something", {"else": "entirely"}]}`,
			asks:        []string{"foo", "1", "oops"},
			expectedErr: ErrNotFound,
		}, {
			description: "Invalid array value.",
			in:          `{"foo":["something", {"else": "entirely"}]}`,
			asks:        []string{"foo", "0-1.2", "oops"},
			expectedErr: unknownErr,
		}, {
			description: "Array index out of bounds.",
			in:          `{"foo":["something", {"else": "entirely"}]}`,
			asks:        []string{"foo", "10", "oops"},
			expectedErr: unknownErr,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			in := decode(tc.in)
			got, err := in.Fetch(tc.asks, ".")

			if tc.expectedErr == nil {
				assert.NoError(err)
				assert.Empty(cmp.Diff(tc.expected, got))
				return
			}
			assert.Error(err)
			if tc.expectedErr != unknownErr {
				assert.ErrorIs(err, tc.expectedErr)
			}
		})
	}
}

func TestToRaw(t *testing.T) {
	tests := []struct {
		description string
		in          Object
		expected    map[string]any
	}{
		{
			description: "Output an empty tree.",
			in: Object{
				Origins: []Origin{},
				Type:    Map,
			},
			expected: make(map[string]any),
		}, {
			description: "Output an small tree.",
			in: Object{
				Origins: []Origin{},
				Type:    Map,
				Map: map[string]Object{
					"foo": {},
				},
			},
			expected: map[string]any{
				"foo": nil,
			},
		}, {
			description: "Output an larger tree.",
			in: Object{
				Origins: []Origin{},
				Type:    Map,
				Map: map[string]Object{
					"foo": {
						Origins: []Origin{},
						Type:    Map,
						Map: map[string]Object{
							"bar": {
								Origins: []Origin{},
								Type:    Value,
								Value:   int(123),
							},
							"car": {
								Origins: []Origin{},
								Type:    Array,
								Array: []Object{
									{
										Origins: []Origin{},
										Type:    Map,
										Map: map[string]Object{
											"sam": {
												Origins: []Origin{},
												Type:    Value,
												Value:   "cart",
											},
										},
									},
									{
										Origins: []Origin{},
										Type:    Value,
										Value:   "golf",
									},
								},
							},
						},
					},
				},
			},
			expected: map[string]any{
				"foo": map[string]any{
					"bar": int(123),
					"car": []any{
						map[string]any{
							"sam": "cart",
						},
						"golf",
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			got := tc.in.ToRaw()

			assert.Empty(cmp.Diff(tc.expected, got))
		})
	}
}

func TestToRedacted(t *testing.T) {
	tests := []struct {
		description string
		in          Object
		expected    Object
	}{
		{
			description: "Output an small tree.",
			in: Object{
				Origins: []Origin{},
				Type:    Map,
				Map: map[string]Object{
					"foo": {
						Origins:  []Origin{},
						Type:     Value,
						IsSecret: true,
						Value:    "very secret.",
					},
				},
			},
			expected: Object{
				Origins: []Origin{},
				Type:    Map,
				Map: map[string]Object{
					"foo": {
						Origins:  []Origin{},
						Type:     Value,
						IsSecret: true,
						Value:    redactedText,
					},
				},
			},
		}, {
			description: "Output an larger tree.",
			in: Object{
				Origins: []Origin{},
				Type:    Map,
				Map: map[string]Object{
					"foo": {
						Origins: []Origin{},
						Type:    Map,
						Map: map[string]Object{
							"bar": {
								Origins: []Origin{},
								Type:    Value,
								Value:   int(123),
							},
							"car": {
								Origins: []Origin{},
								Type:    Array,
								Array: []Object{
									{
										Origins: []Origin{},
										Type:    Map,
										Map: map[string]Object{
											"sam": {
												Origins:  []Origin{},
												Type:     Value,
												IsSecret: true,
												Value:    "cart",
											},
										},
									},
									{
										Origins: []Origin{},
										Type:    Value,
										Value:   "golf",
									},
								},
							},
						},
					},
				},
			},
			expected: Object{
				Origins: []Origin{},
				Type:    Map,
				Map: map[string]Object{
					"foo": {
						Origins: []Origin{},
						Type:    Map,
						Map: map[string]Object{
							"bar": {
								Origins: []Origin{},
								Type:    Value,
								Value:   int(123),
							},
							"car": {
								Origins: []Origin{},
								Type:    Array,
								Array: []Object{
									{
										Origins: []Origin{},
										Type:    Map,
										Map: map[string]Object{
											"sam": {
												Origins:  []Origin{},
												Type:     Value,
												IsSecret: true,
												Value:    redactedText,
											},
										},
									},
									{
										Origins: []Origin{},
										Type:    Value,
										Value:   "golf",
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

			got := tc.in.ToRedacted()

			assert.Empty(cmp.Diff(tc.expected, got))
		})
	}
}

func TestAlterKeyCase(t *testing.T) {
	tests := []struct {
		description string
		in          string
		expected    string
	}{
		{
			description: "Output a small tree.",
			in:          `{"foo":"something"}`,
			expected:    `{"foo":"something"}`,
		}, {
			description: "Output a larger tree.",
			in:          `{"FOO":{ "BAR": "oNe", "Car": [ { "SAM": "CarT"}, "Golf" ] } }`,
			expected:    `{"foo":{ "bar": "oNe", "car": [ { "sam": "CarT"}, "Golf" ] } }`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			in := decode(tc.in)
			expected := decode(tc.expected)

			got := in.AlterKeyCase(strings.ToLower)

			assert.Empty(cmp.Diff(expected, got))
		})
	}
}

func TestResolveCommands(t *testing.T) {
	tests := []struct {
		description string
		in          string
		expected    Object
		expectedErr error
	}{
		{
			description: "Output a small tree.",
			in:          `{"foo ((secret))":"very secret."}`,
			expected: Object{
				Origins: []Origin{},
				Type:    Map,
				Map: map[string]Object{
					"foo": {
						Origins:  []Origin{},
						Type:     Value,
						IsSecret: true,
						Value:    "very secret.",
					},
				},
			},
		}, {
			description: "Output a larger tree.",
			in:          `{"foo ((secret,splice))": {"bar":"123", "car((secret))":[{"sam((secret))":"cart"},"golf"]}}`,
			expected: Object{
				Origins: []Origin{},
				Type:    Map,
				Map: map[string]Object{
					"foo": {
						Origins:  []Origin{},
						Type:     Map,
						IsSecret: true,
						Map: map[string]Object{
							"bar": {
								Origins: []Origin{},
								Type:    Value,
								Value:   "123",
							},
							"car": {
								Origins:  []Origin{},
								Type:     Array,
								IsSecret: true,
								Array: []Object{
									{
										Origins: []Origin{},
										Type:    Map,
										Map: map[string]Object{
											"sam": {
												Origins:  []Origin{},
												Type:     Value,
												IsSecret: true,
												Value:    "cart",
											},
										},
									},
									{
										Origins: []Origin{},
										Type:    Value,
										Value:   "golf",
									},
								},
							},
						},
					},
				},
			},
		}, {
			description: "Output an larger tree.",
			expectedErr: ErrInvalidCommand,
			in:          `{"foo ((secret,splice))": {"bar":"123", "car((secret))":[{"sam((secret, invalid, command))":"cart"},"golf"]}}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			got, err := decode(tc.in).ResolveCommands()

			if tc.expectedErr == nil {
				assert.NoError(err)
				assert.Empty(cmp.Diff(tc.expected, got))
				return
			}

			assert.ErrorIs(tc.expectedErr, err)
		})
	}
}

func TestMerge(t *testing.T) {
	tests := []struct {
		description string
		in          string
		next        string
		expected    Object
		expectedErr error
	}{
		{
			description: "Cover basically all the valid conditions.",
			in: `{	"foo":"cats",
				   	"bar":"food",
				   	"squid":"ink",
				   	"wolf":{
				   		"man":"jack"
					},
				   	"fox":{
				   		"is":"red"
					},
				   	"cow":{
						"noise":"bark"
					},
				   	"ox":{
						"cart":["2"]
					},
				   	"mad":{
					   "cat":"crazy",
					   "money":[ "usd","euro" ]
				   	},
					"dogs":[
						"fido",
						"pluto"
					]}`,
			next: `{	"foo((replace, secret))": "very secret.",
						"bar((keep))": "bats",
						"squid((keep))": {
							"giant": [ "20000" ]
						},
						"soap": "bar",
						"wolf": "of a different type",
						"fox((keep))": {
							"run": "fast"
						},
						"cow((replace))": {
							"eats": "hay"
						},
						"ox": {
							"cart((secret, append))": ["3"]
						},
				   		"mad":{
					   		"cat":"crazier",
					   		"money((prepend))":[ "cad","gbp" ]
						},
						"dogs((replace))":[
							{ "snoopy ((keep))": "beagle" }
						]}`,
			expected: Object{
				Origins: []Origin{},
				Type:    Map,
				Map: map[string]Object{
					"foo": {
						Origins:  []Origin{},
						Type:     Value,
						IsSecret: true,
						Value:    "very secret.",
					},
					"bar": {
						Origins: []Origin{},
						Type:    Value,
						Value:   "food",
					},
					"squid": {
						Origins: []Origin{},
						Type:    Value,
						Value:   "ink",
					},
					"soap": {
						Origins: []Origin{},
						Type:    Value,
						Value:   "bar",
					},
					"wolf": {
						Origins: []Origin{},
						Type:    Value,
						Value:   "of a different type",
					},
					"fox": {
						Origins: []Origin{},
						Type:    Map,
						Map: map[string]Object{
							"is": {
								Origins: []Origin{},
								Type:    Value,
								Value:   "red",
							},
						},
					},
					"cow": {
						Origins: []Origin{},
						Type:    Map,
						Map: map[string]Object{
							"eats": {
								Origins: []Origin{},
								Type:    Value,
								Value:   "hay",
							},
						},
					},
					"ox": {
						Origins: []Origin{},
						Type:    Map,
						Map: map[string]Object{
							"cart": {
								Origins:  []Origin{},
								Type:     Array,
								IsSecret: true,
								Array: []Object{
									{
										Origins: []Origin{},
										Type:    Value,
										Value:   "2",
									},
									{
										Origins: []Origin{},
										Type:    Value,
										Value:   "3",
									},
								},
							},
						},
					},
					"mad": {
						Origins: []Origin{},
						Type:    Map,
						Map: map[string]Object{
							"cat": {
								Origins: []Origin{},
								Type:    Value,
								Value:   "crazier",
							},
							"money": {
								Origins: []Origin{},
								Type:    Array,
								Array: []Object{
									{
										Origins: []Origin{},
										Type:    Value,
										Value:   "cad",
									},
									{
										Origins: []Origin{},
										Type:    Value,
										Value:   "gbp",
									},
									{
										Origins: []Origin{},
										Type:    Value,
										Value:   "usd",
									},
									{
										Origins: []Origin{},
										Type:    Value,
										Value:   "euro",
									},
								},
							},
						},
					},
					"dogs": {
						Origins: []Origin{},
						Type:    Array,
						Array: []Object{
							{
								Origins: []Origin{},
								Type:    Map,
								Map: map[string]Object{
									"snoopy": {
										Origins: []Origin{},
										Type:    Value,
										Value:   "beagle",
									},
								},
							},
						},
					},
				},
			},
		}, {
			description: "Clear the tree.",
			in:          `{"foo":"bar"}`,
			next:        `{"ignored((clear))":"ignored"}`,
			expected: Object{
				Origins: []Origin{},
				Type:    Map,
			},
		}, {
			description: "Append to a secret node and make sure it stays secret.",
			in:          `{"foo((secret))":["bar"]}`,
			next:        `{"foo((append))":["car"]}`,
			expected: Object{
				Origins: []Origin{},
				Type:    Map,
				Map: map[string]Object{
					"foo": {
						Origins:  []Origin{},
						Type:     Array,
						IsSecret: true,
						Array: []Object{
							{
								Origins: []Origin{},
								Type:    Value,
								Value:   "bar",
							},
							{
								Origins: []Origin{},
								Type:    Value,
								Value:   "car",
							},
						},
					},
				},
			},
		}, {
			description: "Prepend to a secret node and make sure it stays secret.",
			in:          `{"foo((secret))":["bar"]}`,
			next:        `{"foo((prepend))":["car"]}`,
			expected: Object{
				Origins: []Origin{},
				Type:    Map,
				Map: map[string]Object{
					"foo": {
						Origins:  []Origin{},
						Type:     Array,
						IsSecret: true,
						Array: []Object{
							{
								Origins: []Origin{},
								Type:    Value,
								Value:   "car",
							},
							{
								Origins: []Origin{},
								Type:    Value,
								Value:   "bar",
							},
						},
					},
				},
			},
		}, {
			description: "Error attempting to clear.",
			in:          `{"foo":"bar"}`,
			next:        `{"ignored((clear, secret, invalid))":"ignored"}`,
			expectedErr: ErrInvalidCommand,
		}, {
			description: "Error fail on conflict with value.",
			in:          `{"foo":"bar"}`,
			next:        `{"foo((fail))":"new"}`,
			expectedErr: ErrConflict,
		}, {
			description: "Error invalid command with value.",
			in:          `{"foo":"bar"}`,
			next:        `{"foo((invalid))":"new"}`,
			expectedErr: ErrInvalidCommand,
		}, {
			description: "Error fail on conflict with array.",
			in:          `{"foo":["bar"]}`,
			next:        `{"foo((fail))":["new"]}`,
			expectedErr: ErrConflict,
		}, {
			description: "Error invalid command with array.",
			in:          `{"foo":["bar"]}`,
			next:        `{"foo((invalid))":["new"]}`,
			expectedErr: ErrInvalidCommand,
		}, {
			description: "Error fail on conflict with map.",
			in:          `{"foo":{"goo":"bar"}}`,
			next:        `{"foo((fail))":{"goo":"new"}}`,
			expectedErr: ErrConflict,
		}, {
			description: "Error invalid command with map.",
			in:          `{"foo":{"goo":"bar"}}`,
			next:        `{"foo((invalid))":{"goo":"new"}}`,
			expectedErr: ErrInvalidCommand,
		}, {
			description: "Error invalid command with nested replace map.",
			in:          `{"foo":{"goo":"bar"}}`,
			next:        `{"foo((replace))":{"goo((invalid))":"new"}}`,
			expectedErr: ErrInvalidCommand,
		}, {
			description: "Error invalid command with nested array.",
			in:          `{"foo":{"goo":"bar"}}`,
			next:        `{"foo((replace))":[{"goo((invalid))":"new"}]}`,
			expectedErr: ErrInvalidCommand,
		}, {
			description: "Error invalid command with nested array.",
			in:          `{"foo":{"goo":["bar", {"cat":"dog"}]}}`,
			next:        `{"foo":{"goo":[{"sam((invalid))":"eagle"}]}}`,
			expectedErr: ErrInvalidCommand,
		}, {
			description: "Error invalid command with new key.",
			in:          `{"foo":{"goo":"bar"}}`,
			next:        `{"cat":{"goo((invalid))":"new"}}`,
			expectedErr: ErrInvalidCommand,
		}, {
			description: "Error conflict due to fail.",
			in:          `{"foo":{"goo":"crab"}}`,
			next:        `{"foo":{"goo((fail))":{"new": "car"}}}`,
			expectedErr: ErrConflict,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			in, err := decode(tc.in).resolveCommands(false)
			require.NoError(err)
			next := decode(tc.next)

			got, err := in.Merge(next)

			if tc.expectedErr == nil {
				assert.NoError(err)
				assert.Empty(cmp.Diff(tc.expected, got))
				return
			}

			assert.ErrorIs(tc.expectedErr, err)
		})
	}
}

func TestOrigin_OriginString(t *testing.T) {
	tests := []struct {
		description string
		obj         Object
		expected    string
	}{
		{
			description: "Output an empty object.",
			expected:    "",
		}, {
			description: "Output an empty origin.",
			obj: Object{
				Origins: []Origin{Origin{}},
			},
			expected: "unknown:???[???]",
		}, {
			description: "Output an filled out origin.",
			obj: Object{
				Origins: []Origin{
					Origin{
						File: "magic.file",
						Line: 42,
						Col:  88,
					},
				},
			},
			expected: "magic.file:42[88]",
		}, {
			description: "Output an filled out origin.",
			obj: Object{
				Origins: []Origin{
					Origin{
						File: "magic.file",
						Line: 42,
						Col:  88,
					},
					Origin{
						File: "foo.json",
						Line: 96,
						Col:  32,
					},
				},
			},
			expected: "magic.file:42[88], foo.json:96[32]",
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			got := tc.obj.OriginString()

			assert.Equal(tc.expected, got)
		})
	}
}
