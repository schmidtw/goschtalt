// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/goschtalt/goschtalt"
	"github.com/stretchr/testify/assert"
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

type adapterToCfg interface {
	To(from reflect.Value) (any, error)
}

type adapterFromCfg interface {
	From(from, to reflect.Value) (any, error)
}

func TestUnmarshalAdapterInternals(t *testing.T) {
	unknownErr := errors.New("unknownErr")
	tests := []struct {
		description string
		from        any
		to          any
		obj         adapterFromCfg
		expect      any
		expectErr   error
	}{
		{
			description: "marshalDuration",
			from:        "1s",
			to:          time.Duration(1),
			obj:         marshalDuration{},
			expect:      time.Second,
		}, {
			description: "marshalDuration ptr",
			from:        "1s",
			to:          new(time.Duration),
			obj:         marshalDuration{},
			expect: func() *time.Duration {
				d := time.Second
				return &d
			}(),
		}, {
			description: "marshalDuration - fail",
			from:        "dogs",
			to:          time.Duration(1),
			obj:         marshalDuration{},
			expectErr:   unknownErr,
		}, {
			description: "marshalDuration - didn't match",
			from:        "dogs",
			to:          time.Time{},
			obj:         marshalDuration{},
			expectErr:   goschtalt.ErrNotApplicable,
		}, {
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
			expectErr:   unknownErr,
		}, {
			description: "stringToIP - didn't match",
			from:        "dogs",
			to:          time.Time{},
			obj:         textMarshaler{matcher: All},
			expectErr:   goschtalt.ErrNotApplicable,
		}, {
			description: "marshalTime",
			from:        "2022-01-30",
			to:          time.Time{},
			obj:         marshalTime{layout: "2006-01-02"},
			expect:      time.Date(2022, time.January, 30, 0, 0, 0, 0, time.UTC),
		}, {
			description: "marshalTime - fail",
			from:        "dogs",
			to:          time.Time{},
			obj:         marshalTime{layout: "2006-01-02"},
			expectErr:   unknownErr,
		}, {
			description: "marshalTime - didn't match",
			from:        "dogs",
			to:          net.IP{},
			obj:         marshalTime{layout: "2006-01-02"},
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
			expectErr:   unknownErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			got, err := tc.obj.From(reflect.ValueOf(tc.from), reflect.ValueOf(tc.to))

			if errors.Is(unknownErr, tc.expectErr) {
				// Accept nil, or the zero value of the type
				if nil != got {
					assert.Equal(reflect.Zero(reflect.TypeOf(got)).Interface(), got)
				}
				assert.Error(err)
				return
			}

			if tc.expectErr != nil {
				// Accept nil, or the zero value of the type
				if nil != got {
					assert.Equal(reflect.Zero(reflect.TypeOf(got)).Interface(), got)
				}
				assert.ErrorIs(err, tc.expectErr)
				return
			}

			assert.EqualValues(tc.expect, got)
			assert.NoError(err)
		})
	}
}

func TestValueAdapterInternals(t *testing.T) {
	unknownErr := errors.New("unknownErr")
	tests := []struct {
		description string
		from        any
		obj         adapterToCfg
		expect      any
		expectErr   error
	}{
		{
			description: "marshalDuration",
			from:        time.Second,
			obj:         marshalDuration{},
			expect:      "1s",
		}, {
			description: "marshalDuration - didn't match",
			from:        time.Time{},
			obj:         marshalDuration{},
			expectErr:   goschtalt.ErrNotApplicable,
		}, {
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
			description: "marshalTime",
			from:        time.Date(2022, time.January, 30, 0, 0, 0, 0, time.UTC),
			obj:         marshalTime{layout: "2006-01-02"},
			expect:      "2022-01-30",
		}, {
			description: "marshalTime - didn't match",
			from:        net.IP{127, 0, 0, 1},
			obj:         marshalTime{layout: "2006-01-02"},
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
			expectErr:   unknownErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			got, err := tc.obj.To(reflect.ValueOf(tc.from))

			if errors.Is(unknownErr, tc.expectErr) {
				// Accept nil, or the zero value of the type
				if nil != got {
					assert.Equal(reflect.Zero(reflect.TypeOf(got)).Interface(), got)
				}
				assert.Error(err)
				return
			}

			if tc.expectErr != nil {
				// Accept nil, or the zero value of the type
				if nil != got {
					assert.Equal(reflect.Zero(reflect.TypeOf(got)).Interface(), got)
				}
				assert.ErrorIs(err, tc.expectErr)
				return
			}

			assert.Equal(tc.expect, got)
			assert.NoError(err)
		})
	}
}

func TestEndToEnd(t *testing.T) {
	type all struct {
		D   time.Duration
		T   time.Time
		IP  net.IP
		Obj TestObj
	}
	tests := []struct {
		description string
		from        all
		unmarshal   []goschtalt.UnmarshalOption
		value       []goschtalt.ValueOption
		expectErr   error
	}{
		{
			description: "String <-> time.Duration",
			from:        all{D: time.Second},
			unmarshal:   []goschtalt.UnmarshalOption{DurationUnmarshal()},
			value:       []goschtalt.ValueOption{MarshalDuration()},
		}, {
			description: "String <-> time.Time",
			from:        all{T: time.Date(2022, time.August, 15, 11, 10, 9, 0, time.UTC)},
			unmarshal:   []goschtalt.UnmarshalOption{TimeUnmarshal(time.RFC3339)},
			value:       []goschtalt.ValueOption{MarshalTime(time.RFC3339)},
		}, {
			description: "String <-> net.IP",
			from:        all{IP: net.ParseIP("127.0.0.1")},
			unmarshal:   []goschtalt.UnmarshalOption{TextUnmarshal(All)},
			value:       []goschtalt.ValueOption{MarshalText(All)},
		}, {
			description: "String <-> TestObj (all)",
			from:        all{Obj: TestObj{Name: "dog", Value: 12}},
			unmarshal:   []goschtalt.UnmarshalOption{TextUnmarshal(All)},
			value:       []goschtalt.ValueOption{MarshalText(All)},
		}, {
			description: "String <-> TestObj (limited)",
			from:        all{Obj: TestObj{Name: "dog", Value: 12}},
			unmarshal:   []goschtalt.UnmarshalOption{TextUnmarshal(All)},
			value:       []goschtalt.ValueOption{MarshalText(All)},
		}, {
			description: "String <-> all",
			from: all{
				D:  time.Hour,
				T:  time.Date(2002, time.August, 15, 0, 0, 0, 0, time.UTC),
				IP: net.ParseIP("192.168.1.1"),
			},
			unmarshal: []goschtalt.UnmarshalOption{
				DurationUnmarshal(),
				TextUnmarshal(AllButTime),
				TimeUnmarshal("2006-01-02"),
			},
			value: []goschtalt.ValueOption{
				MarshalDuration(),
				MarshalText(AllButTime),
				MarshalTime("2006-01-02"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			cfg, err := goschtalt.New(
				goschtalt.AutoCompile(),
				goschtalt.DefaultUnmarshalOptions(tc.unmarshal...),
				goschtalt.DefaultValueOptions(tc.value...),
				goschtalt.AddValue("rec", goschtalt.Root, tc.from),
			)

			assert.NoError(err)

			var got all

			err = cfg.Unmarshal(goschtalt.Root, &got)

			if tc.expectErr != nil {
				assert.Error(err)
				return
			}

			assert.Equal(tc.from, got)
			assert.NoError(err)
		})
	}
}
