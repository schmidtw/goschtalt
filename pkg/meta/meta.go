// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package meta

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const redactedText = "REDACTED"

const (
	cmdSecret  = "secret"
	cmdReplace = "replace"
	cmdKeep    = "keep"
	cmdFail    = "fail"
	cmdAppend  = "append"
	cmdPrepend = "prepend"
	cmdSplice  = "splice"
	cmdClear   = "clear"
)

var (
	ErrConflict         = errors.New("a conflict has been detected")
	ErrInvalidCommand   = errors.New("invalid command")
	ErrNotFound         = errors.New("not found")
	ErrArrayOutOfBounds = errors.New("array index is out of bounds")
)

type Origin struct {
	File string
	Line int
	Col  int
}

// String returns a useful representation for the origin.
func (o Origin) String() string {
	file := "unknown"
	line := "???"
	col := "???"

	if len(o.File) > 0 {
		file = o.File
	}
	if o.Line > 0 {
		line = strconv.Itoa(o.Line)
	}
	if o.Col > 0 {
		col = strconv.Itoa(o.Col)
	}
	return fmt.Sprintf("%s:%s[%s]", file, line, col)
}

const (
	Array = iota + 1
	Map
	Value
)

type Object struct {
	Origins  []Origin
	IsSecret bool
	Type     int
	Map      map[string]Object
	Array    []Object
	Value    any
}

func (obj Object) OriginString() string {
	var list []string
	for _, v := range obj.Origins {
		list = append(list, v.String())
	}

	return strings.Join(list, ", ")
}

func (obj Object) Fetch(asks []string, separater string) (Object, error) {
	return obj.fetch(asks, asks, separater)
}

func getPath(asks, path []string, separater string) string {
	// Trim off the same number of elements that are left in the asks from the
	// path before joining them and returning them.
	path = path[:len(path)-len(asks)]
	return strings.Join(path, separater)
}

func (obj Object) fetch(asks, path []string, separater string) (Object, error) {
	if len(asks) == 0 {
		return obj, nil
	}

	switch obj.Type {
	case Map:
		key := asks[0]
		next, found := obj.Map[key]
		if found {
			return next.fetch(asks[1:], path, separater)
		}
	case Array:
		idx, err := strconv.Atoi(asks[0])
		if err != nil {
			return Object{}, err
		}
		if 0 <= idx && idx < len(obj.Array) {
			return obj.Array[idx].fetch(asks[1:], path, separater)
		} else {
			return Object{},
				fmt.Errorf("with array len of %d and '%s' %w",
					len(obj.Array),
					getPath(asks[1:], path, separater),
					ErrArrayOutOfBounds)
		}
	}

	return Object{}, fmt.Errorf("with '%s' %w", getPath(asks[1:], path, separater), ErrNotFound)
}

func (obj Object) ToRaw() any {
	switch obj.Type {
	case Array:
		rv := make([]any, len(obj.Array))
		for i, val := range obj.Array {
			rv[i] = val.ToRaw()
		}
		return rv
	case Map:
		rv := make(map[string]any)

		for key, val := range obj.Map {
			rv[key] = val.ToRaw()
		}
		return rv
	case Value:
		return obj.Value
	}

	return nil
}

func ObjectFromRaw(in any) (obj Object) {
	obj.Origins = []Origin{}

	switch in := in.(type) {
	case []any:
		obj.Type = Array
		obj.Array = make([]Object, len(in))
		for i, val := range in {
			obj.Array[i] = ObjectFromRaw(val)
		}
	case map[string]any:
		obj.Type = Map
		obj.Map = make(map[string]Object)
		for key, val := range in {
			obj.Map[key] = ObjectFromRaw(val)
		}
	default:
		obj.Type = Value
		obj.Value = in
	}

	return obj
}

func (obj Object) ToRedacted() Object {
	if obj.IsSecret {
		return Object{
			Origins:  []Origin{},
			Type:     Value,
			Value:    redactedText,
			IsSecret: true,
		}
	}

	switch obj.Type {
	case Array:
		array := make([]Object, len(obj.Array))
		for i, val := range obj.Array {
			array[i] = val.ToRedacted()
		}
		obj.Array = array
	case Map:
		m := make(map[string]Object)

		for key, val := range obj.Map {
			m[key] = val.ToRedacted()
		}
		obj.Map = m
	}

	return obj
}

func (obj Object) AlterKeyCase(to func(string) string) Object {
	switch obj.Type {
	case Array:
		array := make([]Object, len(obj.Array))
		for i, val := range obj.Array {
			array[i] = val.AlterKeyCase(to)
		}
		obj.Array = array
	case Map:
		m := make(map[string]Object)

		for key, val := range obj.Map {
			m[to(key)] = val.AlterKeyCase(to)
		}
		obj.Map = m
	}

	return obj
}

