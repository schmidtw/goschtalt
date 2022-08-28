// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt_test

import (
	"fmt"

	"github.com/schmidtw/goschtalt"
	"github.com/schmidtw/goschtalt/pkg/decoder"
	"github.com/schmidtw/goschtalt/pkg/meta"
)

// This is a fake decoder that allows the OrderList() to correctly work without
// needing to bring any external decoders into goschtalt.
type fake struct{}

func (f fake) Decode(_ decoder.Context, _ []byte, _ *meta.Object) error { return nil }
func (f fake) Extensions() []string                                     { return []string{"yml", "yaml", "json"} }

func ExampleConfig_OrderList() {
	g, err := goschtalt.New(goschtalt.DecoderRegister(fake{}))
	if err != nil {
		panic(err)
	}

	err = g.Compile()
	if err != nil {
		panic(err)
	}

	files := []string{
		"file_2.json",
		"file_1.yml",
		"file_1.txt",
		"S99_file.yml",
		"S58_file.json",
	}
	list := g.OrderList(files)

	fmt.Println("files:")
	for _, file := range list {
		fmt.Printf("\t%s\n", file)
	}

	// Output:
	// files:
	// 	S58_file.json
	// 	S99_file.yml
	// 	file_1.yml
	// 	file_2.json
}
