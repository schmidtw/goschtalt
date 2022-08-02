// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

// annotated is a helper interface that lets us treat all 3 of the structs
// the same way with regard to getting and appending the origin files.
type annotated interface {
	// Files simply returns the array of files for the struct.
	Files() []string
}

// annotatedMap provides a place to store the map based configurations and the
// files associated with the configuration.  Because of how we are using this
// struct we know that the 'any' is only ever going to be these types:
// annotatedMap, annotatedArray, annotatedValue
type annotatedMap struct {
	files []string
	m     map[string]any
}

func (a annotatedMap) Files() []string {
	return a.files
}

func (a *annotatedMap) Append(other annotated) {
	a.files = dedupedAppend(a.files, other.Files()...)
}

// annotatedArray provides a place to store the array based configurations and
// the files associated with the configuration.  Because of how we are using
// this struct we know that the 'any' is only ever going to be these types:
// annotatedMap, annotatedArray, annotatedValue
type annotatedArray struct {
	files []string
	array []any
}

func (a annotatedArray) Files() []string {
	return a.files
}

func (a *annotatedArray) Append(other annotated) {
	a.files = dedupedAppend(a.files, other.Files()...)
}

// annotatedValue provides a place to store the value based configurations and
// the files associated with the configuration.  Unlike the others in this
// group of structures, the any field can be anything.
type annotatedValue struct {
	files []string
	value any
}

func (a annotatedValue) Files() []string {
	return a.files
}

func (a *annotatedValue) Append(other annotated) {
	a.files = dedupedAppend(a.files, other.Files()...)
}

//
// The next 3 functions all work together recursively to build the annotated
// configuraiton tree using the annotated structs above.
//

// toAnnotatedMap builds the map & the tree that the map represents using
// the annotated structs so we can show where the configuration value is from.
// This function recurses into itself via toAnnotatedVal() and
// toAnnotatedArray().
func toAnnotatedMap(file string, src map[string]any) annotatedMap {
	m := annotatedMap{
		files: []string{file},
		m:     make(map[string]any),
	}

	for key, val := range src {
		m.m[key] = toAnnotatedVal(file, val)
	}

	return m
}

// toAnnotatedVal build the value and substructure portion of the tree using
// the annotated structs so we can show where the configuration value is from.
// This function recurses into itself via toAnnotatedVal() and
// toAnnotatedArray().
func toAnnotatedVal(file string, val any) any {
	switch val := val.(type) {
	case map[string]any:
		return toAnnotatedMap(file, val)
	case []any:
		return toAnnotatedArray(file, val)
	}

	return annotatedValue{
		files: []string{file},
		value: val,
	}
}

// toAnnotatedArray build the array and substructure portion of the tree using
// the annotated structs so we can show where the configuration value is from.
// This function recurses into itself via toAnnotatedVal() and
// toAnnotatedArray().
func toAnnotatedArray(file string, a []any) annotatedArray {
	rv := annotatedArray{
		files: []string{file},
		array: make([]any, len(a)),
	}
	for i, val := range a {
		rv.array[i] = toAnnotatedVal(file, val)
	}

	return rv
}

//
// The next 3 functions all work together recursively to build the final
// configuraiton tree using the annotated tree.
//

func toFinalMap(src annotatedMap) map[string]any {
	m := make(map[string]any)

	for key, val := range src.m {
		m[key] = toFinalVal(val)
	}

	return m
}

func toFinalVal(src any) any {
	switch src := src.(type) {
	case annotatedMap:
		return toFinalMap(src)
	case annotatedArray:
		return toFinalArray(src)
	case annotatedValue:
		return src.value
	}

	return src
}

func toFinalArray(src annotatedArray) []any {
	rv := make([]any, len(src.array))
	for i, val := range src.array {
		rv[i] = toFinalVal(val)
	}

	return rv
}

// dedupedAppend takes a list of strings along with a slice of strings and
// creates a unified list with the strings in the same order, and any added
// strings are checked to make sure there are no duplicates.  Duplicates are
// not checked for in the list base list.
func dedupedAppend(base []string, added ...string) []string {
	keys := make(map[string]bool)
	for _, item := range base {
		keys[item] = true
	}

	for _, want := range added {
		if _, found := keys[want]; !found {
			keys[want] = true
			base = append(base, want)
		}
	}
	return base
}
