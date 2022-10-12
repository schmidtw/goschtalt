// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

// yaml package is a goschtalt decoder package.
//
// The yaml package automatically registers the decoder as a default decoder
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
//		_ "github.com/schmidtw/goschtalt/extensions/decoders/yaml"
//	)
//
// See the example for how to use this extension package.
package yaml

import (
	"github.com/schmidtw/goschtalt"
	"github.com/schmidtw/goschtalt/pkg/decoder"
	"github.com/schmidtw/goschtalt/pkg/meta"
	yml "gopkg.in/yaml.v3"
)

var _ decoder.Decoder = (*Decoder)(nil)

// Use init to automatically wire this encoder as one available for goschtalt
// simply by including this package.
func init() {
	var d Decoder
	goschtalt.DefaultOptions = append(goschtalt.DefaultOptions, goschtalt.RegisterDecoder(d))
}

// Decoder is a class for the yaml decoder.
type Decoder struct{}

// Extensions returns the supported extensions.
func (d Decoder) Extensions() []string {
	return []string{"yaml", "yml"}
}

// Decode decodes a byte arreay into the meta.Object tree.
func (d Decoder) Decode(ctx decoder.Context, b []byte, m *meta.Object) error {
	var raw map[string]any

	// Use the official unmarshal() function to prevent introducing defects
	// here since the yml.Node tree requires us to in effect do half of the
	// structure processing.  What caused me to stop was the need to
	// re-implement the anchor/alias/merge logic.  By doing it this way, if
	// there is a bug, the bug is in the line number resolution & not the
	// actual configuration.
	if err := yml.Unmarshal(b, &raw); err != nil {
		return err
	}

	if len(raw) == 0 {
		*m = meta.Object{}
		return nil
	}

	objs := meta.ObjectFromRaw(raw)

	// Now let's try to annotate our tree by examining the yml.Node tree.
	var node yml.Node
	// This has been decoded once without issue, it can't error.
	_ = yml.Unmarshal(b, &node)

	*m = annotate(0, ctx.Filename, objs, node, node)
	return nil
}

func annotate(level int, filename string, obj meta.Object, prev, node yml.Node) meta.Object {
	kind := obj.Kind()

	origin := meta.Origin{
		File: filename,
		Line: node.Line,
		Col:  node.Column,
	}

	if node.Kind == yml.DocumentNode {
		node = *node.Content[0]
	}

	if node.Kind == yml.AliasNode {
		node = *node.Alias
	}

	if kind == meta.Value {
		obj.Origins = []meta.Origin{origin}
		return obj
	}

	if kind == meta.Array {
		if node.Kind != yml.SequenceNode {
			return obj
		}
		for i := range obj.Array {
			val := obj.Array[i]
			obj.Array[i] = annotate(level+1, filename, val, *node.Content[i], *node.Content[i])
		}

		origin.Line = prev.Line
		origin.Col = prev.Column
		obj.Origins = []meta.Origin{origin}
		return obj

	}

	// Map
	obj.Origins = []meta.Origin{
		{
			File: filename,
			Line: prev.Line,
			Col:  prev.Column,
		},
	}

	if node.Kind != yml.MappingNode {
		return obj
	}

	keyMap := make(map[string]yml.Node)
	nodeMap := make(map[string]yml.Node)
	for i := 0; i < len(node.Content); i += 2 {
		key := *node.Content[i]
		val := *node.Content[i+1]
		if key.Kind != yml.ScalarNode {
			return obj
		}
		if key.Tag == "!!merge" {
			// TODO Handle merged data better.  This appears to get complicated
			// because this library needs to replicate what the yaml decode is
			// doing to do it right.  It's not clear how to do that without
			// creating a large potential for bugs here.
		} else {
			keyMap[key.Value] = key
			nodeMap[key.Value] = val
		}
	}

	for k, v := range obj.Map {
		if _, found := nodeMap[k]; !found {
			continue
		}
		obj.Map[k] = annotate(level+1, filename, v, keyMap[k], nodeMap[k])
	}

	return obj
}
