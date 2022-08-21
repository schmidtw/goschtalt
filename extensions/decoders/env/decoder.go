// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

// env package is a goschtalt decoder package for processing environment variables.
//
// See the example for how to use this extension package.
package env

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/psanford/memfs"
	"github.com/schmidtw/goschtalt"
	"github.com/schmidtw/goschtalt/pkg/decoder"
	"github.com/schmidtw/goschtalt/pkg/meta"
)

const extension = `ENVIRONMENT_VARIABLE`

var _ decoder.Decoder = (*envDecoder)(nil)

// envDecoder is a class for the property decoder.
type envDecoder struct{}

// Extensions returns the supported extensions.
func (d envDecoder) Extensions() []string {
	return []string{extension}
}

type instructions struct {
	Prefix    string
	Delimiter string
}

// Decode decodes a byte arreay into the meta.Object tree.
func (d envDecoder) Decode(filename, keyDelimiter string, b []byte, m *meta.Object) error {
	var inst instructions
	err := json.Unmarshal(b, &inst)
	if err != nil {
		return err
	}

	origin := meta.Origin{
		File: filename,
	}

	tree := meta.Object{
		Origins: []meta.Origin{origin},
		Map:     make(map[string]meta.Object),
	}
	list := os.Environ()
	for _, item := range list {
		kvp := strings.Split(item, "=")
		if len(kvp) > 1 && strings.HasPrefix(kvp[0], inst.Prefix) {
			key := kvp[0]
			val := os.Getenv(key)
			key = strings.TrimPrefix(key, inst.Prefix)
			tree, err = tree.Add(inst.Delimiter, key, meta.StringToBestType(val), origin)
			if err != nil {
				return err
			}
		}
	}

	*m = tree.ConvertMapsToArrays()

	return nil
}

// EnvVarConfig provides a way to collect configuration values from environment
// variables passed into the program.  The filename is used to sort prior to
// the merge step, so the order of operations can be specified.  The prefix is
// the environment variable name prefix to look for when collecting them.  The
// delimiter is what string should be used to split the tree structure set on.
//
// For some environment variable environments like bash the allowable characters
// in the names is limited to: `[a-zA-Z_][a-zA-Z0-9_]*`
//
// If you need multiple prefix values, this option is safe to use multiple times.
func EnvVarConfig(filename, prefix, delimiter string) []goschtalt.Option {
	fn := fmt.Sprintf("%s.%s", filename, extension)

	inst := instructions{
		Prefix:    prefix,
		Delimiter: delimiter,
	}
	b, err := json.Marshal(inst)
	if err != nil {
		return []goschtalt.Option{
			func(_ *goschtalt.Config) error {
				return err
			},
		}
	}

	envfs := memfs.New()
	err = envfs.WriteFile(fn, b, 0755)
	if err != nil {
		return []goschtalt.Option{
			func(_ *goschtalt.Config) error {
				return err
			},
		}
	}

	group := goschtalt.Group{
		FS:    envfs,
		Paths: []string{"."},
	}

	return []goschtalt.Option{
		goschtalt.DecoderRemove(extension),
		goschtalt.DecoderRegister(envDecoder{}),
		goschtalt.FileGroup(group),
	}
}
