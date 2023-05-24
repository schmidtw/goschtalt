// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"github.com/goschtalt/goschtalt"
)

// IntUnmarshal converts a string to a int/int8/int16/int32/int64 or a pointer
// version of them if possible, or returns an error indicating the failure.  Up
// to triple pointers are supported.
func IntUnmarshal() goschtalt.UnmarshalOption {
	return goschtalt.AdaptFromCfg(
		marshalBuiltin{
			typ: "int",
		},
		"IntUnmarshal")
}

// MarshalInt converts a int/int8/int16/int32/int64 into its string
// configuration form.
func MarshalInt() goschtalt.ValueOption {
	return goschtalt.AdaptToCfg(
		marshalBuiltin{
			typ: "int",
		},
		"MarshalInt")
}
