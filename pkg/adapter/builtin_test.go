// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/goschtalt/goschtalt"
	"github.com/stretchr/testify/assert"
)

func TestOnly(t *testing.T) {
	tests := []struct {
		description string
		s           string
		list        []string
		expect      bool
	}{
		{
			description: "empty list",
			expect:      true,
		}, {
			description: "item in the list",
			s:           "foo",
			list:        []string{"foo", "bar"},
			expect:      true,
		}, {
			description: "item not in the list",
			s:           "wolf",
			list:        []string{"foo", "bar"},
			expect:      false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			got := only(tc.s, tc.list)
			assert.Equal(tc.expect, got)
		})
	}
}

func TestNumToString(t *testing.T) {
	tests := []struct {
		description string
		in          any
		only        string
		expect      string
	}{
		// Simple positive cases
		{
			description: "int(123) --> 123",
			in:          int(123),
			expect:      "123",
		}, {
			description: "int(-55) --> -55",
			in:          int(-55),
			expect:      "-55",
		}, {
			description: "uint(255) --> 255",
			in:          uint(255),
			expect:      "255",
		}, {
			description: "float32(255.0) --> 255",
			in:          float32(255.0),
			expect:      "255",
		}, {
			description: "float64(255.1) --> 255.1",
			in:          float64(255.1),
			expect:      "255.1",
		},

		// Check that only works
		{
			description: "int(123), only=int --> 123",
			in:          int(123),
			only:        "int",
			expect:      "123",
		}, {
			description: "uint(123), only=uint --> 123",
			in:          uint(123),
			only:        "uint",
			expect:      "123",
		}, {
			description: "float32(123), only=float --> 255",
			in:          float32(255),
			only:        "float",
			expect:      "255",
		}, {
			description: "int(123), only=uint --> ''",
			in:          int(123),
			only:        "uint",
			expect:      "",
		}, {
			description: "int(123), only=float --> ''",
			in:          int(123),
			only:        "float",
			expect:      "",
		}, {
			description: "uint(123), only=float --> ''",
			in:          uint(123),
			only:        "float",
			expect:      "",
		},

		// Check that the pointer handling works
		{
			description: "*int(-55) --> -55",
			in:          toPtr(int(-55)),
			expect:      "-55",
		}, {
			description: "**int(-55) --> -55",
			in:          toPtr(toPtr(int(-55))),
			expect:      "-55",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			only := make([]string, 0, 1)
			if tc.only != "" {
				only = append(only, tc.only)
			}

			got := numToString(reflect.ValueOf(tc.in), only...)

			assert.Equal(tc.expect, got)
		})
	}
}

func TestNumOrStringToString(t *testing.T) {
	tests := []struct {
		description string
		in          any
		only        string
		expect      string
	}{
		// Simple checks
		{
			description: "string(foo) --> 'foo'",
			in:          "foo",
			expect:      "foo",
		}, {
			description: "int(-55) --> -55",
			in:          int(-55),
			expect:      "-55",
		},

		// Check that the pointer handling works
		{
			description: "*int(-55) --> -55",
			in:          toPtr(int(-55)),
			only:        "int",
			expect:      "-55",
		}, {
			description: "*string(bar) --> 'bar'",
			in:          toPtr("bar"),
			expect:      "bar",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			only := make([]string, 0, 1)
			if tc.only != "" {
				only = append(only, tc.only)
			}

			got := numOrStringToString(reflect.ValueOf(tc.in), only...)

			assert.Equal(tc.expect, got)
		})
	}
}

func TestNumToStringErr(t *testing.T) {
	tests := []struct {
		description string
		in          any
		only        string
		expect      string
		errExpected error
	}{
		// Simple checks
		{
			description: "string(foo) --> \"\", ErrNotApplicable",
			in:          "foo",
			errExpected: goschtalt.ErrNotApplicable,
		}, {
			description: "int(-55) --> -55",
			in:          int(-55),
			expect:      "-55",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			only := make([]string, 0, 1)
			if tc.only != "" {
				only = append(only, tc.only)
			}

			got, err := numToStringErr(reflect.ValueOf(tc.in), only...)

			if tc.errExpected != nil {
				assert.Equal("", got)
				assert.ErrorIs(tc.errExpected, err)
				return
			}

			assert.Equal(tc.expect, got)
			assert.NoError(err)
		})
	}
}

