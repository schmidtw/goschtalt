// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"github.com/goschtalt/goschtalt"
)

// FloatUnmarshal converts a string to a float32/float64 or *float32/*float64 if
// possible, or returns an error indicating the failure.
func FloatUnmarshal() goschtalt.UnmarshalOption {
	return goschtalt.AdaptFromCfg(
		marshalNumber{
			typ: "float",
		},
		"FloatUnmarshal")
}

// MarshalFloat converts a float32/float64 into its string configuration form.
func MarshalFloat() goschtalt.ValueOption {
	return goschtalt.AdaptToCfg(
		marshalNumber{
			typ: "float",
		},
		"MarshalFloat")
}
