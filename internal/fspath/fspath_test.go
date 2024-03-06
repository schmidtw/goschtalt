// SPDX-FileCopyrightText: 2024 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package fspath

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type toRelTest struct {
	description string
	path        string
	abs         func(string) (string, error)
	want        string
	expectedErr error
}

func Test_toRel(t *testing.T) {
	// Use the tests defined by the platform-specific tests.
	for _, tc := range toRelTests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			got, err := toRel(tc.path, tc.abs)

			if tc.expectedErr != nil {
				assert.Empty(got)
				assert.ErrorIs(err, tc.expectedErr)
				return
			}

			require.NoError(err)
			assert.Equal(tc.want, got)
		})
	}
}

func TestToRel(t *testing.T) {
	assert := assert.New(t)

	got, err := ToRel("./a")

	assert.NotEmpty(got)
	assert.NoError(err)
}

func TestMustToRel(t *testing.T) {
	assert := assert.New(t)

	got := MustToRel("./a")

	assert.NotEmpty(got)
}
