// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/schmidtw/goschtalt/internal/encoding"
	"github.com/schmidtw/goschtalt/internal/encoding/json"
	"github.com/schmidtw/goschtalt/internal/encoding/yaml"
)

var (
	ErrConflict              = errors.New("a conflict has been detected")
	ErrInvalidOption         = errors.New("invalid option")
	ErrNotFound              = errors.New("not found")
	ErrArrayIndexOutOfBounds = errors.New("array index is out of bounds")
	ErrTypeMismatch          = errors.New("type mismatch")
	ErrNotCompiled           = errors.New("the Compile() function must be called first")
)

// Goschtalt is a configurable, prioritized, merging configuration registry.
type Goschtalt struct {
	codecs          *encoding.Registry
	groups          []Group
	annotated       annotatedMap
	final           map[string]any
	hasBeenCompiled bool
	keyDelimiter    string
	mutex           sync.Mutex
	annotatedSorter func([]annotatedMap)
	keySwizzler     caseChanger
	typeMappers     map[string]TypeMapper
	valueConflictFn func(cur, next annotatedValue) (annotatedValue, error)
	arrayConflictFn func(cur, next annotatedArray) (annotatedArray, error)
	mapConflictFn   func(any, any) (any, error)
}

// Option is the type used for options.
type Option func(g *Goschtalt) error

func (fn Option) apply(g *Goschtalt) error {
	return fn(g)
}

// New creates a new goschtalt configuration instance.
func New(opts ...Option) (*Goschtalt, error) {
	r, _ := encoding.NewRegistry()
	g := &Goschtalt{
		final:       make(map[string]any),
		typeMappers: make(map[string]TypeMapper),
		codecs:      r,
	}

	/* set the defaults */
	_ = g.With(
		Codec(json.Codec{}),
		Codec(yaml.Codec{}),
		SortOrder(Natural),
		KeyCase(Lower),
		MergeStrategy(Map, Fail),
		MergeStrategy(Array, Append),
		MergeStrategy(Value, Latest),
		KeyDelimiter("."),
	)

	err := g.With(opts...)
	if err != nil {
		return nil, err
	}

	return g, nil
}

// With takes a list of options and applies them.
func (g *Goschtalt) With(opts ...Option) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	for _, opt := range opts {
		if err := opt.apply(g); err != nil {
			return err
		}
	}
	return nil
}

// Compile reads in all the files configured using the options provided,
// and merges the configuration trees into a single map for later use.
func (g *Goschtalt) Compile() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	full, err := g.collect()
	if err != nil {
		return err
	}

	return g.merge(full)
}

// Marshal renders the into the format specified ('json', 'yaml' or other extensions
// the Codecs provide and if adding comments should be attempted.  If a format
// does not support comments, an error is returned.  The result of the call is
// a slice of bytes with the information rendered into it.
func (g *Goschtalt) Marshal(format string, comments bool) ([]byte, error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// TODO support outputing origin via comments
	return g.codecs.Encode(format, &g.final)
}

// Fetch provides a generic based strict typed approach to fetching parts of the
// configuration tree.  The Goschtalt and key parameters are fairly
// straighforward, but the want may not be.  The want parameter is used to
// determine the type of the output object desired.  This allows this function
// to do to handy things the g.Fetch() method can't do:
//
//  - Thes function is able to validate the type returned is the type desired,
//    or return a descriptive error about why it can't do what was asked for.
//
//  - This function is also able to perform remapping from an existing type to
//    what you want based on the typeMappers provided.  This allows you to
//    automatically convert and cast a string to a time.Duration if you provide
//    the mapper.
//
// Here is an example showing how to add a duration caster based on spf13/cast:
//
//    import (
//        "github.com/schmidtw/goschtalt"
//        "github.com/spf13/cast"
//    )
//
//    func DurationMapper() goschtalt.Option {
//        var d time.Duration
//        return goschtalt.CustomMapper(d, func(i any) (any, error) {
//            return cast.ToDurationE(i)
//        })
//    }
//
//    ...
//
//    g := goschtalt.New(DurationMapper())
//
//    ...
func Fetch[T any](g *Goschtalt, key string, want T) (T, error) {
	var zeroVal T
	rv, err := g.Fetch(key)
	if err != nil {
		return zeroVal, err
	}

	if fn, found := g.typeMappers[reflect.TypeOf(want).String()]; found {
		rv, err = fn(rv)
		if err != nil {
			return zeroVal, err
		}
	}

	if reflect.TypeOf(want) != reflect.TypeOf(rv) {
		return zeroVal, fmt.Errorf("%w: expected type '%s' does not match type found '%s'",
			ErrTypeMismatch,
			reflect.TypeOf(want),
			reflect.TypeOf(rv))
	}

	return rv.(T), nil
}

