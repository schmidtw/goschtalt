// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import "github.com/mitchellh/mapstructure"

// DecoderOption is used for configuring the mapstructure unmarshal operation.
//
// All of these options are directly concerned with the mitchellh/mapstructure
// package.  For additional details please see: https://github.com/mitchellh/mapstructure
type DecoderOption func(*mapstructure.DecoderConfig)

// DecodeHook, will be called before any decoding and any type conversion (if
// WeaklyTypedInput is on). This lets you modify the values before they're set
// down onto the resulting struct. The DecodeHook is called for every map and
// value in the input. This means that if a struct has embedded fields with
// squash tags the decode hook is called only once with all of the input data,
// not once for each embedded struct.
//
// If an error is returned, the entire decode will fail with that error.
//
// Defaults to:
//   mapstructure.ComposeDecodeHookFunc(
//       mapstructure.StringToTimeDurationHookFunc(),
//       mapstructure.StringToSliceHookFunc(",")
//   )
func DecodeHook(hook mapstructure.DecodeHookFunc) DecoderOption {
	return func(cfg *mapstructure.DecoderConfig) {
		cfg.DecodeHook = hook
	}
}

// If ErrorUnused is true, then it is an error for there to exist
// keys in the original map that were unused in the decoding process
// (extra keys).
//
// Defaults to false.
func ErrorUnused(unused bool) DecoderOption {
	return func(cfg *mapstructure.DecoderConfig) {
		cfg.ErrorUnused = unused
	}
}

// If ErrorUnset is true, then it is an error for there to exist
// fields in the result that were not set in the decoding process
// (extra fields). This only applies to decoding to a struct. This
// will affect all nested structs as well.
//
// Defaults to false.
func ErrorUnset(unset bool) DecoderOption {
	return func(cfg *mapstructure.DecoderConfig) {
		cfg.ErrorUnset = unset
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
// Defaults to true.
func WeaklyTypedInput(weak bool) DecoderOption {
	return func(cfg *mapstructure.DecoderConfig) {
		cfg.WeaklyTypedInput = weak
	}
}

// The tag name that mapstructure reads for field names.
//
// This defaults to "mapstructure".
func TagName(name string) DecoderOption {
	return func(cfg *mapstructure.DecoderConfig) {
		cfg.TagName = name
	}
}

// IgnoreUntaggedFields ignores all struct fields without explicit
// TagName, comparable to `mapstructure:"-"` as default behaviour.
//
// Defaults to false.
func IgnoreUntaggedFields(ignore bool) DecoderOption {
	return func(cfg *mapstructure.DecoderConfig) {
		cfg.IgnoreUntaggedFields = ignore
	}
}

// MatchName is the function used to match the map key to the struct
// field name or tag. Defaults to `strings.EqualFold`. This can be used
// to implement case-sensitive tag values, support snake casing, etc.
//
// Defaults to nil.
func MatchName(fn func(mapKey, fieldName string) bool) DecoderOption {
	return func(cfg *mapstructure.DecoderConfig) {
		cfg.MatchName = fn
	}
}

// Unmarshal performs the act of looking up the specified section of the tree
// and decoding the tree into the result.  Additional options can be specified
// to adjust the behavior.
func (c *Config) Unmarshal(key string, result any, opts ...DecoderOption) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if !c.hasBeenCompiled {
		return ErrNotCompiled
	}

	tree, err := c.Fetch(key)
	if err != nil {
		return err
	}

	cfg := &mapstructure.DecoderConfig{
		Result:           result,
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
	}
	for _, opt := range opts {
		opt(cfg)
	}

	decoder, err := mapstructure.NewDecoder(cfg)
	if err != nil {
		return err
	}
	return decoder.Decode(tree)
}
