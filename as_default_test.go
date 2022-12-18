// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAsDefaultOptions(t *testing.T) {
	tests := []struct {
		description string
		opt         AsDefaultOption
		asDefault   bool
		str         string
	}{
		{
			description: "Verify AsDefault()",
			opt:         AsDefault(),
			asDefault:   true,
			str:         "AsDefault()",
		}, {
			description: "Verify AsDefault(true)",
			opt:         AsDefault(true),
			asDefault:   true,
			str:         "AsDefault()",
		}, {
			description: "Verify AsDefault(false)",
			opt:         AsDefault(false),
			str:         "AsDefault(false)",
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			assert.Equal(tc.asDefault, tc.opt.isDefault())

			assert.Equal(tc.str, tc.opt.String())
		})
	}
}
