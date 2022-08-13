// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/schmidtw/goschtalt/pkg/decoder"
	"github.com/schmidtw/goschtalt/pkg/meta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecoderRegistry_Most(t *testing.T) {
	yml := &testDecoder{
		extensions: []string{"yml", "yaml"},
	}
	json := &testDecoder{
		extensions: []string{"json"},
	}
	duplicate := &testDecoder{
		extensions: []string{"duplicate", "Duplicate"},
	}

	tests := []struct {
		description string
		decoders    []decoder.Decoder
		remove      []string
		add         decoder.Decoder
		expected    []string
		expectedErr error
	}{
		{
			description: "Successfully add a single decoder.",
			decoders:    []decoder.Decoder{yml},
			expected:    []string{"yaml", "yml"},
		}, {
			description: "Successfully add a two decoders.",
			decoders:    []decoder.Decoder{json, yml},
			expected:    []string{"json", "yaml", "yml"},
		}, {
			description: "Fail to add a duplicate decoders.",
			decoders:    []decoder.Decoder{json, yml, yml},
			expectedErr: ErrDuplicateFound,
		}, {
			description: "Fail to add a duplicate decoders.",
			decoders:    []decoder.Decoder{json, yml, duplicate},
			expectedErr: ErrDuplicateFound,
		}, {
			description: "Successfully add, then remove, then add decoders.",
			decoders:    []decoder.Decoder{json, yml},
			remove:      []string{"json", "yml", "non-existent"},
			add:         json,
			expected:    []string{"json", "yaml"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			dr := newDecoderRegistry()
			require.NotNil(dr)

			var err error
			if len(tc.decoders) > 0 {
				for i := 0; i < len(tc.decoders); i++ {
					err = dr.register(tc.decoders[i])
					if i < len(tc.decoders)-1 {
						require.NoError(err)
					}
				}
			}

			dr.deregister(tc.remove...)

			if tc.add != nil {
				require.NoError(err)
				err = dr.register(tc.add)
			}

			if tc.expectedErr == nil {
				assert.NoError(err)
				assert.Empty(cmp.Diff(tc.expected, dr.extensions()))
				return
			}
			assert.ErrorIs(err, tc.expectedErr)
		})
	}
}

func TestDecoderRegistry_Decode(t *testing.T) {
	tests := []struct {
		description string
		extension   string
		bytes       string
		expected    meta.Object
		expectedErr error
	}{
		{
			description: "Successfully decode.",
			extension:   "json",
			bytes:       `{ "test": "123" }`,
			expected: meta.Object{
				Origins: []meta.Origin{
					{
						File: "file",
						Line: 1,
						Col:  123,
					},
				},
				Type: meta.Map,
				Map: map[string]meta.Object{
					"test": {
						Origins: []meta.Origin{
							{
								File: "file",
								Line: 2,
								Col:  123,
							},
						},
						Type:  meta.Value,
						Value: "123",
					},
				},
			},
		}, {
			description: "Successfully decode uppercase.",
			extension:   "JSON",
			bytes:       `{ "test": "123" }`,
			expected: meta.Object{
				Origins: []meta.Origin{
					{
						File: "file",
						Line: 1,
						Col:  123,
					},
				},
				Type: meta.Map,
				Map: map[string]meta.Object{
					"test": {
						Origins: []meta.Origin{
							{
								File: "file",
								Line: 2,
								Col:  123,
							},
						},
						Type:  meta.Value,
						Value: "123",
					},
				},
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

			json := &testDecoder{
				extensions: []string{"JSON"},
			}

			dr := newDecoderRegistry()
			require.NotNil(dr)
			require.NoError(dr.register(json))

			var obj meta.Object
			err := dr.decode(tc.extension, "file", []byte(tc.bytes), &obj)
			if tc.expectedErr == nil {
				require.NoError(err)
				assert.Empty(cmp.Diff(tc.expected, obj))
				return
			}
			assert.ErrorIs(err, tc.expectedErr)
		})
	}
}
