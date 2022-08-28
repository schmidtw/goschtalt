// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package simple

import (
	"bufio"
	"bytes"
	"fmt"
	"runtime"
	"testing"

	"github.com/schmidtw/goschtalt"
	"github.com/schmidtw/goschtalt/pkg/decoder"
	"github.com/schmidtw/goschtalt/pkg/meta"
	"github.com/stretchr/testify/assert"
)

type fake struct{}

func (f fake) Decode(_ decoder.Context, _ []byte, _ *meta.Object) error { return nil }
func (f fake) Extensions() []string                                     { return []string{"yml", "yaml", "json"} }

func TestGetConfig(t *testing.T) {
	helpText := `Usage: app [OPTION]...

  -f, --file file       File to process for configuration.  May be repeated.
  -d, --dir dir         Directory to walk for configuration.  Does not recurse.  May be repeated.
  -r, --recurse dir     Recursively walk directory for configuration.  May be repeated.
      --kvp key value   Set a key/value pair as configuration.  May be repeated.  Filename: '1000.cli'

  -s, --show-all        Show all configuration details, then exit.
      --show-cfg        Show the redacted final configuration including the origin of the value, then exit.
      --show-cfg-doc    Show built in config file with all options documentated.  Filename: '0.yml'
      --show-exts       Show the supported configuration file extensions, then exit.
      --show-files      Show files in the order processed, then exit.

  -v, --version         Print the version information, then exit.
  -h, --help            Output this text, then exit.
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

Files Processed first (top) to last (bottom):
	1. 0.yml
	2. 1000.cli

Default Configuration:
-- vvv -------------------------------------------------------------------------
---
  Foo: bar #comments
-- ^^^ -------------------------------------------------------------------------

Unified Configuration:
-- vvv -------------------------------------------------------------------------
foo: # 1000.cli
    bar: cat # 1000.cli
    bear: brown # 1000.cli

-- ^^^ -------------------------------------------------------------------------

`

	defCfg := DefaultConfig{
		Text: "---\n  Foo: bar #comments",
		Ext:  "yml",
	}
	tests := []struct {
		description string
		name        string
		args        []string
		opts        []goschtalt.Option
		defCfg      DefaultConfig
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
			description: "Show everything.",
			name:        "app",
			defCfg:      defCfg,
			opts:        []goschtalt.Option{goschtalt.DecoderRegister(fake{})},
			args:        []string{"--kvp", "Foo.bar", "cat", "--kvp", "Foo.bear", "brown", "-s"},
			expect:      showAll,
		}, {
			description: "Show files only.",
			name:        "app",
			defCfg:      defCfg,
			opts:        []goschtalt.Option{goschtalt.DecoderRegister(fake{})},
			args:        []string{"--show-files"},
			expect:      "Files Processed first (top) to last (bottom):\n\t1. 0.yml\n\t2. 1000.cli\n\n",
		}, {
			description: "Show files only.",
			name:        "app",
			defCfg:      defCfg,
			opts:        []goschtalt.Option{goschtalt.DecoderRegister(fake{})},
			args:        []string{"--show-files"},
			expect:      "Files Processed first (top) to last (bottom):\n\t1. 0.yml\n\t2. 1000.cli\n\n",
		}, {
			description: "Show config doc only.",
			name:        "app",
			defCfg:      defCfg,
			opts:        []goschtalt.Option{goschtalt.DecoderRegister(fake{})},
			args:        []string{"--show-cfg-doc"},
			expect:      "---\n  Foo: bar #comments\n",
		}, {
			description: "Show config only.",
			name:        "app",
			defCfg:      defCfg,
			opts:        []goschtalt.Option{goschtalt.DecoderRegister(fake{})},
			args:        []string{"--show-cfg", "--kvp", "Foo.bar", "cat"},
			expect:      "foo: # 1000.cli\n    bar: cat # 1000.cli\n\n",
		}, {
			description: "Show extensions only.",
			name:        "app",
			defCfg:      defCfg,
			opts:        []goschtalt.Option{goschtalt.DecoderRegister(fake{})},
			args:        []string{"--show-exts"},
			expect:      "Supported File Extensions:\n\t'cli', 'json', 'yaml', 'yml'\n\n",
		}, {
			description: "End to end.",
			name:        "app",
			defCfg:      defCfg,
			opts:        []goschtalt.Option{goschtalt.DecoderRegister(fake{})},
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

			cfg, err := getConfig(tc.name, tc.defCfg, tc.args, w, tc.opts...)
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
				return
			}
			assert.ErrorIs(err, tc.expectedErr)
		})
	}
}
