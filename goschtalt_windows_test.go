// SPDX-FileCopyrightText: 2022-2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStdCfgLayoutWin(t *testing.T) {
	assert := assert.New(t)

	got := StdCfgLayout("name")

	// Only make sure the other things are called.  Other tests ensure the
	// functionality works.
	assert.NotNil(got)
}

func TestCompileWin(t *testing.T) {
	assert := assert.New(t)

	_, err := New(stdCfgLayout("name", nil))

	assert.Error(err)
}
