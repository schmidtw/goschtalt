// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package json

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		description string
		input       map[string]any
		expected    string
		expectedErr bool
	}{
		{
			description: "Simple test.",
			input: map[string]any{
				"foo": 123,
				"bar": 18.9,
			},
			expected: `{
    "bar": 18.9,
    "foo": 123
}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			c := Codec{}
			got, err := c.Encode(&tc.input)
			if tc.expectedErr == false {
				assert.NoError(err)
				assert.True(reflect.DeepEqual(tc.expected, string(got)))
				return
			}
			// encoder/json doesn't throw catchable errors, so just ensure an
			// error is thrown.
			assert.Error(err)
		})
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		description string
		json        string
		expected    map[string]any
		expectedErr bool
	}{
		{
			description: "Simple test.",
			json:        `{"foo": 123, "bar": 18.9}`,
			expected: map[string]any{
				"foo": int64(123),
				"bar": float64(18.9),
			},
		}, {
			description: "Invalid json test.",
			json:        `{"foo": 123, "bar" 18.9}`,
			expectedErr: true,
		}, {
			description: "Array test.",
			json:        `{"foo": [-44, 12], "bar": [-18.9, 99.2e5], "car": [19, 1.9] }`,
			expected: map[string]any{
				"foo": []any{int64(-44), int64(12)},
				"bar": []any{float64(-18.9), float64(99.2e5)},
				"car": []any{int64(19), float64(1.9)},
			},
		}, {
			description: "Number size test.",
			json: `{"foo": 9223372036854775807,
                           "bar": 10000000000000000000,
						   "car": 10e99999999 }`,
			expected: map[string]any{
				"foo": int64(9223372036854775807),
				"bar": float64(10000000000000000000.0),
				"car": string("10e99999999"),
			},
		}, {
			description: "A bit of everything.",
			json: `{
		       "racing": "stripe",
		       "struct": {
		           "foo": [44, 12],
		           "mobile": {
		               "bar": [18.9, 99.2],
		               "car": [19, 1.9, 3.0]
		           },
		           "rabbit": [
		               {
		                   "dogs": ["cats", 12, 99.2]
		               }, {
		                   "mice": ["crows", { "valid": "maybe" }, 99.2]
		               }, {
		                   "1234": 77
		               }
		           ]
		       }
		   }`,
			expected: map[string]any{
				"racing": "stripe",
				"struct": map[string]any{
					"foo": []any{int64(44), int64(12)},
					"mobile": map[string]any{
						"bar": []any{float64(18.9), float64(99.2)},
						"car": []any{int64(19), float64(1.9), float64(3)},
					},
					"rabbit": []any{
						map[string]any{
							"dogs": []any{"cats", int64(12), float64(99.2)},
						},
						map[string]any{
							"mice": []any{"crows", map[string]any{"valid": "maybe"}, float64(99.2)},
						},
						map[string]any{
							"1234": int64(77),
						},
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			got := map[string]any{}
			c := Codec{}
			err := c.Decode([]byte(tc.json), &got)
			if tc.expectedErr == false {
				assert.NoError(err)
				assert.True(reflect.DeepEqual(tc.expected, got))
				return
			}
			// encoder/json doesn't throw catchable errors, so just ensure an
			// error is thrown.
			assert.Error(err)
		})
	}
}

func TestExtensions(t *testing.T) {
	c := Codec{}
	exts := c.Extensions()
	assert.True(t, reflect.DeepEqual([]string{"json"}, exts))
}
