// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/goschtalt/goschtalt/internal/natsort"
	"github.com/goschtalt/goschtalt/internal/print"
	"github.com/goschtalt/goschtalt/pkg/decoder"
	"github.com/goschtalt/goschtalt/pkg/encoder"
)

// Option configures specific behavior of Config as well as the locations used
// for the configurations compiled.  There are 3 basic groups of options:
//
//   - Configuration of locations where to collect configuration
//   - Addition of encoders/decoders
//   - Define default behaviors
type Option interface {
	fmt.Stringer

	// apply applies the changes to the options structure and returns an error
	// if one occurs.
	//
	// Generally options should try to always succeed unless they are is no other
	// way, or they are a validator option.  Validator options are a handy way
	// to easily ensure accuracy or exit without alot of code.
	apply(*options) error

	// ignoreDefaults returns false for all options that do not cause the default
	// options to be ignored.  This function lets us alter the options used
	// without the need to run the options twice to see if the defaults are
	// ignored.
	ignoreDefaults() bool
}

type options struct {
	// Settings where there are one.
	autoCompile  bool
	keyDelimiter string
	keySwizzler  func(string) string
	sorter       func(a, b string) bool

	// Codecs where there can be many.
	decoders *codecRegistry[decoder.Decoder]
	encoders *codecRegistry[encoder.Encoder]

	// Behaviors where there can be many.
	marshalOptions   []MarshalOption
	unmarshalOptions []UnmarshalOption
	valueOptions     []ValueOption

	// Defaults where there can be many.
	defaults []record

	// General configurations; there can be many.
	filegroups []filegroup
	values     []record

	// Expansions; there can be many.
	expansions []expand
}

// ---- Options follow ---------------------------------------------------------

// AddFile adds exactly one file to the list of files to be compiled into a
// configuration.  The filename must be relative to the fs.  If the file
// specified cannot be processed it is considered an error.
func AddFile(fs fs.FS, filename string) Option {
	return &groupOption{
		name: "AddFile",
		grp: filegroup{
			fs:        fs,
			paths:     []string{filename},
			exactFile: true,
		},
	}
}

// AddFiles adds any number of files to the list of files to be compiled into a
// configuration.  The filenames must be relative to the fs.  Any files that
// cannot be processed will be ignored.  It is not an error if any files are
// missing or if all the files cannot be processed.
//
// Use [AddFile]() if you need to require a file to be present.
//
// All the files that can be processed with a decoder will be compiled into the
// configuration.
func AddFiles(fs fs.FS, filenames ...string) Option {
	return &groupOption{
		name: "AddFiles",
		grp: filegroup{
			fs:    fs,
			paths: filenames,
		},
	}
}

// AddTree adds a directory tree (including all subdirectories) for inclusion
// when compiling the configuration.  Any files that cannot be processed will be
// ignored.  It is not an error if any files are missing, or if all the files
// cannot be processed.
//
// Use [AddFile]() if you need to require a file to be present.
//
// All the files that can be processed with a decoder will be compiled into the
// configuration.
func AddTree(fs fs.FS, path string) Option {
	return &groupOption{
		name: "AddTree",
		grp: filegroup{
			fs:      fs,
			paths:   []string{path},
			recurse: true,
		},
	}
}

// AddTrees adds a list of directory trees (including all subdirectories) for
// inclusion when compiling the configuration.  Any files that cannot be
// processed will be ignored.  It is not an error if any files are missing, or
// if all the files cannot be processed.
//
// Use [AddFile]() if you need to require a file to be present.
//
// All the files that can be processed with a decoder will be compiled into the
// configuration.
func AddTrees(fs fs.FS, paths ...string) Option {
	return &groupOption{
		name: "AddTrees",
		grp: filegroup{
			fs:      fs,
			paths:   paths,
			recurse: true,
		},
	}
}

// AddDir adds a directory (excluding all subdirectories) for inclusion
// when compiling the configuration.  Any files that cannot be processed will be
// ignored.  It is not an error if any files are missing, or if all the files
// cannot be processed.
//
// Use [AddFile]() if you need to require a file to be present.
//
// All the files that can be processed with a decoder will be compiled into the
// configuration.
func AddDir(fs fs.FS, path string) Option {
	return &groupOption{
		name: "AddDir",
		grp: filegroup{
			fs:    fs,
			paths: []string{path},
		},
	}
}

