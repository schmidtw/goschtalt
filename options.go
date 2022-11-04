// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/schmidtw/goschtalt/internal/natsort"
	"github.com/schmidtw/goschtalt/pkg/decoder"
	"github.com/schmidtw/goschtalt/pkg/encoder"
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
	autoCompile      bool
	keyDelimiter     string
	keySwizzler      func(string) string
	sorter           func(a, b string) bool
	decoders         *codecRegistry[decoder.Decoder]
	encoders         *codecRegistry[encoder.Encoder]
	marshalOptions   []MarshalOption
	unmarshalOptions []UnmarshalOption
	valueOptions     []DecoderConfigOption
	groups           []group
	values           []value
	expansions       []expand
}

// ---- Options follow ---------------------------------------------------------

// WithError provides a way for plugins to return an error during option
// processing.
func WithError(err error) Option {
	return errorOption{err: err}
}

type errorOption struct {
	err error
}

var _ Option = (*errorOption)(nil)

func (opt errorOption) apply(_ *options) error {
	return opt.err
}

func (_ errorOption) ignoreDefaults() bool {
	return false
}

func (o errorOption) String() string {
	if o.err == nil {
		return "WithError( '' )"
	}

	return fmt.Sprintf("WithError( '%v' )", o.err)
}

// AddFile adds exactly one file to the list of files to be compiled into a
// configuration.  The filename must be relative to the fs.
func AddFile(fs fs.FS, filename string) Option {
	return &groupOption{
		name: "AddFile",
		grp: group{
			fs:    fs,
			paths: []string{filename},
		},
	}
}

// AddFiles adds any number of files to the list of files to be compiled into a
// configuration.  The filenames must be relative to the fs.
//
// All the files that can be processed with a decoder will be compiled into the
// configuration.
func AddFiles(fs fs.FS, filenames ...string) Option {
	return &groupOption{
		name: "AddFiles",
		grp: group{
			fs:    fs,
			paths: filenames,
		},
	}
}

// AddTree adds a directory tree (including all subdirectories) for inclusion
// when compiling the configuration.
//
// All the files that can be processed with a decoder will be compiled into the
// configuration.
func AddTree(fs fs.FS, path string) Option {
	return &groupOption{
		name: "AddTree",
		grp: group{
			fs:      fs,
			paths:   []string{path},
			recurse: true,
		},
	}
}

// AddDir adds a directory (excluding all subdirectories) for inclusion
// when compiling the configuration.
//
// All the files that can be processed with a decoder will be compiled into the
// configuration.
func AddDir(fs fs.FS, path string) Option {
	return &groupOption{
		name: "AddDir",
		grp: group{
			fs:    fs,
			paths: []string{path},
		},
	}
}

type groupOption struct {
	name string
	grp  group
}

var _ Option = (*groupOption)(nil)

func (g groupOption) apply(opts *options) error {
	opts.groups = append(opts.groups, g.grp)
	return nil
}

func (_ groupOption) ignoreDefaults() bool {
	return false
}

func (o groupOption) String() string {
	return o.name + "( '" + strings.Join(o.grp.paths, "', '") + "' )"
}

// AutoCompile instructs New() and With() to also compile the configuration
// after all the options are applied if enable is true or omitted.  Passing
// an enable value of false disables the extra behavior.
func AutoCompile(enable ...bool) Option {
	enable = append(enable, true)
	return autoCompileOption(enable[0])
}

type autoCompileOption bool

var _ Option = (*autoCompileOption)(nil)

func (a autoCompileOption) apply(opts *options) error {
	opts.autoCompile = bool(a)
	return nil
}

func (_ autoCompileOption) ignoreDefaults() bool { return false }
func (a autoCompileOption) String() string {
	if bool(a) {
		return "AutoCompile()"
	}

	return "AutoCompile( false )"
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
func AlterKeyCase(alter func(string) string) Option {
	return alterKeyCaseOption(alter)
}

type alterKeyCaseOption func(string) string

var _ Option = (*alterKeyCaseOption)(nil)

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
	if a == nil {
		return "AlterKeyCase( none )"
	}

	return "AlterKeyCase( custom )"
}

// SetKeyDelimiter provides the delimiter used for determining key parts.  A
// string with length of at least 1 must be provided.  The default value is '.'.
func SetKeyDelimiter(delimiter string) Option {
	return setKeyDelimiterOption(delimiter)
}

type setKeyDelimiterOption string

var _ Option = (*setKeyDelimiterOption)(nil)

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
	return fmt.Sprintf("SetKeyDelimiter( '%s' )", string(s))
}

// SortRecordsCustomFn provides a way to specify how you want the files sorted
// prior to their merge.  This function provides a way to provide a completely
// custom sorting algorithm.
//
// The default is SortRecordsNaturally.
//
// See also: [SortRecordsLexically], [SortRecordsNaturally]
func SortRecordsCustomFn(less func(a, b string) bool) Option {
	return &sortRecordsCustomFnOption{name: "SortRecordsCustomFn( custom )", fn: less}
}

