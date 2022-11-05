// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

// DecoderConfigOption is used to configure the mapstructure used for decoding
// structures into the values used by the internal goschtalt tree.
//
// All of these options are directly concerned with the mitchellh/mapstructure
// package.  For additional details please see: https://github.com/mitchellh/mapstructure
type DecoderConfigOption interface {
	fmt.Stringer

	// decoderApply applies the options to the DecoderConfig used by several
	// parts of goschtalt.
	decoderApply(*mapstructure.DecoderConfig)
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

func (d decodeHookOption) decoderApply(m *mapstructure.DecoderConfig) {
	m.DecodeHook = d.fn
}

func (d decodeHookOption) String() string {
	if d.fn == nil {
		return "DecodeHook('')"
	}

	return "DecodeHook(custom)"
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

func (val errorUnusedOption) decoderApply(m *mapstructure.DecoderConfig) {
	m.ErrorUnused = bool(val)
}

func (val errorUnusedOption) String() string {
	if val {
		return "ErrorUnused()"
	}
	return "ErrorUnused(false)"
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

func (val errorUnsetOption) decoderApply(m *mapstructure.DecoderConfig) {
	m.ErrorUnset = bool(val)
}

func (val errorUnsetOption) String() string {
	if val {
		return "ErrorUnset()"
	}
	return "ErrorUnset(false)"
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

func (val weaklyTypedInputOption) decoderApply(m *mapstructure.DecoderConfig) {
	m.WeaklyTypedInput = bool(val)
}

func (val weaklyTypedInputOption) String() string {
	if val {
		return "WeaklyTypedInput()"
	}
	return "WeaklyTypedInput(false)"
}

// The tag name that mapstructure reads for field names.
//
// This defaults to "mapstructure".
func TagName(name string) DecoderConfigOption {
	return tagNameOption(name)
}

type tagNameOption string

func (val tagNameOption) decoderApply(m *mapstructure.DecoderConfig) {
	m.TagName = string(val)
}

func (val tagNameOption) String() string {
	return fmt.Sprintf("TagName('%s')", string(val))
}

// IgnoreUntaggedFields ignores all struct fields without explicit
// TagName, comparable to `mapstructure:"-"` as default behavior.
func IgnoreUntaggedFields(ignore ...bool) DecoderConfigOption {
	ignore = append(ignore, true)
	return ignoreUntaggedFieldsOption(ignore[0])
}

type ignoreUntaggedFieldsOption bool

func (val ignoreUntaggedFieldsOption) decoderApply(m *mapstructure.DecoderConfig) {
	m.IgnoreUntaggedFields = bool(val)
}

func (val ignoreUntaggedFieldsOption) String() string {
	if val {
		return "IgnoreUntaggedFields()"
	}
	return "IgnoreUntaggedFields(false)"
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

func (match matchNameOption) decoderApply(m *mapstructure.DecoderConfig) {
	m.MatchName = match.fn
}

func (match matchNameOption) String() string {
	if match.fn == nil {
		return "MatchName('')"
	}
	return "MatchName(custom)"
}