// AddDirs adds a list of directories (excluding all subdirectories) for inclusion
// when compiling the configuration.  Any files that cannot be processed will be
// ignored.  It is not an error if any files are missing, or if all the files
// cannot be processed.
//
// Use [AddFile]() if you need to require a file to be present.
//
// All the files that can be processed with a decoder will be compiled into the
// configuration.
func AddDirs(fs fs.FS, paths ...string) Option {
	return &groupOption{
		name: "AddDirs",
		grp: filegroup{
			fs:    fs,
			paths: paths,
		},
	}
}

// AddJumbledFiles adds any number of files or directories (excluding all
// subdirectories) for inclusion when compiling the configuration.  The files
// and directories are sorted into either a relative based filesystem or an
// absolute path based filesystem.  Any files or directories that cannot be
// processed will be ignored.  It is not an error if any files are missing or if
// all the files cannot be processed.
//
// Use [AddFile]() if you need to require a file to be present.
//
// Generally this option is useful when processing files from the same filesystem
// but some are absolute path based and others are relative path based.  Instead
// of needing to sort the files into two buckets, this option will handle that
// for you.
func AddJumbled(abs, rel fs.FS, paths ...string) Option {
	absPaths := make([]string, 0, len(paths))
	relPaths := make([]string, 0, len(paths))

	for _, path := range paths {
		if filepath.IsAbs(path) {
			absPaths = append(absPaths, path)
		} else {
			relPaths = append(relPaths, path)
		}
	}

	return &optionsOption{
		text: print.P("AddJumbled",
			print.Literal("abs"),
			print.Literal("rel"),
			print.Strings(paths),
		),
		opts: []Option{
			&groupOption{
				grp: filegroup{
					fs:    abs,
					paths: absPaths,
				},
			},
			&groupOption{
				grp: filegroup{
					fs:    rel,
					paths: relPaths,
				},
			},
		},
	}
}

type groupOption struct {
	name string
	grp  filegroup
}

var _ Option = (*groupOption)(nil)

func (g groupOption) apply(opts *options) error {
	opts.filegroups = append(opts.filegroups, g.grp)
	return nil
}

func (_ groupOption) ignoreDefaults() bool {
	return false
}

func (o groupOption) String() string {
	return print.P(o.name, print.Literal("fs"), print.Strings(o.grp.paths))
}

// AutoCompile instructs [New]() and [With]() to also compile the configuration
// after all the options are applied if enable is true or omitted.  Passing
// an enable value of false disables the extra behavior.
//
// The enable bool value is optional & assumed to be `true` if omitted.  The
// first specified value is used if provided.  A value of `false` disables the
// option.
//
// # Default
//
// AutoCompile is not enabled.
func AutoCompile(enable ...bool) Option {
	enable = append(enable, true)
	return autoCompileOption(enable[0])
}

type autoCompileOption bool

func (a autoCompileOption) apply(opts *options) error {
	opts.autoCompile = bool(a)
	return nil
}

func (_ autoCompileOption) ignoreDefaults() bool { return false }
func (a autoCompileOption) String() string {
	return print.P("AutoCompile", print.BoolSilentTrue(bool(a)))
}

// AlterKeyCase defines how the keys should be altered prior to use.  This
// option enables enforcing key case to be all upper case, all lower case,
// no change or whatever is needed.
//
// Passing nil alter value is interpreted as "do not alter" the key case and
// is the same as passing:
//
//	func(s string) string { return s }
//
// Examples:
//
//	AlterKeyCase(strings.ToLower)
//	AlterKeyCase(strings.ToUpper)
//
// # Default
//
// AlterKeyCase is set to "do not alter".
func AlterKeyCase(alter func(string) string) Option {
	return alterKeyCaseOption(alter)
}

type alterKeyCaseOption func(string) string

func (alter alterKeyCaseOption) apply(opts *options) error {
	if alter == nil {
		alter = func(s string) string { return s }
	}
	opts.keySwizzler = alter
	return nil
}

func (_ alterKeyCaseOption) ignoreDefaults() bool {
	return false
}

func (a alterKeyCaseOption) String() string {
	return print.P("AlterKeyCase", print.FnAltNil(a, "none"))
}

