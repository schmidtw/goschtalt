// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
)

func TestDecoderConfig(t *testing.T) {
	thing := "string"

	tests := []struct {
		description string
		str         string
		opt         DecoderConfigOption
		want        mapstructure.DecoderConfig
		check       func(*mapstructure.DecoderConfig) bool
	}{
		{
			description: "DecodeHook",
			str:         "DecodeHook(custom)",
			opt:         DecodeHook(func() {}),
			check: func(m *mapstructure.DecoderConfig) bool {
				return m.DecodeHook != nil
			},
		}, {
			description: "ErrorUnused()",
			str:         "ErrorUnused()",
			opt:         ErrorUnused(),
			want: mapstructure.DecoderConfig{
				ErrorUnused: true,
			},
		}, {
			description: "ErrorUnused(true)",
			str:         "ErrorUnused()",
			opt:         ErrorUnused(true),
			want: mapstructure.DecoderConfig{
				ErrorUnused: true,
			},
		}, {
			description: "ErrorUnused(false)",
			str:         "ErrorUnused(false)",
			opt:         ErrorUnused(false),
			want:        mapstructure.DecoderConfig{},
		}, {
			description: "ErrorUnset()",
			str:         "ErrorUnset()",
			opt:         ErrorUnset(),
			want: mapstructure.DecoderConfig{
				ErrorUnset: true,
			},
		}, {
			description: "ErrorUnset(true)",
			str:         "ErrorUnset()",
			opt:         ErrorUnset(true),
			want: mapstructure.DecoderConfig{
				ErrorUnset: true,
			},
		}, {
			description: "ErrorUnset(false)",
			str:         "ErrorUnset(false)",
			opt:         ErrorUnset(false),
			want:        mapstructure.DecoderConfig{},
		}, {
			description: "WeaklyTypedInput()",
			str:         "WeaklyTypedInput()",
			opt:         WeaklyTypedInput(),
			want: mapstructure.DecoderConfig{
				WeaklyTypedInput: true,
			},
		}, {
			description: "WeaklyTypedInput(true)",
			str:         "WeaklyTypedInput()",
			opt:         WeaklyTypedInput(true),
			want: mapstructure.DecoderConfig{
				WeaklyTypedInput: true,
			},
		}, {
			description: "WeaklyTypedInput(false)",
			str:         "WeaklyTypedInput(false)",
			opt:         WeaklyTypedInput(false),
			want:        mapstructure.DecoderConfig{},
		}, {
			description: "TagName('')",
			str:         "TagName('')",
			opt:         TagName(""),
			want:        mapstructure.DecoderConfig{},
		}, {
			description: "TagName('foo')",
			str:         "TagName('foo')",
			opt:         TagName("foo"),
			want: mapstructure.DecoderConfig{
				TagName: "foo",
			},
		}, {
			description: "IgnoreUntaggedFields()",
			str:         "IgnoreUntaggedFields()",
			opt:         IgnoreUntaggedFields(),
			want: mapstructure.DecoderConfig{
				IgnoreUntaggedFields: true,
			},
		}, {
			description: "IgnoreUntaggedFields(true)",
			str:         "IgnoreUntaggedFields()",
			opt:         IgnoreUntaggedFields(true),
			want: mapstructure.DecoderConfig{
				IgnoreUntaggedFields: true,
			},
		}, {
			description: "IgnoreUntaggedFields(false)",
			str:         "IgnoreUntaggedFields(false)",
			opt:         IgnoreUntaggedFields(false),
			want:        mapstructure.DecoderConfig{},
		}, {
			description: "ZeroFields()",
			str:         "ZeroFields()",
			opt:         ZeroFields(),
			want: mapstructure.DecoderConfig{
				ZeroFields: true,
			},
		}, {
			description: "ZeroFields(true)",
			str:         "ZeroFields()",
			opt:         ZeroFields(true),
			want: mapstructure.DecoderConfig{
				ZeroFields: true,
			},
		}, {
			description: "ZeroFields(false)",
			str:         "ZeroFields(false)",
			opt:         ZeroFields(false),
			want:        mapstructure.DecoderConfig{},
		}, {
			description: "MatchName(fn)",
			str:         "MatchName(custom)",
			opt:         MatchName(func(key, field string) bool { return true }),
			check: func(m *mapstructure.DecoderConfig) bool {
				return m.MatchName != nil
			},
		}, {
			description: "Exactly(1)",
			str:         "Exactly(DecodeHook: '', ErrorUnused: false, ErrorUnset: true, ZeroFields: true, WeaklyTypedInput: false, TagName: '', IgnoreUntaggedFields: false, MatchName: '')",
			opt: Exactly(mapstructure.DecoderConfig{
				ErrorUnset: true,
				Squash:     true, // ignored
				ZeroFields: true,
			}),
			want: mapstructure.DecoderConfig{
				ErrorUnset: true,
				ZeroFields: true,
			},
		}, {
			description: "Exactly(2)",
			str:         "Exactly(DecodeHook: '', ErrorUnused: true, ErrorUnset: false, ZeroFields: false, WeaklyTypedInput: true, TagName: 'tags', IgnoreUntaggedFields: true, MatchName: '')",
			opt: Exactly(mapstructure.DecoderConfig{
				ErrorUnused:          true,
				Result:               &thing, // ignored
				TagName:              "tags",
				WeaklyTypedInput:     true,
				IgnoreUntaggedFields: true,
			}),
			want: mapstructure.DecoderConfig{
				ErrorUnused:          true,
				TagName:              "tags",
				WeaklyTypedInput:     true,
				IgnoreUntaggedFields: true,
			},
		}, {
			description: "Exactly(3)",
			str:         "Exactly(DecodeHook: custom, ErrorUnused: false, ErrorUnset: false, ZeroFields: false, WeaklyTypedInput: false, TagName: '', IgnoreUntaggedFields: false, MatchName: custom)",
			opt: Exactly(mapstructure.DecoderConfig{
				DecodeHook: func() {},
				MatchName:  func(k, f string) bool { return true },
			}),
			check: func(m *mapstructure.DecoderConfig) bool {
				return m.MatchName != nil && m.DecodeHook != nil
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			assert.Equal(tc.str, tc.opt.String())

			// These are not the default option.
			assert.False(tc.opt.isDefault())

			var opt mapstructure.DecoderConfig
			tc.opt.decoderApply(&opt)

			un := unmarshalOptions{}
			tc.opt.unmarshalApply(&un)

			vo := valueOptions{}
			tc.opt.valueApply(&vo)
			assert.Equal(vo, valueOptions{})

			if tc.check == nil {
				assert.Equal(opt, tc.want)
				assert.Equal(un.decoder, tc.want)
			} else {
				tc.check(&opt)
				tc.check(&un.decoder)
			}
		})
	}
}
