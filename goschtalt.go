// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"sync"

	"github.com/schmidtw/goschtalt/pkg/meta"
)

var DefaultOptions = []Option{
	FileSortOrderNatural(),
	KeyCaseLower(),
	KeyDelimiter("."),
}

// Config is a configurable, prioritized, merging configuration registry.
type Config struct {
	mutex           sync.Mutex
	tree            meta.Object
	hasBeenCompiled bool

	// options based things
	decoders         *decoderRegistry
	encoders         *encoderRegistry
	groups           []Group
	sorter           func([]meta.Object)
	keyDelimiter     string
	keySwizzler      func(string) string
	unmarshalOptions []UnmarshalOption
	typeMappers      map[string]typeMapper
}

// Option is the type used for options.
type Option func(c *Config) error

// New creates a new goschtalt configuration instance.
func New(opts ...Option) (*Config, error) {
	c := &Config{
		tree:        meta.Object{Type: meta.Map},
		decoders:    newDecoderRegistry(),
		encoders:    newEncoderRegistry(),
		typeMappers: make(map[string]typeMapper),
	}

	/* set the defaults */
	_ = c.With(DefaultOptions...)

	err := c.With(opts...)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// With takes a list of options and applies them.
func (c *Config) With(opts ...Option) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return err
		}
	}
	return nil
}

// Compile reads in all the files configured using the options provided,
// and merges the configuration trees into a single map for later use.
func (c *Config) Compile() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var cfgs []meta.Object

	for _, group := range c.groups {
		tmp, err := group.walk(c.decoders)
		if err != nil {
			return err
		}
		cfgs = append(cfgs, tmp...)
	}

	merged := meta.Object{
		Type: meta.Map,
		Map:  make(map[string]meta.Object),
	}
	if len(cfgs) == 0 {
		c.tree = merged
		c.hasBeenCompiled = true
		return nil
	}

	c.sorter(cfgs)

	for _, cfg := range cfgs {
		var err error
		cfg = cfg.AlterKeyCase(c.keySwizzler)
		merged, err = merged.Merge(cfg)
		if err != nil {
			return err
		}
	}
	c.tree = merged
	c.hasBeenCompiled = true
	return nil
}
