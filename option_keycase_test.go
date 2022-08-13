// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyKaseUnchanged(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	var c Config

	require.NoError(c.With(KeyCaseUnchanged()))
	require.NotNil(c.keySwizzler)
	assert.Equal("AbCdEf.XyZ", c.keySwizzler("AbCdEf.XyZ"))
}

func TestKeyKaseLower(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	var c Config

	require.NoError(c.With(KeyCaseLower()))
	require.NotNil(c.keySwizzler)
	assert.Equal("abcdef.xyz", c.keySwizzler("AbCdEf.XyZ"))
}

func TestKeyKaseUpper(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	var c Config

	require.NoError(c.With(KeyCaseUpper()))
	require.NotNil(c.keySwizzler)
	assert.Equal("ABCDEF.XYZ", c.keySwizzler("AbCdEf.XyZ"))
}
