// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package meta

import (
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMap_getCmd(t *testing.T) {
	tests := []struct {
		description string
		input       string
		expected    command
		expectedErr error
	}{
		{
			description: "Simple, no commands found.",
			input:       "foo",
			expected: command{
				full:  "foo",
				final: "foo",
			},
		}, {
			description: "Simple, single command found. #1",
			input:       "foo (( bar ))",
			expected: command{
				full:  "foo (( bar ))",
				cmd:   "bar",
				final: "foo",
			},
		}, {
			description: "Simple, single command found. #2",
			input:       "foo((bar))",
			expected: command{
				full:  "foo((bar))",
				cmd:   "bar",
				final: "foo",
			},
		}, {
			description: "Simple, single command found. #3",
			input:       "foo(( bar))   ",
			expected: command{
				full:  "foo(( bar))   ",
				cmd:   "bar",
				final: "foo",
			},
		}, {
			description: "Simple, single command found. #4",
			input:       "foo	(( bar	))	   ",
			expected: command{
				full:  "foo	(( bar	))	   ",
				cmd:   "bar",
				final: "foo",
			},
		}, {
			description: "Just the secret is found.",
			input:       "foo(( secret))",
			expected: command{
				full:   "foo(( secret))",
				cmd:    "",
				secret: true,
				final:  "foo",
			},
		}, {
			description: "Harder, command and secret found. #1",
			input:       "foo(( bar, secret))",
			expected: command{
				full:   "foo(( bar, secret))",
				cmd:    "bar",
				secret: true,
				final:  "foo",
			},
		}, {
			description: "Harder, command and secret found. #2",
			input:       "foo(( bar secret))",
			expected: command{
				full:   "foo(( bar secret))",
				cmd:    "bar",
				secret: true,
				final:  "foo",
			},
		}, {
			description: "Harder, command and secret found. #3",
			input:       "foo(( secret bar ))",
			expected: command{
				full:   "foo(( secret bar ))",
				cmd:    "bar",
				secret: true,
				final:  "foo",
			},
		}, {
			description: "Harder, command and secret found. #4",
			input:       "foo(( secret, bar ))",
			expected: command{
				full:   "foo(( secret, bar ))",
				cmd:    "bar",
				secret: true,
				final:  "foo",
			},
		}, {
			description: "Harder, command and secret found. #5",
			input:       "foo(( secret , bar ))",
			expected: command{
				full:   "foo(( secret , bar ))",
				cmd:    "bar",
				secret: true,
				final:  "foo",
			},
		}, {
			description: "Harder, command and secret found. #6",
			input:       "foo((secret,bar))",
			expected: command{
				full:   "foo((secret,bar))",
				cmd:    "bar",
				secret: true,
				final:  "foo",
			},
		}, {
			description: "Harder, command and secret found. #7",
			input:       "foo((secret,bar))",
			expected: command{
				full:   "foo((secret,bar))",
				cmd:    "bar",
				secret: true,
				final:  "foo",
			},
		}, {
			description: "Invalid because secret can only be present once.",
			input:       "foo((secret,secret))",
			expectedErr: ErrInvalidCommand,
		}, {
			description: "Invalid because secret must be one of two commands.",
			input:       "foo((bob,cat))",
			expectedErr: ErrInvalidCommand,
		}, {
			description: "Invalid because only two commands are allowed.",
			input:       "foo((secret,cat,dog))",
			expectedErr: ErrInvalidCommand,
		}, {
			description: "Fuzz finding: '0(())0'.",
			input:       "0(())0",
			expected: command{
				full:  "0(())0",
				final: "0(())0",
			},
		}, {
			description: "Nested parentheses.",
			input:       "0((()))",
			expectedErr: ErrInvalidCommand,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			got, err := getCmd(tc.input)

			if tc.expectedErr == nil {
				assert.NoError(err)
				assert.True(reflect.DeepEqual(tc.expected, got))
				return
			}
			assert.ErrorIs(err, tc.expectedErr)
		})
	}
}

func FuzzMap_getCmd(f *testing.F) {
	f.Add("foo((bar))")
	f.Fuzz(func(t *testing.T, in string) {
		r, err := getCmd(in)
		if err != nil {
			return
		}
		if len(r.full) < len(r.final) {
			t.Errorf("the full %q is less than the final %q\n", r.full, r.final)
		}
		if len(r.full) < len(r.cmd) {
			t.Errorf("the full %q is less than the cmd %q\n", r.full, r.cmd)
		}
		if r.final == r.full && r.secret {
			t.Errorf("the full %q and final indicate no command found, but secret is set.\n", r.full)
		}
		if r.final == r.full && len(r.cmd) != 0 {
			t.Errorf("the full %q and final indicate no command found, but there is a cmd %q set.\n", r.full, r.cmd)
		}
		if strings.ContainsAny(r.cmd, " 	,") {
			t.Errorf("the cmd %q contains invalid characters.\n", r.cmd)
		}
		if strings.Contains(r.cmd, "secret") {
			t.Errorf("the cmd %q contains secret.\n", r.cmd)
		}
		if r.final != r.full {
			tmp := strings.TrimSpace(in)
			if !strings.HasPrefix(tmp, r.final) {
				t.Errorf("the final %q doesn't start the original string.\n", r.final)
			}
			if !strings.HasSuffix(tmp, "))") {
				t.Errorf("the string doesn't end with '))' %q", in)
			}
			if !strings.Contains(tmp, "((") {
				t.Errorf("the string doesn't contain with '((' %q", in)
			}
			if !regexp.MustCompile(`^[a-zA-Z0-9_-]*$`).MatchString(r.cmd) {
				t.Errorf("the cmd contains invalid characters %s", r.cmd)
			}
		}

	})
}
