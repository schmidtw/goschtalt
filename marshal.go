// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

type encoderOptions struct {
	redactSecrets bool
	withOrigins   bool
	format        string
}

// MarshalOption provides specific configuration for the process of producing
// a document based on the present information in the goschtalt object.
type MarshalOption func(*encoderOptions)

// RedactSecrets enables or disables the replacement of secret portions of the
// tree with REDACTED.
func RedactSecrets(redact bool) MarshalOption {
	return func(e *encoderOptions) {
		e.redactSecrets = redact
	}
}

// IncludeOrigins enables or disables providing the origin for each configuration
// value present.
func IncludeOrigins(origins bool) MarshalOption {
	return func(e *encoderOptions) {
		e.withOrigins = origins
	}
}

// UseFormat specifies the final document format to use when performing the
// operation.
func UseFormat(extension string) MarshalOption {
	return func(e *encoderOptions) {
		e.format = extension
	}
}

// Marshal renders the into the format specified ('json', 'yaml' or other extensions
// the Codecs provide and if adding comments should be attempted.  If a format
// does not support comments, an error is returned.  The result of the call is
// a slice of bytes with the information rendered into it.
func (c *Config) Marshal(opts ...MarshalOption) ([]byte, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if !c.hasBeenCompiled {
		return nil, ErrNotCompiled
	}

	var cfg encoderOptions
	exts := c.encoders.extensions()
	if len(exts) > 0 {
		cfg.format = exts[0]
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	tree := c.tree
	if cfg.redactSecrets {
		tree = tree.ToRedacted()
	}

	if cfg.withOrigins {
		return c.encoders.encodeExtended(cfg.format, tree)
	}

	return c.encoders.encode(cfg.format, tree.ToRaw())
}
