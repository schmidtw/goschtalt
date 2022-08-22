// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package cli_test

import (
	"fmt"
	"strings"

	"github.com/schmidtw/goschtalt"
	"github.com/schmidtw/goschtalt/extensions/decoders/cli"
)

func Example() {
	args := []string{
		"--kvp", "example.version", "1",
		"--kvp", "example.colors.0", "red",
		"--kvp", "example.colors.1", "green",
		"--kvp", "example.colors.2", "blue",
	}
	g, err := goschtalt.New(cli.Options("cli", ".", args)...)
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
