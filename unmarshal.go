// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/schmidtw/goschtalt/pkg/meta"
)

// Unmarshal provides a generics based strict typed approach to fetching parts
// of the configuration tree.
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
func UnmarshalFn[T any](key string, opts ...UnmarshalOption) func(*Config) (T, error) {
	return func(cfg *Config) (T, error) {
		return Unmarshal[T](cfg, key, opts...)
	}
}

// Unmarshal performs the act of looking up the specified section of the tree
// and decoding the tree into the result.  Additional options can be specified
// to adjust the behavior.
func (c *Config) Unmarshal(key string, result any, opts ...UnmarshalOption) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.compiled {
		return ErrNotCompiled
	}

	var options unmarshalOptions
	options.decoder.Result = result

	full := append(c.opts.unmarshalOptions, opts...)
	for _, opt := range full {
		if opt != nil {
			opt.unmarshalApply(&options)
		}
	}

	obj := c.tree
	if len(key) > 0 {
		key = c.opts.keySwizzler(key)
		path := strings.Split(key, c.opts.keyDelimiter)

		var err error
		obj, err = c.tree.Fetch(path, c.opts.keyDelimiter)
		if err != nil {
			if !errors.Is(err, meta.ErrNotFound) && !options.optional {
				return err
			}
		}
	}
	tree := obj.ToRaw()

	decoder, err := mapstructure.NewDecoder(&options.decoder)
	if err != nil {
		return err
	}
	return decoder.Decode(tree)
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
	optional bool
	decoder  mapstructure.DecoderConfig
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

// UnmarshalWith provides a way to configure the [mapstructure.DecoderConfig]
// structure that controls the unmarshalling process.
func UnmarshalWith(opts ...DecoderConfigOption) UnmarshalOption {
	return unmarshalWithOption(opts)
}

type unmarshalWithOption []DecoderConfigOption

func (u unmarshalWithOption) unmarshalApply(opts *unmarshalOptions) {
	for _, opt := range []DecoderConfigOption(u) {
		opt.decoderApply(&opts.decoder)
	}
}

func (u unmarshalWithOption) String() string {
	opts := []DecoderConfigOption(u)
	if len(opts) == 0 {
		return "UnmarshalWith()"
	}

	s := make([]string, len(opts))
	for i, opt := range opts {
		s[i] = opt.String()
	}
	return "UnmarshalWith(" + strings.Join(s, ", ") + ")"
}
