// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package strs

// Contains returns if a string is in the list.
func Contains(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

// ContainsAll returns if all the wanted strings are in the list.
func ContainsAll(list []string, want []string) bool {
	m := make(map[string]int, len(list))

	for _, s := range list {
		m[s] = 1
	}

	for _, w := range want {
		if _, ok := m[w]; !ok {
			return false
		}
	}
	return true
}

// Missing returns the list of all the missing strings that were wanted but not
// present in the list.
func Missing(list []string, want []string) []string {
	missing := make([]string, 0, len(want))

	m := make(map[string]int, len(list))

	for _, s := range list {
		m[s] = 1
	}

	for _, w := range want {
		if _, ok := m[w]; !ok {
			missing = append(missing, w)
		}
	}

	return missing
}
