// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"github.com/schmidtw/goschtalt/pkg/decoder"
	"github.com/schmidtw/goschtalt/pkg/encoder"
)

// CompileNow instructs New() to also compile the configuration if successful
// up to that point.  The error could be from creating the Config object or
// from the call to Compile().
func CompileNow() Option {
	return func(c *Config) error {
		c.compileNow = true
		return nil
	}
}

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

// NoDefaults provides a way to explicitly not use any preconfigured default
// values and instead use just the ones specified as options.
func NoDefaults() Option {
	return func(c *Config) error {
		c.ignoreDefaults = true
		return nil
	}
}

// ExpandVarsOpts controls how variables are identified and processed.
type ExpandVarsOpts struct {
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

// ExpandVars provides a way to expand variables in values throughout the
// configuration tree.  ExpandVars() can be called multiple times to expand
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
//	ExpandVars(&ExpandVarsOpts{Mapper: os.Getenv})
func ExpandVars(cfg *ExpandVarsOpts) Option {
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
