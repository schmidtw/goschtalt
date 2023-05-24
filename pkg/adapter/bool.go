// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"github.com/goschtalt/goschtalt"
)

// BoolUnmarshal converts a string to a bool or *bool if possible, or returns an
// error indicating the failure.  The case is ignored, but only the following
// values are accepted: 'f', 'false', 't', 'true'
func BoolUnmarshal() goschtalt.UnmarshalOption {
	return goschtalt.AdaptFromCfg(
		marshalBuiltin{
			typ: "bool",
		},
		"BoolUnmarshal")
}

// MarshalBool converts a bool into its configuration form.  The
// configuration form is a string of value 'true' or 'false'.
func MarshalBool() goschtalt.ValueOption {
	return goschtalt.AdaptToCfg(
		marshalBuiltin{
			typ: "bool",
		},
		"MarshalBool")
}
