// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	//pp "github.com/k0kubun/pp/v3"
	"github.com/schmidtw/goschtalt/pkg/meta"
	"github.com/stretchr/testify/assert"
)

func TestFetch(t *testing.T) {
	type Example struct {
		Value []string
	}

	tests := []struct {
		description     string
		json            string
		key             string
		want            any
		opts            []UnmarshalOption
		expected        any
		expectedErr     error
		notCompiled     bool
		expectedErrText string
	}{
		{
			description: "Fetch string with a matching type.",
			json:        `{"foo":"bar"}`,
			key:         "foo",
			want:        "string",
			expected:    "bar",
		}, {
			description: "Fetch float64 with a matching type.",
			json:        `{"foo":[0.1, 0.2]}`,
			key:         "foo.1",
			want:        "float64",
			expected:    float64(0.2),
		}, {
			description: "Fetch int64 with a matching type.",
			json:        `{"foo":1}`,
			key:         "foo",
			want:        "int64",
			expected:    int64(1),
		}, {
			description: "Fetch int with a matching type.",
			json:        `{"foo":1}`,
			key:         "foo",
			want:        "int",
			expected:    int(1),
		}, {
			description: "Fetch full tree.",
			json:        `{"foo":["car", "bat"]}`,
			key:         "",
			want:        "map[string]any",
			expected: map[string]any{
				"foo": []any{
					"car", "bat",
				},
			},
		}, {
			description: "Fetch an array of strings.",
			json:        `{"foo":["car", "bat"]}`,
			key:         "foo",
			want:        "[]string",
			expected:    []string{"car", "bat"},
		}, {
			description: "Fetch an array of strings.",
			json:        `{"foo":["car", "bat"]}`,
			key:         "foo",
			want:        "[]string",
			expected:    []string{"car", "bat"},
		}, {
			description: "Fetch an slice of structs.",
			json:        `{"foo":[{"value":["a", "b"]},{"value":["c","d"]}]}`,
			key:         "foo",
			want:        "[]Example",
			expected: []Example{
				{
					Value: []string{"a", "b"},
				}, {
					Value: []string{"c", "d"},
				},
			},
		}, {
			description: "Fetch something that isn't there.",
			json:        `{"foo":"bar"}`,
			key:         "goofy",
			want:        "string",
			expectedErr: meta.ErrNotFound,
		}, {
			description: "Fetch float64 with a non-matching type and a mapper.",
			json:        `{"foo":["0.1", "0.2"]}`,
			key:         "foo.1",
			want:        "float64",
			expected:    float64(0.2),
			opts:        []UnmarshalOption{WeaklyTypedInput(true)},
		}, {
			description: "Not compile yet.",
			notCompiled: true,
			json:        `{"foo":[0.1, 0.2]}`,
			key:         "foo.1",
			want:        "string",
			expectedErr: ErrNotCompiled,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			c := Config{
				tree:            decode("file", tc.json),
				hasBeenCompiled: !tc.notCompiled,
				keySwizzler:     strings.ToLower,
				keyDelimiter:    ".",
			}

			var got any
			var err error

			switch tc.want {
			case "int":
				got, err = Fetch[int](&c, tc.key, tc.opts...)
			case "int64":
				got, err = Fetch[int64](&c, tc.key, tc.opts...)
			case "string":
				got, err = Fetch[string](&c, tc.key, tc.opts...)
			case "[]string":
				got, err = Fetch[[]string](&c, tc.key, tc.opts...)
			case "float64":
				got, err = Fetch[float64](&c, tc.key, tc.opts...)
			case "map[string]any":
				got, err = Fetch[map[string]any](&c, tc.key, tc.opts...)
			case "[]Example":
				got, err = Fetch[[]Example](&c, tc.key, tc.opts...)
			}

			if tc.expectedErr == nil {
				assert.NoError(err)
				//pp.Printf("Got:\n%s\n", got)
				//pp.Printf("Expected:\n%s\n", tc.expected)
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
