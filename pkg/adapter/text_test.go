// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/goschtalt/goschtalt"
)

type TestObj struct {
	Name  string
	Value int
}

func (t *TestObj) UnmarshalText(b []byte) error {
	var name string
	var val int

	s := string(b)

	_, err := fmt.Sscanf(s, "%s %d", &name, &val)
	if err != nil {
		return err
	}

	t.Name = name
	t.Value = val

	return nil
}

func (t *TestObj) MarshalText() ([]byte, error) {
	if t.Name == "error" {
		return nil, errors.New("some error")
	}
	s := fmt.Sprintf("%s %d", t.Name, t.Value)
	return []byte(s), nil
}

func TestTextUnmarshalAdapterInternals(t *testing.T) {
	tests := []unmarshalAdapterTest{
		{
			description: "stringToIP",
			from:        "127.0.0.1",
			to:          net.IP{},
			obj:         textMarshaler{matcher: All},
			expect:      net.ParseIP("127.0.0.1"),
		}, {
			description: "stringToIP - fail",
			from:        "dogs",
			to:          net.IP{},
			obj:         textMarshaler{matcher: All},
			expectErr:   errUnknown,
		}, {
			description: "stringToIP - didn't match",
			from:        "dogs",
			to:          time.Time{},
			obj:         textMarshaler{matcher: All},
			expectErr:   goschtalt.ErrNotApplicable,
		}, {
			description: "textMarshaler ptr",
			from:        "dogs 10",
			to:          &TestObj{},
			obj:         textMarshaler{matcher: All},
			expect:      &TestObj{Name: "dogs", Value: 10},
		}, {
			description: "textMarshaler",
			from:        "dogs 10",
			to:          TestObj{},
			obj:         textMarshaler{matcher: All},
			expect:      TestObj{Name: "dogs", Value: 10},
		}, {
			description: "textMarshaler ptr specific type",
			from:        "dogs 10",
			to:          &TestObj{},
			obj:         textMarshaler{matcher: All},
			expect:      &TestObj{Name: "dogs", Value: 10},
		}, {
			description: "textMarshaler specific type",
			from:        "dogs 10",
			to:          TestObj{},
			obj:         textMarshaler{matcher: All},
			expect:      TestObj{Name: "dogs", Value: 10},
		}, {
			description: "textMarshaler, not a string",
			from:        12,
			to:          TestObj{},
			obj:         textMarshaler{matcher: All},
			expectErr:   goschtalt.ErrNotApplicable,
		}, {
			description: "textMarshaler, obj doesn't implement the interface",
			from:        "dogs 10",
			to:          "string",
			obj:         textMarshaler{matcher: All},
			expectErr:   goschtalt.ErrNotApplicable,
		}, {
			description: "textMarshaler, invalid string",
			from:        "dogs_10",
			to:          TestObj{},
			obj:         textMarshaler{matcher: All},
			expectErr:   errUnknown,
		}}

	testUnmarshalAdapters(t, tests)
}

func TestTextValueAdapterInternals(t *testing.T) {
	tests := []valueAdapterTest{
		{
			description: "ipToCfg",
			from:        net.ParseIP("127.0.0.1"),
			obj:         textMarshaler{matcher: All},
			expect:      "127.0.0.1",
		}, {
			description: "ipToCfg - didn't match",
			from:        "dogs",
			obj:         textMarshaler{matcher: All},
			expectErr:   goschtalt.ErrNotApplicable,
		}, {
			description: "withMarshalTextptr",
			from:        &TestObj{"hi", 99},
			obj:         textMarshaler{matcher: All},
			expect:      "hi 99",
		}, {
			description: "textMarshaler",
			from:        TestObj{"hi", 99},
			obj:         textMarshaler{matcher: All},
			expect:      "hi 99",
		}, {
			description: "textMarshaler only TestObj",
			from:        TestObj{"hi", 99},
			obj:         textMarshaler{matcher: All},
			expect:      "hi 99",
		}, {
			description: "textMarshaler only TestObj, fail",
			from:        "invalid",
			obj:         textMarshaler{matcher: All},
			expectErr:   goschtalt.ErrNotApplicable,
		}, {
			description: "textMarshaler, wrong interface",
			from:        "invalid",
			obj:         textMarshaler{matcher: All},
			expectErr:   goschtalt.ErrNotApplicable,
		}, {
			description: "textMarshaler, wrong interface",
			from:        TestObj{"error", 99},
			obj:         textMarshaler{matcher: All},
			expectErr:   errUnknown,
		},
	}

	testValueAdapters(t, tests)
}
