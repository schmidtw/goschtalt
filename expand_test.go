// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpand(t *testing.T) {
	fn := func(_ string) string {
		return ""
	}
	tests := []struct {
		description string
		in          Option
		want        expand
		count       int
		fnNil       bool
	}{
		{
			description: "Simple success",
			in:          Expand(fn),
			count:       1,
			want: expand{
				start:   "${",
				end:     "}",
				maximum: 10000,
			},
		}, {
			description: "Fully defined",
			in:          Expand(fn, WithOrigin("origin"), WithDelimiters("${{", "}}"), WithMaximum(10)),
			count:       1,
			want: expand{
				origin:  "origin",
				start:   "${{",
				end:     "}}",
				maximum: 10,
			},
		}, {
			description: "Fully defined",
			in:          ExpandEnv(WithOrigin("origin"), WithDelimiters("${{", "}}"), WithMaximum(-1)),
			count:       1,
			want: expand{
				origin:  "origin",
				start:   "${{",
				end:     "}}",
				maximum: 10000,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			var c Config
			err := c.With(tc.in)
			assert.NoError(err)

			assert.Equal(tc.count, len(c.opts.expansions))
			if tc.count == 1 {
				if tc.fnNil {
					assert.Nil(c.opts.expansions[0].mapper)
				}
				c.opts.expansions[0].mapper = nil
				c.opts.expansions[0].name = ""

				assert.True(reflect.DeepEqual(tc.want, c.opts.expansions[0]))
			}
		})
	}
}
