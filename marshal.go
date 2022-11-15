// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"fmt"
	"time"
)

// Marshal renders the into the format specified ('json', 'yaml' or other extensions
// the Codecs provide and if adding comments should be attempted.  If a format
// does not support comments, an error is returned.  The result of the call is
// a slice of bytes with the information rendered into it.
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
	for _, opt := range opts {
		if opt != nil {
			opt.marshalApply(&cfg)
		}
	}

	tree := c.tree
	if cfg.redactSecrets {
		tree = tree.ToRedacted()
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
	marshalApply(*marshalOptions)
}

type marshalOptions struct {
	redactSecrets bool
	withOrigins   bool
	format        string
}

// RedactSecrets enables the replacement of secret portions of the tree with
// REDACTED.  Passing a redact value of false disables this behavior.  The
// default behavior is to redact secrets.
func RedactSecrets(redact ...bool) MarshalOption {
	redact = append(redact, true)
	return redactSecretsOption(redact[0])
}

type redactSecretsOption bool

func (r redactSecretsOption) marshalApply(opts *marshalOptions) {
	opts.redactSecrets = bool(r)
}

func (r redactSecretsOption) String() string {
	if bool(r) {
		return "RedactSecrets()"
	}

	return "RedactSecrets(false)"
}

// IncludeOrigins enables or disables providing the origin for each configuration
// value present.  The default behavior is not to include origins.
func IncludeOrigins(origins ...bool) MarshalOption {
	origins = append(origins, true)
	return includeOriginsOption(origins[0])
}

type includeOriginsOption bool

func (w includeOriginsOption) marshalApply(opts *marshalOptions) {
	opts.withOrigins = bool(w)
}

func (i includeOriginsOption) String() string {
	if bool(i) {
		return "IncludeOrigins()"
	}

	return "IncludeOrigins(false)"
}

// FormatAs specifies the final document format extension to use when performing
// the operation.
func FormatAs(extension string) MarshalOption {
	return formatAsOption(extension)
}

type formatAsOption string

func (f formatAsOption) marshalApply(opts *marshalOptions) {
	opts.format = string(f)
}

func (f formatAsOption) String() string {
	return "FormatAs('" + string(f) + "')"
}
