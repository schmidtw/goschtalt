// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"fmt"

	"github.com/goschtalt/goschtalt/internal/print"
	"github.com/mitchellh/mapstructure"
)

// DecoderConfigOption is used to configure the mapstructure used for decoding
// structures into the values used by the internal goschtalt tree.
//
// All of these options are directly concerned with the mitchellh/mapstructure
// package.  For additional details please see: https://github.com/mitchellh/mapstructure
type DecoderConfigOption interface {
	fmt.Stringer

	UnmarshalOption
	ValueOption
}

// DecodeHook, will be called before any decoding and any type conversion (if
// WeaklyTypedInput is on). This lets you modify the values before they're set
// down onto the resulting struct. The DecodeHook is called for every map and
// value in the input. This means that if a struct has embedded fields with
// squash tags the decode hook is called only once with all of the input data,
// not once for each embedded struct.
//
// If an error is returned, the entire decode will fail with that error.
//
// Defaults to nothing set.
func DecodeHook(hook mapstructure.DecodeHookFunc) DecoderConfigOption {
	return &decodeHookOption{fn: hook}
}

type decodeHookOption struct {
	fn mapstructure.DecodeHookFunc
}

func (d decodeHookOption) unmarshalApply(opts *unmarshalOptions) error {
	opts.decoder.DecodeHook = d.fn
	return nil
}

func (d decodeHookOption) valueApply(opts *valueOptions) error {
	opts.decoder.DecodeHook = d.fn
	return nil
}

func (d decodeHookOption) String() string {
	return print.P("DecodeHook", print.Fn(d.fn), print.SubOpt())
}

// If ErrorUnused is true, then it is an error for there to exist
// keys in the original map that were unused in the decoding process
// (extra keys).
//
// Defaults to false.
func ErrorUnused(unused ...bool) DecoderConfigOption {
	unused = append(unused, true)
	return errorUnusedOption(unused[0])
}

type errorUnusedOption bool

func (val errorUnusedOption) unmarshalApply(opts *unmarshalOptions) error {
	opts.decoder.ErrorUnused = bool(val)
	return nil
}

func (val errorUnusedOption) valueApply(opts *valueOptions) error {
	opts.decoder.ErrorUnused = bool(val)
	return nil
}

func (val errorUnusedOption) String() string {
	return print.P("ErrorUnused", print.BoolSilentTrue(bool(val)), print.SubOpt())
}

// If ErrorUnset is true, then it is an error for there to exist
// fields in the result that were not set in the decoding process
// (extra fields). This only applies to decoding to a struct. This
// will affect all nested structs as well.
//
// Defaults to false.
func ErrorUnset(unset ...bool) DecoderConfigOption {
	unset = append(unset, true)
	return errorUnsetOption(unset[0])
}

type errorUnsetOption bool

func (val errorUnsetOption) unmarshalApply(opts *unmarshalOptions) error {
	opts.decoder.ErrorUnset = bool(val)
	return nil
}

func (val errorUnsetOption) valueApply(opts *valueOptions) error {
	opts.decoder.ErrorUnset = bool(val)
	return nil
}

func (val errorUnsetOption) String() string {
	return print.P("ErrorUnset", print.BoolSilentTrue(bool(val)), print.SubOpt())
}

// If WeaklyTypedInput is true, the decoder will make the following
// "weak" conversions:
//
//   - bools to string (true = "1", false = "0")
//   - numbers to string (base 10)
//   - bools to int/uint (true = 1, false = 0)
//   - strings to int/uint (base implied by prefix)
//   - int to bool (true if value != 0)
//   - string to bool (accepts: 1, t, T, TRUE, true, True, 0, f, F,
//     FALSE, false, False. Anything else is an error)
//   - empty array = empty map and vice versa
//   - negative numbers to overflowed uint values (base 10)
//   - slice of maps to a merged map
//   - single values are converted to slices if required. Each
//     element is weakly decoded. For example: "4" can become []int{4}
//     if the target type is an int slice.
//
// Defaults to false.
func WeaklyTypedInput(weak ...bool) DecoderConfigOption {
	weak = append(weak, true)
	return weaklyTypedInputOption(weak[0])
}

type weaklyTypedInputOption bool

func (val weaklyTypedInputOption) unmarshalApply(opts *unmarshalOptions) error {
	opts.decoder.WeaklyTypedInput = bool(val)
	return nil
}

func (val weaklyTypedInputOption) valueApply(opts *valueOptions) error {
	opts.decoder.WeaklyTypedInput = bool(val)
	return nil
}

func (val weaklyTypedInputOption) String() string {
	return print.P("WeaklyTypedInput", print.BoolSilentTrue(bool(val)), print.SubOpt())
}

// The tag name that mapstructure reads for field names.
//
// This defaults to "mapstructure".
func TagName(name string) DecoderConfigOption {
	return tagNameOption(name)
}

type tagNameOption string

func (val tagNameOption) unmarshalApply(opts *unmarshalOptions) error {
	opts.decoder.TagName = string(val)
	return nil
}

