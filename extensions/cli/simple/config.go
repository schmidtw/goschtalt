// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

// simple package is a simple but opinionated package that combines common
// chores for server startup into a single package.  It is still highly flexible
// based on the application's needs and keeps the dependencies as low as possible.
//
//   - processing command line input
//   - providing a default configuration
//   - providing a way for users to discover configuration documentation
//   - providing a place for storing build details
//   - unopinionated about decoders ... you must specify your own
//   - opinionated about encoders ... you will get a yaml encoder
//
// See the example for how to use this package in your code.
package simple

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/psanford/memfs"
	"github.com/schmidtw/goschtalt"
	"github.com/schmidtw/goschtalt/extensions/decoders/cli"
	_ "github.com/schmidtw/goschtalt/extensions/encoders/yaml"
)

const (
	showCfg = 1 << iota
	showCfgDoc
	showExts
	showFiles
)

// Use the -ldflags commandline to set the values of these three values at build
// time so they can be produced upon request from the command line interface upon
// request.
//
// Example:
//
//	go build -ldflags "\
//		-X 'simple.BuildTime=$(date --rfc-3339=seconds)' \
//		-X 'simple.Version=v1.2.3' \
//		-X 'simple.GitCommit=abc1234'"
var (
	GitCommit = "undefined" // The git commit hash of the source the program was built from.
	Version   = "undefined" // The version of the program.
	BuildTime = "undefined" // The time the program was built.
)

var (
	ErrDefaultConfigInvalid = errors.New("default config invalid")
)

// DefaultConfig defines the built in configuration document and extension.
type DefaultConfig struct {
	Text string // The text of the default configuration.
	Ext  string // The extension of the default configuration.
}

// GetConfig takes the application name, a default configuration text that is built
// in, the extension type of the configuration file and the goschtalt options to use for
// handling the configuration for a simple server interface.  This call collects
// the command line arguments and produces a goschtalt.Config object representing
// what the caller has specified.
//
// Notes:
//   - The caller must ensure there is a decoder available for the
//     config data passed in.  For example, if the config data is yaml format, make
//     sure there is a yaml decoder specified.
//   - A fully documented (with comments describing the configuration file options)
//     config file is a great way to provide documentation to your users.
//   - If a nil configuration is returned, that means the program should exit.  The
//     error value indicates if an error occured or a graceful exit should happen.
func GetConfig(name string, cfg DefaultConfig, opts ...goschtalt.Option) (*goschtalt.Config, error) {
	return getConfig(name, cfg, os.Args[1:], os.Stderr, opts...)
}