// SetKeyDelimiter provides the delimiter used for determining key parts.  A
// string with length of at least 1 must be provided.
//
// # Default
//
// The default value is '.'.
func SetKeyDelimiter(delimiter string) Option {
	return setKeyDelimiterOption(delimiter)
}

type setKeyDelimiterOption string

func (s setKeyDelimiterOption) apply(opts *options) error {
	if len(s) == 0 {
		return fmt.Errorf("%w: a KeyDelimiter with length > 0 must be specified.", ErrInvalidInput)
	}

	opts.keyDelimiter = string(s)
	return nil
}

func (_ setKeyDelimiterOption) ignoreDefaults() bool {
	return false
}

func (s setKeyDelimiterOption) String() string {
	return print.P("SetKeyDelimiter", print.String(string(s)))
}

// SortRecordsCustomFn provides a way to specify how you want the files sorted
// prior to their merge.  This function provides a way to provide a completely
// custom sorting algorithm.
//
// See also: [SortRecordsLexically], [SortRecordsNaturally]
//
// # Default
//
// The default is [SortRecordsNaturally].
func SortRecordsCustomFn(less func(a, b string) bool) Option {
	return &sortRecordsCustomFnOption{
		text: print.P("SortRecordsCustomFn", print.Fn(less)),
		fn:   less,
	}
}

// SortRecordsLexically provides a built in sorter based on lexical order.
//
// See also: [SortRecordsCustomFn], [SortRecordsNaturally]
//
// # Default
//
// The default is [SortRecordsNaturally].
func SortRecordsLexically() Option {
	return &sortRecordsCustomFnOption{
		text: print.P("SortRecordsLexically"),
		fn: func(a, b string) bool {
			return a < b
		},
	}
}

// SortRecordsNaturally provides a built in sorter based on natural order.
// More information about natural sort order: https://en.wikipedia.org/wiki/Natural_sort_order
//
// Notes:
//
//   - Don't use floating point numbers.  They are treated like 2 integers separated
//     by the '.' rune.
//   - Any leading 0 values are dropped from the number.
//
// Example sort order:
//
//	01_foo.yml
//	2_foo.yml
//	98_foo.yml
//	99 dogs.yml
//	99_Abc.yml
//	99_cli.yml
//	99_mine.yml
//	100_alpha.yml
//
// See also: [SortRecordsCustomFn], [SortRecordsLexically]
//
// # Default
//
// The default is [SortRecordsNaturally].
func SortRecordsNaturally() Option {
	return &sortRecordsCustomFnOption{
		text: print.P("SortRecordsNaturally"),
		fn:   natsort.Compare,
	}
}

type sortRecordsCustomFnOption struct {
	text string
	fn   func(a, b string) bool
}

func (s sortRecordsCustomFnOption) apply(opts *options) error {
	if s.fn == nil {
		return fmt.Errorf("%w: a SortRecords function/option must be specified", ErrInvalidInput)
	}

	opts.sorter = s.fn
	return nil
}

func (_ sortRecordsCustomFnOption) ignoreDefaults() bool { return false }
func (s sortRecordsCustomFnOption) String() string       { return s.text }

// WithDecoder registers a Decoder for the specific file extensions provided.
// Attempting to register a duplicate extension is not supported.
//
// See also: [WithEncoder]
func WithDecoder(d decoder.Decoder) Option {
	return &withDecoderOption{decoder: d}
}

type withDecoderOption struct {
	decoder decoder.Decoder
}

func (w withDecoderOption) apply(opts *options) error {
	if w.decoder != nil {
		opts.decoders.register(w.decoder)
	}
	return nil
}

func (_ withDecoderOption) ignoreDefaults() bool {
	return false
}

func (w withDecoderOption) String() string {
	var s []string
	if w.decoder != nil {
		s = w.decoder.Extensions()
	}

	return print.P("WithDecoder", print.Strings(s))
}

// WithEncoder registers a Encoder for the specific file extensions provided.
// Attempting to register a duplicate extension is not supported.
//
// See also: [WithDecoder]
func WithEncoder(enc encoder.Encoder) Option {
	return &withEncoderOption{enc: enc}
}

type withEncoderOption struct {
	enc encoder.Encoder
}

func (w withEncoderOption) apply(opts *options) error {
	if w.enc != nil {
		opts.encoders.register(w.enc)
	}
	return nil
}

func (_ withEncoderOption) ignoreDefaults() bool {
	return false
}

