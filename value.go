// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/schmidtw/goschtalt/pkg/meta"
)

// AddValues provides a simple way to set additional configuration values at
// runtime.
func AddValue(recordName, key string, val any, opts ...ValueOption) Option {
	return &value{
		text:       "AddValue",
		recordName: recordName,
		key:        key,
		fn: func(_ string, _ UnmarshalFunc) (any, error) {
			return val, nil
		},
		opts: opts,
	}
}

// AddValues provides a simple way to set additional configuration values at
// runtime via a function call.  Note that the provided fn will be called each
// time the configuration is compiled, allowing the value returned to change if
// desired.
func AddValueFn(recordName, key string, fn func(recordName string, unmarshal UnmarshalFunc) (any, error), opts ...ValueOption) Option {
	return &value{
		text:       "AddValueFn",
		recordName: recordName,
		key:        key,
		fn:         fn,
		opts:       opts,
	}
}

// value defines a key and value that is injected into the configuration tree.
type value struct {
	text string

	// The record to use for sorting this configuration.
	recordName string

	// The key to set the value at.
	key string

	// The fn to use to get the value.
	fn func(recordName string, unmarshal UnmarshalFunc) (any, error)

	// Options that configure how mapstructure will process the Value provided.
	// These options are in addition to any default settings set with
	// AddDefaultValueOptions().
	opts []ValueOption
}

func (v value) toTree(delimiter string, um UnmarshalFunc, defaultOpts ...ValueOption) (meta.Object, error) {
	tree := make(map[string]any)
	cfg := mapstructure.DecoderConfig{
		Result: &tree,
	}

	all := append(defaultOpts, v.opts...)
	for _, opt := range all {
		opt.decoderApply(&cfg)
	}

	decoder, err := mapstructure.NewDecoder(&cfg)
	if err == nil {
		var data any
		data, err = v.fn(v.recordName, um)
		if err == nil {
			err = decoder.Decode(data)
		}
	}
	if err != nil {
		return meta.Object{}, err
	}

	origin := []meta.Origin{{File: v.recordName}}

	return meta.ObjectFromRawWithOrigin(tree, origin, strings.Split(v.key, delimiter)...), nil
}

func (v value) apply(opts *options) error {
	if len(v.recordName) == 0 {
		return fmt.Errorf("%w: no valid record name provided", ErrInvalidInput)
	}

	r := record{
		name: v.recordName,
		val:  &v,
	}

	for _, opt := range v.opts {
		if opt.isDefault() {
			opts.defaults = append(opts.defaults, r)
			return nil
		}
	}

	opts.values = append(opts.values, r)

	return nil
}

func (_ value) ignoreDefaults() bool {
	return false
}

func (v value) String() string {
	s := make([]string, len(v.opts))
	for i, opt := range v.opts {
		s[i] = opt.String()
	}

	if len(s) == 0 {
		s = append(s, "none")
	}

	fn := ""
	if v.text == "AddValueFn" {
		fn = "'', "
		if v.fn != nil {
			fn = "custom, "
		}
	}
	return fmt.Sprintf("%s( '%s', '%s', %s%s )",
		v.text, v.recordName, v.key, fn, strings.Join(s, ", "))
}

// -- ValueOption options follow -----------------------------------------------

// ValueOption provides the means to configure options around variable mapping
// as well as if the specific value being added should be a default or a normal
// configuration value.
//
// See also DecoderConfigOption which can be used as ValueOption options.
type ValueOption interface {
	fmt.Stringer

	isDefault() bool

	// decoderApply applies the options to the DecoderConfig used by several
	// parts of goschtalt.
	decoderApply(*mapstructure.DecoderConfig)
}

// AsDefault specifies that this value is a default value & is applied prior to
// any other configuration values.  Default values are applied in the order the
// options are specified.
func AsDefault(asDefault ...bool) ValueOption {
	asDefault = append(asDefault, true)

	return optionalAsDefault(asDefault[0])
}

type optionalAsDefault bool

func (o optionalAsDefault) isDefault() bool {
	return bool(o)
}

func (_ optionalAsDefault) decoderApply(_ *mapstructure.DecoderConfig) {}

func (o optionalAsDefault) String() string {
	if o {
		return "AsDefault()"
	}
	return "AsDefault(false)"
}
