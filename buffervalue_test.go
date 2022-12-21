// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBufferValueOptions(t *testing.T) {
	testErr := errors.New("test error")
	tests := []struct {
		description string
		opt         BufferValueOption
		asDefault   bool
		str         string
		expectedErr error
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
			description: "Verify WithError(testErr)",
			opt:         WithError(testErr),
			str:         "WithError( 'test error' )",
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			var bo bufferOptions
			var vo valueOptions

			bErr := tc.opt.bufferApply(&bo)
			vErr := tc.opt.valueApply(&vo)

			if tc.expectedErr == nil {
				assert.Equal(tc.asDefault, bo.isDefault)
				assert.Equal(tc.asDefault, vo.isDefault)

				assert.Equal(tc.str, tc.opt.String())
				return
			}

			assert.ErrorIs(bErr, tc.expectedErr)
			assert.ErrorIs(vErr, tc.expectedErr)
		})
	}
}
