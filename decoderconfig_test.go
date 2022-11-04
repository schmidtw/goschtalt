// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
)

func TestDecoderConfig(t *testing.T) {
	tests := []struct {
		description string
		opt         DecoderConfigOption
		want        mapstructure.DecoderConfig
		check       func(*mapstructure.DecoderConfig) bool
	}{
		{
			description: "DecodeHook",
			opt:         DecodeHook(func() {}),
			check: func(m *mapstructure.DecoderConfig) bool {
				return m.DecodeHook != nil
			},
		}, {
			description: "ErrorUnused()",
			opt:         ErrorUnused(),
			want: mapstructure.DecoderConfig{
				ErrorUnused: true,
			},
		}, {
			description: "ErrorUnused(true)",
			opt:         ErrorUnused(true),
			want: mapstructure.DecoderConfig{
				ErrorUnused: true,
			},
		}, {
			description: "ErrorUnused(false)",
			opt:         ErrorUnused(false),
			want:        mapstructure.DecoderConfig{},
		}, {
			description: "ErrorUnset()",
			opt:         ErrorUnset(),
			want: mapstructure.DecoderConfig{
				ErrorUnset: true,
			},
		}, {
			description: "ErrorUnset(true)",
			opt:         ErrorUnset(true),
			want: mapstructure.DecoderConfig{
				ErrorUnset: true,
			},
		}, {
			description: "ErrorUnset(false)",
			opt:         ErrorUnset(false),
			want:        mapstructure.DecoderConfig{},
		}, {
			description: "WeaklyTypedInput()",
			opt:         WeaklyTypedInput(),
			want: mapstructure.DecoderConfig{
				WeaklyTypedInput: true,
			},
		}, {
			description: "WeaklyTypedInput(true)",
			opt:         WeaklyTypedInput(true),
			want: mapstructure.DecoderConfig{
				WeaklyTypedInput: true,
			},
		}, {
			description: "WeaklyTypedInput(false)",
			opt:         WeaklyTypedInput(false),
			want:        mapstructure.DecoderConfig{},
		}, {
			description: "TagName('')",
			opt:         TagName(""),
			want:        mapstructure.DecoderConfig{},
		}, {
			description: "TagName('foo')",
			opt:         TagName("foo"),
			want: mapstructure.DecoderConfig{
				TagName: "foo",
			},
		}, {
			description: "IgnoreUntaggedFields()",
			opt:         IgnoreUntaggedFields(),
			want: mapstructure.DecoderConfig{
				IgnoreUntaggedFields: true,
			},
		}, {
			description: "IgnoreUntaggedFields(true)",
			opt:         IgnoreUntaggedFields(true),
			want: mapstructure.DecoderConfig{
				IgnoreUntaggedFields: true,
			},
		}, {
			description: "IgnoreUntaggedFields(false)",
			opt:         IgnoreUntaggedFields(false),
			want:        mapstructure.DecoderConfig{},
		}, {
			description: "MatchName(fn)",
			opt:         MatchName(func(key, field string) bool { return true }),
			check: func(m *mapstructure.DecoderConfig) bool {
				return m.MatchName != nil
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			var opt mapstructure.DecoderConfig
			tc.opt.decoderApply(&opt)

			if tc.check == nil {
				assert.Equal(opt, tc.want)
			} else {
				tc.check(&opt)
			}
		})
	}
}
