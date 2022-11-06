// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"sort"
)

// group is a filesystem and paths to examine for configuration files.
type group struct {
	// fs is the filesystem to examine.
	fs fs.FS

	// paths are either exact files, or directories to examine for configuration.
	paths []string

	// recurse specifies if directories encoutered in the paths should be examined
	// recursively or not.
	recurse bool
}

func groupsToRecords(groups []group) ([]record, error) {
	rv := make([]record, 0, len(groups))
	for _, grp := range groups {
		tmp, err := grp.walk()
		if err != nil {
			return nil, err
		}
		rv = append(rv, tmp...)
	}

	return rv, nil
}

func (g group) walk() ([]record, error) {
	files, err := g.enumerate()
	if err != nil {
		return nil, err
	}

	list := make([]record, 0, len(files))
	for _, file := range files {
		f, err := g.fs.Open(file)
		if err != nil {
			return nil, err
		}
		stat, err := f.Stat()
		if err != nil {
			return nil, err
		}

		r := record{
			name: stat.Name(),
			fn: func(_ string) (io.ReadCloser, error) {
				return f, nil
			},
		}

		list = append(list, r)
	}
	return list, nil
}

// enumerate walks the specified paths and collects the files it finds that match
// the specified extensions.
func (g group) enumerate() ([]string, error) {
	var files []string

	// By default include everything in the base directory if nothing is specified.
	if len(g.paths) == 0 {
		g.paths = []string{"."}
	}

	for _, path := range g.paths {
		found, err := g.enumeratePath(path)
		if err != nil {
			return nil, err
		}
		files = append(files, found...)
	}
	sort.Strings(files)

	return files, nil
}

// enumeratePath examines a specific path and collects all the appropriate files.
// If the path ends up being a specific file return exactly that file.
func (g group) enumeratePath(path string) ([]string, error) {
	// Make sure the paths are consistent across FS implementations with
	// go's documentation.  This prevents errors due to some FS accepting
	// invalid paths while others correctly reject them.
	if !fs.ValidPath(path) {
		return nil, fmt.Errorf("path '%s' %w", path, fs.ErrInvalid)
	}

	file, err := g.fs.Open(path)
	if err != nil {
		// Fail if the path is not found, otherwise, continue
		if errors.Is(err, fs.ErrInvalid) || errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
		return nil, nil
	}

	stat, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, nil
	}
	isDir := stat.IsDir()
	_ = file.Close()

	if !isDir {
		return []string{path}, nil
	}

	var files []string
	var walker fs.WalkDirFunc
	if g.recurse {
		walker = func(file string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			files = append(files, file)
			return nil
		}
	} else {
		walker = func(file string, d fs.DirEntry, err error) error {
			if err == nil {
				// Don't proceed into any directories except the top directory
				// specified.
				if file != path {
					if d.IsDir() {
						return fs.SkipDir
					}
					files = append(files, file)
				}
			}
			return err
		}
	}

	err = fs.WalkDir(g.fs, path, walker)

	if err != nil {
		files = []string{}
	}

	return files, err
}
