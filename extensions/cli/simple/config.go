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
	"reflect"
	"runtime"
	"strings"

	"github.com/psanford/memfs"
	"github.com/schmidtw/goschtalt"
	"github.com/schmidtw/goschtalt/extensions/decoders/cli"
	_ "github.com/schmidtw/goschtalt/extensions/encoders/yaml"
)

const (
	cmdShowCfg = 1 << iota
	cmdShowCfgDoc
	cmdShowCfgUnsafe
	cmdShowExts
	cmdShowFiles
	cmdSkipValidation
	cmdExit
	cmdExitNow
)

const (
	filenameDefault = "000"
	filenameEnviron = "800"
	filenameCLI     = "900"
)

// Use the -ldflags command line option to set the values of these three values
// at build time.  These values are produced upon request from the command line
// interface upon request.
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

func (d DefaultConfig) getFileGroup() (goschtalt.Option, error) {
	if len(d.Text) == 0 || len(d.Ext) == 0 {
		return nil, fmt.Errorf("%w: default configuration text must not be empty", ErrDefaultConfigInvalid)
	}

	fn := fmt.Sprintf("%s.%s", filenameDefault, d.Ext)

	fs := memfs.New()
	err := fs.WriteFile(fn, []byte(d.Text), 0755)
	if err != nil {
		return nil, fmt.Errorf("%w: probably an invalid extension: '%s' ... %v",
			ErrDefaultConfigInvalid, d.Ext, err)
	}

	return goschtalt.FileGroup(goschtalt.Group{
		FS:    fs,
		Paths: []string{"."},
	}), nil
}