func TestSToB(t *testing.T) {
	tests := []struct {
		description string
		kind        reflect.Kind
		in          string
		ptr         int
		expect      any
		errExpected error
	}{
		// Simple checks
		{
			description: "true",
			kind:        reflect.Bool,
			in:          "true",
			expect:      true,
		},

		// Pointer checks
		{
			description: "single pointer",
			kind:        reflect.Bool,
			in:          "true",
			ptr:         1,
			expect:      toPtr(true),
		}, {
			description: "double pointer",
			kind:        reflect.Bool,
			in:          "true",
			ptr:         2,
			expect:      toPtr(toPtr(true)),
		}, {
			description: "triple pointer",
			kind:        reflect.Bool,
			in:          "true",
			ptr:         3,
			expect:      toPtr(toPtr(toPtr(true))),
		}, {
			description: "quad pointer",
			kind:        reflect.Bool,
			in:          "true",
			ptr:         4,
			errExpected: goschtalt.ErrUnsupported,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			var got any
			var err error
			switch tc.kind {
			case reflect.Bool:
				got, err = sToB(tc.in, tc.ptr)
			default:
				panic("unsupported kind")
			}

			if tc.errExpected != nil {
				assert.Nil(got)
				assert.ErrorIs(err, tc.errExpected)
				return
			}

			assert.NoError(err)
			assert.Equal(tc.expect, got)
		})
	}
}

func TestSToF(t *testing.T) {
	tests := []struct {
		description string
		kind        reflect.Kind
		in          string
		ptr         int
		expect      any
		errExpected error
	}{
		// Simple checks
		{
			description: "float32(23.5) --> 23.5",
			kind:        reflect.Float32,
			in:          "23.5",
			expect:      float32(23.5),
		}, {
			description: "float64(255) --> 255",
			kind:        reflect.Float64,
			in:          "255",
			expect:      float64(255),
		},

		// Error conditions
		{
			description: "overflow check",
			kind:        reflect.Float32,
			in:          "4.9e294967296",
			errExpected: strconv.ErrRange,
		}, {
			description: "syntax check",
			kind:        reflect.Float32,
			in:          "some other string",
			errExpected: goschtalt.ErrNotApplicable,
		},

		// Pointer checks
		{
			description: "single pointer",
			kind:        reflect.Float32,
			in:          "255",
			ptr:         1,
			expect:      toPtr(float32(255)),
		}, {
			description: "double pointer",
			kind:        reflect.Float32,
			in:          "255",
			ptr:         2,
			expect:      toPtr(toPtr(float32(255))),
		}, {
			description: "triple pointer",
			kind:        reflect.Float32,
			in:          "255",
			ptr:         3,
			expect:      toPtr(toPtr(toPtr(float32(255)))),
		}, {
			description: "quad pointer",
			kind:        reflect.Float32,
			in:          "255",
			ptr:         4,
			errExpected: goschtalt.ErrUnsupported,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			var got any
			var err error
			switch tc.kind {
			case reflect.Float32:
				got, err = sToF[float32](tc.in, 32, tc.ptr)
			case reflect.Float64:
				got, err = sToF[float64](tc.in, 64, tc.ptr)
			default:
				panic("unsupported kind")
			}

			if tc.errExpected != nil {
				assert.Nil(got)
				assert.ErrorIs(err, tc.errExpected)
				return
			}

			assert.NoError(err)
			assert.Equal(tc.expect, got)
		})
	}
}

func TestSToI(t *testing.T) {
	tests := []struct {
		description string
		kind        reflect.Kind
		in          string
		ptr         int
		expect      any
		errExpected error
	}{
		// Simple checks
		{
			description: "int(99) --> 99",
			kind:        reflect.Int,
			in:          "99",
			expect:      int(99),
		}, {
			description: "int8(8) --> 8",
			kind:        reflect.Int8,
			in:          "8",
			expect:      int8(8),
		}, {
			description: "int16(16) --> 16",
			kind:        reflect.Int16,
			in:          "16",
			expect:      int16(16),
		}, {
			description: "int32(32) --> 32",
			kind:        reflect.Int32,
			in:          "32",
			expect:      int32(32),
		}, {
			description: "int64(-64) --> -64",
			kind:        reflect.Int64,
			in:          "-64",
			expect:      int64(-64),
		},

		// Error conditions
		{
			description: "overflow check",
			kind:        reflect.Int8,
			in:          "300",
			errExpected: strconv.ErrRange,
		}, {
			description: "syntax check",
			kind:        reflect.Int,
			in:          "some other string",
			errExpected: goschtalt.ErrNotApplicable,
		},

		// Pointer checks
		{
			description: "single pointer",
			kind:        reflect.Int,
			in:          "255",
			ptr:         1,
			expect:      toPtr(int(255)),
		}, {
			description: "double pointer",
			kind:        reflect.Int,
			in:          "255",
			ptr:         2,
			expect:      toPtr(toPtr(int(255))),
		}, {
			description: "triple pointer",
			kind:        reflect.Int,
			in:          "255",
			ptr:         3,
			expect:      toPtr(toPtr(toPtr(int(255)))),
		}, {
			description: "quad pointer",
			kind:        reflect.Int,
			in:          "255",
			ptr:         4,
			errExpected: goschtalt.ErrUnsupported,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			var got any
			var err error
			switch tc.kind {
			case reflect.Int:
				got, err = sToI[int](tc.in, reflect.TypeOf(int(0)).Bits(), tc.ptr)
			case reflect.Int8:
				got, err = sToI[int8](tc.in, 8, tc.ptr)
			case reflect.Int16:
				got, err = sToI[int16](tc.in, 16, tc.ptr)
			case reflect.Int32:
				got, err = sToI[int32](tc.in, 32, tc.ptr)
			case reflect.Int64:
				got, err = sToI[int64](tc.in, 64, tc.ptr)
			default:
				panic("unsupported kind")
			}

			if tc.errExpected != nil {
				assert.Nil(got)
				assert.ErrorIs(err, tc.errExpected)
				return
			}

			assert.NoError(err)
			assert.Equal(tc.expect, got)
		})
	}
}

