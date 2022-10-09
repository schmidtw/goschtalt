// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import "fmt"

// validateSorter ensures a sorter has been set.
func validateSorter() Option {
	return func(c *Config) error {
		if c.sorter == nil {
			return fmt.Errorf("%w: a FileSortOrder... option must be specified.", ErrConfigMissing)
		}
		return nil
	}
}

// validateKeyDelimiter ensures a keyDelimiter has been set.
func validateKeyDelimiter() Option {
	return func(c *Config) error {
		if len(c.keyDelimiter) == 0 {
			return fmt.Errorf("%w: KeyDelimiter() option must be specified.", ErrConfigMissing)
		}
		return nil
	}
}

// validateKeySwizzler ensures a keySwizzler has been set.
func validateKeySwizzler() Option {
	return func(c *Config) error {
		if c.keySwizzler == nil {
			return fmt.Errorf("%w: a KeyCase... option must be specified.", ErrConfigMissing)
		}
		return nil
	}
}
