// SPDX-FileCopyrightText: 2022-2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/goschtalt/goschtalt/internal/print"
	"github.com/goschtalt/goschtalt/pkg/meta"
	"github.com/mitchellh/mapstructure"
)

// UnmarshalFunc provides a special use [Unmarshal]() function during [AddBufferFn]()
// and [AddValueFn]() option provided callbacks.  This pattern allows the specified
// function access to the configuration values up to this point.  Expansion of
// any [Expand]() or [ExpandEnv]() options is also applied to the configuration tree
// provided.
type UnmarshalFunc func(key string, result any, opts ...UnmarshalOption) error

// Unmarshal provides a generics based strict typed approach to fetching parts
// of the configuration tree.
//
// To read the entire configuration tree, use `goschtalt.Root` [Root] instead of
// "" for more clarity.
//
// Valid Option Types:
//   - [GlobalOption]
//   - [UnmarshalOption]
//   - [UnmarshalValueOption]
func Unmarshal[T any](c *Config, key string, opts ...UnmarshalOption) (T, error) {
	var rv T
	err := c.Unmarshal(key, &rv, opts...)
	if err != nil {
		var zeroVal T
		return zeroVal, err
	}

	return rv, nil
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
//
// To read the entire configuration tree, use `goschtalt.Root` [Root] instead of
// "" for more clarity.
//
// Valid Option Types:
//   - [GlobalOption]
//   - [UnmarshalOption]
//   - [UnmarshalValueOption]
func UnmarshalFn[T any](key string, opts ...UnmarshalOption) func(*Config) (T, error) {
	return func(cfg *Config) (T, error) {
		return Unmarshal[T](cfg, key, opts...)
	}
}

// Unmarshal performs the act of looking up the specified section of the tree
// and decoding the tree into the result.  Additional options can be specified
// to adjust the behavior.
//
// To read the entire configuration tree, use `goschtalt.Root` [Root] instead of
// "" for more clarity.
//
// Valid Option Types:
//   - [GlobalOption]
//   - [UnmarshalOption]
//   - [UnmarshalValueOption]
func (c *Config) Unmarshal(key string, result any, opts ...UnmarshalOption) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.compiledAt.Equal(time.Time{}) {
		return ErrNotCompiled
	}

	return c.unmarshal(key, result, c.tree, opts...)
}

// adapter is a function that maps a value from one form (from) to a different
// form (to) if possible.  Generally they are short, simple functions.  The
// result is returned with no error, or a nil result is returned with
// ErrNotApplicable, or nil/zero is returned with an error.
type adapter func(from, to reflect.Value) (any, error)

// adapterIterator takes an array of adapters and applies them one at a time
// until one of the following happens:
//   - an adapter succeeds
//   - all adapters are executed & there are error(s) (excluding ErrNotApplicable)
//     resulting in an overall error
//   - all adapters are executed & there are no error(s) (excluding ErrNotApplicable)
//     resulting in the from value being returned
func adapterIterator(ff []adapter) func(reflect.Value, reflect.Value) (any, error) {
	return func(from, to reflect.Value) (any, error) {
		var out any
		var err error

		errs := make([]error, 0, len(ff))
		for _, f := range ff {
			out, err = f(from, to)
			if err == nil {
				return out, nil
			}

			if !errors.Is(err, ErrNotApplicable) {
				errs = append(errs, err)
			}
		}

		if len(errs) == 0 {
			return from.Interface(), nil
		}

		strs := make([]string, len(errs))
		for i, e := range errs {
			strs[i] = e.Error()
		}
		return nil, fmt.Errorf("%w: %s", ErrAdaptFailure, strings.Join(strs, ", "))
	}
}

func (c *Config) unmarshal(key string, result any, tree meta.Object, opts ...UnmarshalOption) error {
	options := unmarshalOptions{
		decoder: mapstructure.DecoderConfig{
			Result:  result,
			TagName: defaultTag,
		},
	}

	full := append(c.opts.unmarshalOptions, opts...)
	for _, opt := range full {
		if opt != nil {
			err := opt.unmarshalApply(&options)
			if err != nil {
				return err
			}
		}
	}

	options.decoder.DecodeHook = adapterIterator(options.adapters)

	options.decoder.MatchName = func(key, field string) bool {
		encoded := options.mapper(field)
		if "-" == encoded {
			return false
		}
		return encoded == key
	}

	obj := tree
	if len(key) > 0 {
		path := strings.Split(key, c.opts.keyDelimiter)

		var err error
		obj, err = tree.Fetch(path, c.opts.keyDelimiter)
		if err != nil {
			if !options.optional || !errors.Is(err, meta.ErrNotFound) {
				return err
			}
		}
	}
	raw := obj.ToRaw()

	decoder, err := mapstructure.NewDecoder(&options.decoder)
	if err != nil {
		return err
	}
	if err := decoder.Decode(raw); err != nil {
		return err
	}
	if options.validator != nil {
		if err := options.validator(result); err != nil {
			return err
		}
	}
	return nil
}

// -- UnmarshalOption options follow -------------------------------------------

// UnmarshalOption provides specific configuration for the process of producing
// a document based on the present information in the goschtalt object.
type UnmarshalOption interface {
	fmt.Stringer

	// marshalApply applies the options to the Marshal function.
	unmarshalApply(*unmarshalOptions) error
}

type unmarshalOptions struct {
	optional  bool
	mappers   []Mapper
	adapters  []adapter
	decoder   mapstructure.DecoderConfig
	validator func(any) error
}

// mapper is a helper function that applies the mapper function behavior
// uniformly.
func (u unmarshalOptions) mapper(s string) string {
	for _, m := range u.mappers {
		if rv := m(s); rv != "" {
			s = rv
		}
	}

	return s
}

