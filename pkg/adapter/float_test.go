// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"testing"

	"github.com/goschtalt/goschtalt"
)

func TestFloatValueAdapterInternals(t *testing.T) {
	tests := []valueAdapterTest{
		{
			description: "float32(0.0) - valid",
			from:        float32(0.0),
			obj:         marshalFloat{},
			expect:      "0",
		}, {
			description: "float32(1.1) - valid",
			from:        float32(1.1),
			obj:         marshalFloat{},
			expect:      "1.1",
		}, {
			description: "float64(0.0) - valid",
			from:        float64(0.0),
			obj:         marshalFloat{},
			expect:      "0",
		}, {
			description: "float64(1.1) - valid",
			from:        float64(1.1),
			obj:         marshalFloat{},
			expect:      "1.1",
		}, {
			description: "*float32(0.0) - valid",
			from:        toPtr(float32(0.0)),
			obj:         marshalFloat{},
			expect:      "0",
		}, {
			description: "float32(1.1) - valid",
			from:        toPtr(float32(1.1)),
			obj:         marshalFloat{},
			expect:      "1.1",
		}, {
			description: "*float64(0.0) - valid",
			from:        toPtr(float64(0.0)),
			obj:         marshalFloat{},
			expect:      "0",
		}, {
			description: "_float64_1_1_ - valid",
			from:        toPtr(float64(1.1)),
			obj:         marshalFloat{},
			expect:      "1.1",
		}, {
			description: "invalid",
			from:        "string",
			obj:         marshalFloat{},
			expectErr:   goschtalt.ErrNotApplicable,
		},
	}

	testValueAdapters(t, tests)
}

func TestFloatUnmarshalAdapterInternals(t *testing.T) {
	tests := []unmarshalAdapterTest{
		{
			description: "marshalFloat (32), a valid string",
			from:        "1.1",
			to:          float32(1.1),
			obj:         marshalFloat{},
			expect:      float32(1.1),
		}, {
			description: "marshalFloat (64), a valid string",
			from:        "1.1",
			to:          float64(1.1),
			obj:         marshalFloat{},
			expect:      float64(1.1),
		}, {
			description: "marshalFloat (32), a valid string to a pointer",
			from:        "12.34",
			to:          new(float32),
			obj:         marshalFloat{},
			expect:      toPtr(float32(12.34)),
		}, {
			description: "marshalFloat (64), a valid string to a pointer",
			from:        "12.34",
			to:          new(float64),
			obj:         marshalFloat{},
			expect:      toPtr(float64(12.34)),
		}, {
			description: "marshalFloat, not a string",
			from:        12,
			to:          float32(1.1),
			obj:         marshalFloat{},
			expectErr:   goschtalt.ErrNotApplicable,
		}, {
			description: "marshalFloat, not a float",
			from:        "true",
			to:          "string",
			obj:         marshalFloat{},
			expectErr:   goschtalt.ErrNotApplicable,
		}, {
			description: "marshalFloat, not a valid float option",
			from:        "invalid",
			to:          float32(1.1),
			obj:         marshalFloat{},
			expectErr:   goschtalt.ErrNotApplicable,
		},
	}

	testUnmarshalAdapters(t, tests)
}
