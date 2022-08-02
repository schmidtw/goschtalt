// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import "github.com/schmidtw/goschtalt/internal/encoding"

// Codec registers a Codec for the specific file extensions provided.
// Attempting to register a duplicate extension is not supported.
func Codec(enc encoding.Codec) Option {
	return func(c *Config) error {
		opt := encoding.DecoderEncoder(enc)
		return c.codecs.With(opt)
	}
}

// ExcludedExtensions provides a mechanism for effectively removing the codecs
// from use for specific file types.
func ExcludedExtensions(exts ...string) Option {
	return func(c *Config) error {
		opt := encoding.ExcludedExtensions(exts...)
		return c.codecs.With(opt)
	}
}

// FileGroup provides a group of files, directories or both to examine for
// configuration files.
func FileGroup(group Group) Option {
	return func(c *Config) error {
		c.groups = append(c.groups, group)
		return nil
	}
}

// KeyDelimiter provides the delimiter used for determining key parts.
func KeyDelimiter(delimiter string) Option {
	return func(c *Config) error {
		c.keyDelimiter = delimiter
		return nil
	}
}
