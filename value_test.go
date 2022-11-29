// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"reflect"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
)

func TestValueOptions(t *testing.T) {
	tests := []struct {
		description string
		opt         ValueOption
		decoder     mapstructure.DecoderConfig
		asDefault   bool
		want        valueOptions
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
		}, {
			description: "Verify FailOnNonSerializable()",
			opt:         FailOnNonSerializable(),
			want: valueOptions{
				failOnNonSerializable: true,
			},
			str: "FailOnNonSerializable()",
		}, {
			description: "Verify FailOnNonSerializable(true)",
			opt:         FailOnNonSerializable(true),
			want: valueOptions{
				failOnNonSerializable: true,
			},
			str: "FailOnNonSerializable()",
		}, {
			description: "Verify FailOnNonSerializable(false)",
			opt:         FailOnNonSerializable(false),
			str:         "FailOnNonSerializable(false)",
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			assert.Equal(tc.asDefault, tc.opt.isDefault())

			var opts valueOptions
			tc.opt.valueApply(&opts)
			assert.True(reflect.DeepEqual(tc.want, opts))

			var dec mapstructure.DecoderConfig
			tc.opt.decoderApply(&dec)
			assert.True(reflect.DeepEqual(tc.decoder, dec))

			assert.Equal(tc.str, tc.opt.String())
		})
	}
}