func (d DefaultConfig) wereDefaultsUsed(g *goschtalt.Config) error {
	var found bool

	exts := g.Extensions()
	for _, ext := range exts {
		if d.Ext == ext {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("%w: extension '%s' not supported: '%s'",
			ErrDefaultConfigInvalid, d.Ext, strings.Join(exts, "', '"))
	}

	files, err := g.ShowOrder()
	if err != nil {
		return err
	}

	fn := fmt.Sprintf("%s.%s", filenameDefault, d.Ext)
	for _, file := range files {
		if file == fn {
			return nil
		}
	}

	// This should never happen, but if it does, we at least have an error.
	return fmt.Errorf("%w: default configuration file '%s' was not included.",
		ErrDefaultConfigInvalid, fn)
}

type Program struct {
	Name      string        // Required - name of the application
	Default   DefaultConfig // Required - default configuration
	Prefix    string        // Optional - defaults to strings.ToUpper(Name)+"_" if empty
	Licensing string        // Optional - if you want to include licensing details and option
	Output    io.Writer     // Optional - defaults to os.Stderr

	// An optional map of label to struct that will be used to validate
	// the default configuration against.  The configuration is validated if
	// exactly the same number of fields in the struct matches the fields in
	// the configuration (honoring mapstruct instructions).  No more or no
	// fewer are permitted.  This validation will be done at program start
	// unless explicitly instructed to skip via the --skip-validation option
	// is passed.
	//
	// Why provide this map?  This helps ensure your configuration documentation
	// and your program are close, if not the same.  If there are areas that
	// does not fit into this model well, feel free to leave them out.
	Validate map[string]any
}

// GetConfig takes the assorted inputs and merges them into goschtalt.Config
// objet.
//
//   - name is the application name.
//   - prefix is the environment variable prefix to search for.
//   - cfg is the default configuration text and extension.
//   - opts are any options you would like to customize.
//
// Notes:
//   - The caller must ensure there is a decoder available for the
//     config data passed in.  For example, if the config data is yaml format, make
//     sure there is a yaml decoder specified.
//   - A fully documented (with comments describing the configuration file options)
//     config file is a great way to provide documentation to your users.
//   - If a nil configuration is returned, that means the program should exit.  The
//     error value indicates if an error occurred or a graceful exit should happen.
func (p Program) GetConfig(opts ...goschtalt.Option) (*goschtalt.Config, error) {
	// Mostly this function lets us test the rest of the code easier.
	return p.applyDefaults(os.Args[1:], opts...)
}

// applyDefaults applies the defaults for the program.  Mostly here to make testing
// easier.
func (p Program) applyDefaults(args []string, opts ...goschtalt.Option) (*goschtalt.Config, error) {
	if len(p.Prefix) == 0 {
		p.Prefix = strings.ToUpper(p.Name) + "_"
	}
	if p.Output == nil {
		p.Output = os.Stderr
	}

	return p.getConfig(args, opts...)
}

func (p Program) getConfig(args []string, opts ...goschtalt.Option) (*goschtalt.Config, error) {
	w := p.Output

	defFG, err := p.Default.getFileGroup()
	if err != nil {
		return nil, err
	}

	extra, cmd := p.processArgs(args, w)
	if cmd&cmdExitNow != 0 {
		return nil, nil
	}

	if err = validateDefault(cmd, defFG, p.Validate, opts...); err != nil {
		return nil, err
	}

	// Append the specified options after these options in case they really
	// want to overwrite something.
	var allOpts []goschtalt.Option
	allOpts = append(allOpts, defFG)
	allOpts = append(allOpts, cli.Options(filenameCLI, ".", extra)...)
	allOpts = append(allOpts, opts...)

	g, err := goschtalt.New(allOpts...)
	if err != nil {
		return nil, err
	}

	if err = g.Compile(); err != nil {
		return nil, err
	}

	if err = p.Default.wereDefaultsUsed(g); err != nil {
		return nil, err
	}

	showExtensions(cmd, w, g)
	if err = showFileList(cmd, w, g); err != nil {
		return nil, err
	}
	showCfgDoc(cmd, w, p.Default)
	if err = showCfg(cmd, w, g); err != nil {
		return nil, err
	}

	if cmd&cmdExit != 0 {
		return nil, nil
	}

	return g, nil
}

// processArgs handles the input arguments and groups them into a set of next
// step commands.  The extra return value are the args that goschtalt should
// process.  The cmd value indicates what should be shown/done.
func (p Program) processArgs(args []string, w io.Writer) (extra []string, cmd int) {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-h", "--help":
			extWidth := getExtWidth(p.Default.Ext)

			fmt.Fprintf(w, "Usage: %s [OPTION]...\n", p.Name)
			fmt.Fprintf(w, "\n")
			fmt.Fprintf(w, "  -f, --file file        File to process for configuration.  May be repeated.\n")
			fmt.Fprintf(w, "  -d, --dir dir          Directory to walk for configuration.  Does not recurse.  May be repeated.\n")
			fmt.Fprintf(w, "  -r, --recurse dir      Recursively walk directory for configuration.  May be repeated.\n")
			fmt.Fprintf(w, "      --kvp key value    Set a key/value pair as configuration.  May be repeated.\n")
			fmt.Fprintf(w, "\n")
			fmt.Fprintf(w, "  -s, --show-all         Show all configuration details, then exit.\n")
			fmt.Fprintf(w, "      --show-cfg         Show the redacted final configuration including the origin of the value, then exit.\n")
			fmt.Fprintf(w, "      --show-cfg-doc     Show built in config file with all options documented.\n")
			fmt.Fprintf(w, "      --show-cfg-unsafe  Show the non-redacted version of --show-cfg.  Not included in --show-all or -s.\n")
			fmt.Fprintf(w, "      --show-exts        Show the supported configuration file extensions, then exit.\n")
			fmt.Fprintf(w, "      --show-files       Show files in the order processed, then exit.\n")
			fmt.Fprintf(w, "\n")
			fmt.Fprintf(w, "      --skip-validation  Skips the validation of the default configuration against expected structs.\n")
			fmt.Fprintf(w, "  -l, --licensing        Show licensing details, then exit.\n")
			fmt.Fprintf(w, "  -v, --version          Print the version information, then exit.\n")
			fmt.Fprintf(w, "  -h, --help             Output this text, then exit.\n")
			fmt.Fprintf(w, "\n")
			fmt.Fprintf(w, "\n")
			fmt.Fprintf(w, "Automatic configuration files:\n")
			fmt.Fprintf(w, "\n")
			fmt.Fprintf(w, "  %s.%-*s   built in configuration document\n", filenameDefault, extWidth, p.Default.Ext)
			fmt.Fprintf(w, "  %s.%-*s   command line arguments\n", filenameCLI, extWidth, cli.Extension)
			return []string{}, cmdExitNow

		case "-l", "--licensing":
			if len(p.Licensing) == 0 {
				fmt.Fprintf(w, "Licensing information is unavailable.\n")
			} else {
				txt := ensureTrailingNewline(p.Licensing)
				fmt.Fprintf(w, "Licensing for %s:\n%s", p.Name, txt)
			}
			return []string{}, cmdExitNow

		case "-s", "--show-all":
			cmd |= cmdShowCfg | cmdShowCfgDoc | cmdShowExts | cmdShowFiles | cmdExit

		case "--show-cfg":
			cmd |= cmdShowCfg | cmdExit

		case "--show-cfg-doc":
			cmd |= cmdShowCfgDoc | cmdExit

		case "--show-exts":
			cmd |= cmdShowExts | cmdExit

		case "--show-files":
			cmd |= cmdShowFiles | cmdExit

		case "--show-cfg-unsafe":
			cmd |= cmdShowCfgUnsafe | cmdExit

		case "--skip-validation":
			cmd |= cmdSkipValidation

		case "-v", "--version":
			fmt.Fprintf(w, "%s:\n", p.Name)
			fmt.Fprintf(w, "  version:    %s\n", Version)
			fmt.Fprintf(w, "  built time: %s\n", BuildTime)
			fmt.Fprintf(w, "  git commit: %s\n", GitCommit)
			fmt.Fprintf(w, "  go version: %s\n", runtime.Version())
			fmt.Fprintf(w, "  os/arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
			return []string{}, cmdExitNow

		// The following are handled by the cli decoder:
		//	-f, --file
		//  -d, --dir
		//  -r, --recurse
		//      --kvp
		default:
			extra = append(extra, args[i])
		}
	}

	return extra, cmd
}

