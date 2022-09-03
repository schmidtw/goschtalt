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

func TestAndCompile(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	var c Config

	assert.Equal(c.compileNow, false)
	require.NoError(c.With(AndCompile()))
	assert.Equal(c.compileNow, true)
}

func TestNoDefaults(t *testing.T) {
	tests := []struct {
		description string
		opts        []Option
		delimiter   string
		ignore      bool
		expectedErr error
	}{
		{
			description: "NoDefaults and required options",
			delimiter:   "|",
			ignore:      true,
			opts: []Option{
				NoDefaults(),
				FileSortOrderNatural(),
				KeyCaseLower(),
				KeyDelimiter("|"),
			},
		}, {
			description: "NoDefaults and missing FileSortOrder",
			opts: []Option{
				NoDefaults(),
				KeyCaseLower(),
				KeyDelimiter("|"),
			},
			expectedErr: ErrConfigMissing,
		}, {
			description: "NoDefaults and missing KeyCase",
			opts: []Option{
				NoDefaults(),
				FileSortOrderNatural(),
				KeyDelimiter("|"),
			},
			expectedErr: ErrConfigMissing,
		}, {
			description: "NoDefaults and missing KeyDelimiter",
			opts: []Option{
				NoDefaults(),
				FileSortOrderNatural(),
				KeyCaseLower(),
			},
			expectedErr: ErrConfigMissing,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			c, err := New(tc.opts...)
			if tc.expectedErr == nil {
				assert.NoError(err)
				require.NotNil(c)
				assert.Equal(tc.delimiter, c.keyDelimiter)
				assert.Equal(tc.ignore, c.ignoreDefaults)
				return
			}

			assert.ErrorIs(err, tc.expectedErr)
		})
	}
}
