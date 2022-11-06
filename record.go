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
	name string
	fn   func(recordName string) (io.ReadCloser, error)
	tree meta.Object
}

func (rec record) ext() string {
	return strings.TrimPrefix(filepath.Ext(rec.name), ".")
}

func (rec record) getData() ([]byte, error) {
	if rec.fn == nil {
		return nil, fmt.Errorf("%w no function to acquire data provided", ErrInvalidInput)
	}

	stream, err := rec.fn(rec.name)
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

func (rec *record) decode(decoders *codecRegistry[decoder.Decoder], keyDelimiter string) error {
	rec.tree = meta.Object{}

	data, err := rec.getData()
	if err != nil {
		return err
	}

	ctx := decoder.Context{
		Filename:  rec.name,
		Delimiter: keyDelimiter,
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
		_, err := decoders.find(rec.ext())
		if err == nil {
			rv = append(rv, rec)
		}
	}

	return rv
}