func validateDefault(cmd int, def goschtalt.Option, what map[string]any, optsIn ...goschtalt.Option) error {
	if cmd&cmdSkipValidation == 0 && len(what) > 0 {
		// create a single use goschtalt.Config to just validate the default.
		opts := []goschtalt.Option{def}
		opts = append(opts, cli.Options(filenameCLI, ".", []string{})...)
		opts = append(opts, optsIn...)

		g, err := goschtalt.New(opts...)
		if err != nil {
			return err
		}

		if err = g.Compile(); err != nil {
			return err
		}

		for k, v := range what {
			// Take either a pointer or direct object and ensure we have a
			// pointer to a unique object.  This prevents a partial configuration
			// value from being leaked back to the caller by accident.
			p := &v
			if reflect.ValueOf(v).Kind() == reflect.Pointer {
				// Get the type of the thing being pointed to
				t := reflect.TypeOf(v).Elem()
				// Make a new thing & convert it to an any so the object is right
				n := reflect.New(t).Interface()
				p = &n
			}

			if err = g.Unmarshal(k, p, goschtalt.ErrorUnused(true),
				goschtalt.ErrorUnset(true)); err != nil {
				return fmt.Errorf("%w: default configuration issue: %v",
					ErrDefaultConfigInvalid, err)
			}
		}
	}

	return nil
}

func showExtensions(cmd int, w io.Writer, g *goschtalt.Config) {
	if cmd&cmdShowExts != 0 {
		exts := g.Extensions()
		fmt.Fprintf(w, "Supported File Extensions:\n\t'%s'\n\n",
			strings.Join(exts, "', '"))
	}
}

func showFileList(cmd int, w io.Writer, g *goschtalt.Config) error {
	if cmd&cmdShowFiles != 0 {
		files, err := g.ShowOrder()
		if err != nil {
			return err
		}

		fmt.Fprintf(w, "Files Processed first (1) to last (%d):\n", len(files))
		for i, file := range files {
			fmt.Fprintf(w, "\t%d. %s\n", i+1, file)
		}
		fmt.Fprintf(w, "\n")
	}
	return nil
}

func showCfgDoc(cmd int, w io.Writer, cfg DefaultConfig) {
	cmd = cmd ^ cmdExit
	if cmd&cmdShowCfgDoc != 0 {
		upper := "-----BEGIN DEFAULT CONFIGURATION-----\n"
		lower := "-----END DEFAULT CONFIGURATION-----\n\n"
		if cmd == cmdShowCfgDoc {
			upper = ""
			lower = ""
		}
		txt := ensureTrailingNewline(cfg.Text)
		fmt.Fprintf(w, "%s%s%s", upper, txt, lower)
	}
}

func showCfg(cmd int, w io.Writer, g *goschtalt.Config) error {
	cmd = cmd ^ cmdExit
	if cmd&(cmdShowCfg|cmdShowCfgUnsafe) != 0 {
		// Only show the non-redacted version if it's the only item requested
		redact := goschtalt.RedactSecrets(true)
		if cmd == cmdShowCfgUnsafe {
			redact = nil
		}

		b, err := g.Marshal(goschtalt.IncludeOrigins(true), redact)
		if err != nil {
			return err
		}

		upper := "-----BEGIN REDACTED UNIFIED CONFIGURATION-----\n"
		lower := "-----END REDACTED UNIFIED CONFIGURATION-----\n\n"
		if cmd == cmdShowCfg || cmd == cmdShowCfgUnsafe {
			upper = ""
			lower = ""
		}
		txt := ensureTrailingNewline(string(b))
		fmt.Fprintf(w, "%s%s%s", upper, txt, lower)
	}

	return nil
}

// ensureTrailingNewline makes the last character in a string a '\n' without
// duplicating it if already present.  It will not alter what was there, so if
// '\n\n' was already present, it will stay that way.
func ensureTrailingNewline(s string) string {
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	return s
}

// getExtWidth finds the extension with the longest width and returns that width.
func getExtWidth(ext string) int {
	s := []string{cli.Extension, ext}

	rv := 0
	for _, val := range s {
		l := len(val)
		if rv < l {
			rv = l
		}
	}

	return rv
}