func getConfig(name string, cfg DefaultConfig, args []string, w io.Writer, opts ...goschtalt.Option) (*goschtalt.Config, error) {
	var extra []string
	var show int

	if len(cfg.Text) == 0 || len(cfg.Ext) == 0 {
		return nil, fmt.Errorf("%w: default config text must not be empty", ErrDefaultConfigInvalid)
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-h", "--help":
			fmt.Fprintf(w, "Usage: %s [OPTION]...\n", name)
			fmt.Fprintf(w, "\n")
			fmt.Fprintf(w, "  -f, --file file       File to process for configuration.  May be repeated.\n")
			fmt.Fprintf(w, "  -d, --dir dir         Directory to walk for configuration.  Does not recurse.  May be repeated.\n")
			fmt.Fprintf(w, "  -r, --recurse dir     Recursively walk directory for configuration.  May be repeated.\n")
			fmt.Fprintf(w, "      --kvp key value   Set a key/value pair as configuration.  May be repeated.  Filename: '1000.cli'\n")
			fmt.Fprintf(w, "\n")
			fmt.Fprintf(w, "  -s, --show-all        Show all configuration details, then exit.\n")
			fmt.Fprintf(w, "      --show-cfg        Show the redacted final configuration including the origin of the value, then exit.\n")
			fmt.Fprintf(w, "      --show-cfg-doc    Show built in config file with all options documentated.  Filename: '0.%s'\n", cfg.Ext)
			fmt.Fprintf(w, "      --show-exts       Show the supported configuration file extensions, then exit.\n")
			fmt.Fprintf(w, "      --show-files      Show files in the order processed, then exit.\n")
			fmt.Fprintf(w, "\n")
			fmt.Fprintf(w, "  -v, --version         Print the version information, then exit.\n")
			fmt.Fprintf(w, "  -h, --help            Output this text, then exit.\n")
			return nil, nil

		case "-s", "--show-all":
			show |= showCfg | showCfgDoc | showExts | showFiles
		case "--show-cfg":
			show |= showCfg
		case "--show-cfg-doc":
			show |= showCfgDoc
		case "--show-exts":
			show |= showExts
		case "--show-files":
			show |= showFiles

		case "-v", "--version":
			fmt.Fprintf(w, "%s:\n", name)
			fmt.Fprintf(w, "  version:    %s\n", Version)
			fmt.Fprintf(w, "  built time: %s\n", BuildTime)
			fmt.Fprintf(w, "  git commit: %s\n", GitCommit)
			fmt.Fprintf(w, "  go version: %s\n", runtime.Version())
			fmt.Fprintf(w, "  os/arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
			return nil, nil

		// The following are handled by the cli decoder:
		//	-f, --file
		//  -d, --dir
		//  -r, --recurse
		//      --kvp
		default:
			extra = append(extra, args[i])
		}
	}

	// Create the FS & FileGroup for the default configuration.
	defaultsFS := memfs.New()
	err := defaultsFS.WriteFile(fmt.Sprintf("0.%s", cfg.Ext), []byte(cfg.Text), 0755)
	if err != nil {
		return nil, fmt.Errorf("%w: probably an invalid extension: '%s' ... %v",
			ErrDefaultConfigInvalid, cfg.Ext, err)
	}
	defaultGroup := goschtalt.FileGroup(goschtalt.Group{
		FS:    defaultsFS,
		Paths: []string{"."},
	})

	// Append the specified options after these options in case they really
	// want to overwrite something.
	var allOpts []goschtalt.Option

	allOpts = append(allOpts, defaultGroup)
	allOpts = append(allOpts, cli.Options("1000", ".", extra)...)
	allOpts = append(allOpts, opts...)

	g, err := goschtalt.New(allOpts...)
	if err != nil {
		return nil, err
	}

	var found bool
	exts := g.Extensions()
	for _, ext := range exts {
		if cfg.Ext == ext {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("%w: extension '%s' not supported: '%s'",
			ErrDefaultConfigInvalid, cfg.Ext, strings.Join(exts, "', '"))
	}

	err = g.Compile()
	if err != nil {
		return nil, err
	}

	if show&showExts != 0 {
		fmt.Fprintf(w, "Supported File Extensions:\n\t'%s'\n\n",
			strings.Join(exts, "', '"))
	}

	if show&showFiles != 0 {
		files, err := g.ShowOrder()
		if err != nil {
			return nil, err
		}
		fmt.Fprintf(w, "Files Processed first (top) to last (bottom):\n")
		for i, file := range files {
			fmt.Fprintf(w, "\t%d. %s\n", i+1, file)
		}
		fmt.Fprintf(w, "\n")
	}

	if show&showCfgDoc != 0 {
		if show == showCfgDoc {
			fmt.Fprintln(w, cfg.Text)
		} else {
			upper := "-- vvv -------------------------------------------------------------------------\n"
			lower := "-- ^^^ -------------------------------------------------------------------------\n\n"
			fmt.Fprintf(w, "Default Configuration:\n%s%s\n%s", upper, cfg.Text, lower)
		}
	}

	if show&showCfg != 0 {
		b, err := g.Marshal(goschtalt.IncludeOrigins(true), goschtalt.RedactSecrets(true))
		if err != nil {
			return nil, err
		}
		if show == showCfg {
			fmt.Fprintln(w, string(b))
		} else {
			upper := "-- vvv -------------------------------------------------------------------------\n"
			lower := "-- ^^^ -------------------------------------------------------------------------\n\n"
			fmt.Fprintf(w, "Unified Configuration:\n%s%s\n%s", upper, string(b), lower)
		}
	}

	// If we showed anything stop processing.
	if show != 0 {
		return nil, nil
	}

	return g, nil
}