func (obj Object) ResolveCommands() (Object, error) {
	return obj.resolveCommands(false)
}

func (obj Object) resolveCommands(secret bool) (Object, error) {
	if secret {
		obj.IsSecret = true
	}

	switch obj.Type {
	case Array:
		array := make([]Object, len(obj.Array))
		for i, val := range obj.Array {
			v, err := val.resolveCommands(false)
			if err != nil {
				return Object{}, err
			}
			array[i] = v
		}
		obj.Array = array
	case Map:
		m := make(map[string]Object)

		for key, val := range obj.Map {
			cmd, err := getValidCmd(key, val)
			if err != nil {
				return Object{}, err
			}
			tmp, err := val.resolveCommands(cmd.secret)
			if err != nil {
				return Object{}, err
			}
			m[cmd.final] = tmp
		}
		obj.Map = m
	}

	return obj, nil
}

func (obj Object) Merge(next Object) (Object, error) {
	// The 'clear' command is special in that if it is found at all, it
	// overwrites everything else in the existing tree and exists the merge.
	for k := range next.Map {
		cmd, err := getCmd(k)
		if err != nil {
			return Object{}, err
		}
		if cmd.cmd == cmdClear {
			return Object{Origins: []Origin{}, Type: Map}, nil
		}
	}

	return obj.merge(command{}, next)
}

func (obj Object) merge(cmd command, next Object) (Object, error) {
	if obj.Type == Value {
		rv := obj
		switch cmd.cmd {
		case cmdReplace, "":
			rv = next
		case cmdKeep:
		case cmdFail:
			return Object{}, ErrConflict
		}

		rv.IsSecret = cmd.secret
		return rv, nil
	}

	if obj.Type == Array {
		rv := obj
		next, err := next.resolveCommands(obj.IsSecret)
		if err != nil {
			return Object{}, err
		}
		switch cmd.cmd {
		case cmdAppend, "":
			if obj.IsSecret || next.IsSecret || cmd.secret {
				rv.IsSecret = true
			}
			rv.Origins = append(obj.Origins, next.Origins...)
			rv.Array = append(obj.Array, next.Array...)
		case cmdPrepend:
			if obj.IsSecret || next.IsSecret || cmd.secret {
				rv.IsSecret = true
			}
			rv.Origins = append(next.Origins, obj.Origins...)
			rv.Array = append(next.Array, obj.Array...)
		case cmdReplace:
			rv.IsSecret = cmd.secret
			rv = next
		case cmdKeep:
		case cmdFail:
			return Object{}, ErrConflict
		}
		return rv, nil
	}

	rv := obj
	switch cmd.cmd {
	case cmdSplice, "":
		for key, val := range next.Map {
			newCmd, err := getValidCmd(key, val)
			if err != nil {
				return Object{}, err
			}

			existing, found := obj.Map[newCmd.final]
			if !found {
				// Merging with no conflicts.
				v, err := val.resolveCommands(newCmd.secret)
				if err != nil {
					return Object{}, err
				}
				obj.Map[newCmd.final] = v
				continue
			}

			if existing.Type == val.Type {
				v, err := existing.merge(newCmd, val)
				if err != nil {
					return Object{}, err
				}
				obj.Map[newCmd.final] = v
				continue
			}

			switch newCmd.cmd {
			case cmdSplice, cmdReplace, "":
				v, err := val.resolveCommands(newCmd.secret)
				if err != nil {
					return Object{}, err
				}
				obj.Map[newCmd.final] = v
			case cmdKeep:
				obj.Map[newCmd.final] = existing
			case cmdFail:
				return Object{}, ErrConflict
			}
		}
	case cmdReplace:
		v, err := next.resolveCommands(false)
		if err != nil {
			return Object{}, err
		}
		rv = v
	case cmdKeep:
	case cmdFail:
		return Object{}, ErrConflict
	}

	rv.IsSecret = cmd.secret
	return rv, nil
}

func getValidCmd(key string, obj Object) (command, error) {
	cmd, err := getCmd(key)
	if err != nil {
		return command{}, err
	}

	list := map[int][]string{
		Map:   {"", cmdFail, cmdKeep, cmdReplace, cmdSplice},
		Array: {"", cmdFail, cmdKeep, cmdReplace, cmdAppend, cmdPrepend},
		Value: {"", cmdFail, cmdKeep, cmdReplace},
	}

	opts, found := list[obj.Type]
	if found {
		for _, opt := range opts {
			if cmd.cmd == opt {
				return cmd, nil
			}
		}
	}

	return command{}, ErrInvalidCommand
}
