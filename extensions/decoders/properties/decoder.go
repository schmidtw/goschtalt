// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

// properties package is a goschtalt decoder package.
//
// The properties package automatically registers the decoder as a default decoder
// with the goschtalt package so the usage is as simple as possible ... simply
// import the package and it should just work.
//
// Import the package like you do for pprof - like this:
//
//	import (
//		"fmt"
//		"os"
//		...
//
//		"github.com/schmidtw/goschtalt"
//		_ "github.com/schmidtw/goschtalt/extensions/decoders/properties"
//	)
//
// See the example for how to use this extension package.
package properties

import (
	"strings"

	"github.com/magiconair/properties"
	"github.com/schmidtw/goschtalt"
	"github.com/schmidtw/goschtalt/pkg/decoder"
	"github.com/schmidtw/goschtalt/pkg/meta"
)

var _ decoder.Decoder = (*Decoder)(nil)

// Use init to automatically wire this encoder as one available for goschtalt
// simply by including this package.
func init() {
	var d Decoder
	goschtalt.DefaultOptions = append(goschtalt.DefaultOptions, goschtalt.DecoderRegister(d))
}

// Decoder is a class for the property decoder.
type Decoder struct{}

// Extensions returns the supported extensions.
func (d Decoder) Extensions() []string {
	return []string{"properties"}
}

// Decode decodes a byte arreay into the meta.Object tree.
func (d Decoder) Decode(filename, keyDelimiter string, b []byte, m *meta.Object) error {
	props, err := properties.Load(b, properties.UTF8)
	if err != nil {
		return err
	}
	lines := strings.Split(string(b), "\n")

	keys := props.Keys()
	//sort.Strings(keys)

	if len(keys) == 0 {
		return nil
	}

	tree := meta.Object{
		Origins: []meta.Origin{{
			File: filename,
			Line: 1,
			Col:  1,
		}},
		Map: make(map[string]meta.Object),
	}
	for _, key := range keys {
		origin := meta.Origin{
			File: filename,
			Col:  1,
		}
		for i, line := range lines {
			for _, c := range []string{"=", ":", " ", "\t", "\f"} {
				k := key + c
				if strings.HasPrefix(strings.TrimSpace(line), k) {
					origin.Line = i + 1
					break
				}
			}
		}

		val, _ := props.Get(key)
		tree = tree.Add(keyDelimiter, key, val, origin)
	}

	*m = tree.ConvertMapsToArrays()
	return nil
}
