// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package env_test

import (
	"fmt"
	"os"
	"strings"

	"github.com/schmidtw/goschtalt"
	"github.com/schmidtw/goschtalt/extensions/decoders/env"
)

func Example() {
	_ = os.Setenv("EXAMexample_version", "1")
	_ = os.Setenv("EXAMexample_colors_0", "red")
	_ = os.Setenv("EXAMexample_colors_1", "green")
	_ = os.Setenv("EXAMexample_colors_2", "blue")
	g, err := goschtalt.New(env.EnvVarConfig("OrderFilename", "EXAM", "_")...)
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
