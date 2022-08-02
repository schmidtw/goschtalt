// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package encoding

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCodec struct {
	extensions []string
}

func (t *testCodec) Decode(b []byte, v *map[string]any) error {
	return json.Unmarshal(b, v)
}

func (t *testCodec) Encode(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (t *testCodec) Extensions() []string {
	return t.extensions
}

func TestRegistry_Register(t *testing.T) {
	yml := &testCodec{
		extensions: []string{"yml", "yaml"},
	}
	json := &testCodec{
		extensions: []string{"json"},
	}
	duplicate := &testCodec{
		extensions: []string{"duplicate", "Duplicate"},
	}

	tests := []struct {
		description string
		codecs      []Option
		final       Option
		expectedErr error
	}{
		{
			description: "Successfully add a single codec.",
			final:       DecoderEncoder(yml),
		}, {
			description: "Successfully add a two codecs.",
			codecs:      []Option{DecoderEncoder(json)},
			final:       DecoderEncoder(yml),
		}, {
			description: "Fail to add a duplicate codecs.",
			codecs:      []Option{DecoderEncoder(json), DecoderEncoder(yml)},
			final:       DecoderEncoder(yml),
			expectedErr: ErrDuplicateFound,
		}, {
			description: "Fail to add a duplicate codecs.",
			codecs:      []Option{DecoderEncoder(json), DecoderEncoder(yml)},
			final:       DecoderEncoder(duplicate),
			expectedErr: ErrDuplicateFound,
		}, {
			description: "Fail to add a duplicate codecs.",
			codecs:      []Option{DecoderEncoder(json), DecoderEncoder(yml), DecoderEncoder(duplicate)},
			expectedErr: ErrDuplicateFound,
		}, {
			description: "Fail to add a duplicate codecs.",
			codecs:      []Option{DecoderEncoder(json), DecoderEncoder(json)},
			expectedErr: ErrDuplicateFound,
		}, {
			description: "Successfully add, then remove, then add codecs.",
			codecs:      []Option{DecoderEncoder(json), ExcludedExtensions("json", "yml"), DecoderEncoder(json)},
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			dr, err := NewRegistry(tc.codecs...)

			if tc.final != nil {
				require.NotNil(dr)
				require.NoError(err)
				err = dr.With(tc.final)
			}

			if tc.expectedErr == nil {
				assert.NoError(err)
				return
			}
			assert.ErrorIs(err, tc.expectedErr)
		})
	}
}

func TestRegistry_Extensions(t *testing.T) {
	yml := &testCodec{
		extensions: []string{"yml", "yaml"},
	}
	json := &testCodec{
		extensions: []string{"JSON"},
	}
	tests := []struct {
		description string
		codecs      []Option
		expected    []string
	}{
		{
			description: "Fail to add a duplicate codecs.",
			codecs:      []Option{DecoderEncoder(json), DecoderEncoder(yml)},
			expected:    []string{"json", "yaml", "yml"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			dr, err := NewRegistry(tc.codecs...)
			require.NotNil(dr)
			require.NoError(err)

			got := dr.Extensions()
			assert.True(reflect.DeepEqual(tc.expected, got))
		})
	}
}

func TestRegistry_Decode(t *testing.T) {
	tests := []struct {
		description string
		extension   string
		bytes       string
		expected    map[string]any
		expectedErr error
	}{
		{
			description: "Successfully decode.",
			extension:   "json",
			bytes:       `{ "test": "123" }`,
			expected: map[string]any{
				"test": "123",
			},
		}, {
			description: "Successfully decode uppercase.",
			extension:   "JSON",
			bytes:       `{ "test": { "hello": "world" }, "testing": "123" }`,
			expected: map[string]any{
				"test": map[string]any{
					"hello": "world",
				},
				"testing": "123",
			},
		}, {
			description: "Fail to find a decoder.",
			extension:   "invalid",
			expectedErr: ErrNotFound,
		}, {
			description: "Decoding error.",
			extension:   "JSON",
			bytes:       `{ invalid }`,
			expectedErr: ErrDecoding,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			json := &testCodec{
				extensions: []string{"JSON"},
			}

			dr, err := NewRegistry(DecoderEncoder(json))
			require.NotNil(dr)
			require.NoError(err)

			v := &map[string]any{}
			err = dr.Decode(tc.extension, []byte(tc.bytes), v)
			if tc.expectedErr == nil {
				require.NoError(err)
				assert.True(reflect.DeepEqual(&tc.expected, v))
				return
			}
			assert.ErrorIs(err, tc.expectedErr)
		})
	}
}

func TestRegistry_Encode(t *testing.T) {
	tests := []struct {
		description string
		extension   string
		input       map[string]any
		expected    string
		expectedErr error
	}{
		{
			description: "Successfully encode.",
			extension:   "json",
			input: map[string]any{
				"test": "123",
			},
			expected: `{"test":"123"}`,
		}, {
			description: "Successfully encode mixed case.",
			extension:   "JsOn",
			input: map[string]any{
				"test": "123",
			},
			expected: `{"test":"123"}`,
		}, {
			description: "Fail to find a encoder.",
			extension:   "invalid",
			expectedErr: ErrNotFound,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			json := &testCodec{
				extensions: []string{"JSON"},
			}

			dr, err := NewRegistry(DecoderEncoder(json))
			require.NotNil(dr)
			require.NoError(err)

			got, err := dr.Encode(tc.extension, &tc.input)
			if tc.expectedErr == nil {
				require.NoError(err)
				assert.Equal(tc.expected, string(got))
				return
			}
			assert.ErrorIs(err, tc.expectedErr)
		})
	}
}
