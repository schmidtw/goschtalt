// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecoderOptions(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	c := Config{
		decoders: newDecoderRegistry(),
	}
	require.NoError(c.With(DecoderRegister(&testDecoder{extensions: []string{"json"}})))
	assert.Empty(cmp.Diff([]string{"json"}, c.decoders.extensions()))

	require.NoError(c.With(DecoderRegister(&testDecoder{extensions: []string{"yml"}})))
	assert.Empty(cmp.Diff([]string{"json", "yml"}, c.decoders.extensions()))

	require.NoError(c.With(DecoderRemove("json")))
	assert.Empty(cmp.Diff([]string{"yml"}, c.decoders.extensions()))
}

func TestEncoderOptions(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	c := Config{
		encoders: newEncoderRegistry(),
	}
	require.NoError(c.With(EncoderRegister(&testEncoder{extensions: []string{"json"}})))
	assert.Empty(cmp.Diff([]string{"json"}, c.encoders.extensions()))

	require.NoError(c.With(EncoderRegister(&testEncoder{extensions: []string{"yml"}})))
	assert.Empty(cmp.Diff([]string{"json", "yml"}, c.encoders.extensions()))

	require.NoError(c.With(EncoderRemove("json")))
	assert.Empty(cmp.Diff([]string{"yml"}, c.encoders.extensions()))
}

func TestFileGroup(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	var c Config

	g0 := Group{
		Paths: []string{"group 1"},
	}
	g1 := Group{
		Paths: []string{"group 2"},
	}

	require.NoError(c.With(FileGroup(g0)))
	require.Equal(1, len(c.groups))
	assert.Empty(cmp.Diff(g0, c.groups[0]))

	require.NoError(c.With(FileGroup(g1)))
	require.Equal(2, len(c.groups))
	assert.Empty(cmp.Diff(g1, c.groups[1]))
}

func TestKeyDelimiter(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	var c Config

	assert.Equal(c.keyDelimiter, "")
	require.NoError(c.With(KeyDelimiter(".")))
	assert.Equal(c.keyDelimiter, ".")
	require.NoError(c.With(KeyDelimiter("<crazy>")))
	assert.Equal(c.keyDelimiter, "<crazy>")
}
