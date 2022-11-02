// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

// json package is a goschtalt decoder package.
//
// The json package automatically registers the decoder as a default decoder
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
//		_ "github.com/schmidtw/goschtalt/extensions/decoders/json"
//	)
//
// See the example for how to use this extension package.
package json

import (
	"bytes"
	stljson "encoding/json"

	"github.com/schmidtw/goschtalt"
	"github.com/schmidtw/goschtalt/pkg/decoder"
	"github.com/schmidtw/goschtalt/pkg/meta"
)

var _ decoder.Decoder = (*Decoder)(nil)

// Use init to automatically wire this encoder as one available for goschtalt
// simply by including this package.
func init() {
	var d Decoder
	goschtalt.DefaultOptions = append(goschtalt.DefaultOptions, goschtalt.WithDecoder(d))
}

// Decoder is a class for the json decoder.
type Decoder struct{}

// Extensions returns the supported extensions.
func (d Decoder) Extensions() []string {
	return []string{"json"}
}

// Decode decodes a byte arreay into the meta.Object tree.
func (d Decoder) Decode(ctx decoder.Context, b []byte, m *meta.Object) error {
	var raw map[string]any

	if len(b) == 0 {
		*m = meta.Object{}
		return nil
	}

	dec := stljson.NewDecoder(bytes.NewBuffer(b))
	dec.UseNumber()
	if err := dec.Decode(&raw); err != nil {
		return err
	}

	objs := meta.ObjectFromRaw(raw)

	// TODO use the stream processor to find the line and column information.
	*m = annotate(ctx.Filename, objs)
	return nil
}

func annotate(filename string, obj meta.Object) meta.Object {
	kind := obj.Kind()

	origin := meta.Origin{
		File: filename,
	}

	if kind == meta.Value {
		obj.Origins = []meta.Origin{origin}
		return obj
	}

	if kind == meta.Array {
		for i := range obj.Array {
			val := obj.Array[i]
			obj.Array[i] = annotate(filename, val)
		}
		obj.Origins = []meta.Origin{origin}
		return obj

	}

	// Map
	obj.Origins = []meta.Origin{origin}
	for k, v := range obj.Map {
		obj.Map[k] = annotate(filename, v)
	}

	return obj
}
