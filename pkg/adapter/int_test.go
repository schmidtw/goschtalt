// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"strconv"
	"testing"

	"github.com/goschtalt/goschtalt"
)

func TestIntValueAdapterInternals(t *testing.T) {
	tests := []valueAdapterTest{
		{
			description: "int",
			from:        int(1234),
			obj:         marshalBuiltin{typ: "int"},
			expect:      "1234",
		}, {
			description: "*int",
			from:        toPtr(int(1234)),
			obj:         marshalBuiltin{typ: "int"},
			expect:      "1234",
		}, {
			description: "**int",
			from:        toPtr(toPtr(int(1234))),
			obj:         marshalBuiltin{typ: "int"},
			expect:      "1234",
		}, {
			description: "string",
			from:        "string",
			obj:         marshalBuiltin{typ: "int"},
			expectErr:   goschtalt.ErrNotApplicable,
		}, {
			description: "*string",
			from:        toPtr("string"),
			obj:         marshalBuiltin{typ: "int"},
			expectErr:   goschtalt.ErrNotApplicable,
		},
	}

	testValueAdapters(t, tests)
}

func TestIntUnmarshalAdapterInternals(t *testing.T) {
	tests := []unmarshalAdapterTest{
		{
			description: "int(123)",
			from:        "123",
			to:          int(0),
			obj:         marshalBuiltin{typ: "int"},
			expect:      int(123),
		}, {
			description: "int8(123)",
			from:        "123",
			to:          int8(0),
			obj:         marshalBuiltin{typ: "int"},
			expect:      int8(123),
		}, {
			description: "int16(123)",
			from:        "123",
			to:          int16(0),
			obj:         marshalBuiltin{typ: "int"},
			expect:      int16(123),
		}, {
			description: "int32(123)",
			from:        "123",
			to:          int32(0),
			obj:         marshalBuiltin{typ: "int"},
			expect:      int32(123),
		}, {
			description: "int64(123)",
			from:        "123",
			to:          int64(0),
			obj:         marshalBuiltin{typ: "int"},
			expect:      int64(123),
		}, {
			description: "*int(123)",
			from:        "123",
			to:          new(int),
			obj:         marshalBuiltin{typ: "int"},
			expect:      toPtr(int(123)),
		}, {
			description: "*int8(123)",
			from:        "123",
			to:          new(int8),
			obj:         marshalBuiltin{typ: "int"},
			expect:      toPtr(int8(123)),
		}, {
			description: "*int16(123)",
			from:        "123",
			to:          new(int16),
			obj:         marshalBuiltin{typ: "int"},
			expect:      toPtr(int16(123)),
		}, {
			description: "*int32(123)",
			from:        "123",
			to:          new(int32),
			obj:         marshalBuiltin{typ: "int"},
			expect:      toPtr(int32(123)),
		}, {
			description: "int64(123)",
			from:        "123",
			to:          new(int64),
			obj:         marshalBuiltin{typ: "int"},
			expect:      toPtr(int64(123)),
		}, {
			description: "from int to int",
			from:        int(5),
			to:          int(0),
			obj:         marshalBuiltin{typ: "int"},
			expect:      int(5),
		}, {
			description: "invalid to type",
			from:        "123",
			to:          bool(true),
			obj:         marshalBuiltin{typ: "int"},
			expectErr:   goschtalt.ErrNotApplicable,
		}, {
			description: "not a number",
			from:        "99redballons99",
			to:          int(0),
			obj:         marshalBuiltin{typ: "int"},
			expectErr:   goschtalt.ErrNotApplicable,
		}, {
			description: "invalid number",
			from:        "2_",
			to:          int(0),
			obj:         marshalBuiltin{typ: "int"},
			expectErr:   goschtalt.ErrNotApplicable,
		}, {
			description: "range is too big",
			from:        "300",
			to:          int8(0),
			obj:         marshalBuiltin{typ: "int"},
			expectErr:   strconv.ErrRange,
		},
	}

	testUnmarshalAdapters(t, tests)
}
