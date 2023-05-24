// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"strconv"
	"testing"

	"github.com/goschtalt/goschtalt"
)

func TestUintValueAdapterInternals(t *testing.T) {
	tests := []valueAdapterTest{
		{
			description: "uint",
			from:        uint(1234),
			obj:         marshalBuiltin{typ: "uint"},
			expect:      "1234",
		}, {
			description: "*uint",
			from:        toPtr(uint(1234)),
			obj:         marshalBuiltin{typ: "uint"},
			expect:      "1234",
		}, {
			description: "**uint",
			from:        toPtr(toPtr(uint(1234))),
			obj:         marshalBuiltin{typ: "uint"},
			expect:      "1234",
		}, {
			description: "string",
			from:        "string",
			obj:         marshalBuiltin{typ: "uint"},
			expectErr:   goschtalt.ErrNotApplicable,
		}, {
			description: "*string",
			from:        toPtr("string"),
			obj:         marshalBuiltin{typ: "uint"},
			expectErr:   goschtalt.ErrNotApplicable,
		},
	}

	testValueAdapters(t, tests)
}

func TestUintUnmarshalAdapterInternals(t *testing.T) {
	tests := []unmarshalAdapterTest{
		{
			description: "uint(123)",
			from:        "123",
			to:          uint(0),
			obj:         marshalBuiltin{typ: "uint"},
			expect:      uint(123),
		}, {
			description: "uint8(123)",
			from:        "123",
			to:          uint8(0),
			obj:         marshalBuiltin{typ: "uint"},
			expect:      uint8(123),
		}, {
			description: "uint16(123)",
			from:        "123",
			to:          uint16(0),
			obj:         marshalBuiltin{typ: "uint"},
			expect:      uint16(123),
		}, {
			description: "uint32(123)",
			from:        "123",
			to:          uint32(0),
			obj:         marshalBuiltin{typ: "uint"},
			expect:      uint32(123),
		}, {
			description: "uint64(123)",
			from:        "123",
			to:          uint64(0),
			obj:         marshalBuiltin{typ: "uint"},
			expect:      uint64(123),
		}, {
			description: "uintptr(123)",
			from:        "123",
			to:          uintptr(0),
			obj:         marshalBuiltin{typ: "uint"},
			expect:      uintptr(123),
		}, {
			description: "*uint(123)",
			from:        "123",
			to:          new(uint),
			obj:         marshalBuiltin{typ: "uint"},
			expect:      toPtr(uint(123)),
		}, {
			description: "*uint8(123)",
			from:        "123",
			to:          new(uint8),
			obj:         marshalBuiltin{typ: "uint"},
			expect:      toPtr(uint8(123)),
		}, {
			description: "*uint16(123)",
			from:        "123",
			to:          new(uint16),
			obj:         marshalBuiltin{typ: "uint"},
			expect:      toPtr(uint16(123)),
		}, {
			description: "*uint32(123)",
			from:        "123",
			to:          new(uint32),
			obj:         marshalBuiltin{typ: "uint"},
			expect:      toPtr(uint32(123)),
		}, {
			description: "uint64(123)",
			from:        "123",
			to:          new(uint64),
			obj:         marshalBuiltin{typ: "uint"},
			expect:      toPtr(uint64(123)),
		}, {
			description: "from a uint8 type",
			from:        uint8(5),
			to:          uint(0),
			obj:         marshalBuiltin{typ: "uint"},
			expect:      uint(5),
		}, {
			description: "invalid to type",
			from:        "123",
			to:          bool(true),
			obj:         marshalBuiltin{typ: "uint"},
			expectErr:   goschtalt.ErrNotApplicable,
		}, {
			description: "not a number",
			from:        "99redballons99",
			to:          uint(0),
			obj:         marshalBuiltin{typ: "uint"},
			expectErr:   goschtalt.ErrNotApplicable,
		}, {
			description: "invalid number",
			from:        "2_",
			to:          uint(0),
			obj:         marshalBuiltin{typ: "uint"},
			expectErr:   goschtalt.ErrNotApplicable,
		}, {
			description: "range is too big",
			from:        "300",
			to:          uint8(0),
			obj:         marshalBuiltin{typ: "uint"},
			expectErr:   strconv.ErrRange,
		},
	}

	testUnmarshalAdapters(t, tests)
}
