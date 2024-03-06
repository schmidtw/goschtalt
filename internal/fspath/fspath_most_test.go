// SPDX-FileCopyrightText: 2024 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

//go:build !windows

package fspath

import "errors"

var errUnknown = errors.New("unknown error")

var toRelTests = []toRelTest{
	{
		description: "abs case",
		path:        "/a/b/c",
		abs: func(string) (string, error) {
			return "/a/b", nil
		},
		want: "a/b/c",
	}, {
		description: "rel case",
		path:        "../c",
		abs: func(string) (string, error) {
			return "/a/b", nil
		},
		want: "a/c",
	}, {
		description: "empty case",
		path:        "",
		abs: func(string) (string, error) {
			return "/a/b", nil
		},
		expectedErr: ErrInvalidPath,
	}, {
		description: "error case",
		path:        "a/b/c",
		abs: func(string) (string, error) {
			return "", errUnknown
		},
		expectedErr: errUnknown,
	},
}
