// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package simple

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/schmidtw/goschtalt"
	"github.com/schmidtw/goschtalt/pkg/decoder"
	"github.com/schmidtw/goschtalt/pkg/meta"
	"github.com/stretchr/testify/assert"
)

type fake struct{}

func (f fake) Decode(ctx decoder.Context, b []byte, m *meta.Object) error {
	// Just eat the yaml ones.
	if !strings.HasSuffix(ctx.Filename, "json") {
		return nil
	}

	var data any
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	*m = meta.ObjectFromRaw(data)
	return nil
}

func (f fake) Extensions() []string { return []string{"yml", "yaml", "json"} }

func TestGetConfig(t *testing.T) {
	unknown := fmt.Errorf("unknown")
	helpText := `Usage: app [OPTION]...

  -f, --file file        File to process for configuration.  May be repeated.
  -d, --dir dir          Directory to walk for configuration.  Does not recurse.  May be repeated.
  -r, --recurse dir      Recursively walk directory for configuration.  May be repeated.
      --kvp key value    Set a key/value pair as configuration.  May be repeated.

  -s, --show-all         Show all configuration details, then exit.
      --show-cfg         Show the redacted final configuration including the origin of the value, then exit.
      --show-cfg-doc     Show built in config file with all options documented.
      --show-cfg-unsafe  Show the non-redacted version of --show-cfg.  Not included in --show-all or -s.
      --show-exts        Show the supported configuration file extensions, then exit.
      --show-files       Show files in the order processed, then exit.

      --skip-validation  Skips the validation of the default configuration against expected structs.
  -l, --licensing        Show licensing details, then exit.
  -v, --version          Print the version information, then exit.
  -h, --help             Output this text, then exit.


Automatic configuration files:

  000.yml   built in configuration document
  900.cli   command line arguments
`

	versionText := `app:
  version:    undefined
  built time: undefined
  git commit: undefined
  go version: %s
  os/arch:    %s/%s
`

	showAll := `Supported File Extensions:
	'cli', 'json', 'yaml', 'yml'

Files Processed first (1) to last (2):
	1. 000.yml
	2. 900.cli

-----BEGIN DEFAULT CONFIGURATION-----
---
  Foo: bar #comments
-----END DEFAULT CONFIGURATION-----

-----BEGIN REDACTED UNIFIED CONFIGURATION-----
foo: # 900.cli
    bar: cat # 900.cli
    bear: brown # 900.cli
-----END REDACTED UNIFIED CONFIGURATION-----

`
	type FooStruct struct {
		Bar string
	}
	empty := FooStruct{}

	defCfg := DefaultConfig{
		Text: "---\n  Foo: bar #comments",
		Ext:  "yml",
	}
	secretCfg := DefaultConfig{
		Text: `{ "Foo": {"bar((secret))": "car"}}`,
		Ext:  "json",
	}
	tests := []struct {
		description string
		name        string
		prefix      string
		license     string
		args        []string
		opts        []goschtalt.Option
		defCfg      DefaultConfig
		validate    map[string]any
		expectedCfg bool
		expectedErr error
		expect      string
	}{
		{
			description: "Empty config.",
			name:        "app",
			expectedErr: ErrDefaultConfigInvalid,
		}, {
			description: "Get help, ignoring everything else.",
			name:        "app",
			defCfg:      defCfg,
			args:        []string{"-h", "-d", "ignored"},
			expect:      helpText,
		}, {
			description: "Get version, ignoring everything else.",
			name:        "app",
			defCfg:      defCfg,
			args:        []string{"-v", "-h", "-d", "ignored"},
			expect:      fmt.Sprintf(versionText, runtime.Version(), runtime.GOOS, runtime.GOARCH),
		}, {
			description: "Show the licensing output with no license.",
			name:        "app",
			defCfg:      defCfg,
			args:        []string{"-l"},
			expect:      "Licensing information is unavailable.\n",
		}, {
			description: "Show the licensing output with a license.",
			name:        "app",
			license:     "Apache license.",
			defCfg:      defCfg,
			args:        []string{"-l"},
			expect:      "Licensing for app:\nApache license.\n",
		}, {
			description: "Show everything.",
			name:        "app",
			defCfg:      defCfg,
			opts:        []goschtalt.Option{goschtalt.RegisterDecoder(fake{})},
			args:        []string{"--kvp", "Foo.bar", "cat", "--kvp", "Foo.bear", "brown", "-s"},
			expect:      showAll,
		}, {
			description: "Show files only.",
			name:        "app",
			defCfg:      defCfg,
			opts:        []goschtalt.Option{goschtalt.RegisterDecoder(fake{})},
			args:        []string{"--show-files"},
			expect:      "Files Processed first (1) to last (2):\n\t1. 000.yml\n\t2. 900.cli\n\n",
		}, {
			description: "Show files only.",
			name:        "app",
			defCfg:      defCfg,
			opts:        []goschtalt.Option{goschtalt.RegisterDecoder(fake{})},
			args:        []string{"--show-files"},
			expect:      "Files Processed first (1) to last (2):\n\t1. 000.yml\n\t2. 900.cli\n\n",
		}, {
			description: "Show config doc only.",
			name:        "app",
			defCfg:      defCfg,
			opts:        []goschtalt.Option{goschtalt.RegisterDecoder(fake{})},
			args:        []string{"--show-cfg-doc"},
			expect:      "---\n  Foo: bar #comments\n",
		}, {
			description: "Show config only.",
			name:        "app",
			defCfg:      defCfg,
			opts:        []goschtalt.Option{goschtalt.RegisterDecoder(fake{})},
			args:        []string{"--show-cfg", "--kvp", "Foo.bar", "cat"},
			expect:      "foo: # 900.cli\n    bar: cat # 900.cli\n",
		}, {
			description: "Show extensions only.",
			name:        "app",
			defCfg:      defCfg,
			opts:        []goschtalt.Option{goschtalt.RegisterDecoder(fake{})},
			args:        []string{"--show-exts"},
			expect:      "Supported File Extensions:\n\t'cli', 'json', 'yaml', 'yml'\n\n",
		}, {
			description: "Show redacted.",
			name:        "app",
			defCfg:      secretCfg,
			opts:        []goschtalt.Option{goschtalt.RegisterDecoder(fake{})},
			args:        []string{"--show-cfg"},
			expect:      "foo:\n    bar: REDACTED\n",
		}, {
			description: "Show un-redacted.",
			name:        "app",
			defCfg:      secretCfg,
			opts:        []goschtalt.Option{goschtalt.RegisterDecoder(fake{})},
			args:        []string{"--show-cfg-unsafe"},
			expect:      "foo:\n    bar: car\n",
		}, {
			description: "Validate the structure as a pointer.",
			name:        "app",
			defCfg: DefaultConfig{
				Text: `{ "Foo": {"bar": "cat"}}`,
				Ext:  "json",
			},
			opts:        []goschtalt.Option{goschtalt.RegisterDecoder(fake{})},
			args:        []string{},
			validate:    map[string]any{"foo": &FooStruct{}},
			expectedCfg: true,
		}, {
			description: "Validate the pointer value isn't altered.",
			name:        "app",
			defCfg: DefaultConfig{
				Text: `{ "Foo": {"bar": "cat"}}`,
				Ext:  "json",
			},
			opts:        []goschtalt.Option{goschtalt.RegisterDecoder(fake{})},
			args:        []string{},
			validate:    map[string]any{"foo": &empty},
			expectedCfg: true,
		}, {
			description: "Validate the structure directly.",
			name:        "app",
			defCfg: DefaultConfig{
				Text: `{ "Foo": {"bar": "cat"}}`,
				Ext:  "json",
			},
			opts:        []goschtalt.Option{goschtalt.RegisterDecoder(fake{})},
			args:        []string{},
			validate:    map[string]any{"foo": FooStruct{}},
			expectedCfg: true,
		}, {
			description: "Validate the structure, but missing bar so fail.",
			name:        "app",
			defCfg: DefaultConfig{
				Text: `{ "Foo": {"bart": "cat"}}`,
				Ext:  "json",
			},
			opts:        []goschtalt.Option{goschtalt.RegisterDecoder(fake{})},
			args:        []string{},
			validate:    map[string]any{"foo": &FooStruct{}},
			expectedErr: ErrDefaultConfigInvalid,
		}, {
			description: "Validate the structure, missing bar, but skip validation.",
			name:        "app",
			defCfg: DefaultConfig{
				Text: `{ "Foo": {"bart": "cat"}}`,
				Ext:  "json",
			},
			opts:        []goschtalt.Option{goschtalt.RegisterDecoder(fake{})},
			args:        []string{"--skip-validation"},
			validate:    map[string]any{"foo": &FooStruct{}},
			expectedCfg: true,
		}, {
			description: "Invalid default file during validation.",
			name:        "app",
			defCfg: DefaultConfig{
				Text: `{ "Foo": {"bart": "cat"}`,
				Ext:  "json",
			},
			opts:        []goschtalt.Option{goschtalt.RegisterDecoder(fake{})},
			args:        []string{},
			validate:    map[string]any{"foo": &FooStruct{}},
			expectedErr: unknown,
		}, {
			description: "Invalid default file no validation.",
			name:        "app",
			defCfg: DefaultConfig{
				Text: `{ "Foo": {"bart": "cat"}`,
				Ext:  "json",
			},
			opts:        []goschtalt.Option{goschtalt.RegisterDecoder(fake{})},
			args:        []string{},
			expectedErr: unknown,
		}, {
			description: "Invalid options during validation.",
			name:        "app",
			defCfg: DefaultConfig{
				Text: `{ "Foo": {"bart": "cat"}}`,
				Ext:  "json",
			},
			opts:        []goschtalt.Option{goschtalt.RegisterDecoder(fake{}), goschtalt.RegisterDecoder(fake{})},
			args:        []string{},
			validate:    map[string]any{"foo": &FooStruct{}},
			expectedErr: unknown,
		}, {
			description: "Invalid options no validation.",
			name:        "app",
			defCfg: DefaultConfig{
				Text: `{ "Foo": {"bart": "cat"}}`,
				Ext:  "json",
			},
			opts:        []goschtalt.Option{goschtalt.RegisterDecoder(fake{}), goschtalt.RegisterDecoder(fake{})},
			args:        []string{},
			expectedErr: unknown,
		}, {
			description: "End to end.",
			name:        "app",
			defCfg:      defCfg,
			opts:        []goschtalt.Option{goschtalt.RegisterDecoder(fake{})},
			args:        []string{"--kvp", "Foo.bar", "cat"},
			expectedCfg: true,
		}, {
			description: "Verify the default extension check.",
			name:        "app",
			defCfg:      defCfg,
			expectedErr: ErrDefaultConfigInvalid,
		}, {
			description: "Verify the extension is sane.",
			name:        "app",
			defCfg: DefaultConfig{
				Text: "---",
				Ext:  "foo/bar",
			},
			expectedErr: ErrDefaultConfigInvalid,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			var b bytes.Buffer
			w := bufio.NewWriter(&b)

			p := Program{
				Name:      tc.name,
				Default:   tc.defCfg,
				Prefix:    tc.prefix,
				Licensing: tc.license,
				Output:    w,
				Validate:  tc.validate,
			}

			if len(p.Prefix) == 0 {
				p.Prefix = "ILLEGAL.SO.NONE.MATCH"
			}

			cfg, err := p.getConfig(tc.args, tc.opts...)
			_ = w.Flush()

			if tc.expectedErr == nil {
				if tc.expectedCfg {
					assert.NotNil(cfg)
				} else {
					assert.Nil(cfg)
				}
				got := b.String()
				assert.Equal(tc.expect, got)
				if tc.expect != got {
					fmt.Printf("got:\n%s", got)
				}
				assert.NoError(err)
				assert.Empty(empty)
				return
			}

			assert.Error(err)
			if tc.expectedErr != unknown {
				assert.ErrorIs(err, tc.expectedErr)
			}
		})
	}
}
