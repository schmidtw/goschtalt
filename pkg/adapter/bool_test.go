// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"testing"

	"github.com/goschtalt/goschtalt"
)

func TestBoolValueAdapterInternals(t *testing.T) {
	tests := []valueAdapterTest{
		{
			description: "bool - true",
			from:        bool(true),
			obj:         marshalBool{},
			expect:      "true",
		}, {
			description: "*bool - true",
			from:        toPtr(true),
			obj:         marshalBool{},
			expect:      "true",
		}, {
			description: "bool - false",
			from:        bool(false),
			obj:         marshalBool{},
			expect:      "false",
		}, {
			description: "*bool - false",
			from:        toPtr(false),
			obj:         marshalBool{},
			expect:      "false",
		}, {
			description: "bool - invalid",
			from:        "string",
			obj:         marshalBool{},
			expectErr:   goschtalt.ErrNotApplicable,
		},
	}

	testValueAdapters(t, tests)
}

func TestBoolUnmarshalAdapterInternals(t *testing.T) {
	tests := []unmarshalAdapterTest{
		{
			description: "marshalBool, a valid string",
			from:        "true",
			to:          true,
			obj:         marshalBool{},
			expect:      true,
		}, {
			description: "marshalBool, a valid string to a pointer",
			from:        "true",
			to:          new(bool),
			obj:         marshalBool{},
			expect:      toPtr(true),
		}, {
			description: "marshalBool, not a string",
			from:        12,
			to:          true,
			obj:         marshalBool{},
			expectErr:   goschtalt.ErrNotApplicable,
		}, {
			description: "marshalBool, not a bool",
			from:        "true",
			to:          "string",
			obj:         marshalBool{},
			expectErr:   goschtalt.ErrNotApplicable,
		}, {
			description: "marshalBool, not a valid bool option",
			from:        "invalid",
			to:          true,
			obj:         marshalBool{},
			expectErr:   goschtalt.ErrNotApplicable,
		},
	}

	testUnmarshalAdapters(t, tests)
}
