// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/schmidtw/goschtalt/pkg/decoder"
	"github.com/schmidtw/goschtalt/pkg/encoder"
	"github.com/schmidtw/goschtalt/pkg/meta"
)

// These options must always be present to prevent panics, etc.
var alwaysOptions = []Option{
	SortRecordsNaturally(),
	AlterKeyCase(nil),
	SetKeyDelimiter("."),
}

var DefaultOptions = []Option{}

// Config is a configurable, prioritized, merging configuration registry.
type Config struct {
	mutex          sync.Mutex
	files          []string
	tree           meta.Object
	compiled       bool
	explainOptions strings.Builder
	explainCompile strings.Builder

	rawOpts []Option
	opts    options
}

// New creates a new goschtalt configuration instance.
func New(opts ...Option) (*Config, error) {
	c := Config{
		tree: meta.Object{},
		opts: options{
			decoders: newRegistry[decoder.Decoder](),
			encoders: newRegistry[encoder.Encoder](),
		},
	}

	if err := c.With(opts...); err != nil {
		return nil, err
	}

	return &c, nil
}

// With takes a list of options and applies them.
func (c *Config) With(opts ...Option) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	cfg := options{
		decoders: newRegistry[decoder.Decoder](),
		encoders: newRegistry[encoder.Encoder](),
	}

	c.explainOptions.Reset()
	c.explainCompile.Reset()

	fmt.Fprintf(&c.explainOptions, "Start of options processing.\n\n")

	raw := append(c.rawOpts, opts...)

	full := alwaysOptions

	if !ignoreDefaultOpts(raw) {
		full = append(full, DefaultOptions...)
	}

	full = append(full, c.rawOpts...)

	full = append(full, opts...)

	fmt.Fprintln(&c.explainOptions, "Options in effect:")
	i := 1
	for _, opt := range full {
		if opt != nil {
			fmt.Fprintf(&c.explainOptions, "  %d. %s\n", i, opt.String())
			i++
			if err := opt.apply(&cfg); err != nil {
				return err
			}
		}
	}

	// The options are valid, record them.
	c.opts = cfg
	c.rawOpts = raw

	fmt.Fprintf(&c.explainOptions, "\nFile extensions supported:\n")
	exts := c.opts.decoders.extensions()
	if len(exts) == 0 {
		fmt.Fprintln(&c.explainOptions, "  none")
	} else {
		for _, ext := range exts {
			fmt.Fprintf(&c.explainOptions, "  - %s\n", ext)
		}
	}

	if c.opts.autoCompile {
		if err := c.compile(); err != nil {
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

	return c.compile()
}

// compile is the internal compile function.
func (c *Config) compile() error {
	var cfgs []fileObject

	c.explainCompile.Reset()

	fmt.Fprintf(&c.explainCompile, "Start of compilation.\n\n")

	for _, group := range c.opts.groups {
		tmp, err := group.walk(c.opts.decoders, c.opts.keyDelimiter)
		if err != nil {
			return err
		}
		cfgs = append(cfgs, tmp...)
	}

	for _, val := range c.opts.values {
		tmp, err := val.decode(c.opts.keyDelimiter, c.opts.valueOptions...)
		if err != nil {
			return err
		}
		cfgs = append(cfgs, tmp)
	}

	merged := meta.Object{
		Map: make(map[string]meta.Object),
	}
	if len(cfgs) == 0 {
		c.tree = merged
		c.compiled = true
		return nil
	}

	sorter := c.getSorter()
	sorter(cfgs)

	var files []string

	fmt.Fprintln(&c.explainCompile, "Records processed in order.")
	i := 1
	if len(cfgs) == 0 {
		fmt.Fprintln(&c.explainCompile, "  none")
	}
	for _, cfg := range cfgs {
		fmt.Fprintf(&c.explainCompile, "  %d. %s\n", i, cfg.File)
		i++

		var err error
		subtree := cfg.Obj.AlterKeyCase(c.opts.keySwizzler)
		merged, err = merged.Merge(subtree)
		if err != nil {
			return err
		}
		files = append(files, cfg.File)
	}

	fmt.Fprintf(&c.explainCompile, "\nVariable expansions processed in order.\n")
	i = 1
	if len(c.opts.expansions) == 0 {
		fmt.Fprintln(&c.explainCompile, "  none")
	}
	for _, exp := range c.opts.expansions {
		fmt.Fprintf(&c.explainCompile, "  %d. %s\n", i, exp.String())
		i++

		var err error
		merged, err = merged.ToExpanded(exp.maximum, exp.origin, exp.start, exp.end, exp.mapper)
		if err != nil {
			return err
		}
	}

	c.files = files
	c.tree = merged
	c.compiled = true
	return nil
}

func (c *Config) getSorter() func([]fileObject) {
	return func(a []fileObject) {
		sort.SliceStable(a, func(i, j int) bool {
			return c.opts.sorter(a[i].File, a[j].File)
		})
	}
}

// ShowOrder is a helper function that provides the order the configuration
// files were combined based on the present configuration.  This can only
// be called after the Compile() has been called.
func (c *Config) ShowOrder() ([]string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.compiled {
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

	sorter := c.getSorter()
	sorter(cfgs)

	for _, cfg := range cfgs {
		file := cfg.File

		// Only include the file if there is a decoder for it.
		ext := strings.TrimPrefix(filepath.Ext(file), ".")
		_, err := c.opts.decoders.find(ext)
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

	return c.opts.decoders.extensions()
}

func (c *Config) Explain() string {
	return c.explainOptions.String() + "\n" + c.explainCompile.String()
}
