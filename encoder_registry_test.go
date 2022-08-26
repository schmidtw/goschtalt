// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/schmidtw/goschtalt/pkg/encoder"
	"github.com/schmidtw/goschtalt/pkg/meta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncoderRegistry_Most(t *testing.T) {
	yml := &testEncoder{
		extensions: []string{"yml", "yaml"},
	}
	json := &testEncoder{
		extensions: []string{"json"},
	}
	duplicate := &testEncoder{
		extensions: []string{"duplicate", "Duplicate"},
	}

	tests := []struct {
		description string
		encoders    []encoder.Encoder
		remove      []string
		add         encoder.Encoder
		expected    []string
		expectedErr error
	}{
		{
			description: "Successfully add a single encoder.",
			encoders:    []encoder.Encoder{yml},
			expected:    []string{"yaml", "yml"},
		}, {
			description: "Successfully add a two encoders.",
			encoders:    []encoder.Encoder{json, yml},
			expected:    []string{"json", "yaml", "yml"},
		}, {
			description: "Fail to add a duplicate encoders.",
			encoders:    []encoder.Encoder{json, yml, yml},
			expectedErr: ErrDuplicateFound,
		}, {
			description: "Fail to add a duplicate encoders.",
			encoders:    []encoder.Encoder{json, yml, duplicate},
			expectedErr: ErrDuplicateFound,
		}, {
			description: "Successfully add, then remove, then add encoders.",
			encoders:    []encoder.Encoder{json, yml},
			remove:      []string{"json", "yml", "non-existent"},
			add:         json,
			expected:    []string{"json", "yaml"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			er := newEncoderRegistry()
			require.NotNil(er)

			var err error

			if len(tc.encoders) > 0 {
				for i := 0; i < len(tc.encoders); i++ {
					err = er.register(tc.encoders[i])
					if i < len(tc.encoders)-1 {
						require.NoError(err)
					}
				}
			}

			er.deregister(tc.remove...)

			if tc.add != nil {
				require.NoError(err)
				err = er.register(tc.add)
			}

			if tc.expectedErr == nil {
				assert.NoError(err)
				assert.Empty(cmp.Diff(tc.expected, er.extensions()))
				return
			}
			assert.ErrorIs(err, tc.expectedErr)
		})
	}
}

func TestEncoderRegistry_Encode(t *testing.T) {
	tests := []struct {
		description string
		extension   string
		in          any
		expected    string
		expectedErr error
	}{
		{
			description: "Successfully encode.",
			extension:   "json",
			in: map[string]any{
				"test": "123",
			},
			expected: `{"test":"123"}`,
		}, {
			description: "Successfully encode uppercase.",
			extension:   "JSON",
			in: map[string]any{
				"test": "123",
			},
			expected: `{"test":"123"}`,
		}, {
			description: "Fail to find an encoder.",
			extension:   "invalid",
			expectedErr: ErrCodecNotFound,
		}, {
			description: "Encoding error.",
			extension:   "JSON",
			in:          nil,
			expectedErr: ErrEncoding,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			json := &testEncoder{
				extensions: []string{"JSON"},
			}

			er := newEncoderRegistry()
			require.NotNil(er)
			require.NoError(er.register(json))

			got, err := er.encode(tc.extension, tc.in)
			if tc.expectedErr == nil {
				require.NoError(err)
				assert.Empty(cmp.Diff([]byte(tc.expected), got))
				return
			}
			assert.ErrorIs(err, tc.expectedErr)
		})
	}
}

func TestEncoderRegistry_EncodeExtended(t *testing.T) {
	tests := []struct {
		description string
		extension   string
		in          meta.Object
		expected    string
		expectedErr error
	}{
		{
			description: "Successfully encode.",
			extension:   "json",
			in:          meta.Object{},
			expected:    `{"Origins":null,"Array":null,"Map":null,"Value":null}`,
		}, {
			description: "Successfully encode uppercase.",
			extension:   "JSON",
			in:          meta.Object{},
			expected:    `{"Origins":null,"Array":null,"Map":null,"Value":null}`,
		}, {
			description: "Fail to find an encoder.",
			extension:   "invalid",
			expectedErr: ErrCodecNotFound,
		}, {
			description: "Encoding error.",
			extension:   "JSON",
			in: meta.Object{
				Value: "cause error",
			},
			expectedErr: ErrEncoding,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			json := &testEncoder{
				extensions: []string{"JSON"},
			}

			er := newEncoderRegistry()
			require.NotNil(er)
			require.NoError(er.register(json))

			got, err := er.encodeExtended(tc.extension, tc.in)
			if tc.expectedErr == nil {
				require.NoError(err)
				assert.Empty(cmp.Diff([]byte(tc.expected), got))
				return
			}
			assert.ErrorIs(err, tc.expectedErr)
		})
	}
}
