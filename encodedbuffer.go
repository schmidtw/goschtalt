// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/schmidtw/goschtalt/pkg/decoder"
	"github.com/schmidtw/goschtalt/pkg/meta"
)

// AddBuffer adds a buffer of bytes for inclusion when compiling the configuration.
// The format of the bytes is determined by the extension of the recordName field.
// The recordName field is also used for sorting this configuration value relative
// to other configuration values.
func AddBuffer(recordName string, in []byte) Option {
	bytesText := "[]byte"
	if in == nil {
		bytesText = "nil"
	}
	return &encodedBuffer{
		text:       fmt.Sprintf("AddBuffer( '%s', %s )", recordName, bytesText),
		recordName: recordName,
		fn: func(_ string, _ UnmarshalFunc) ([]byte, error) {
			return in, nil
		},
		//return io.NopCloser(bytes.NewReader(in)), nil
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
func AddBufferFn(recordName string, fn func(recordName string, un UnmarshalFunc) ([]byte, error)) Option {
	if fn == nil {
		return &encodedBuffer{
			text:       fmt.Sprintf("AddBufferFn( '%s', '' )", recordName),
			recordName: recordName,
		}
	}

	return &encodedBuffer{
		text:       fmt.Sprintf("AddBufferFn( '%s', custom )", recordName),
		recordName: recordName,
		fn: func(name string, un UnmarshalFunc) ([]byte, error) {
			return fn(name, un)
		},
	}
}

type encodedBuffer struct {
	// The text to use when String() is called.
	text string

	// The red
	recordName string

	// The fn to use to get the value.
	fn func(recordName string, unmarshal UnmarshalFunc) ([]byte, error)
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
	opts.values = append(opts.values, r)
	return nil
}

func (_ encodedBuffer) ignoreDefaults() bool { return false }
func (eb encodedBuffer) String() string      { return eb.text }

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
