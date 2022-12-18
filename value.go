// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"fmt"
	"strings"

	"github.com/goschtalt/goschtalt/pkg/meta"
	"github.com/mitchellh/mapstructure"
)

// AddValues provides a simple way to set additional configuration values at
// runtime.
//
// To place the configuration at the root use `goschtalt.Root` instead of ""
// for more clarity.
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
//
// To place the configuration at the root use `goschtalt.Root` instead of ""
// for more clarity.
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

// toTree does the work of converting from a structure of some sort to the
// normalized object tree goschtalt uses.
func (v value) toTree(delimiter string, um UnmarshalFunc, defaultOpts ...ValueOption) (meta.Object, error) {
	raw := make(map[string]any)
	cfg := mapstructure.DecoderConfig{
		Result: &raw,
	}
	var valopts valueOptions

	all := append(defaultOpts, v.opts...)
	for _, opt := range all {
		opt.decoderApply(&cfg)
		opt.valueApply(&valopts)
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

	tree := meta.ObjectFromRawWithOrigin(raw,
		[]meta.Origin{{File: v.recordName}},
		strings.Split(v.key, delimiter)...)

	if valopts.failOnNonSerializable {
		if err = tree.ErrOnNonSerializable(); err != nil {
			return meta.Object{}, err
		}
	}

	return tree.FilterNonSerializable(), nil
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

	valueApply(*valueOptions)
}

type valueOptions struct {
	failOnNonSerializable bool
}

// FailOnNonSerializable specifies that an error should be returned if any
// non-serializable objects (channels, functions, unsafe pointers) are
// encountered in the resulting configuration tree.  Non-serializable objects
// cannot be in the configuration sets that goschtalt works with.
//
// The default behavior is to ignore and drop any non-serializable objects.
func FailOnNonSerializable(fail ...bool) ValueOption {
	fail = append(fail, true)
	return failOnNonSerializableOption(fail[0])
}

type failOnNonSerializableOption bool

func (_ failOnNonSerializableOption) isDefault() bool {
	return false
}

func (_ failOnNonSerializableOption) decoderApply(_ *mapstructure.DecoderConfig) {}

func (e failOnNonSerializableOption) valueApply(opt *valueOptions) {
	opt.failOnNonSerializable = bool(e)
}

func (e failOnNonSerializableOption) String() string {
	if bool(e) {
		return "FailOnNonSerializable()"
	}

	return "FailOnNonSerializable(false)"
}
