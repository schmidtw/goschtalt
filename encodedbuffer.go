// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/goschtalt/goschtalt/internal/print"
	"github.com/goschtalt/goschtalt/pkg/decoder"
	"github.com/goschtalt/goschtalt/pkg/meta"
)

// AddBuffer adds a buffer of bytes for inclusion when compiling the configuration.
// The format of the bytes is determined by the extension of the recordName field.
// The recordName field is also used for sorting this configuration value relative
// to other configuration values.
func AddBuffer(recordName string, in []byte, opts ...BufferOption) Option {
	return &encodedBuffer{
		text:       print.P("AddBuffer", print.String(recordName), print.Bytes(in), print.LiteralStringers(opts)),
		recordName: recordName,
		fn: func(_ string, _ UnmarshalFunc) ([]byte, error) {
			return in, nil
		},
		opts: opts,
	}
}

// AddBufferFn adds a function that is called during compile time of the
// configuration.  The recordName of this record is passed into the fn function
// that is called as well as an UnmarshalFunc that represents the existing state
// of the merged configuration prior to adding the buffer that results in the
// call to fn.
//
// The format of th ebytes is determined by the extension of the recordName field.
// The recordName field is also used for sorting this configuration value relative
// to other configuration values.
func AddBufferFn(recordName string, fn func(recordName string, un UnmarshalFunc) ([]byte, error), opts ...BufferOption) Option {
	rv := encodedBuffer{
		text:       print.P("AddBufferFn", print.String(recordName), print.Fn(fn), print.LiteralStringers(opts)),
		recordName: recordName,
		opts:       opts,
	}

	if fn != nil {
		rv.fn = func(name string, un UnmarshalFunc) ([]byte, error) {
			return fn(name, un)
		}
	}

	return &rv
}

type encodedBuffer struct {
	// The text to use when String() is called.
	text string

	// The record name.
	recordName string

	// The fn to use to get the value.
	fn func(recordName string, unmarshal UnmarshalFunc) ([]byte, error)

	// Options that configure how this buffer is treated and processed.
	// These options are in addition to any default settings set with
	// AddDefaultValueOptions().
	opts []BufferOption
}

func (eb encodedBuffer) apply(opts *options) error {
	if len(eb.recordName) == 0 {
		return fmt.Errorf("%w: a recordName with length > 0 must be specified.", ErrInvalidInput)
	}

	if eb.fn == nil {
		return fmt.Errorf("%w: a non-nil func must be specified.", ErrInvalidInput)
	}

	r := record{
		name:    eb.recordName,
		encoded: &eb,
	}

	for _, opt := range eb.opts {
		if opt.isDefault() {
			opts.defaults = append(opts.defaults, r)
			return nil
		}
	}

	opts.values = append(opts.values, r)
	return nil
}

func (_ encodedBuffer) ignoreDefaults() bool {
	return false
}

func (eb encodedBuffer) String() string {
	return eb.text
}

// toTree converts an encodedBuffer into a meta.Object tree.  This will happen
// during the compilation stage.
func (eb *encodedBuffer) toTree(delimiter string, umf UnmarshalFunc, decoders *codecRegistry[decoder.Decoder]) (meta.Object, error) {
	data, err := eb.fn(eb.recordName, umf)
	if err != nil {
		return meta.Object{}, err
	}

	ext := strings.TrimPrefix(filepath.Ext(eb.recordName), ".")

	dec, err := decoders.find(ext)
	if err != nil {
		return meta.Object{}, err
	}

	ctx := decoder.Context{
		Filename:  eb.recordName,
		Delimiter: delimiter,
	}

	var tree meta.Object
	err = dec.Decode(ctx, data, &tree)
	if err != nil {
		err = fmt.Errorf("decoder error for extension '%s' processing buffer '%s' %w %v",
			ext, eb.recordName, ErrDecoding, err)

		return meta.Object{}, err
	}

	return tree, nil
}

// -- BufferOption options follow ----------------------------------------------

// BufferOption provides the means to configure options for handling of the
// buffer configuration values.
type BufferOption interface {
	fmt.Stringer

	isDefault() bool
}
