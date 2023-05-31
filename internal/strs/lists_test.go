// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package strs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	tests := []struct {
		description string
		list        []string
		want        string
		found       bool
	}{
		{
			description: "found at the front",
			list:        []string{"foo", "bar", "goo"},
			want:        "foo",
			found:       true,
		}, {
			description: "found in the middle",
			list:        []string{"foo", "bar", "goo"},
			want:        "bar",
			found:       true,
		}, {
			description: "found at the end",
			list:        []string{"foo", "bar", "goo"},
			want:        "goo",
			found:       true,
		}, {
			description: "not found",
			list:        []string{"foo", "bar", "goo"},
			want:        "oops",
		}, {
			description: "empty list",
			list:        []string{},
			want:        "nothing",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			assert.Equal(tc.found, Contains(tc.list, tc.want))
		})
	}
}

func TestContainsAll(t *testing.T) {
	tests := []struct {
		description string
		list        []string
		want        []string
		found       bool
	}{
		{
			description: "found all, simple",
			list:        []string{"foo", "bar", "goo"},
			want:        []string{"foo"},
			found:       true,
		}, {
			description: "found all",
			list:        []string{"foo", "bar", "goo"},
			want:        []string{"bar", "foo", "goo"},
			found:       true,
		}, {
			description: "found with duplicates",
			list:        []string{"foo", "bar", "goo"},
			want:        []string{"goo", "bar", "goo", "goo"},
			found:       true,
		}, {
			description: "not all found",
			list:        []string{"foo", "bar", "goo"},
			want:        []string{"foo", "oops"},
		}, {
			description: "empty all list",
			list:        []string{},
			want:        []string{"things"},
		}, {
			description: "empty want list",
			list:        []string{"things"},
			want:        []string{},
			found:       true,
		}, {
			description: "empty all and want list",
			list:        []string{},
			want:        []string{},
			found:       true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			assert.Equal(tc.found, ContainsAll(tc.list, tc.want))
		})
	}
}

func TestMissing(t *testing.T) {
	tests := []struct {
		description string
		list        []string
		want        []string
		missing     []string
	}{
		{
			description: "found all, simple",
			list:        []string{"foo", "bar", "goo"},
			want:        []string{"foo"},
		}, {
			description: "found all",
			list:        []string{"foo", "bar", "goo"},
			want:        []string{"bar", "foo", "goo"},
		}, {
			description: "found with duplicates",
			list:        []string{"foo", "bar", "goo"},
			want:        []string{"goo", "bar", "goo", "goo"},
		}, {
			description: "not all found",
			list:        []string{"foo", "bar", "goo"},
			want:        []string{"foo", "oops"},
			missing:     []string{"oops"},
		}, {
			description: "empty all list",
			list:        []string{},
			want:        []string{"things"},
			missing:     []string{"things"},
		}, {
			description: "empty want list",
			list:        []string{"things"},
			want:        []string{},
		}, {
			description: "empty all and want list",
			list:        []string{},
			want:        []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			got := Missing(tc.list, tc.want)
			if tc.missing == nil {
				assert.Empty(got)
			} else {
				assert.Equal(tc.missing, got)
			}
		})
	}
}