// Fetch pulls the specified portion of the configuration tree and returns it to
// the caller as an any, since it could be a map node or a specific value.
func (g *Goschtalt) Fetch(key string) (any, error) {
	if len(key) == 0 {
		return g.final, nil
	}

	key = g.keySwizzler(key)
	path := strings.Split(key, g.keyDelimiter)

	val, at, err := g.searchMap(g.final, path)
	if err != nil {
		return nil, fmt.Errorf("with '%s' %w", strings.Join(at, g.keyDelimiter), err)
	}

	return val, nil
}

func (g *Goschtalt) searchMap(src map[string]any, path []string) (any, []string, error) {
	var err error

	at := []string{path[0]}

	next, found := src[path[0]]
	if !found {
		return nil, at, ErrNotFound
	}

	// If there is more to the path, continue traversing, otherwise return.
	if len(path) > 1 {
		var up []string

		switch typedNext := next.(type) {
		case map[string]any:
			next, up, err = g.searchMap(typedNext, path[1:])
		case []any:
			next, up, err = g.searchArray(typedNext, path[1:])
		default:
		}

		at = append(at, up...)
	}

	return next, at, err
}

func (g *Goschtalt) searchArray(src []any, path []string) (any, []string, error) {
	at := []string{path[0]}

	idx, err := strconv.Atoi(path[0])
	if err != nil {
		return nil, at, err
	}
	if len(src) <= idx {
		return nil, at, fmt.Errorf("%w len(array) is %d", ErrArrayIndexOutOfBounds, len(src))
	}

	next := src[idx]
	if len(path) > 1 {
		var up []string

		switch typedNext := next.(type) {
		case map[string]any:
			next, up, err = g.searchMap(typedNext, path[1:])
		case []any:
			next, up, err = g.searchArray(typedNext, path[1:])
		default:
		}

		at = append(at, up...)
	}

	return next, at, err
}

func (g *Goschtalt) collect() ([]annotatedMap, error) {
	full := []annotatedMap{}

	for _, group := range g.groups {
		cfgs, err := group.walk(g.codecs)
		if err != nil {
			return nil, err
		}

		full = append(full, cfgs...)
	}

	for i := range full {
		// Apply any key mangling that is needed.
		keycaseMap(g.keySwizzler, full[i].m)
	}
	g.annotatedSorter(full)

	return full, nil
}

func (g *Goschtalt) merge(cfgs []annotatedMap) error {
	if len(cfgs) == 0 {
		return nil
	}

	g.annotated = cfgs[0]
	for _, cfg := range cfgs[1:] {
		if err := g.mergeMap(cfg, &g.annotated); err != nil {
			return err
		}
	}
	g.final = toFinalMap(g.annotated)
	g.hasBeenCompiled = true
	return nil
}

func (g *Goschtalt) mergeMap(src annotatedMap, dest *annotatedMap) error {
	for key, next := range src.m {
		current, found := dest.m[key]
		if !found {
			// No conflicts - easy merge
			dest.Append(next.(annotated))
			dest.m[key] = next
			continue
		}

		if reflect.TypeOf(current) == reflect.TypeOf(next) {
			// Types match.
			dest.Append(next.(annotated))

			switch next := next.(type) {
			case annotatedMap:
				// This one is a bit of a special case in that there isn't a
				// conflict as much as we ran into another structural map.
				// Because of that, the conflict resolver doesn't get called
				// here.
				c := current.(annotatedMap)
				if err := g.mergeMap(next, &c); err != nil {
					return err
				}
				dest.m[key] = c
			case annotatedArray: // Both arrays, resolve.
				tmp, err := g.arrayConflictFn(current.(annotatedArray), next)
				if err != nil {
					return err
				}
				dest.m[key] = tmp
			case annotatedValue: // Both values, resolve.
				tmp, err := g.valueConflictFn(current.(annotatedValue), next)
				if err != nil {
					return err
				}
				dest.m[key] = tmp
			}
			continue
		}

		// The types don't match, resolve.
		tmp, err := g.mapConflictFn(current, next)
		if err != nil {
			return err
		}

		// Adjust the file metrics
		dest.Append(tmp.(annotated))
		dest.m[key] = tmp
	}

	return nil
}
