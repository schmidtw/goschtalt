// SPDX-FileCopyrightText: 2022-2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalValueOption(t *testing.T) {
	tests := []struct {
		description string
		str         string
		opt         UnmarshalValueOption
		want        string
	}{
		{
			description: "TagName('')",
			str:         "TagName('')",
			opt:         TagName(""),
			want:        defaultTag,
		}, {
			description: "TagName('foo')",
			str:         "TagName('foo')",
			opt:         TagName("foo"),
			want:        "foo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			assert.Equal(tc.str, tc.opt.String())

			un := unmarshalOptions{}
			assert.NoError(tc.opt.unmarshalApply(&un))

			vo := valueOptions{}
			assert.NoError(tc.opt.valueApply(&vo))

			assert.Equal(tc.want, un.decoder.TagName)
			assert.Equal(tc.want, vo.tagName)
		})
	}
}
