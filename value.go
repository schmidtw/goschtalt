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
func AddValue(recordName, key string, val any, opts ...DecoderConfigOption) Option {
	return &value{
		text:       "AddValue",
		recordName: recordName,
		key:        key,
		fn: func(_ string) (any, error) {
			return val, nil
		},
		opts: opts,
	}
}

func AddValueFn(recordName, key string, fn func(recordName string) (any, error), opts ...DecoderConfigOption) Option {
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
	fn func(string) (any, error)

	// Optional options that configure how mapstructure will process the Value
	// provided.  These options are in addition to any default settings set with
	// AddDefaultValueOptions().
	opts []DecoderConfigOption
}

func (v value) decode(delimiter string, opts ...DecoderConfigOption) (record, error) {
	tree := make(map[string]any)
	cfg := mapstructure.DecoderConfig{
		Result: &tree,
	}

	all := append(opts, v.opts...)
	for _, opt := range all {
		opt.decoderApply(&cfg)
	}

	decoder, err := mapstructure.NewDecoder(&cfg)
	if err == nil {
		data, err := v.fn(v.recordName)
		if err == nil {
			err = decoder.Decode(data)
		}
	}
	if err != nil {
		return record{}, err
	}

	return record{
		name: v.recordName,
		tree: meta.ObjectFromRaw(tree, strings.Split(v.key, delimiter)...),
	}, nil
}

func (v value) apply(opts *options) error {
	if len(v.recordName) == 0 {
		return fmt.Errorf("%w: no valid record name provided", ErrInvalidInput)
	}
	opts.values = append(opts.values, v)
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

	return fmt.Sprintf("%s( recordName: '%s', key: '%s', opts: %s )",
		v.text, v.recordName, v.key, strings.Join(s, ", "))
}
