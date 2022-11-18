// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"github.com/goschtalt/goschtalt/pkg/decoder"
	"github.com/goschtalt/goschtalt/pkg/meta"
)

// record is the basic unit needed to define a configuration and it's name.
// With this information all the records can be decoded.
type record struct {
	name    string
	val     *value
	encoded *encodedBuffer
	tree    meta.Object
}

// fetch normalizes the calls to the val or encoded types of records.
func (rec *record) fetch(delimiter string, umf UnmarshalFunc, decoders *codecRegistry[decoder.Decoder], defaultOpts []ValueOption) error {
	if rec.val != nil {
		tree, err := rec.val.toTree(delimiter, umf, defaultOpts...)
		if err != nil {
			return err
		}
		rec.tree = tree
	}

	if rec.encoded != nil {
		tree, err := rec.encoded.toTree(delimiter, umf, decoders)
		if err != nil {
			return err
		}
		rec.tree = tree
	}

	return nil
}
