// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"strings"

	"github.com/schmidtw/goschtalt/pkg/meta"
)

// Fetch pulls the specified portion of the configuration tree and returns it to
// the caller as an any, since it could be a map node or a specific value.
func (c *Config) Fetch(key string) (any, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	a, _, err := c.fetchWithOrigin(key)
	return a, err
}

// FetchWithOrigin pulls the specified portion of the configuration tree as
// well as the origin information and returns both sets of data to the caller
// as an any and array of meta.Origin values.
func (c *Config) FetchWithOrigin(key string) (any, []meta.Origin, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.fetchWithOrigin(key)
}

// fetchWithOrigin is the non-mutex wrapped function that does the work, but
// can be called by the other mutex protected functions without deadlocking.
func (c *Config) fetchWithOrigin(key string) (any, []meta.Origin, error) {
	if !c.hasBeenCompiled {
		return nil, nil, ErrNotCompiled
	}

	if len(key) == 0 {
		return c.tree.ToRaw(), c.tree.Origins, nil
	}

	key = c.keySwizzler(key)
	path := strings.Split(key, c.keyDelimiter)

	obj, err := c.tree.Fetch(path, c.keyDelimiter)
	if err != nil {
		return nil, nil, err
	}

	return obj.ToRaw(), obj.Origins, nil
}
