// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"path/filepath"
	"strings"
	"sync"

	"github.com/schmidtw/goschtalt/pkg/meta"
)

var DefaultOptions = []Option{
	FileSortOrderNatural(),
	KeyCaseLower(),
	KeyDelimiter("."),
}

// Add options to validate things so the code is a cleaner.
var validatorOptions = []Option{
	validateSorter(),
	validateKeyDelimiter(),
	validateKeySwizzler(),
}

// Config is a configurable, prioritized, merging configuration registry.
type Config struct {
	mutex           sync.Mutex
	files           []string
	tree            meta.Object
	hasBeenCompiled bool

	// options based things
	compileNow       bool
	ignoreDefaults   bool
	decoders         *decoderRegistry
	encoders         *encoderRegistry
	groups           []Group
	sorter           func([]fileObject)
	keyDelimiter     string
	keySwizzler      func(string) string
	unmarshalOptions []UnmarshalOption
	expansions       []ExpandVarsOpts
}

// Option is the type used for options.
type Option func(c *Config) error

func newConfig() *Config {
	return &Config{
		tree:     meta.Object{},
		decoders: newDecoderRegistry(),
		encoders: newEncoderRegistry(),
	}
}

// New creates a new goschtalt configuration instance.
func New(opts ...Option) (*Config, error) {
	// Check to see if an option indicates to not apply defaults on a
	// throw-away object.
	tmp := newConfig()
	err := tmp.with(opts...)
	if err != nil {
		return nil, err
	}

	var allOpts []Option
	if !tmp.ignoreDefaults {
		allOpts = append(allOpts, DefaultOptions...)
	}
	allOpts = append(allOpts, opts...)
	allOpts = append(allOpts, validatorOptions...)

	c := newConfig()

	err = c.With(allOpts...)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// With takes a list of options and applies them.
func (c *Config) With(opts ...Option) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	err := c.with(opts...)
	if err != nil {
		return err
	}

	if c.compileNow {
		if err := c.compile(); err != nil {
			return err
		}
	}

	return nil
}

// with is the internal looper for applying options.
func (c *Config) with(opts ...Option) error {
	for _, opt := range opts {
		if opt != nil {
			if err := opt(c); err != nil {
				return err
			}
		}
	}

	return nil
}

// Compile reads in all the files configured using the options provided,
// and merges the configuration trees into a single map for later use.
func (c *Config) Compile() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.compile()
}

// compile is the internal compile function.
func (c *Config) compile() error {
	var cfgs []fileObject

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
	var files []string

	for _, cfg := range cfgs {
		var err error
		subtree := cfg.Obj.AlterKeyCase(c.keySwizzler)
		merged, err = merged.Merge(subtree)
		if err != nil {
			return err
		}
		files = append(files, cfg.File)
	}

	for _, e := range c.expansions {
		var err error
		merged, err = merged.ToExpanded(e.Maximum, e.Name, e.Start, e.End, e.Mapper)
		if err != nil {
			return err
		}
	}

	c.files = files
	c.tree = merged
	c.hasBeenCompiled = true
	return nil
}

// ShowOrder is a helper function that provides the order the configuration
// files were combined based on the present configuration.  This can only
// be called after the Compile() has been called.
func (c *Config) ShowOrder() ([]string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.hasBeenCompiled {
		return []string{}, ErrNotCompiled
	}

	return c.files, nil
}

// OrderList is a helper function that sorts a caller provided list of filenames
// exectly the same way the Config object would sort them when reading and
// merging the files when the configuration is being compiled.  It also filters
// the list based on the decoders present.
func (c *Config) OrderList(list []string) (files []string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var cfgs []fileObject

	for _, item := range list {
		cfgs = append(cfgs, fileObject{File: item})
	}

	c.sorter(cfgs)

	for _, cfg := range cfgs {
		file := cfg.File

		// Only include the file if there is a decoder for it.
		ext := strings.TrimPrefix(filepath.Ext(file), ".")
		_, err := c.decoders.find(ext)
		if err == nil {
			files = append(files, file)
		}
	}

	return files
}

// Extensions returns the extensions this config object supports.
func (c *Config) Extensions() []string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.decoders.extensions()
}
