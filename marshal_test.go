// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/goschtalt/goschtalt/pkg/encoder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshal(t *testing.T) {
	var zeroOpt MarshalOption
	testErr := errors.New("test error")

	tests := []struct {
		description string
		input       string
		opts        []MarshalOption
		notCompiled bool
		noEncoders  bool
		expected    string
		expectedErr error
	}{
		{
			description: "Import and export a normal tree.",
			input:       `{"foo":"bar"}`,
			opts:        []MarshalOption{FormatAs("json")},
			expected:    `{"foo":"bar"}`,
		}, {
			description: "Import and export a normal tree, with an empty option.",
			input:       `{"foo":"bar"}`,
			opts:        []MarshalOption{FormatAs("json"), zeroOpt},
			expected:    `{"foo":"bar"}`,
		}, {
			description: "Import and export a tree with a secret.",
			input:       `{"foo((secret))":"bar"}`,
			opts:        []MarshalOption{FormatAs("json")},
			expected:    `{"foo":"bar"}`,
		}, {
			description: "Import and export a tree with a redacted secret.",
			input:       `{"foo((secret))":"bar"}`,
			opts:        []MarshalOption{FormatAs("json"), RedactSecrets(true)},
			expected:    `{"foo":"REDACTED"}`,
		}, {
			description: "Import and export a tree with orgins.",
			input:       `{"foo":"bar"}`,
			opts:        []MarshalOption{FormatAs("json"), IncludeOrigins(true)},
			expected:    `{"Origins":[{"File":"file","Line":1,"Col":123}],"Array":null,"Map":{"foo":{"Origins":[{"File":"file","Line":2,"Col":123}],"Array":null,"Map":null,"Value":"bar"}},"Value":null}`,
		}, {
			description: "Not compiled.",
			input:       `{"foo":"bar"}`,
			notCompiled: true,
			opts:        []MarshalOption{FormatAs("json")},
			expectedErr: ErrNotCompiled,
		}, {
			description: "No format exporter found.",
			input:       `{"foo":"bar"}`,
			opts:        []MarshalOption{FormatAs("unsupported")},
			expectedErr: ErrCodecNotFound,
		}, {
			description: "No format exporter found.",
			input:       `{"foo":"bar"}`,
			noEncoders:  true,
			opts:        []MarshalOption{FormatAs("json")},
			expectedErr: ErrCodecNotFound,
		}, {
			description: "Handle an error.",
			input:       `{"foo":"bar"}`,
			opts:        []MarshalOption{WithError(testErr)},
			expectedErr: testErr,
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
					encoders:     newRegistry[encoder.Encoder](),
					keySwizzler:  strings.ToLower,
					keyDelimiter: ".",
				},
			}

			if !tc.noEncoders {
				c.opts.encoders.register(&testEncoder{extensions: []string{"json"}})
			}

			got, err := c.Marshal(tc.opts...)

			if tc.expectedErr == nil {
				assert.NoError(err)
				assert.Empty(cmp.Diff(tc.expected, string(got)))
				return
			}

			assert.ErrorIs(err, tc.expectedErr)
		})
	}
}
