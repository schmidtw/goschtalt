// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package decoder

import "github.com/goschtalt/goschtalt/pkg/meta"

// Context is a way to pass additional information that the decoder may need
// access to in a more future proof way.
type Context struct {
	Filename  string // The filename (not full path) of the file being decoded.
	Delimiter string // The key delimiter string to use if needed.
}

// Decoder provides the decoder interface for goschtalt to use.
type Decoder interface {
	// Decode is called to decode a collection of bytes based on the formats
	// provided via the Extensions() function.  The filename is provided so it
	// can be included in the metadata about where the item came from.  The
	// keyDelimiter is provided in case the key needs to be split.
	//
	// A decoder should strive to provide as much information as it can about
	// where a value originated.  The filename, line number and column number
	// are very helpful when attempting to diagnose where a configuration
	// value originated from.
	Decode(ctx Context, b []byte, m *meta.Object) error

	// Extensions provides the list of extensions this decoder is able to decode.
	Extensions() []string
}