func TestSToU(t *testing.T) {
	tests := []struct {
		description string
		kind        reflect.Kind
		in          string
		ptr         int
		expect      any
		errExpected error
	}{
		// Simple checks
		{
			description: "uint(99) --> 99",
			kind:        reflect.Uint,
			in:          "99",
			expect:      uint(99),
		}, {
			description: "uint8(8) --> 8",
			kind:        reflect.Uint8,
			in:          "8",
			expect:      uint8(8),
		}, {
			description: "uint16(16) --> 16",
			kind:        reflect.Uint16,
			in:          "16",
			expect:      uint16(16),
		}, {
			description: "uint32(32) --> 32",
			kind:        reflect.Uint32,
			in:          "32",
			expect:      uint32(32),
		}, {
			description: "uint64(64) --> 64",
			kind:        reflect.Uint64,
			in:          "64",
			expect:      uint64(64),
		}, {
			description: "uintptr(99) --> 99",
			kind:        reflect.Uintptr,
			in:          "99",
			expect:      uintptr(99),
		},

		// Error conditions
		{
			description: "overflow check",
			kind:        reflect.Uint8,
			in:          "300",
			errExpected: strconv.ErrRange,
		}, {
			description: "syntax check",
			kind:        reflect.Uint,
			in:          "some other string",
			errExpected: goschtalt.ErrNotApplicable,
		},

		// Pointer checks
		{
			description: "single pointer",
			kind:        reflect.Uint,
			in:          "255",
			ptr:         1,
			expect:      toPtr(uint(255)),
		}, {
			description: "double pointer",
			kind:        reflect.Uint,
			in:          "255",
			ptr:         2,
			expect:      toPtr(toPtr(uint(255))),
		}, {
			description: "triple pointer",
			kind:        reflect.Uint,
			in:          "255",
			ptr:         3,
			expect:      toPtr(toPtr(toPtr(uint(255)))),
		}, {
			description: "quad pointer",
			kind:        reflect.Uint,
			in:          "255",
			ptr:         4,
			errExpected: goschtalt.ErrUnsupported,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			var got any
			var err error
			switch tc.kind {
			case reflect.Uint:
				got, err = sToU[uint](tc.in, reflect.TypeOf(uint(0)).Bits(), tc.ptr)
			case reflect.Uint8:
				got, err = sToU[uint8](tc.in, 8, tc.ptr)
			case reflect.Uint16:
				got, err = sToU[uint16](tc.in, 16, tc.ptr)
			case reflect.Uint32:
				got, err = sToU[uint32](tc.in, 32, tc.ptr)
			case reflect.Uint64:
				got, err = sToU[uint64](tc.in, 64, tc.ptr)
			case reflect.Uintptr:
				got, err = sToU[uintptr](tc.in, reflect.TypeOf(uintptr(0)).Bits(), tc.ptr)
			default:
				panic("unsupported kind")
			}

			if tc.errExpected != nil {
				assert.Nil(got)
				assert.ErrorIs(err, tc.errExpected)
				return
			}

			assert.NoError(err)
			assert.Equal(tc.expect, got)
		})
	}
}

func TestSToS(t *testing.T) {
	tests := []struct {
		description string
		kind        reflect.Kind
		in          string
		ptr         int
		expect      any
		errExpected error
	}{
		// Simple checks
		{
			description: "some string",
			kind:        reflect.String,
			in:          "some string",
			expect:      "some string",
		},

		// Pointer checks
		{
			description: "single pointer",
			kind:        reflect.String,
			in:          "255",
			ptr:         1,
			expect:      toPtr("255"),
		}, {
			description: "double pointer",
			kind:        reflect.String,
			in:          "255",
			ptr:         2,
			expect:      toPtr(toPtr("255")),
		}, {
			description: "triple pointer",
			kind:        reflect.String,
			in:          "255",
			ptr:         3,
			expect:      toPtr(toPtr(toPtr("255"))),
		}, {
			description: "quad pointer",
			kind:        reflect.String,
			in:          "255",
			ptr:         4,
			errExpected: goschtalt.ErrUnsupported,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			var got any
			var err error
			switch tc.kind {
			case reflect.String:
				got, err = sToS(tc.in, tc.ptr)
			default:
				panic("unsupported kind")
			}

			if tc.errExpected != nil {
				assert.Nil(got)
				assert.ErrorIs(err, tc.errExpected)
				return
			}

			assert.NoError(err)
			assert.Equal(tc.expect, got)
		})
	}
}
