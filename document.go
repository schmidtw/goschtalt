// SPDX-FileCopyrightText: 2025 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"fmt"
	"sort"
	"strings"

	"github.com/goschtalt/goschtalt/internal/encoding"
	"github.com/goschtalt/goschtalt/internal/encoding/yaml"
	"github.com/goschtalt/goschtalt/pkg/meta"
	"github.com/k0kubun/pp"
)

var documentFormats = map[string]encoding.Encoder{
	"yml": &yaml.Renderer{
		MaxLineLength:         80,
		TrailingCommentColumn: 80,
		SpacesPerIndent:       2,
	},
}

var documentTypes = map[string]struct{}{
	"full":     {},
	"defaults": {},
	"provided": {},
}

func (c *Config) document(format, typ string) (string, error) {
	enc := documentFormats[format]
	if enc == nil {
		fmts := make([]string, 0, len(documentFormats))
		for k := range documentFormats {
			fmts = append(fmts, k)
		}
		sort.Strings(fmts)
		return "", fmt.Errorf("unknown document format %q", strings.Join(fmts, ", "))
	}

	if _, ok := documentTypes[typ]; !ok {
		types := make([]string, 0, len(documentTypes))
		for k := range documentTypes {
			types = append(types, k)
		}
		sort.Strings(types)
		return "", fmt.Errorf("unknown document type %q", strings.Join(types, ", "))
	}

	if typ == "full" {
		return c.docFull(enc)
	}
	return "", nil
}

func (c *Config) docFull(enc encoding.Encoder) (string, error) {
	var defaults meta.Object
	var out strings.Builder

	// Fetch the defaults.
	{
		results, err := c.compileInternal(true)
		if err != nil {
			return "", err
		}

		defaults = results.merged
	}

	if &defaults == nil {
	}

	d := c.opts.doc

	got, err := calcUnified(&d, &c.tree)
	if err != nil {
		return "", fmt.Errorf("failed to calculate unified document: %w", err)
	}

	fmt.Println("")
	fmt.Println("")
	fmt.Println("")
	fmt.Println("=======================================================")
	fmt.Println("")
	fmt.Println("")
	fmt.Println("")

	// Print the unified document for debugging.
	pp.Println("unified document:", got)
	fmt.Println("")
	fmt.Println("")
	fmt.Println("")
	fmt.Println("--------------------------------------------------------")
	fmt.Println("")
	fmt.Println("")
	fmt.Println("")

	err = enc.Encode(&out, &got)
	/*
		fmt.Fprintln(&out, "---")
		for _, key := range got.childrenKeys() {
			child := got.children[key]
			err := child.toYML(&out, []string{key}, 0, false)
			if err != nil {
				return "", fmt.Errorf("failed to write unified document: %w", err)
			}
		}
		//if err := fullYaml(&out, nil, d, defaults, c.tree, 0); err != nil {
		//return "", fmt.Errorf("failed to write full YAML: %w", err)
		//}
	*/
	return out.String(), nil
}
