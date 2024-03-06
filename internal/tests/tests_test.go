// SPDX-FileCopyrightText: 2024 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/goschtalt/goschtalt"
	"github.com/goschtalt/goschtalt/internal/fspath"
	"github.com/stretchr/testify/assert"
)

func findRoot() (string, error) {
	root, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}

	for {
		up, err := filepath.Abs(filepath.Join(root, ".."))
		if err != nil || up == root {
			break
		}
		root = up
	}

	return root, nil
}

func TestEndToEndWithPathOutsidePWD(t *testing.T) {
	assert := assert.New(t)

	root, err := findRoot()
	assert.NoError(err)

	rel := os.DirFS(".")
	abs := os.DirFS(root)

	file := "../../.golangci.yml"

	wd, _ := os.Getwd()
	fmt.Printf("Current working directory: %s\n", wd)
	fmt.Printf("Root: %s\n", root)
	fmt.Printf("File: %s\n", file)
	g, err := goschtalt.New(
		goschtalt.AddJumbledHalt(abs, rel, file),
	)

	assert.NoError(err)
	assert.NotNil(g)
}

func TestEndToEndWithAbsPath(t *testing.T) {
	assert := assert.New(t)

	root, err := findRoot()
	assert.NoError(err)

	rel := os.DirFS(".")
	abs := os.DirFS(root)

	file := fspath.MustToRel("../../.golangci.yml")

	g, err := goschtalt.New(
		goschtalt.AddJumbledHalt(abs, rel, file),
	)

	assert.NoError(err)
	assert.NotNil(g)
}
