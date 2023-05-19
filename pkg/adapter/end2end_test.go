// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"net"
	"testing"
	"time"

	"github.com/goschtalt/goschtalt"
	"github.com/stretchr/testify/assert"
)

func TestEndToEnd(t *testing.T) {
	type all struct {
		B   bool
		D   time.Duration
		F   float64
		I   int
		IP  net.IP
		Obj TestObj
		T   time.Time
		U   uint
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
				B:  true,
				D:  time.Hour,
				F:  1234.56,
				I:  -99,
				IP: net.ParseIP("192.168.1.1"),
				T:  time.Date(2002, time.August, 15, 0, 0, 0, 0, time.UTC),
				U:  73,
			},
			unmarshal: []goschtalt.UnmarshalOption{
				BoolUnmarshal(),
				DurationUnmarshal(),
				FloatUnmarshal(),
				IntUnmarshal(),
				TextUnmarshal(AllButTime),
				TimeUnmarshal("2006-01-02"),
				UintUnmarshal(),
			},
			value: []goschtalt.ValueOption{
				MarshalBool(),
				MarshalDuration(),
				MarshalFloat(),
				MarshalInt(),
				MarshalText(AllButTime),
				MarshalTime("2006-01-02"),
				MarshalUint(),
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
