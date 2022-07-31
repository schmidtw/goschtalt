// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import "github.com/schmidtw/goschtalt/internal/encoding"

// WithCodec registers a Codec for the specific file extensions provided.
// Attempting to register a duplicate extension is not supported.
func WithCodec(enc encoding.Codec) Option {
	return func(g *Goschtalt) error {
		opt := encoding.WithCodec(enc)
		return g.codecs.Options(opt)
	}
}

// WithoutExtensions provides a mechanism for effectively removing the codecs
// from use for specific file types.
func WithoutExtensions(exts ...string) Option {
	return func(g *Goschtalt) error {
		opt := encoding.WithoutExtensions(exts...)
		return g.codecs.Options(opt)
	}
}

// WithFileGroup provides a group of files, directories or both to examine for
// configuration files.
func WithFileGroup(group Group) Option {
	return func(g *Goschtalt) error {
		g.groups = append(g.groups, group)
		return nil
	}
}
