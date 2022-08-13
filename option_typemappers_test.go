// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomMapper(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	c := Config{
		typeMappers: make(map[string]typeMapper),
	}

	var s string
	mapper0 := func(_ any) (any, error) {
		return "mapper0", nil
	}

	var i int
	mapper1 := func(_ any) (any, error) {
		return "mapper1", nil
	}

	assert.Equal(0, len(c.typeMappers))

	require.NoError(c.With(CustomMapper(s, mapper0)))
	assert.Equal(1, len(c.typeMappers))
	got, err := c.typeMappers[reflect.TypeOf(s).String()](nil)
	assert.NoError(err)
	assert.Equal("mapper0", got)

	require.NoError(c.With(CustomMapper(i, mapper1)))
	assert.Equal(2, len(c.typeMappers))
	got, err = c.typeMappers[reflect.TypeOf(i).String()](nil)
	assert.NoError(err)
	assert.Equal("mapper1", got)

	require.NoError(c.With(CustomMapper(s, nil)))
	assert.Equal(1, len(c.typeMappers))
	got, err = c.typeMappers[reflect.TypeOf(i).String()](nil)
	assert.NoError(err)
	assert.Equal("mapper1", got)
}
