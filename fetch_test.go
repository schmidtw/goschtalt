// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/schmidtw/goschtalt/pkg/meta"
	"github.com/stretchr/testify/assert"
)

func TestFetch(t *testing.T) {
	failureErr := fmt.Errorf("always fails.")

	tests := []struct {
		description     string
		json            string
		key             string
		addGoodMapper   bool
		addErrorMapper  bool
		expected        any
		expectedErr     error
		notCompiled     bool
		expectedErrText string
	}{
		{
			description: "Fetch string with a matching type.",
			json:        `{"foo":"bar"}`,
			key:         "foo",
			expected:    "bar",
		}, {
			description: "Fetch float64 with a matching type.",
			json:        `{"foo":[0.1, 0.2]}`,
			key:         "foo.1",
			expected:    float64(0.2),
		}, {
			description: "Fetch full tree.",
			json:        `{"foo":["car", "bat"]}`,
			key:         "",
			expected: map[string]any{
				"foo": []any{
					"car", "bat",
				},
			},
		}, {
			description: "Fetch something that isn't there.",
			json:        `{"foo":"bar"}`,
			key:         "goofy",
			expected:    "ignored, but used for type",
			expectedErr: meta.ErrNotFound,
		}, {
			description:   "Fetch float64 with a non-matching type and a mapper.",
			json:          `{"foo":[0.1, 0.2]}`,
			key:           "foo.1",
			expected:      "0.2",
			addGoodMapper: true,
		}, {
			description:    "Fetch with a mapper that always return error.",
			json:           `{"foo":[0.1, 0.2]}`,
			key:            "foo.1",
			expected:       "ignored, but used for type",
			addErrorMapper: true,
			expectedErr:    failureErr,
		}, {
			description: "Not compile yet.",
			notCompiled: true,
			json:        `{"foo":[0.1, 0.2]}`,
			key:         "foo.1",
			expected:    "ignored, but used for type",
			expectedErr: ErrNotCompiled,
		}, {
			description: "Fetch float64 with a non-matching type.",
			json:        `{"foo":[0.1, 0.2]}`,
			key:         "foo.1",
			expected:    "ignored, but used for type",
			expectedErr: ErrTypeMismatch,
			// Normally checking error text is a bad idea, but since this will
			// be hard to debug for the user, I think it's worth it in this case.
			expectedErrText: "type mismatch: expected type 'string' does not match type found 'float64'",
		}, {
			description: "Fetch a missing value to check the error.",
			json:        `{"foo":[{"car":"map"},{"dog": "cat"}]}`,
			key:         "foo.1.rat",
			expected:    "ignored, but used for type",
			expectedErr: meta.ErrNotFound,
			// Normally checking error text is a bad idea, but since this will
			// be hard to debug for the user, I think it's worth it in this case.
			expectedErrText: "with 'foo.1.rat' not found",
		}, {
			description: "Fetch float64 with a non-matching type.",
			json:        `{"foo":[0.1, 0.2]}`,
			key:         "foo.2.dog",
			expected:    "ignored, but used for type",
			expectedErr: meta.ErrArrayOutOfBounds,
			// Normally checking error text is a bad idea, but since this will
			// be hard to debug for the user, I think it's worth it in this case.
			expectedErrText: "with array len of 2 and 'foo.2' array index is out of bounds",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			c := Config{
				tree:            decode("file", tc.json),
				hasBeenCompiled: !tc.notCompiled,
				typeMappers:     make(map[string]typeMapper),
				keySwizzler:     strings.ToLower,
				keyDelimiter:    ".",
			}

			if tc.addGoodMapper {
				c.typeMappers["string"] = func(i any) (any, error) {
					return "0.2", nil
				}
			}
			if tc.addErrorMapper {
				c.typeMappers["string"] = func(i any) (any, error) {
					return nil, failureErr
				}
			}

			got, err := Fetch(&c, tc.key, tc.expected)

			if tc.expectedErr == nil {
				assert.NoError(err)
				assert.Empty(cmp.Diff(tc.expected, got))
				return
			}

			assert.ErrorIs(err, tc.expectedErr)
			if len(tc.expectedErrText) > 0 {
				assert.Equal(tc.expectedErrText, fmt.Sprintf("%s", err))
			}
		})
	}
}

func TestFetchMethod(t *testing.T) {
	tests := []struct {
		description string
		json        string
		key         string
		expected    any
		expectedErr error
	}{
		{
			description: "Fetch string with a matching type.",
			json:        `{"foo":"bar"}`,
			key:         "foo",
			expected:    "bar",
		}, {
			description: "Fetch something that isn't there.",
			json:        `{"foo":"bar"}`,
			key:         "goofy",
			expected:    "ignored, but used for type",
			expectedErr: meta.ErrNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			c := Config{
				tree:            decode("file", tc.json),
				hasBeenCompiled: true,
				typeMappers:     make(map[string]typeMapper),
				keySwizzler:     strings.ToLower,
				keyDelimiter:    ".",
			}

			got, err := c.Fetch(tc.key)

			if tc.expectedErr == nil {
				assert.NoError(err)
				assert.Empty(cmp.Diff(tc.expected, got))
				return
			}

			assert.ErrorIs(err, tc.expectedErr)
		})
	}
}

func TestFetchWithOrigin(t *testing.T) {
	tests := []struct {
		description string
		json        string
		key         string
		expected    any
		origin      []meta.Origin
		expectedErr error
	}{
		{
			description: "Fetch string with a matching type.",
			json:        `{"foo":"bar"}`,
			key:         "foo",
			expected:    "bar",
			origin: []meta.Origin{
				{
					File: "file",
					Line: 2,
					Col:  123,
				},
			},
		}, {
			description: "Fetch something that isn't there.",
			json:        `{"foo":"bar"}`,
			key:         "goofy",
			expected:    "ignored, but used for type",
			expectedErr: meta.ErrNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			c := Config{
				tree:            decode("file", tc.json),
				hasBeenCompiled: true,
				typeMappers:     make(map[string]typeMapper),
				keySwizzler:     strings.ToLower,
				keyDelimiter:    ".",
			}

			got, origin, err := c.FetchWithOrigin(tc.key)

			if tc.expectedErr == nil {
				assert.NoError(err)
				assert.Empty(cmp.Diff(tc.expected, got))
				assert.Empty(cmp.Diff(tc.origin, origin))
				return
			}

			assert.ErrorIs(err, tc.expectedErr)
		})
	}
}
