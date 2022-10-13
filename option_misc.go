// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"github.com/schmidtw/goschtalt/pkg/decoder"
	"github.com/schmidtw/goschtalt/pkg/encoder"
)

// AutoCompile instructs New() to also compile the configuration if successful
// up to that point.  The error could be from creating the Config object or
// from the call to Compile().
func AutoCompile() Option {
	return func(c *Config) error {
		c.autoCompile = true
		return nil
	}
}

// RegisterDecoder registers a Decoder for the specific file extensions provided.
// Attempting to register a duplicate extension is not supported.
func RegisterDecoder(d decoder.Decoder) Option {
	return func(c *Config) error {
		return c.decoders.register(d)
	}
}

// RemoveDecoder provides a mechanism for removing the decoders from use for
// specific file types.
func RemoveDecoder(exts ...string) Option {
	return func(c *Config) error {
		c.decoders.deregister(exts...)
		return nil
	}
}

// RegisterEncoder registers a Encoder for the specific file extensions provided.
// Attempting to register a duplicate extension is not supported.
func RegisterEncoder(d encoder.Encoder) Option {
	return func(c *Config) error {
		return c.encoders.register(d)
	}
}

// RemoveEncoder provides a mechanism for removing the encoders from use for
// specific file types.
func RemoveEncoder(exts ...string) Option {
	return func(c *Config) error {
		c.encoders.deregister(exts...)
		return nil
	}
}

// AddFileGroup provides a group of files, directories or both to examine for
// configuration files.
func AddFileGroup(group Group) Option {
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

// NoDefaults provides a way to explicitly not use any preconfigured default
// values and instead use just the ones specified as options.
func NoDefaults() Option {
	return func(c *Config) error {
		c.ignoreDefaults = true
		return nil
	}
}

// Expand controls how variables are identified and processed.
type Expand struct {
	// Optional name showing where the value came from.
	Name string

	// The string that prefixes a variable.  "${{" or "${" are common examples.
	// Defaults to "${" if equal to "".
	Start string

	// The string that trails a variable.  "}}" or "}" are common examples.
	// Defaults to "}" if equal to "".
	End string

	// The string to string mapping function.
	// Mapping request ignored if nil.
	Mapper func(string) string

	// The maximum expansions of a value before a recursion error is returned.
	// Defaults to 10000 if set to less than 1.
	Maximum int
}

// AddExpansion provides a way to expand variables in values throughout the
// configuration tree.  AddExpansion() can be called multiple times to expand
// variables based on additional configurations and mappers.  To remove all
// expansion pass in nil for the cfg parameter.
//
// The initial discovery of a variable to expand in the configuration tree
// value is determined by the Start and End delimiters provided. Further
// expansions of values replaces ${var} or $var in the string based on the
// mapping function provided.
//
// Expansions are done in the order specified.
//
// Here is an example of how to expand environment variables:
//
//	AddExpansion(&Expand{Mapper: os.Getenv})
func AddExpansion(cfg *Expand) Option {
	return func(c *Config) error {
		if cfg == nil {
			c.expansions = nil
			return nil
		}

		exp := *cfg

		if len(exp.Start) == 0 {
			exp.Start = "${"
		}
		if len(exp.End) == 0 {
			exp.End = "}"
		}
		if exp.Maximum < 1 {
			exp.Maximum = 10000
		}
		if exp.Mapper != nil {
			c.expansions = append(c.expansions, exp)
		}
		return nil
	}
}