// Optional provides a way to allow the requested configuration to not be present
// and return an empty structure without an error instead of failing.
//
// The optional bool value is optional & assumed to be `true` if omitted.  The
// first specified value is used if provided.  A value of `false` disables the
// option.
//
// See also: [Required]
//
// # Default
//
// The default behavior is to require the request to be present.
func Optional(optional ...bool) UnmarshalOption {
	optional = append(optional, true)
	return &optionalOption{
		text:     print.P("Optional", print.BoolSilentTrue(optional[0]), print.SubOpt()),
		optional: optional[0],
	}
}

// Required provides a way to allow the requested configuration to be required
// and return an error if it is missing.
//
// The required bool value is optional & assumed to be `true` if omitted.  The
// first specified value is used if provided.  A value of `false` disables the
// option.
//
// See also: [Optional]
//
// # Default
//
// The default behavior is to require the request to be present.
func Required(required ...bool) UnmarshalOption {
	required = append(required, true)
	return &optionalOption{
		text:     print.P("Required", print.BoolSilentTrue(required[0]), print.SubOpt()),
		optional: !required[0],
	}
}

type optionalOption struct {
	text     string
	optional bool
}

func (o optionalOption) unmarshalApply(opts *unmarshalOptions) error {
	opts.optional = o.optional
	return nil
}

func (o optionalOption) String() string {
	return o.text
}

// WithValidator provides a way to specify a validator to use after a structure
// has been unmarshaled, but prior to returning the data.  This allows for an
// easy way to consistently validate configuration as it is being consumed.  If
// the validator function returns an error the [Unmarshal]() operation will result
// in a failure and return the error.
//
// Setting the value to nil disables validation.
//
// # Default
//
// The default behavior is to not validate.
func WithValidator(fn func(any) error) UnmarshalOption {
	return &validatorOption{
		fn: fn,
	}
}

type validatorOption struct {
	fn func(any) error
}

func (v validatorOption) unmarshalApply(opts *unmarshalOptions) error {
	opts.validator = v.fn
	return nil
}

func (v validatorOption) String() string {
	return print.P("WithValidator", print.Fn(v.fn), print.SubOpt())
}

// AdaptFromCfg converts a value from one form to another if possible.  The resulting
// form is returned if adapted.  If the combination of f and t are unknown
// or unsupported return ErrNotApplicable as the error with a nil value.
//
// If ErrNotApplicable is returned the value returned will be ignored.
//
// The optional label parameter allows you to provide the function name of the
// adapter so it is more clear which adapters are registered.
//
// All functions provided to Adapt() are called in the order provided until one
// returns no error or the end of the list is encountered.
//
// When used as an option for an Unmarshal() operation, the value of f will
// be the form in the configuration (string, int, etc) and the value of t will
// represent the desired form in the target structure.
//
// When used as an option for a Value() operation, the value of f will be the
// form present in the source structure and the t value will always be the type
// Best.  The returned type should match the type retrieved from the
// configuration decoder.  This generally will be a built in type like string,
// int or bool.
func AdaptFromCfg(fn func(from, to reflect.Value) (any, error), label ...string) UnmarshalOption {
	label = append(label, "")
	return &adaptFromCfgOption{
		label: label[0],
		fn:    fn,
	}
}

type adaptFromCfgOption struct {
	label string
	fn    adapter
}

func (a adaptFromCfgOption) unmarshalApply(opts *unmarshalOptions) error {
	opts.adapters = append(opts.adapters, a.fn)
	return nil
}

func (a adaptFromCfgOption) String() string {
	labels := make([]string, 0, 1)
	if len(a.label) > 0 {
		labels = append(labels, a.label)
	}
	return print.P("AdaptFromCfg", print.Fn(a.fn, labels...), print.SubOpt())
}

// A Level represents a specific degree in which a configuration matches a
// structure's fields.
type Level string

const (
	NONE     Level = "NONE"
	SUBSET   Level = "SUBSET"
	COMPLETE Level = "COMPLETE"
	EXACT    Level = "EXACT"
)

// Strictness defines the relationship between the configuration values and the
// structure fields.  The level parameter defines which mode to use.
//
//   - NONE - Any confiugration values are ok.  Neither missing nor extra values
//     is an error.
//   - SUBSET - The configuration values are limited to the set defined by the
//     structure fields or it is an error.  Missing configuration is ok, but
//     extra configuration values is an error.
//   - COMPLETE - The configuration values must completely fill the structure
//     fields or it is an error.  Extra configuration is ok, but missing
//     configuration values is an error.
//   - EXACT - The configuration values must exactly fill the structure fields
//     or it is an error.  Extra or missing configuration are both errors.
//   - NONE - (default) Both extra or too few configuration values as well as
//
// # Default
//
// NONE
func Strictness(level Level) UnmarshalOption {
	r := remapOption{
		level: string(level),
	}

	switch level {
	case NONE:
	case SUBSET:
		r.errorUnused = true
	case COMPLETE:
		r.errorUnset = true
	case EXACT:
		r.errorUnused = true
		r.errorUnset = true
	default:
		r.err = fmt.Errorf("%w: unsupported strictness level: '%s'", ErrInvalidInput, level)
	}

	return &r
}

type remapOption struct {
	level       string
	err         error
	errorUnused bool
	errorUnset  bool
}

func (r remapOption) unmarshalApply(opts *unmarshalOptions) error {
	opts.decoder.ErrorUnused = r.errorUnused
	opts.decoder.ErrorUnset = r.errorUnset
	return r.err
}

func (r remapOption) String() string {
	return print.P("Strictness", print.String(r.level), print.SubOpt())
}
