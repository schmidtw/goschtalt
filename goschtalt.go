// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"fmt"
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
	ignoreDefaults   bool
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

func newConfig() *Config {
	return &Config{
		tree:        meta.Object{},
		decoders:    newDecoderRegistry(),
		encoders:    newEncoderRegistry(),
		typeMappers: make(map[string]typeMapper),
	}
}

// New creates a new goschtalt configuration instance.
func New(opts ...Option) (*Config, error) {
	// Check to see if an option indicates to not apply defaults on a
	// throw-away object.
	tmp := newConfig()
	err := tmp.With(opts...)
	if err != nil {
		return nil, err
	}

	var allOpts []Option
	if !tmp.ignoreDefaults {
		allOpts = append(allOpts, DefaultOptions...)
	}

	c := newConfig()

	allOpts = append(allOpts, opts...)
	err = c.With(allOpts...)
	if err != nil {
		return nil, err
	}

	if c.sorter == nil {
		return nil, fmt.Errorf("%w: a FileSortOrder... option must be specified.", ErrConfigMissing)
	}

	if len(c.keyDelimiter) == 0 {
		return nil, fmt.Errorf("%w: KeyDelimiter() option must be specified.", ErrConfigMissing)
	}

	if c.keySwizzler == nil {
		return nil, fmt.Errorf("%w: a KeyCase... option must be specified.", ErrConfigMissing)
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
		tmp, err := group.walk(c.decoders, c.keyDelimiter)
		if err != nil {
			return err
		}
		cfgs = append(cfgs, tmp...)
	}

	merged := meta.Object{
		Map: make(map[string]meta.Object),
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
