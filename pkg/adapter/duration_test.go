// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"testing"
	"time"

	"github.com/goschtalt/goschtalt"
)

func TestDurationValueAdapterInternals(t *testing.T) {
	tests := []valueAdapterTest{
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
		},
	}

	testValueAdapters(t, tests)
}

func TestDurationUnmarshalAdapterInternals(t *testing.T) {
	tests := []unmarshalAdapterTest{
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
			expect:      toPtr(time.Second),
		}, {
			description: "marshalDuration - fail",
			from:        "dogs",
			to:          time.Duration(1),
			obj:         marshalDuration{},
			expectErr:   errUnknown,
		}, {
			description: "marshalDuration - didn't match",
			from:        "dogs",
			to:          time.Time{},
			obj:         marshalDuration{},
			expectErr:   goschtalt.ErrNotApplicable,
		},
	}

	testUnmarshalAdapters(t, tests)
}
