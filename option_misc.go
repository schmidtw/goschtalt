// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"github.com/schmidtw/goschtalt/pkg/decoder"
	"github.com/schmidtw/goschtalt/pkg/encoder"
)

// DecoderRegister registers a Decoder for the specific file extensions provided.
// Attempting to register a duplicate extension is not supported.
func DecoderRegister(d decoder.Decoder) Option {
	return func(c *Config) error {
		return c.decoders.register(d)
	}
}

// DecoderRemove provides a mechanism for removing the decoders from use for
// specific file types.
func DecoderRemove(exts ...string) Option {
	return func(c *Config) error {
		c.decoders.deregister(exts...)
		return nil
	}
}

// EncoderRegister registers a Encoder for the specific file extensions provided.
// Attempting to register a duplicate extension is not supported.
func EncoderRegister(d encoder.Encoder) Option {
	return func(c *Config) error {
		return c.encoders.register(d)
	}
}

// EncoderRemove provides a mechanism for removing the encoders from use for
// specific file types.
func EncoderRemove(exts ...string) Option {
	return func(c *Config) error {
		c.encoders.deregister(exts...)
		return nil
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
