// SPDX-FileCopyrightText: 2024 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

//go:build windows

package fspath

var toRelTests = []toRelTest{
	{
		description: "abs case",
		path:        "/a/b/c",
		abs: func(string) (string, error) {
			return "c:\\a\\b", nil
		},
		want: "a/b/c",
	}, {
		description: "rel case",
		path:        "../c",
		abs: func(string) (string, error) {
			return "c:\\a\\b", nil
		},
		want: "a/c",
	},
}
