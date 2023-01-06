// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/goschtalt/goschtalt/pkg/decoder"
	"github.com/goschtalt/goschtalt/pkg/meta"
)

// filegroup is a filesystem and paths to examine for configuration files.
type filegroup struct {
	// fs is the filesystem to examine.
	fs fs.FS

	// paths are either exact files, or directories to examine for configuration.
	paths []string

	// recurse specifies if directories encoutered in the paths should be examined
	// recursively or not.
	recurse bool

	// exactFile means that there should be exactly the same number of records as
	// files found or it is considered a failure.  This is mainly to support the
	// AddFile() use case where the file must be present or it is an error.
	exactFile bool
}

// toRecords walks the filegroup and finds all the records that are present and
// can be processed using the present configuration.
func (g filegroup) toRecords(delimiter string, decoders *codecRegistry[decoder.Decoder]) ([]record, error) {
	files, err := g.enumerate()
	if err != nil {
		return nil, err
	}

	list := make([]record, 0, len(files))
	for _, file := range files {
		r, err := g.toRecord(file, delimiter, decoders)
		if err != nil {
			return nil, err
		}

		list = append(list, r...)
	}

	return list, nil
}

// toRecord handles examining a single file and returning it as part of an array
// of records.  This allows for returning 0 or 1 record easily.
func (g filegroup) toRecord(file, delimiter string, decoders *codecRegistry[decoder.Decoder]) ([]record, error) {
	f, err := g.fs.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	basename := stat.Name()
	ext := strings.TrimPrefix(path.Ext(basename), ".")

	dec, err := decoders.find(ext)
	if dec == nil {
		if g.exactFile {
			// No failures allowed.
			return nil, err
		}

		// The file isn't supported by a decoder, skip it.
		return nil, nil
	}

	// Only read the file after we're pretty sure it can be decoded.
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	ctx := decoder.Context{
		Filename:  basename,
		Delimiter: delimiter,
	}

	var tree meta.Object
	err = dec.Decode(ctx, data, &tree)
	if err != nil {
		err = fmt.Errorf("decoder error for extension '%s' processing file '%s' %w %v",
			ext, basename, ErrDecoding, err)

		return nil, err
	}

	return []record{{
		name: basename,
		tree: tree,
	}}, nil
}

// enumerate walks the specified paths and collects the files it finds that match
// the specified extensions.
func (g filegroup) enumerate() ([]string, error) {
	var files []string

	for _, p := range g.paths {
		found, err := g.enumeratePath(path.Clean(p))
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
func (g filegroup) enumeratePath(path string) ([]string, error) {
	isDir, err := g.isDir(path)
	if err != nil {
		// Fail if the path is not found, otherwise, continue
		if errors.Is(err, fs.ErrInvalid) || errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
		return nil, nil
	}

	if !isDir {
		return []string{path}, nil
	}

	if g.exactFile {
		return nil, fs.ErrInvalid
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

	return files, err
}

// isDir examines a structure to see if it is a directory or something else.
func (g filegroup) isDir(path string) (dir bool, err error) {
	// Make sure the paths are consistent across FS implementations with
	// go's documentation.  This prevents errors due to some FS accepting
	// invalid paths while others correctly reject them.
	if !fs.ValidPath(path) {
		return false, fmt.Errorf("path '%s' %w", path, fs.ErrInvalid)
	}

	var file fs.File

	file, err = g.fs.Open(path)
	if err == nil {
		var stat fs.FileInfo
		stat, err = file.Stat()
		if err == nil {
			dir = stat.IsDir()
		}

		_ = file.Close()
	}

	return dir, err
}

// filegroupsToRecords converts a list of filegroups into a list of records.
func filegroupsToRecords(delimiter string, filegroups []filegroup, decoders *codecRegistry[decoder.Decoder]) ([]record, error) {
	rv := make([]record, 0, len(filegroups))
	for _, grp := range filegroups {
		tmp, err := grp.toRecords(delimiter, decoders)
		if err != nil {
			return nil, err
		}
		rv = append(rv, tmp...)
	}

	return rv, nil
}
