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
			obj:         marshalBuiltin{typ: "bool"},
			expect:      "true",
		}, {
			description: "*bool - true",
			from:        toPtr(true),
			obj:         marshalBuiltin{typ: "bool"},
			expect:      "true",
		}, {
			description: "bool - false",
			from:        bool(false),
			obj:         marshalBuiltin{typ: "bool"},
			expect:      "false",
		}, {
			description: "*bool - false",
			from:        toPtr(false),
			obj:         marshalBuiltin{typ: "bool"},
			expect:      "false",
		}, {
			description: "bool - invalid",
			from:        "string",
			obj:         marshalBuiltin{typ: "bool"},
			expectErr:   goschtalt.ErrNotApplicable,
		},
	}

	testValueAdapters(t, tests)
}

func TestBoolUnmarshalAdapterInternals(t *testing.T) {
	tests := []unmarshalAdapterTest{
		{
			description: "a valid string",
			from:        "true",
			to:          true,
			obj:         marshalBuiltin{typ: "bool"},
			expect:      true,
		}, {
			description: "a valid string to a pointer",
			from:        "true",
			to:          new(bool),
			obj:         marshalBuiltin{typ: "bool"},
			expect:      toPtr(true),
		}, {
			description: "not a string",
			from:        12,
			to:          true,
			obj:         marshalBuiltin{typ: "bool"},
			expectErr:   goschtalt.ErrNotApplicable,
		}, {
			description: "not a bool",
			from:        "true",
			to:          int(0),
			obj:         marshalBuiltin{typ: "bool"},
			expectErr:   goschtalt.ErrNotApplicable,
		}, {
			description: "not a valid bool option",
			from:        "invalid",
			to:          true,
			obj:         marshalBuiltin{typ: "bool"},
			expectErr:   goschtalt.ErrNotApplicable,
		},
	}

	testUnmarshalAdapters(t, tests)
}
