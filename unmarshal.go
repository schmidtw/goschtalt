// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/goschtalt/goschtalt/pkg/meta"
	"github.com/mitchellh/mapstructure"
)

// UnmarshalFunc provides a special use Unmarshal() function during AddBufferFn()
// and AddValueFn() option provided callbacks.  This pattern allows the specified
// function access to the configuration values up to this point.  Expansion of
// any Expand() or ExpandEnv() options is also applied to the configuration tree
// provided.
type UnmarshalFunc func(key string, result any, opts ...UnmarshalOption) error

// Unmarshal provides a generics based strict typed approach to fetching parts
// of the configuration tree.
//
// To read the entire configuration tree, use `goschtalt.Root` instead of "" for
// more clarity.
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
// To read the entire configuration tree, use `goschtalt.Root` instead of "" for
// more clarity.
func UnmarshalFn[T any](key string, opts ...UnmarshalOption) func(*Config) (T, error) {
	return func(cfg *Config) (T, error) {
		return Unmarshal[T](cfg, key, opts...)
	}
}

// Unmarshal performs the act of looking up the specified section of the tree
// and decoding the tree into the result.  Additional options can be specified
// to adjust the behavior.
//
// To read the entire configuration tree, use `goschtalt.Root` instead of "" for
// more clarity.
func (c *Config) Unmarshal(key string, result any, opts ...UnmarshalOption) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.compiledAt.Equal(time.Time{}) {
		return ErrNotCompiled
	}

	return c.unmarshal(key, result, c.tree, opts...)
}

func (c *Config) unmarshal(key string, result any, tree meta.Object, opts ...UnmarshalOption) error {
	var options unmarshalOptions
	options.decoder.Result = result

	full := append(c.opts.unmarshalOptions, opts...)
	for _, opt := range full {
		if opt != nil {
			opt.unmarshalApply(&options)
		}
	}

	obj := tree
	if len(key) > 0 {
		key = c.opts.keySwizzler(key)
		path := strings.Split(key, c.opts.keyDelimiter)

		var err error
		obj, err = tree.Fetch(path, c.opts.keyDelimiter)
		if err != nil {
			if !errors.Is(err, meta.ErrNotFound) && !options.optional {
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
	unmarshalApply(*unmarshalOptions)
}

type unmarshalOptions struct {
	optional  bool
	decoder   mapstructure.DecoderConfig
	validator func(any) error
}

// Optional provides a way to allow the requested configuration to not be present
// and return an empty structure without an error instead of failing.  If the
// optional parameter is not passed, the value is assumed to be true.
//
// The default behavior is to require the request to be present.
//
// See also: [Required]
func Optional(optional ...bool) UnmarshalOption {
	optional = append(optional, true)
	if optional[0] {
		return &optionalOption{
			text:     "Optional()",
			optional: true,
		}
	}

	return &optionalOption{
		text: "Optional(false)",
	}
}

// Required provides a way to allow the requested configuration to be required
// and return an error if it is missing.  If the optional parameter is not
// passed, the value is assumed to be true.
//
// The default behavior is to require the request to be present.
//
// See also: [Optional]
func Required(required ...bool) UnmarshalOption {
	required = append(required, true)
	if required[0] {
		return &optionalOption{
			text: "Required()",
		}
	}

	return &optionalOption{
		text:     "Required(false)",
		optional: true,
	}
}

type optionalOption struct {
	text     string
	optional bool
}

func (o optionalOption) unmarshalApply(opts *unmarshalOptions) {
	opts.optional = o.optional
}

func (o optionalOption) String() string {
	return o.text
}

// WithValidator provides a way to specify a validator to use after a structure
// has been unmarshaled, but prior to returning the data.  This allows for an
// easy way to consistently validate configuration as it is being consumed.  If
// the validator function returns an error the Unmarshal operation will result
// in a failure and return the error.
//
// The default behavior is to not validate.
//
// Setting the value to nil disables validation.
func WithValidator(fn func(any) error) UnmarshalOption {
	fnType := "nil"
	if fn != nil {
		fnType = "custom"
	}
	return &validatorOption{
		text: fnType,
		fn:   fn,
	}
}

type validatorOption struct {
	text string
	fn   func(any) error
}

func (v validatorOption) unmarshalApply(opts *unmarshalOptions) {
	opts.validator = v.fn
}

func (v validatorOption) String() string {
	return fmt.Sprintf("WithValidator(%s)", v.text)
}