// SortRecordsLexically provides a built in sorter based on lexical order.
//
// The default is SortRecordsNaturally.
//
// See also: [SortRecordsCustomFn], [SortRecordsNaturally]
func SortRecordsLexically() Option {
	return &sortRecordsCustomFnOption{
		name: "SortRecordsLexically()",
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
// The default is SortRecordsNaturally.
//
// See also: [SortRecordsCustomFn], [SortRecordsLexically]
func SortRecordsNaturally() Option {
	return &sortRecordsCustomFnOption{
		name: "SortRecordsNaturally()",
		fn:   natsort.Compare,
	}
}

type sortRecordsCustomFnOption struct {
	name string
	fn   func(a, b string) bool
}

var _ Option = (*sortRecordsCustomFnOption)(nil)

func (s sortRecordsCustomFnOption) apply(opts *options) error {
	if s.fn == nil {
		return fmt.Errorf("%w: a SortRecords function/option must be specified", ErrInvalidInput)
	}

	opts.sorter = s.fn
	return nil
}

func (_ sortRecordsCustomFnOption) ignoreDefaults() bool { return false }
func (s sortRecordsCustomFnOption) String() string       { return s.name }

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

var _ Option = (*withDecoderOption)(nil)

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
	if w.decoder == nil {
		return "WithDecoder( '' )"
	}

	return "WithDecoder( '" + strings.Join(w.decoder.Extensions(), "', '") + "' )"
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

var _ Option = (*withEncoderOption)(nil)

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
	if w.enc == nil {
		return "WithEncoder( '' )"
	}

	return "WithEncoder( '" + strings.Join(w.enc.Extensions(), "', '") + "' )"
}

// DisableDefaultPackageOptions provides a way to explicitly not use any preconfigured
// default values by this package and instead use just the options specified.
func DisableDefaultPackageOptions() Option {
	return disableDefaultPackageOption{}
}

type disableDefaultPackageOption struct{}

var _ Option = (*disableDefaultPackageOption)(nil)

func (_ disableDefaultPackageOption) apply(opts *options) error { return nil }
func (_ disableDefaultPackageOption) ignoreDefaults() bool      { return true }
func (_ disableDefaultPackageOption) String() string            { return "DisableDefaultPackageOptions()" }

// DefaultMarshalOptions allows customization of the desired options for all
// invocations of the Marshal() function.  This should make consistent use
// use of the Marshal() call easier.
func DefaultMarshalOptions(opts ...MarshalOption) Option {
	return &defaultMarshalOption{opts: opts}
}

type defaultMarshalOption struct {
	opts []MarshalOption
}

var _ Option = (*defaultMarshalOption)(nil)

func (d defaultMarshalOption) apply(opts *options) error {
	opts.marshalOptions = append(opts.marshalOptions, d.opts...)
	return nil
}

func (_ defaultMarshalOption) ignoreDefaults() bool {
	return false
}

func (d defaultMarshalOption) String() string {
	if len(d.opts) == 0 {
		return "DefaultMarshalOptions()"
	}

	var s []string
	for _, opt := range d.opts {
		s = append(s, opt.String())
	}
	return "DefaultMarshalOptions( " + strings.Join(s, ", ") + " )"
}

// DefaultUnmarshalOptions allows customization of the desired options for all
// invocations of the Unmarshal() function.  This should make consistent use
// use of the Unmarshal() call easier.
func DefaultUnmarshalOptions(opts ...UnmarshalOption) Option {
	return &defaultUnmarshalOption{opts: opts}
}

type defaultUnmarshalOption struct {
	opts []UnmarshalOption
}

var _ Option = (*defaultUnmarshalOption)(nil)

func (d defaultUnmarshalOption) apply(opts *options) error {
	opts.unmarshalOptions = append(opts.unmarshalOptions, d.opts...)
	return nil
}

func (_ defaultUnmarshalOption) ignoreDefaults() bool {
	return false
}

func (d defaultUnmarshalOption) String() string {
	if len(d.opts) == 0 {
		return "DefaultUnmarshalOptions()"
	}

	var s []string
	for _, opt := range d.opts {
		s = append(s, opt.String())
	}
	return "DefaultUnmarshalOptions( " + strings.Join(s, ", ") + " )"
}

// DefaultValueOptions allows customization of the desired options for all
// invocations of the TODO:() function.  This should make consistent use
// use of the TODO:() call easier.
func DefaultValueOptions(opts ...DecoderConfigOption) Option {
	return &defaultValueOption{opts: opts}
}

type defaultValueOption struct {
	opts []DecoderConfigOption
}

var _ Option = (*defaultValueOption)(nil)

func (d defaultValueOption) apply(opts *options) error {
	opts.valueOptions = append(opts.valueOptions, d.opts...)
	return nil
}

func (_ defaultValueOption) ignoreDefaults() bool {
	return false
}

func (d defaultValueOption) String() string {
	if len(d.opts) == 0 {
		return "DefaultValueOptions()"
	}

	var s []string
	for i := range d.opts {
		s = append(s, d.opts[i].String())
	}
	return "DefaultValueOptions( " + strings.Join(s, ", ") + " )"
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