func (val tagNameOption) valueApply(opts *valueOptions) error {
	opts.decoder.TagName = string(val)
	return nil
}

func (val tagNameOption) String() string {
	return print.P("TagName", print.String(string(val)), print.SubOpt())
}

// IgnoreUntaggedFields ignores all struct fields without explicit
// TagName, comparable to `mapstructure:"-"` as default behavior.
func IgnoreUntaggedFields(ignore ...bool) DecoderConfigOption {
	ignore = append(ignore, true)
	return ignoreUntaggedFieldsOption(ignore[0])
}

type ignoreUntaggedFieldsOption bool

func (val ignoreUntaggedFieldsOption) unmarshalApply(opts *unmarshalOptions) error {
	opts.decoder.IgnoreUntaggedFields = bool(val)
	return nil
}

func (val ignoreUntaggedFieldsOption) valueApply(opts *valueOptions) error {
	opts.decoder.IgnoreUntaggedFields = bool(val)
	return nil
}

func (val ignoreUntaggedFieldsOption) String() string {
	return print.P("IgnoreUntaggedFields", print.BoolSilentTrue(bool(val)), print.SubOpt())
}

// MatchName is the function used to match the map key to the struct
// field name or tag. Defaults to `strings.EqualFold`. This can be used
// to implement case-sensitive tag values, support snake casing, etc.
//
// Defaults to nil.
func MatchName(fn func(key, field string) bool) DecoderConfigOption {
	return &matchNameOption{fn: fn}
}

type matchNameOption struct {
	fn func(mapKey, fieldName string) bool
}

func (match matchNameOption) unmarshalApply(opts *unmarshalOptions) error {
	opts.decoder.MatchName = match.fn
	return nil
}

func (match matchNameOption) valueApply(opts *valueOptions) error {
	opts.decoder.MatchName = match.fn
	return nil
}

func (match matchNameOption) String() string {
	return print.P("MatchName", print.Fn(match.fn), print.SubOpt())
}

// ZeroFields, if set to true, will zero fields before writing them.
// For example, a map will be emptied before decoded values are put in
// it. If this is false, a map will be merged.
//
// Defaults to false.
func ZeroFields(zero ...bool) DecoderConfigOption {
	zero = append(zero, true)
	return zeroFieldsOption(zero[0])
}

type zeroFieldsOption bool

func (z zeroFieldsOption) unmarshalApply(opts *unmarshalOptions) error {
	opts.decoder.ZeroFields = bool(z)
	return nil
}

func (z zeroFieldsOption) valueApply(opts *valueOptions) error {
	opts.decoder.ZeroFields = bool(z)
	return nil
}

func (z zeroFieldsOption) String() string {
	return print.P("ZeroFields", print.BoolSilentTrue(bool(z)), print.SubOpt())
}

// Exactly allows setting nearly all the mapstructure.DecoderConfig values to
// whatever value is desired.  A few fields aren't available (Metadata, Squash,
// Result) but the rest are honored.
//
// This option will mainly be useful in a scope where the code has no idea what
// options have been set & needs something very specific.
func Exactly(this mapstructure.DecoderConfig) DecoderConfigOption {
	return &exactlyOption{dc: this}
}

type exactlyOption struct {
	dc mapstructure.DecoderConfig
}

func (exact exactlyOption) decoderApply(m *mapstructure.DecoderConfig) error {
	m.DecodeHook = exact.dc.DecodeHook
	m.ErrorUnused = exact.dc.ErrorUnused
	m.ErrorUnset = exact.dc.ErrorUnset
	m.ZeroFields = exact.dc.ZeroFields
	m.WeaklyTypedInput = exact.dc.WeaklyTypedInput
	// Squash ... I don't think we can use it
	// Metadata isn't supported
	// Result is needed by goschtalt
	m.TagName = exact.dc.TagName
	m.IgnoreUntaggedFields = exact.dc.IgnoreUntaggedFields
	m.MatchName = exact.dc.MatchName
	return nil
}

func (exact exactlyOption) unmarshalApply(opts *unmarshalOptions) error {
	exact.decoderApply(&opts.decoder)
	return nil
}

func (exact exactlyOption) valueApply(opts *valueOptions) error {
	exact.decoderApply(&opts.decoder)
	return nil
}

func (exact exactlyOption) String() string {
	return print.P("Exactly",
		print.Fn(exact.dc.DecodeHook, "DecodeHook"),
		print.Bool(exact.dc.ErrorUnused, "ErrorUnused"),
		print.Bool(exact.dc.ErrorUnset, "ErrorUnset"),
		print.Bool(exact.dc.ZeroFields, "ZeroFields"),
		print.Bool(exact.dc.WeaklyTypedInput, "WeaklyTypedInput"),
		print.String(exact.dc.TagName, "TagName"),
		print.Bool(exact.dc.IgnoreUntaggedFields, "IgnoreUntaggedFields"),
		print.Fn(exact.dc.MatchName, "MatchName"),
		print.SubOpt())
}
