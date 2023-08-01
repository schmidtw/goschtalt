// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"fmt"
	"time"

	"github.com/goschtalt/goschtalt/internal/print"
)

// Marshal renders the into the format specified ('json', 'yaml' or other extensions
// the Codecs provide and if adding comments should be attempted.  If a format
// does not support comments, an error is returned.  The result of the call is
// a slice of bytes with the information rendered into it.
//
// Valid Option Types:
//   - [GlobalOption]
//   - [MarshalOption]
func (c *Config) Marshal(opts ...MarshalOption) ([]byte, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.compiledAt.Equal(time.Time{}) {
		return nil, ErrNotCompiled
	}

	var cfg marshalOptions
	exts := c.opts.encoders.extensions()
	if len(exts) > 0 {
		cfg.format = exts[0]
	}

	full := append(c.opts.marshalOptions, opts...)
	for _, opt := range full {
		if opt != nil {
			if err := opt.marshalApply(&cfg); err != nil {
				return nil, err
			}
		}
	}

	tree := c.tree
	if cfg.redactSecrets {
		tree = tree.ToRedacted()
	}

	// Issue 52 - depending on encoders, they may encode a nil or null object
	// instead of returning an expected empty array of bytes.
	if tree.IsEmpty() {
		return []byte{}, nil
	}

	enc, err := c.opts.encoders.find(cfg.format)
	if err != nil {
		return nil, err
	}

	if cfg.withOrigins {
		return enc.EncodeExtended(tree)
	}

	return enc.Encode(tree.ToRaw())
}

// ---- MarshalOption options follow -------------------------------------------

// MarshalOption provides specific configuration for the process of producing
// a document based on the present information in the goschtalt object.
type MarshalOption interface {
	fmt.Stringer

	// marshalApply applies the options to the Marshal function.
	marshalApply(*marshalOptions) error
}

type marshalOptions struct {
	redactSecrets bool
	withOrigins   bool
	format        string
}

// RedactSecrets enables the replacement of secret portions of the tree with
// REDACTED.  Passing a redact value of false disables this behavior.
//
// The unused bool value is optional & assumed to be `true` if omitted.  The
// first specified value is used if provided.  A value of `false` disables the
// option.
//
// # Default
//
// Secret values are redacted.
func RedactSecrets(redact ...bool) MarshalOption {
	redact = append(redact, true)
	return redactSecretsOption(redact[0])
}

type redactSecretsOption bool

func (r redactSecretsOption) marshalApply(opts *marshalOptions) error {
	opts.redactSecrets = bool(r)
	return nil
}

func (r redactSecretsOption) String() string {
	return print.P("RedactSecrets", print.BoolSilentTrue(bool(r)), print.SubOpt())
}

// IncludeOrigins enables or disables providing the origin for each configuration
// value present.
//
// # Default
//
// Origins are not included by default.
func IncludeOrigins(origins ...bool) MarshalOption {
	origins = append(origins, true)
	return includeOriginsOption(origins[0])
}

type includeOriginsOption bool

func (w includeOriginsOption) marshalApply(opts *marshalOptions) error {
	opts.withOrigins = bool(w)
	return nil
}

func (i includeOriginsOption) String() string {
	return print.P("IncludeOrigins", print.BoolSilentTrue(bool(i)), print.SubOpt())
}

// FormatAs specifies the final document format extension to use when performing
// the operation.
func FormatAs(extension string) MarshalOption {
	return formatAsOption(extension)
}

type formatAsOption string

func (f formatAsOption) marshalApply(opts *marshalOptions) error {
	opts.format = string(f)
	return nil
}

func (f formatAsOption) String() string {
	return print.P("FormatAs", print.String(string(f)), print.SubOpt())
}
