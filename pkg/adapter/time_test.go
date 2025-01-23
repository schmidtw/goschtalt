// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"math"
	"net"
	"testing"
	"time"

	"github.com/goschtalt/goschtalt"
)

func TestTimeValueAdapterInternals(t *testing.T) {
	tests := []valueAdapterTest{
		{
			description: "marshalTime",
			from:        time.Date(2022, time.January, 30, 0, 0, 0, 0, time.UTC),
			obj:         marshalTime{layout: "2006-01-02"},
			expect:      "2022-01-30",
		}, {
			description: "marshalTime - didn't match",
			from:        net.IP{127, 0, 0, 1},
			obj:         marshalTime{layout: "2006-01-02"},
			expectErr:   goschtalt.ErrNotApplicable,
		},
	}

	testValueAdapters(t, tests)
}

func TestTimeUnmarshalAdapterInternals(t *testing.T) {
	tests := []unmarshalAdapterTest{
		{
			description: "marshalTime",
			from:        "2022-01-30",
			to:          time.Time{},
			obj:         marshalTime{layout: "2006-01-02"},
			expect:      time.Date(2022, time.January, 30, 0, 0, 0, 0, time.UTC),
		}, {
			description: "marshalTime from unix time in UTC (int)",
			from:        int(1737483492),
			to:          time.Time{},
			obj:         marshalTime{layout: "2006-01-02"},
			expect:      time.Date(2025, time.January, 21, 18, 18, 12, 0, time.UTC),
		}, {
			description: "marshalTime from unix time in UTC (uint)",
			from:        uint(1737483492),
			to:          time.Time{},
			obj:         marshalTime{},
			expect:      time.Date(2025, time.January, 21, 18, 18, 12, 0, time.UTC),
		}, {
			description: "marshalTime from unix time in UTC (uint)... too big",
			from:        uint64(math.MaxUint64 - 10),
			to:          time.Time{},
			obj:         marshalTime{},
			expectErr:   errUnknown,
		}, {
			description: "marshalTime from unix time in UTC (float64)",
			from:        float64(1737483492.1),
			to:          time.Time{},
			obj:         marshalTime{},
			expect:      time.Date(2025, time.January, 21, 18, 18, 12, 99999904, time.UTC),
		}, {
			description: "marshalTime - fail",
			from:        "dogs",
			to:          time.Time{},
			obj:         marshalTime{},
			expectErr:   errUnknown,
		}, {
			description: "marshalTime - didn't match",
			from:        "dogs",
			to:          net.IP{},
			obj:         marshalTime{},
			expectErr:   goschtalt.ErrNotApplicable,
		},
	}

	testUnmarshalAdapters(t, tests)
}
