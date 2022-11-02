// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpand(t *testing.T) {
	fn := func(_ string) string {
		return ""
	}
	tests := []struct {
		description string
		in          []Option
		count       int
	}{
		{
			description: "Simple success",
			in:          []Option{Expand(fn)},
			count:       1,
		}, {
			description: "Fully defined",
			in:          []Option{Expand(fn, WithDelimiters("${{", "}}"), WithMaximum(10))},
			count:       1,
		}, {
			description: "2 of them",
			in:          []Option{Expand(fn), Expand(fn)},
			count:       2,
		}, {
			description: "1 of them because no mapper in one",
			in:          []Option{Expand(fn), Expand(nil)},
			count:       1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			var c Config

			err := c.With(tc.in...)
			assert.NoError(err)

			assert.Equal(tc.count, len(c.opts.expansions))
		})
	}
}
