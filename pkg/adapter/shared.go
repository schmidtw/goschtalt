// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package adapter

func toPtr[T any](o T) *T {
	rv := new(T)
	*rv = o
	return rv
}
