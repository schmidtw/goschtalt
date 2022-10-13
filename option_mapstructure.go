// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"github.com/mitchellh/mapstructure"
)

type decoderConfig struct {
	optional bool
	cfg      mapstructure.DecoderConfig
}

// MapstructureOption is used for configuring the mapstructure used for decoding
// structures into the values used by the internal goschtalt tree.
//
// All of these options are directly concerned with the mitchellh/mapstructure
// package.  For additional details please see: https://github.com/mitchellh/mapstructure
type MapstructureOption func(*decoderConfig)

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
func DecodeHook(hook mapstructure.DecodeHookFunc) MapstructureOption {
	return func(d *decoderConfig) {
		d.cfg.DecodeHook = hook
	}
}

// If ErrorUnused is true, then it is an error for there to exist
// keys in the original map that were unused in the decoding process
// (extra keys).
//
// Defaults to false.
func ErrorUnused(unused bool) MapstructureOption {
	return func(d *decoderConfig) {
		d.cfg.ErrorUnused = unused
	}
}

// If ErrorUnset is true, then it is an error for there to exist
// fields in the result that were not set in the decoding process
// (extra fields). This only applies to decoding to a struct. This
// will affect all nested structs as well.
//
// Defaults to false.
func ErrorUnset(unset bool) MapstructureOption {
	return func(d *decoderConfig) {
		d.cfg.ErrorUnset = unset
	}
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
func WeaklyTypedInput(weak bool) MapstructureOption {
	return func(d *decoderConfig) {
		d.cfg.WeaklyTypedInput = weak
	}
}

// The tag name that mapstructure reads for field names.
//
// This defaults to "mapstructure".
func TagName(name string) MapstructureOption {
	return func(d *decoderConfig) {
		d.cfg.TagName = name
	}
}

// IgnoreUntaggedFields ignores all struct fields without explicit
// TagName, comparable to `mapstructure:"-"` as default behaviour.
func IgnoreUntaggedFields(ignore bool) MapstructureOption {
	return func(d *decoderConfig) {
		d.cfg.IgnoreUntaggedFields = ignore
	}
}

// MatchName is the function used to match the map key to the struct
// field name or tag. Defaults to `strings.EqualFold`. This can be used
// to implement case-sensitive tag values, support snake casing, etc.
//
// Defaults to nil.
func MatchName(fn func(mapKey, fieldName string) bool) MapstructureOption {
	return func(d *decoderConfig) {
		d.cfg.MatchName = fn
	}
}

// Optional set to true causes the operation to ignore missing parts of the
// tree and simply pass back the object unchanged.
//
// Defaults to false.
func Optional(optional bool) MapstructureOption {
	return func(d *decoderConfig) {
		d.optional = optional
	}
}
