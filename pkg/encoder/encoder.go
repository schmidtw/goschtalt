// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package encoder

import "github.com/goschtalt/goschtalt/pkg/meta"

// Encoder provides the encoder interface for goschtalt to use.
type Encoder interface {
	// EncodeExtended provides a way to encode extra information, specifically
	// the origin information (filename, line number, column) as comments or
	// other metadata about the field along with the rest of the structure.
	EncodeExtended(m meta.Object) ([]byte, error)

	// Encode provides a way to encode a normal map[string]any style of output.
	// Nothing special is expected.
	Encode(v any) ([]byte, error)

	// Extensions provides the list of extensions this decoder is able to decode.
	Extensions() []string
}
