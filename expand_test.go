// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpand(t *testing.T) {
	testErr := errors.New("test error")
	expander := mockExpander{
		f: func(string) (string, bool) { return "", false },
	}

	tests := []struct {
		description string
		str         string
		in          Option
		want        []expand
		expectErr   error
	}{
		{
			description: "Simple success",
			in:          Expand(&expander),
			str:         "Expand( *goschtalt.mockExpander, ... ) --> start: '${', end: '}', origin: '', maximum: 0",
			want: []expand{{
				start:    "${",
				end:      "}",
				expander: &expander,
				maximum:  10000,
			}},
		}, {
			description: "Empty expand",
			in:          Expand(nil),
			str:         "Expand( nil, ... ) --> start: '${', end: '}', origin: '', maximum: 0",
		}, {
			description: "Fully defined",
			in:          Expand(&expander, WithOrigin("origin"), WithDelimiters("${{", "}}"), WithMaximum(10)),
			str:         "Expand( *goschtalt.mockExpander, ... ) --> start: '${{', end: '}}', origin: 'origin', maximum: 10",
			want: []expand{{
				origin:   "origin",
				start:    "${{",
				end:      "}}",
				expander: &expander,
				maximum:  10,
			}},
		}, {
			description: "Env, fully defined",
			in:          ExpandEnv(WithOrigin("origin"), WithDelimiters("${{", "}}"), WithMaximum(-1)),
			str:         "ExpandEnv( ... ) --> start: '${{', end: '}}', origin: 'origin', maximum: -1",
			want: []expand{{
				origin:   "origin",
				start:    "${{",
				end:      "}}",
				expander: envExpander{},
				maximum:  10000,
			}},
		}, {
			description: "Handle an error",
			in:          ExpandEnv(WithError(testErr)),
			str:         "WithError( 'ExpandEnv() err: test error' )",
			expectErr:   testErr,
		}, {
			description: "Handle an error in Expand()",
			in:          Expand(nil, WithError(testErr)),
			str:         "WithError( 'Expand() err: test error' )",
			expectErr:   testErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			assert.Equal(tc.str, tc.in.String())

			var c Config
			err := c.With(tc.in)

			if tc.expectErr == nil {
				assert.NoError(err)

				// Don't compare the text
				for i := range c.opts.expansions {
					c.opts.expansions[i].text = ""
				}
				assert.Equal(tc.want, c.opts.expansions)
			}

			assert.ErrorIs(err, tc.expectErr)
		})
	}
}

func TestExpandFunc(t *testing.T) {
	tests := []struct {
		in    string
		want  string
		found bool
	}{
		{
			in:    "text",
			want:  "text",
			found: true,
		}, {
			in: "frogs",
		},
	}

	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			assert := assert.New(t)
			f := ExpanderFunc(func(s string) (string, bool) {
				if s == "text" {
					return "text", true
				}
				return "", false
			})

			got, found := f.Expand(tc.in)
			assert.Equal(tc.want, got)
			assert.Equal(tc.found, found)
		})
	}
}

func Test_envExpander_Expand(t *testing.T) {
	tests := []struct {
		description string
		in          string
		want        string
		found       bool
	}{
		{
			description: "replace the string",
			in:          "replace",
			want:        "a value",
			found:       true,
		}, {
			description: "do not replace the string",
			in:          "ignored",
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			t.Setenv("replace", "a value")

			e := envExpander{}
			got, found := e.Expand(tc.in)
			assert.Equal(tc.want, got)
			assert.Equal(tc.found, found)
		})
	}
}