func (w withEncoderOption) String() string {
	var s []string
	if w.enc != nil {
		s = w.enc.Extensions()
	}

	return print.P("WithEncoder", print.Strings(s))
}

// DisableDefaultPackageOptions provides a way to explicitly not use any preconfigured
// default values by this package and instead use just the options specified.
//
// See: [DefaultOptions]
func DisableDefaultPackageOptions() Option {
	return disableDefaultPackageOption{}
}

type disableDefaultPackageOption struct{}

func (_ disableDefaultPackageOption) apply(opts *options) error { return nil }
func (_ disableDefaultPackageOption) ignoreDefaults() bool      { return true }
func (_ disableDefaultPackageOption) String() string {
	return print.P("DisableDefaultPackageOptions")
}

// DefaultMarshalOptions allows customization of the desired options for all
// invocations of the [Marshal]() function.  This should make consistent use
// use of the Marshal() call easier.
//
// Valid Option Types:
//   - [GlobalOption]
//   - [MarshalOption]
func DefaultMarshalOptions(opts ...MarshalOption) Option {
	return &defaultMarshalOption{opts: opts}
}

type defaultMarshalOption struct {
	opts []MarshalOption
}

func (d defaultMarshalOption) apply(opts *options) error {
	opts.marshalOptions = append(opts.marshalOptions, d.opts...)
	return nil
}

func (_ defaultMarshalOption) ignoreDefaults() bool {
	return false
}

func (d defaultMarshalOption) String() string {
	return print.P("DefaultMarshalOptions", print.LiteralStringers(d.opts))
}

// DefaultUnmarshalOptions allows customization of the desired options for all
// invocations of the [Unmarshal]() function.  This should make consistent use
// use of the Unmarshal() call easier.
//
// Valid Option Types:
//   - [GlobalOption]
//   - [UnmarshalOption]
//   - [UnmarshalValueOption]
func DefaultUnmarshalOptions(opts ...UnmarshalOption) Option {
	return &defaultUnmarshalOption{opts: opts}
}

type defaultUnmarshalOption struct {
	opts []UnmarshalOption
}

func (d defaultUnmarshalOption) apply(opts *options) error {
	opts.unmarshalOptions = append(opts.unmarshalOptions, d.opts...)
	return nil
}

func (_ defaultUnmarshalOption) ignoreDefaults() bool {
	return false
}

func (d defaultUnmarshalOption) String() string {
	return print.P("DefaultUnmarshalOptions", print.LiteralStringers(d.opts))
}

// DefaultValueOptions allows customization of the desired options for all
// invocations of the [AddValue]() and [AddValueFn]() functions.  This should
// make consistent use use of these functions easier.
//
// Valid Option Types:
//   - [BufferValueOption]
//   - [GlobalOption]
//   - [ValueOption]
//   - [UnmarshalValueOption]
func DefaultValueOptions(opts ...ValueOption) Option {
	return &defaultValueOption{opts: opts}
}

type defaultValueOption struct {
	opts []ValueOption
}

func (d defaultValueOption) apply(opts *options) error {
	opts.valueOptions = append(opts.valueOptions, d.opts...)
	return nil
}

func (_ defaultValueOption) ignoreDefaults() bool {
	return false
}

func (d defaultValueOption) String() string {
	return print.P("DefaultValueOptions", print.LiteralStringers(d.opts))
}

// Options provides a way to group multiple options together into an easy to
// use Option.
func Options(opts ...Option) Option {
	return &optionsOption{
		text: print.P("Options", print.LiteralStringers(opts)),
		opts: opts,
	}
}

// optionsOption allows for returning an option that is actually several
// sub options when needed, without needing to return []Option everywhere.
type optionsOption struct {
	text string
	opts []Option
}

func (m optionsOption) apply(in *options) error {
	for _, opt := range m.opts {
		err := opt.apply(in)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m optionsOption) ignoreDefaults() bool {
	for _, opt := range m.opts {
		if opt.ignoreDefaults() {
			return true
		}
	}
	return false
}

func (m optionsOption) String() string {
	return m.text
}

// ---- Options related helper functions follow --------------------------------

func ignoreDefaultOpts(opts []Option) bool {
	for _, opt := range opts {
		if opt != nil && opt.ignoreDefaults() {
			return true
		}
	}
	return false
}
