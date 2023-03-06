// SPDX-FileCopyrightText: 2022-2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/goschtalt/goschtalt/internal/print"
	"github.com/goschtalt/goschtalt/internal/structs"
	"github.com/goschtalt/goschtalt/pkg/meta"
)

// AddValues provides a simple way to set additional configuration values at
// runtime.
//
// To place the configuration at the root use `goschtalt.Root` ([Root]) instead
// of "" for more clarity.
//
// Valid Option Types:
//   - [BufferValueOption]
//   - [GlobalOption]
//   - [ValueOption]
//   - [UnmarshalValueOption]
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
// To place the configuration at the root use `goschtalt.Root` ([Root]) instead
// of "" for more clarity.
//
// Valid Option Types:
//   - [BufferValueOption]
//   - [GlobalOption]
//   - [ValueOption]
//   - [UnmarshalValueOption]
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

	// Options that configure how to process the Value provided.
	// These options are in addition to any default settings set with
	// AddDefaultValueOptions().
	opts []ValueOption
}

// toTree does the work of converting from a structure of some sort to the
// normalized object tree goschtalt uses.
func (v value) toTree(delimiter string, um UnmarshalFunc, defaultOpts ...ValueOption) (meta.Object, error) {
	cfg := valueOptions{
		tagName: defaultTag,
	}

	all := append(defaultOpts, v.opts...)
	for _, opt := range all {
		if err := opt.valueApply(&cfg); err != nil {
			return meta.Object{}, err
		}
	}

	data, err := v.fn(v.recordName, um)
	if err != nil {
		return meta.Object{}, err
	}

	// Dereference the pointer if it is one.
	if reflect.TypeOf(data).Kind() == reflect.Ptr {
		data = reflect.ValueOf(data).Elem().Interface()
	}

	if reflect.TypeOf(data).Kind() == reflect.Struct {
		s := structs.New(data)
		s.TagName = cfg.tagName
		data = s.Map()
	}

	tree := meta.ObjectFromRawWithOrigin(data,
		[]meta.Origin{{File: v.recordName}},
		strings.Split(v.key, delimiter)...)

	tree = tree.AlterKeyCase(func(s string) string {
		return cfg.mapper(s)
	})

	tree, err = tree.AdaptToRaw(adapterIterator(cfg.adapters))
	if err != nil {
		return meta.Object{}, err
	}

	if cfg.failOnNonSerializable {
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
		var info valueOptions

		if err := opt.valueApply(&info); err != nil {
			return err
		}

		if info.isDefault {
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
// See also [UnmarshalValueOption] which can be used as ValueOption options.
type ValueOption interface {
	fmt.Stringer

	valueApply(*valueOptions) error
}

type valueOptions struct {
	tagName               string
	mappers               []Mapper
	adapters              []adapter
	failOnNonSerializable bool
	isDefault             bool
}

// mapper is a simple helper that does the mapping based on the specified
// options.
func (v valueOptions) mapper(s string) string {
	for _, m := range v.mappers {
		if rv := m(s); rv != "" {
			s = rv
		}
	}

	return s
}

// FailOnNonSerializable specifies that an error should be returned if any
// non-serializable objects (channels, functions, unsafe pointers) are
// encountered in the resulting configuration tree.  Non-serializable objects
// cannot be in the configuration sets that goschtalt works with.
//
// The fail bool value is optional & assumed to be `true` if omitted.  The
// first specified value is used if provided.  A value of `false` disables the
// option.
//
// The default behavior is to ignore and drop any non-serializable objects.
func FailOnNonSerializable(fail ...bool) ValueOption {
	fail = append(fail, true)
	return failOnNonSerializableOption(fail[0])
}

type failOnNonSerializableOption bool

func (e failOnNonSerializableOption) valueApply(opt *valueOptions) error {
	opt.failOnNonSerializable = bool(e)
	return nil
}

func (e failOnNonSerializableOption) String() string {
	return print.P("FailOnNonSerializable", print.BoolSilentTrue(bool(e)), print.SubOpt())
}

// AdaptToCfg converts a value from one form to another if possible.  The resulting
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
func AdaptToCfg(fn func(from reflect.Value) (any, error), label ...string) ValueOption {
	label = append(label, "")
	return &adaptToCfgOption{
		label: label[0],
		fn: func(from, to reflect.Value) (any, error) {
			return fn(from)
		},
	}
}

type adaptToCfgOption struct {
	label string
	fn    adapter
}

func (a adaptToCfgOption) valueApply(opts *valueOptions) error {
	opts.adapters = append(opts.adapters, a.fn)
	return nil
}

func (a adaptToCfgOption) String() string {
	labels := make([]string, 0, 1)
	if len(a.label) > 0 {
		labels = append(labels, a.label)
	}
	return print.P("AdaptToCfg", print.Fn(a.fn, labels...), print.SubOpt())
}
