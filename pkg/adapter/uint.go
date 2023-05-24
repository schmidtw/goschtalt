// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"github.com/goschtalt/goschtalt"
)

// UintUnmarshal converts a string to a uint/uint8/uint16/uint32/uint64/uintptr
// or a pointer version of them if possible, or returns an error indicating the
// failure.  Up to triple pointers are supported.
func UintUnmarshal() goschtalt.UnmarshalOption {
	return goschtalt.AdaptFromCfg(
		marshalNumber{
			typ: "uint",
		},
		"UintUnmarshal")
}

// MarshalUint converts a uint/uint8/uint16/uint32/uint64/uintptr into its
// string configuration form.
func MarshalUint() goschtalt.ValueOption {
	return goschtalt.AdaptToCfg(
		marshalNumber{
			typ: "uint",
		},
		"MarshalUint")
}
