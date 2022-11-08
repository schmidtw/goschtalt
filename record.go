// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/schmidtw/goschtalt/pkg/decoder"
	"github.com/schmidtw/goschtalt/pkg/meta"
)

// record is the basic unit needed to define a configuration and it's name.
// With this information all the records can be decoded.
type record struct {
	name          string
	bufFetcher    func(string, UnmarshalFunc) (io.ReadCloser, error)
	structFetcher func(string, UnmarshalFunc) (any, error)
	tree          meta.Object
}

func (rec record) keep(decoders *codecRegistry[decoder.Decoder]) bool {
	if rec.structFetcher != nil {
		// Keep because no decoder is needed.
		return true
	}

	// Keep if we have a decoder.
	_, err := decoders.find(strings.TrimPrefix(filepath.Ext(rec.name), "."))
	if err == nil {
		return true
	}

	return false
}

func (rec record) getData() ([]byte, error) {
	if rec.bufFetcher == nil {
		return nil, fmt.Errorf("%w no function to acquire data provided", ErrInvalidInput)
	}

	stream, err := rec.bufFetcher(rec.name)
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(stream)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (rec *record) decode(delimiter string, um UnmarshalFunc, decoders *codecRegistry[decoder.Decoder], valOpts []ValueOption) error {
	rec.tree = meta.Object{}

	data, err := rec.getData()
	if err != nil {
		return err
	}

	ctx := decoder.Context{
		Filename:  rec.name,
		Delimiter: delimiter,
	}

	dec, err := decoders.find(rec.ext())
	if err != nil {
		return err
	}

	var m meta.Object
	err = dec.Decode(ctx, data, &m)
	if err != nil {
		err = fmt.Errorf("decoder error for extension '%s' processing file '%s' %w %v",
			rec.ext(), rec.name, ErrDecoding, err)

		return err
	}

	rec.tree = m

	return nil
}

func filterRecords(records []record, decoders *codecRegistry[decoder.Decoder]) []record {
	rv := make([]record, 0, len(records))
	for _, rec := range records {
		if rec.keep(decoders) {
			rv = append(rv, rec)
		}
	}

	return rv
}
