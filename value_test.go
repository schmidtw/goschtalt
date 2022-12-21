// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"errors"
	"reflect"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
)

func TestValueOptions(t *testing.T) {
	testErr := errors.New("test error")
	tests := []struct {
		description string
		opt         ValueOption
		decoder     mapstructure.DecoderConfig
		asDefault   bool
		want        valueOptions
		str         string
		expectedErr error
	}{
		{
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
		}, {
			description: "Verify AsDefault()",
			opt:         AsDefault(),
			asDefault:   true,
			want: valueOptions{
				isDefault: true,
			},
			str: "AsDefault()",
		}, {
			description: "Verify AsDefault(true)",
			opt:         AsDefault(true),
			asDefault:   true,
			want: valueOptions{
				isDefault: true,
			},
			str: "AsDefault()",
		}, {
			description: "Verify AsDefault(false)",
			opt:         WithError(testErr),
			asDefault:   false,
			str:         "WithError( 'test error' )",
			expectedErr: testErr,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			assert.Equal(tc.str, tc.opt.String())

			var opts valueOptions
			err := tc.opt.valueApply(&opts)

			if tc.expectedErr == nil {
				assert.True(reflect.DeepEqual(tc.want, opts))
				assert.Equal(tc.asDefault, opts.isDefault)
			} else {
				assert.ErrorIs(err, tc.expectedErr)
			}
		})
	}
}
