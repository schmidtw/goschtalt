// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0
//go:build !windows

package json_test

import (
	"fmt"
	"os"
	"strings"

	"github.com/schmidtw/goschtalt"
	_ "github.com/schmidtw/goschtalt/extensions/decoders/json"
)

const text = `{
  "example": {
      "version": 1,
      "colors": ["red", "green", "blue"]
  }
}`

func Example() {
	err := os.WriteFile("/tmp/example.json", []byte(text), 0644)
	if err != nil {
		panic(err)
	}

	g, err := goschtalt.New(goschtalt.FileGroup(
		goschtalt.Group{
			FS:    os.DirFS("/tmp"),
			Paths: []string{"."}, // Look in '/tmp/.'
		}))
	if err != nil {
		panic(err)
	}

	err = g.Compile()
	if err != nil {
		panic(err)
	}

	var cfg struct {
		Example struct {
			Version int
			Colors  []string
		}
	}

	err = g.Unmarshal("", &cfg)
	if err != nil {
		panic(err)
	}

	fmt.Println("example")
	fmt.Printf("    version = %d\n", cfg.Example.Version)
	fmt.Printf("    colors  = [ %s ]\n", strings.Join(cfg.Example.Colors, ", "))

	// Output:
	// example
	//     version = 1
	//     colors  = [ red, green, blue ]
}
