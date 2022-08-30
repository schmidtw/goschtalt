// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"errors"

	"github.com/mitchellh/mapstructure"
	"github.com/schmidtw/goschtalt/pkg/meta"
)

type unmarshalConfig struct {
	optional bool
	cfg      mapstructure.DecoderConfig
}

// UnmarshalOption is used for configuring the mapstructure unmarshal operation.
//
// All of these options are directly concerned with the mitchellh/mapstructure
// package.  For additional details please see: https://github.com/mitchellh/mapstructure
type UnmarshalOption func(*unmarshalConfig)

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
func DecodeHook(hook mapstructure.DecodeHookFunc) UnmarshalOption {
	return func(u *unmarshalConfig) {
		u.cfg.DecodeHook = hook
	}
}

// If ErrorUnused is true, then it is an error for there to exist
// keys in the original map that were unused in the decoding process
// (extra keys).
//
// Defaults to false.
func ErrorUnused(unused bool) UnmarshalOption {
	return func(u *unmarshalConfig) {
		u.cfg.ErrorUnused = unused
	}
}

// If ErrorUnset is true, then it is an error for there to exist
// fields in the result that were not set in the decoding process
// (extra fields). This only applies to decoding to a struct. This
// will affect all nested structs as well.
//
// Defaults to false.
func ErrorUnset(unset bool) UnmarshalOption {
	return func(u *unmarshalConfig) {
		u.cfg.ErrorUnset = unset
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
func WeaklyTypedInput(weak bool) UnmarshalOption {
	return func(u *unmarshalConfig) {
		u.cfg.WeaklyTypedInput = weak
	}
}

// The tag name that mapstructure reads for field names.
//
// This defaults to "mapstructure".
func TagName(name string) UnmarshalOption {
	return func(u *unmarshalConfig) {
		u.cfg.TagName = name
	}
}

// IgnoreUntaggedFields ignores all struct fields without explicit
// TagName, comparable to `mapstructure:"-"` as default behaviour.
//
// NOTE: This appears broken upstream in mapstructure.
//
// Defaults to false.
func IgnoreUntaggedFields(ignore bool) UnmarshalOption {
	return func(u *unmarshalConfig) {
		u.cfg.IgnoreUntaggedFields = ignore
	}
}

// MatchName is the function used to match the map key to the struct
// field name or tag. Defaults to `strings.EqualFold`. This can be used
// to implement case-sensitive tag values, support snake casing, etc.
//
// Defaults to nil.
func MatchName(fn func(mapKey, fieldName string) bool) UnmarshalOption {
	return func(u *unmarshalConfig) {
		u.cfg.MatchName = fn
	}
}

// Optional set to true causes the unmarshal operation to ignore missing parts
// of the tree and simply pass back the object unchanged.
//
// Defaults to false.
func Optional(optional bool) UnmarshalOption {
	return func(u *unmarshalConfig) {
		u.optional = optional
	}
}

// DefaultUnmarshalOption allows customization of the desired options for all
// invocations of the Unmarshal() function.  This should make consistent use
// use of the Unmarshal() call easier.
func DefaultUnmarshalOption(opt UnmarshalOption) Option {
	return func(c *Config) error {
		c.unmarshalOptions = append(c.unmarshalOptions, opt)
		return nil
	}
}

// Unmarshal performs the act of looking up the specified section of the tree
// and decoding the tree into the result.  Additional options can be specified
// to adjust the behavior.
func (c *Config) Unmarshal(key string, result any, opts ...UnmarshalOption) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var u unmarshalConfig
	u.cfg.Result = result

	full := append(c.unmarshalOptions, opts...)
	for _, opt := range full {
		if opt != nil {
			opt(&u)
		}
	}

	tree, _, err := c.fetchWithOrigin(key)
	if err != nil {
		if errors.Is(err, meta.ErrNotFound) && u.optional {
			return nil
		}
		return err
	}

	decoder, err := mapstructure.NewDecoder(&u.cfg)
	if err != nil {
		return err
	}
	return decoder.Decode(tree)
}

// UnmarshalFn returns a function that takes a goschtalt Config structure and
// returns a function that allows for unmarshalling of a portion of the tree
// specified by the key into a zero value type.
//
// This function is specifically helpful with DI frameworks like Uber's fx
// framework.
//
// In this short example, the type myStruct is created and populated with the
// configuring values found under the "conf" key in the goschtalt configuration.
//
//	app := fx.New(
//		fx.Provide(
//			goschtalt.UnmarshalFn[myStruct]("conf"),
//		),
//	)
func UnmarshalFn[T any](key string, opts ...UnmarshalOption) func(*Config) (T, error) {
	return func(cfg *Config) (T, error) {
		var zeroVal T
		var obj T

		err := cfg.Unmarshal(key, &obj, opts...)
		if err != nil {
			return zeroVal, err
		}

		return obj, nil
	}
}
