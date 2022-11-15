// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

type codec interface {
	Extensions() []string
}

// registry contains the mapping of extension to codec.
type codecRegistry[C codec] struct {
	mutex  sync.Mutex
	codecs map[string]C
}

// newRegistry creates a new instance of a codecRegistry.
func newRegistry[C codec]() *codecRegistry[C] {
	return &codecRegistry[C]{
		codecs: make(map[string]C),
	}
}

// extensions returns a list of supported extensions.  Note that all extensions
// are lowercase.
func (c *codecRegistry[C]) extensions() (list []string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	codecs := c.codecs

	for key := range codecs {
		list = append(list, key)
	}

	sort.Strings(list)

	return list
}

// find the codec of interest and return it.
func (c *codecRegistry[C]) find(ext string) (C, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	ext = strings.ToLower(ext)

	cod, ok := c.codecs[ext]

	if !ok {
		var rv C
		return rv, fmt.Errorf("extension '%s' %w", ext, ErrCodecNotFound)
	}

	return cod, nil
}

// register registers a codec for the specific file extensions provided.
// Attempting to register a nil codec will result in a panic.
func (c *codecRegistry[C]) register(enc C) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	exts := enc.Extensions()
	for _, ext := range exts {
		ext = strings.ToLower(ext)
		c.codecs[ext] = enc
	}
}
