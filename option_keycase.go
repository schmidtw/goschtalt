// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import "strings"

// KeyCaseCustom specifies a custom alteration to the case of the keys.
func KeyCaseCustom(fn func(string) string) Option {
	return func(c *Config) error {
		c.keySwizzler = fn
		return nil
	}
}

// KeyCaseUnchanged specifies no changes to the case of the keys.
func KeyCaseUnchanged() Option {
	return KeyCaseCustom(func(s string) string {
		return s
	})
}

// KeyCaseLower specifies the keys will be converted to all lowercase.
func KeyCaseLower() Option {
	return KeyCaseCustom(strings.ToLower)
}

// KeyCaseUpper specifies the keys will be converted to all uppercase.
func KeyCaseUpper() Option {
	return KeyCaseCustom(strings.ToUpper)
}
