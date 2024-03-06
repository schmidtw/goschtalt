// SPDX-FileCopyrightText: 2024 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package fspath

import (
	"errors"
	"path"
	"path/filepath"
)

var ErrInvalidPath = errors.New("invalid path")

// IsLocal returns true if the file is a local file.
//
// The file must be a fs.FS formatted path with '/' as the separator.
func IsLocal(file string) bool {
	return filepath.IsLocal(filepath.FromSlash(file))
}

// ToRel converts a path to an absolute path for use in a fs.FS filesystem.
// This means any volume name and leading separator are stripped off, and any
// relative path is converted to an absolute path.
//
// The file must be a fs.FS formatted path with '/' as the separator.
func ToRel(file string) (string, error) {
	return toRel(file, filepath.Abs)
}

// MustToRel is like RelAbs but panics if there is an error.
func MustToRel(file string) string {
	rv, err := ToRel(file)
	if err != nil {
		panic(err)
	}
	return rv
}

// toRel provides a way to test RelToAbs with a custom getwd function.
func toRel(file string, abs func(string) (string, error)) (string, error) {
	if file == "" {
		return "", ErrInvalidPath
	}

	// The simple case is if the path is already absolute.
	if path.IsAbs(file) {
		return path.Clean(file[1:]), nil
	}

	// The path is relative, so we need to convert it to an absolute path.
	targetAbs, err := abs(".")
	if err != nil {
		return "", err
	}

	targetAbs = filepath.Join(targetAbs, file)

	// Just keep going up the tree until we find the root instead of trying
	// to figure out the way root is represented on the current system.
	base := targetAbs
	for {
		up := filepath.Join(base, "..")
		up = filepath.Clean(up)
		if up == base {
			base = up
			break
		}
		base = up
	}

	targetRel, err := filepath.Rel(base, targetAbs)
	if err != nil {
		return "", err
	}

	return path.Clean(filepath.ToSlash(targetRel)), nil
}
