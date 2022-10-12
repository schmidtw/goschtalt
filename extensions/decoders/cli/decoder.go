// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

// cli package is a goschtalt decoder package for processing the command line
// interface.
//
// See the example for how to use this extension package.
package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/psanford/memfs"
	"github.com/schmidtw/goschtalt"
	"github.com/schmidtw/goschtalt/pkg/decoder"
	"github.com/schmidtw/goschtalt/pkg/meta"
)

// The extension this decoder uses since the decode itself isn't public.
const Extension = `cli`

var _ decoder.Decoder = (*cliDecoder)(nil)

var ErrInput = errors.New("invalid input")

// cliDecoder is a class for the property decoder.
type cliDecoder struct{}

// Extensions returns the supported extensions.
func (d cliDecoder) Extensions() []string {
	return []string{Extension}
}

type kvp struct {
	Key   string
	Value string
}

type instructions struct {
	Delimiter string
	Entries   []kvp
}

// Decode decodes a byte arreay into the meta.Object tree.
func (d cliDecoder) Decode(ctx decoder.Context, b []byte, m *meta.Object) error {
	var inst instructions
	err := json.Unmarshal(b, &inst)
	if err != nil {
		return err
	}

	origin := meta.Origin{
		File: ctx.Filename,
	}

	tree := meta.Object{
		Origins: []meta.Origin{origin},
		Map:     make(map[string]meta.Object),
	}
	for _, item := range inst.Entries {
		tree, err = tree.Add(inst.Delimiter, item.Key, meta.StringToBestType(item.Value), origin)
		if err != nil {
			return err
		}
	}

	*m = tree.ConvertMapsToArrays()

	return nil
}

func retErr(err error) []goschtalt.Option {
	return []goschtalt.Option{
		func(_ *goschtalt.Config) error {
			return err
		},
	}
}

func isKey(s string) bool {
	switch s {
	case "-d", "--dir", "-f", "--file", "--kvp", "-r", "--recurse":
		return true
	}

	return false
}

// CLIConfig is a lightweight and opinionated command line processor that
// configures goschtalt.Config based on command line options referencing the
// local filesystem(s) based on relative or absolut paths.  Additionally,
// key/value pairs can be provided as well.
//
//	The list of command options: (each may be repeated any number of times)
//	[-f filename] / [--file filename] provides an exact file to evaluate
//	[-d dir] / [--dir dir] provides a directory to examine for files
//	[-r dir] / [--recurse dir] provides a directory tree to examine for files recursively
//	[--kvp key value] provides a specific key and value pair that should be set
func Options(filename, delimiter string, args []string, dirFS ...func(string) fs.FS) []goschtalt.Option {
	dirfs := os.DirFS
	if len(dirFS) > 0 {
		dirfs = dirFS[0]
	}

	options := []goschtalt.Option{
		goschtalt.RemoveDecoder(Extension),
		goschtalt.RegisterDecoder(cliDecoder{}),
	}

	inst := instructions{
		Delimiter: delimiter,
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-f", "--file":
			// Make i point to the file part of `-f file`
			i++
			if len(args) <= i || isKey(args[i]) {
				return retErr(fmt.Errorf("%w: cli option [-f file] missing the file part after '%s'",
					ErrInput,
					strings.Join(args[:i], " ")))
			}
			file := args[i]
			group := goschtalt.Group{
				FS:    dirfs(path.Dir(file)),
				Paths: []string{path.Base(file)},
			}
			options = append(options, goschtalt.AddFileGroup(group))
		case "-d", "--dir", "-r", "--recurse":
			// Make i point to the dir part of `-d dir` or `-r dir`
			i++
			if len(args) <= i || isKey(args[i]) {
				return retErr(fmt.Errorf("%w: cli option [%s file] missing the file part after '%s'",
					ErrInput, arg,
					strings.Join(args[:i], " ")))
			}

			dir := args[i]
			group := goschtalt.Group{
				FS:    dirfs(dir),
				Paths: []string{"."},
			}
			if arg == "-r" || arg == "--recurse" {
				group.Recurse = true
			}
			options = append(options, goschtalt.AddFileGroup(group))
		case "--kvp":
			if len(args) <= (i + 2) {
				return retErr(fmt.Errorf("%w: cli option [--kvp key value] missing parameters '%s'",
					ErrInput,
					strings.Join(args, " ")))
			}
			entry := kvp{
				Key:   args[i+1],
				Value: args[i+2],
			}
			inst.Entries = append(inst.Entries, entry)
			i += 2
		default:
			return retErr(fmt.Errorf("%w: cli option '%s' unknown at '%s'",
				ErrInput, arg,
				strings.Join(args[:i], " ")))
		}
	}

	fn := fmt.Sprintf("%s.%s", filename, Extension)

	b, err := json.Marshal(inst)
	if err != nil {
		return retErr(err)
	}

	clifs := memfs.New()
	err = clifs.WriteFile(fn, b, 0755)
	if err != nil {
		return retErr(err)
	}

	group := goschtalt.Group{
		FS:    clifs,
		Paths: []string{"."},
	}
	options = append(options, goschtalt.AddFileGroup(group))

	return options
}
