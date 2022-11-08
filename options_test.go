// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"errors"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/k0kubun/pp"
	"github.com/schmidtw/goschtalt/pkg/decoder"
	"github.com/schmidtw/goschtalt/pkg/encoder"
	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	unknown := errors.New("unknown")

	testErr := errors.New("test err")
	fs := os.DirFS("/")
	list := []string{"zeta", "alpha", "19beta", "19alpha", "4tango",
		"1alpha", "7alpha", "bravo", "7alpha10", "7alpha2", "7alpha0"}

	sortCheck := func(cfg *options, want []string) bool {
		if cfg.sorter == nil {
			return false
		}

		got := list

		sorter := func(a []string) {
			sort.SliceStable(a, func(i, j int) bool {
				return cfg.sorter(a[i], a[j])
			})
		}

		sorter(got)

		return reflect.DeepEqual(got, want)
	}

	tests := []struct {
		description string
		opt         Option
		opts        []Option
		goal        options
		check       func(*options) bool
		str         string
		ignore      bool
		initCodecs  bool
		expectErr   error
	}{
		{
			description: "WithError( testErr )",
			opt:         WithError(testErr),
			str:         "WithError( 'test err' )",
			expectErr:   testErr,
		}, {
			description: "WithError( nil )",
			opt:         WithError(nil),
			str:         "WithError( '' )",
		}, {
			description: "AddFile( /, filename )",
			opt:         AddFile(fs, "filename"),
			str:         "AddFile( 'filename' )",
			goal: options{
				groups: []group{
					{
						fs:    fs,
						paths: []string{"filename"},
					},
				},
			},
		}, {
			description: "AddFile( /, a ), AddFile( /, b )",
			opts:        []Option{AddFile(fs, "a"), AddFile(fs, "b")},
			goal: options{
				groups: []group{
					{
						fs:    fs,
						paths: []string{"a"},
					}, {
						fs:    fs,
						paths: []string{"b"},
					},
				},
			},
		}, {
			description: "AddFiles( / )",
			opt:         AddFiles(fs),
			str:         "AddFiles( '' )",
			goal: options{
				groups: []group{{fs: fs}},
			},
		}, {
			description: "AddFiles( /, filename )",
			opt:         AddFiles(fs, "filename"),
			str:         "AddFiles( 'filename' )",
			goal: options{
				groups: []group{
					{
						fs:    fs,
						paths: []string{"filename"},
					},
				},
			},
		}, {
			description: "AddFiles( /, a, b )",
			opt:         AddFiles(fs, "a", "b"),
			str:         "AddFiles( 'a', 'b' )",
			goal: options{
				groups: []group{
					{
						fs:    fs,
						paths: []string{"a", "b"},
					},
				},
			},
		}, {
			description: "AddTree( /, path )",
			opt:         AddTree(fs, "./path"),
			str:         "AddTree( './path' )",
			goal: options{
				groups: []group{
					{
						fs:      fs,
						paths:   []string{"./path"},
						recurse: true,
					},
				},
			},
		}, {
			description: "AddDir( /, path )",
			opt:         AddDir(fs, "./path"),
			str:         "AddDir( './path' )",
			goal: options{
				groups: []group{
					{
						fs:    fs,
						paths: []string{"./path"},
					},
				},
			},
		}, {
			description: "AutoCompile()",
			opt:         AutoCompile(),
			str:         "AutoCompile()",
			goal: options{
				autoCompile: true,
			},
		}, {
			description: "AutoCompile(false)",
			opt:         AutoCompile(false),
			str:         "AutoCompile( false )",
		}, {
			description: "AlterKeyCase(nil)",
			opt:         AlterKeyCase(nil),
			str:         "AlterKeyCase( none )",
			check: func(cfg *options) bool {
				if cfg.keySwizzler == nil {
					return false
				}
				return cfg.keySwizzler("AbCd") == "AbCd"
			},
		}, {
			description: "AlterKeyCase(strings.ToLower)",
			opt:         AlterKeyCase(strings.ToLower),
			str:         "AlterKeyCase( custom )",
			check: func(cfg *options) bool {
				if cfg.keySwizzler == nil {
					return false
				}
				return cfg.keySwizzler("AbCd") == "abcd"
			},
		}, {
			description: "AlterKeyCase(strings.ToUpper)",
			opt:         AlterKeyCase(strings.ToUpper),
			str:         "AlterKeyCase( custom )",
			check: func(cfg *options) bool {
				if cfg.keySwizzler == nil {
					return false
				}
				return cfg.keySwizzler("AbCd") == "ABCD"
			},
		}, {
			description: "SetKeyDelimiter( . )",
			opt:         SetKeyDelimiter("."),
			str:         "SetKeyDelimiter( '.' )",
			goal: options{
				keyDelimiter: ".",
			},
		}, {
			description: "SetKeyDelimiter( '' )",
			opt:         SetKeyDelimiter(""),
			str:         "SetKeyDelimiter( '' )",
			expectErr:   ErrInvalidInput,
		}, {
			description: "SortRecordsCustomFn( '' )",
			opt:         SortRecordsCustomFn(nil),
			str:         "SortRecordsCustomFn( custom )",
			expectErr:   ErrInvalidInput,
		}, {
			description: "SortRecordsCustomFn( '(reverse)' )",
			opt: SortRecordsCustomFn(func(a, b string) bool {
				return a > b
			}),
			str: "SortRecordsCustomFn( custom )",
			check: func(cfg *options) bool {
				return sortCheck(cfg, []string{
					"zeta",
					"bravo",
					"alpha",
					"7alpha2",
					"7alpha10",
					"7alpha0",
					"7alpha",
					"4tango",
					"1alpha",
					"19beta",
					"19alpha",
				})
			},
		}, {
			description: "SortRecordsNaturally()",
			opt:         SortRecordsNaturally(),
			str:         "SortRecordsNaturally()",
			check: func(cfg *options) bool {
				return sortCheck(cfg, []string{
					"1alpha",
					"4tango",
					"7alpha",
					"7alpha0",
					"7alpha2",
					"7alpha10",
					"19alpha",
					"19beta",
					"alpha",
					"bravo",
					"zeta",
				})
			},
		}, {
			description: "SortRecordsLexically()",
			opt:         SortRecordsLexically(),
			str:         "SortRecordsLexically()",
			check: func(cfg *options) bool {
				return sortCheck(cfg, []string{
					"19alpha",
					"19beta",
					"1alpha",
					"4tango",
					"7alpha",
					"7alpha0",
					"7alpha10",
					"7alpha2",
					"alpha",
					"bravo",
					"zeta",
				})
			},
		}, {
			description: "WithEncoder( json, yml )",
			opt:         WithEncoder(&testEncoder{extensions: []string{"json", "yml"}}),
			str:         "WithEncoder( 'json', 'yml' )",
			initCodecs:  true,
			check: func(cfg *options) bool {
				return reflect.DeepEqual([]string{"json", "yml"},
					cfg.encoders.extensions())
			},
		}, {
			description: "WithEncoder( foo )",
			opt:         WithEncoder(&testEncoder{extensions: []string{"foo"}}),
			str:         "WithEncoder( 'foo' )",
			initCodecs:  true,
			check: func(cfg *options) bool {
				return reflect.DeepEqual([]string{"foo"},
					cfg.encoders.extensions())
			},
		}, {
			description: "WithEncoder(nil)",
			opt:         WithEncoder(nil),
			str:         "WithEncoder( '' )",
		}, {
			description: "WithDecoder( json, yml )",
			opt:         WithDecoder(&testDecoder{extensions: []string{"json", "yml"}}),
			str:         "WithDecoder( 'json', 'yml' )",
			initCodecs:  true,
			check: func(cfg *options) bool {
				return reflect.DeepEqual([]string{"json", "yml"},
					cfg.decoders.extensions())
			},
		}, {
			description: "WithDecoder( foo )",
			opt:         WithDecoder(&testDecoder{extensions: []string{"foo"}}),
			str:         "WithDecoder( 'foo' )",
			initCodecs:  true,
			check: func(cfg *options) bool {
				return reflect.DeepEqual([]string{"foo"},
					cfg.decoders.extensions())
			},
		}, {
			description: "WithDecoder(nil)",
			opt:         WithDecoder(nil),
			str:         "WithDecoder( '' )",
		}, {
			description: "DisableDefaultPackageOptions()",
			opt:         DisableDefaultPackageOptions(),
			str:         "DisableDefaultPackageOptions()",
			ignore:      true,
		}, {
			description: "DefaultMarshalOptions()",
			opt:         DefaultMarshalOptions(),
			str:         "DefaultMarshalOptions()",
		}, {
			description: "DefaultMarshalOptions( RedactSecrets(true), IncludeOrigins(true), FormatAs(foo) )",
			opt:         DefaultMarshalOptions(RedactSecrets(true), IncludeOrigins(true), FormatAs("foo")),
			str:         "DefaultMarshalOptions( RedactSecrets(), IncludeOrigins(), FormatAs('foo') )",
			goal: options{
				marshalOptions: []MarshalOption{redactSecretsOption(true), includeOriginsOption(true), formatAsOption("foo")},
			},
		}, {
			description: "DefaultMarshalOptions( RedactSecrets(false), IncludeOrigins(false) )",
			opt:         DefaultMarshalOptions(RedactSecrets(false), IncludeOrigins(false)),
			str:         "DefaultMarshalOptions( RedactSecrets(false), IncludeOrigins(false) )",
			goal: options{
				marshalOptions: []MarshalOption{redactSecretsOption(false), includeOriginsOption(false)},
			},
		}, {
			description: "DefaultMarshalOptions( RedactSecrets(), IncludeOrigins() )",
			opt:         DefaultMarshalOptions(RedactSecrets(), IncludeOrigins()),
			str:         "DefaultMarshalOptions( RedactSecrets(), IncludeOrigins() )",
			goal: options{
				marshalOptions: []MarshalOption{redactSecretsOption(true), includeOriginsOption(true)},
			},
		}, {
			description: "DefaultMarshalOptions( RedactSecrets() ), DefaultMarshalOptions( IncludeOrigins() )",
			opts: []Option{
				DefaultMarshalOptions(RedactSecrets(true)),
				DefaultMarshalOptions(IncludeOrigins(true)),
			},
			goal: options{
				marshalOptions: []MarshalOption{redactSecretsOption(true), includeOriginsOption(true)},
			},
		}, {
			description: "DefaultUnmarshalOptions()",
			opt:         DefaultUnmarshalOptions(),
			str:         "DefaultUnmarshalOptions()",
		}, {
			description: "DefaultUnmarshalOptions( Optional(true), Required(true) )",
			opt:         DefaultUnmarshalOptions(Optional(true), Required(true)),
			str:         "DefaultUnmarshalOptions( Optional(), Required() )",
			goal: options{
				unmarshalOptions: []UnmarshalOption{
					&optionalOption{
						text:     "Optional()",
						optional: true,
					},
					&optionalOption{
						text: "Required()",
					},
				},
			},
		}, {
			description: "DefaultUnmarshalOptions( Optional(false), Required(false) )",
			opt:         DefaultUnmarshalOptions(Optional(false), Required(false)),
			str:         "DefaultUnmarshalOptions( Optional(false), Required(false) )",
			goal: options{
				unmarshalOptions: []UnmarshalOption{
					&optionalOption{
						text: "Optional(false)",
					},
					&optionalOption{
						text:     "Required(false)",
						optional: true,
					},
				},
			},
		}, {
			description: "DefaultUnmarshalOptions( Optional(), Required() )",
			opt:         DefaultUnmarshalOptions(Optional(), Required()),
			str:         "DefaultUnmarshalOptions( Optional(), Required() )",
			goal: options{
				unmarshalOptions: []UnmarshalOption{
					&optionalOption{
						text:     "Optional()",
						optional: true,
					},
					&optionalOption{
						text: "Required()",
					},
				},
			},
		}, {
			description: "DefaultUnmarshalOptions( Optional() ), DefaultUnmarshalOptions( Required() )",
			opts: []Option{
				DefaultUnmarshalOptions(Optional()),
				DefaultUnmarshalOptions(Required()),
			},
			goal: options{
				unmarshalOptions: []UnmarshalOption{
					&optionalOption{
						text:     "Optional()",
						optional: true,
					},
					&optionalOption{
						text: "Required()",
					},
				},
			},
		}, {
			description: "DefaultUnmarshalOptions( DecodeHook() )",
			opts: []Option{
				DefaultUnmarshalOptions(DecodeHook(nil)),
			},
			goal: options{
				unmarshalOptions: []UnmarshalOption{
					&decodeHookOption{},
				},
			},
		}, {
			description: "DefaultUnmarshalOptions( most )",
			opt: DefaultUnmarshalOptions(
				DecodeHook(nil),
				ErrorUnused(),
				ErrorUnset(),
				WeaklyTypedInput(),
				TagName("tag"),
				IgnoreUntaggedFields(),
				MatchName(nil),
			),
			goal: options{
				unmarshalOptions: []UnmarshalOption{
					&decodeHookOption{},
					errorUnusedOption(true),
					errorUnsetOption(true),
					weaklyTypedInputOption(true),
					tagNameOption("tag"),
					ignoreUntaggedFieldsOption(true),
					&matchNameOption{},
				},
			},
			str: "DefaultUnmarshalOptions( DecodeHook(''), ErrorUnused(), ErrorUnset(), WeaklyTypedInput(), TagName('tag'), IgnoreUntaggedFields(), MatchName('') )",
		}, {
			description: "DefaultUnmarshalOptions( most )",
			opt: DefaultUnmarshalOptions(
				ErrorUnused(false),
				ErrorUnset(false),
				WeaklyTypedInput(false),
				IgnoreUntaggedFields(false),
			),
			goal: options{
				unmarshalOptions: []UnmarshalOption{
					errorUnusedOption(false),
					errorUnsetOption(false),
					weaklyTypedInputOption(false),
					ignoreUntaggedFieldsOption(false),
				},
			},
			str: "DefaultUnmarshalOptions( ErrorUnused(false), ErrorUnset(false), WeaklyTypedInput(false), IgnoreUntaggedFields(false) )",
		}, {
			description: "DefaultUnmarshalOptions( most )",
			opt: DefaultUnmarshalOptions(
				DecodeHook(func() {}),
				MatchName(func(k, f string) bool { return true }),
			),
			check: func(cfg *options) bool {
				return len(cfg.unmarshalOptions) == 2
			},
			str: "DefaultUnmarshalOptions( DecodeHook(custom), MatchName(custom) )",
		}, {
			description: "DefaultValueOptions()",
			opt:         DefaultValueOptions(),
			str:         "DefaultValueOptions()",
		}, {
			description: "DefaultValueOptions( DecodeHook() )",
			opt:         DefaultValueOptions(DecodeHook(nil)),
			goal: options{
				valueOptions: []ValueOption{
					&decodeHookOption{},
				},
			},
			str: "DefaultValueOptions( DecodeHook('') )",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			var cfg options

			if tc.initCodecs {
				cfg.encoders = newRegistry[encoder.Encoder]()
				cfg.decoders = newRegistry[decoder.Decoder]()
			}

			var err error
			if len(tc.opts) == 0 {
				err = tc.opt.apply(&cfg)

				s := tc.opt.String()
				assert.Equal(tc.str, s)

				assert.Equal(tc.ignore, tc.opt.ignoreDefaults())
				assert.Equal(tc.ignore, ignoreDefaultOpts([]Option{tc.opt}))
			} else {
				assert.Equal(tc.ignore, ignoreDefaultOpts(tc.opts))
				for _, opt := range tc.opts {
					err = opt.apply(&cfg)
				}
			}

			if tc.expectErr == nil {
				if tc.check != nil {
					assert.True(tc.check(&cfg))
				} else {
					assert.True(reflect.DeepEqual(tc.goal, cfg))
					if !reflect.DeepEqual(tc.goal, cfg) {
						pp.Printf("Want:\n%s\n", tc.goal)
						pp.Printf("Got:\n%s\n", cfg)
					}
				}
				return
			}

			if !errors.Is(unknown, tc.expectErr) {
				assert.ErrorIs(err, tc.expectErr)
				return
			}

		})
	}
}
