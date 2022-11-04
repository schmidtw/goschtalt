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
		recordName: recordName,
		key:        key,
		value:      val,
		opts:       opts,
	}
}

// value defines a key and value that is injected into the configuration tree.
type value struct {
	// The record to use for sorting this configuration.
	recordName string

	// The key to set the value at.
	key string

	// The value to set.  It may be a struct.
	value any

	// Optional options that configure how mapstructure will process the Value
	// provided.  These options are in addition to any default settings set with
	// AddDefaultValueOptions().
	opts []DecoderConfigOption
}

func (v value) decode(delimiter string, opts ...DecoderConfigOption) (fileObject, error) {
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
		err = decoder.Decode(v.value)
	}
	if err != nil {
		return fileObject{}, err
	}

	return fileObject{
		File: v.recordName,
		Obj:  meta.ObjectFromRaw(tree, strings.Split(v.key, delimiter)...),
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
	var s []string
	for i := range v.opts {
		s = append(s, v.opts[i].String())
	}

	if len(s) == 0 {
		s = append(s, "none")
	}

	return fmt.Sprintf("AddValue( recordName: '%s', key: '%s', opts: %s )",
		v.recordName, v.key, strings.Join(s, ", "))
}
