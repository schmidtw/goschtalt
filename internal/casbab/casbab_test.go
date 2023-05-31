// SPDX-FileCopyrightText: 2016 Janoš Guljaš <janos@resenje.org>
// SPDX-FileCopyrightText: 2023 Weston Schmidt
// SPDX-License-Identifier: BSD-3-Clause
//
// This file originated from https://github.com/janos/casbab
// The following modifications were made:
//   - Strings are processed as runes, not bytes so unicode characters are supported.
//   - Trailing 's' characters at the end of the last word is grouped with the prior
//     word instead of being made a new 2 letter word.
//   - A few additional styles were added.

package casbab

import (
	"strings"
	"testing"
)

var (
	cases = []struct {
		In  []string
		Out map[string]string
	}{
		{
			In: []string{
				"camelSnakeKebab",
				"CamelSnakeKebab",
				"camel_snake_kebab",
				"Camel_Snake_Kebab",
				"CAMEL_SNAKE_KEBAB",
				"camel-snake-kebab",
				"Camel-Snake-Kebab",
				"CAMEL-SNAKE-KEBAB",
				"camel snake kebab",
				"Camel Snake Kebab",
				"CAMEL SNAKE KEBAB",
				"camel__snake_kebab",
				"camel___snake_kebab",
				"camel____snake_kebab",
				"camel_ snake_kebab",
				"camel_  snake_kebab",
				"camel_   snake_kebab",
				"camel_-snake_kebab",
				"camel_ -snake_kebab",
				"camel - snake_kebab",
				" camel - snake_kebab",
				"CAMELSnakeKebab",
				"camelSNAKEKebab   ",
				"   camelSnakeKEBAB",
			},
			Out: map[string]string{
				"twoWords":  "camelSnakeKebab",
				"TwoWords":  "CamelSnakeKebab",
				"two_words": "camel_snake_kebab",
				"Two_Words": "Camel_Snake_Kebab",
				"TWO_WORDS": "CAMEL_SNAKE_KEBAB",
				"two-words": "camel-snake-kebab",
				"Two-Words": "Camel-Snake-Kebab",
				"TWO-WORDS": "CAMEL-SNAKE-KEBAB",
				"two words": "camel snake kebab",
				"Two Words": "Camel Snake Kebab",
				"TWO WORDS": "CAMEL SNAKE KEBAB",
				"two_Words": "camel_Snake_Kebab",
				"two-Words": "camel-Snake-Kebab",
				"twowords":  "camelsnakekebab",
				"TWOWORDS":  "CAMELSNAKEKEBAB",
			},
		},

		{
			In: []string{
				"__camel_snake_kebab__",
				"__camel_snakeKEBAB__",
				"__ Camel-snakeKEBAB__",
			},
			Out: map[string]string{
				"twoWords":  "camelSnakeKebab",
				"TwoWords":  "CamelSnakeKebab",
				"two_words": "__camel_snake_kebab__",
				"Two_Words": "__Camel_Snake_Kebab__",
				"TWO_WORDS": "__CAMEL_SNAKE_KEBAB__",
				"two-words": "camel-snake-kebab",
				"Two-Words": "Camel-Snake-Kebab",
				"TWO-WORDS": "CAMEL-SNAKE-KEBAB",
				"two words": "camel snake kebab",
				"Two Words": "Camel Snake Kebab",
				"TWO WORDS": "CAMEL SNAKE KEBAB",
				"two_Words": "__camel_Snake_Kebab__",
				"two-Words": "camel-Snake-Kebab",
				"twowords":  "camelsnakekebab",
				"TWOWORDS":  "CAMELSNAKEKEBAB",
			},
		},
		{
			In: []string{
				"__camel_snake_kebabs__",
				"__camel_snakeKEBABs__",
				"__ Camel-snakeKEBABs__",
			},
			Out: map[string]string{
				"twoWords":  "camelSnakeKebabs",
				"TwoWords":  "CamelSnakeKebabs",
				"two_words": "__camel_snake_kebabs__",
				"Two_Words": "__Camel_Snake_Kebabs__",
				"TWO_WORDS": "__CAMEL_SNAKE_KEBABS__",
				"two-words": "camel-snake-kebabs",
				"Two-Words": "Camel-Snake-Kebabs",
				"TWO-WORDS": "CAMEL-SNAKE-KEBABS",
				"two words": "camel snake kebabs",
				"Two Words": "Camel Snake Kebabs",
				"TWO WORDS": "CAMEL SNAKE KEBABS",
				"two_Words": "__camel_Snake_Kebabs__",
				"two-Words": "camel-Snake-Kebabs",
				"twowords":  "camelsnakekebabs",
				"TWOWORDS":  "CAMELSNAKEKEBABS",
			},
		},
		{
			In: []string{
				"__ camel-snake_kebab__ _",
				"__ camelSnake_Kebab_",
				"__CamelSnake_Kebab_",
				"__CamelSNAKE_Kebab_",
			},
			Out: map[string]string{
				"twoWords":  "camelSnakeKebab",
				"TwoWords":  "CamelSnakeKebab",
				"two_words": "__camel_snake_kebab_",
				"Two_Words": "__Camel_Snake_Kebab_",
				"TWO_WORDS": "__CAMEL_SNAKE_KEBAB_",
				"two-words": "camel-snake-kebab",
				"Two-Words": "Camel-Snake-Kebab",
				"TWO-WORDS": "CAMEL-SNAKE-KEBAB",
				"two words": "camel snake kebab",
				"Two Words": "Camel Snake Kebab",
				"TWO WORDS": "CAMEL SNAKE KEBAB",
				"two_Words": "__camel_Snake_Kebab_",
				"two-Words": "camel-Snake-Kebab",
				"twowords":  "camelsnakekebab",
				"TWOWORDS":  "CAMELSNAKEKEBAB",
			},
		},
		{
			In: []string{
				"--camel-snake-kebab",
				"--CAMELSnake_kebab",
			},
			Out: map[string]string{
				"twoWords":  "camelSnakeKebab",
				"TwoWords":  "CamelSnakeKebab",
				"two_words": "camel_snake_kebab",
				"Two_Words": "Camel_Snake_Kebab",
				"TWO_WORDS": "CAMEL_SNAKE_KEBAB",
				"two-words": "--camel-snake-kebab",
				"Two-Words": "--Camel-Snake-Kebab",
				"TWO-WORDS": "--CAMEL-SNAKE-KEBAB",
				"two words": "camel snake kebab",
				"Two Words": "Camel Snake Kebab",
				"TWO WORDS": "CAMEL SNAKE KEBAB",
				"two_Words": "camel_Snake_Kebab",
				"two-Words": "--camel-Snake-Kebab",
				"twowords":  "camelsnakekebab",
				"TWOWORDS":  "CAMELSNAKEKEBAB",
			},
		},
		{
			In: []string{
				"-camel-snake-kebab----",
				"-CAMEL   Snake_kebab ----",
			},
			Out: map[string]string{
				"twoWords":  "camelSnakeKebab",
				"TwoWords":  "CamelSnakeKebab",
				"two_words": "camel_snake_kebab",
				"Two_Words": "Camel_Snake_Kebab",
				"TWO_WORDS": "CAMEL_SNAKE_KEBAB",
				"two-words": "-camel-snake-kebab----",
				"Two-Words": "-Camel-Snake-Kebab----",
				"TWO-WORDS": "-CAMEL-SNAKE-KEBAB----",
				"two words": "camel snake kebab",
				"Two Words": "Camel Snake Kebab",
				"TWO WORDS": "CAMEL SNAKE KEBAB",
				"two_Words": "camel_Snake_Kebab",
				"two-Words": "-camel-Snake-Kebab----",
				"twowords":  "camelsnakekebab",
				"TWOWORDS":  "CAMELSNAKEKEBAB",
			},
		},
		{
			In: []string{
				"xCamelXXSnakeXXXKebab",
				"XCamelXXSnakeXXXKebab",
				"x_camel_xx_snake_xxx_kebab",
				"X_Camel_XX_Snake_XXX_Kebab",
				"X_CAMEL_XX_SNAKE_XXX_KEBAB",
				"x-camel-xx-snake-xxx-kebab",
				"X-Camel-XX_Snake-XXX-Kebab",
				"X-CAMEL-XX_SNAKE-XXX-KEBAB",
				"x camel xx snake xxx kebab",
				"X Camel XX Snake XXX Kebab",
				"X CAMEL XX SNAKE XXX KEBAB",
			},
			Out: map[string]string{
				"twoWords":  "xCamelXxSnakeXxxKebab",
				"TwoWords":  "XCamelXxSnakeXxxKebab",
				"two_words": "x_camel_xx_snake_xxx_kebab",
				"Two_Words": "X_Camel_Xx_Snake_Xxx_Kebab",
				"TWO_WORDS": "X_CAMEL_XX_SNAKE_XXX_KEBAB",
				"two-words": "x-camel-xx-snake-xxx-kebab",
				"Two-Words": "X-Camel-Xx-Snake-Xxx-Kebab",
				"TWO-WORDS": "X-CAMEL-XX-SNAKE-XXX-KEBAB",
				"two words": "x camel xx snake xxx kebab",
				"Two Words": "X Camel Xx Snake Xxx Kebab",
				"TWO WORDS": "X CAMEL XX SNAKE XXX KEBAB",
				"two_Words": "x_Camel_Xx_Snake_Xxx_Kebab",
				"two-Words": "x-Camel-Xx-Snake-Xxx-Kebab",
				"twowords":  "xcamelxxsnakexxxkebab",
				"TWOWORDS":  "XCAMELXXSNAKEXXXKEBAB",
			},
		},
		{
			In: []string{
				"",
			},
			Out: map[string]string{
				"twoWords":  "",
				"TwoWords":  "",
				"two_words": "",
				"Two_Words": "",
				"TWO_WORDS": "",
				"two-words": "",
				"Two-Words": "",
				"TWO-WORDS": "",
				"two words": "",
				"Two Words": "",
				"TWO WORDS": "",
				"two_Words": "",
				"two-Words": "",
				"twowords":  "",
				"TWOWORDS":  "",
			},
		},
		{
			In: []string{
				"I♥You",
			},
			Out: map[string]string{
				"twoWords":  "i♥You",
				"TwoWords":  "I♥You",
				"two_words": "i♥_you",
				"Two_Words": "I♥_You",
				"TWO_WORDS": "I♥_YOU",
				"two-words": "i♥-you",
				"Two-Words": "I♥-You",
				"TWO-WORDS": "I♥-YOU",
				"two words": "i♥ you",
				"Two Words": "I♥ You",
				"TWO WORDS": "I♥ YOU",
				"two_Words": "i♥_You",
				"two-Words": "i♥-You",
				"twowords":  "i♥you",
				"TWOWORDS":  "I♥YOU",
			},
		},
		{
			In: []string{
				"I ♥ You",
			},
			Out: map[string]string{
				"twoWords":  "i♥You",
				"TwoWords":  "I♥You",
				"two_words": "i_♥_you",
				"Two_Words": "I_♥_You",
				"TWO_WORDS": "I_♥_YOU",
				"two-words": "i-♥-you",
				"Two-Words": "I-♥-You",
				"TWO-WORDS": "I-♥-YOU",
				"two words": "i ♥ you",
				"Two Words": "I ♥ You",
				"TWO WORDS": "I ♥ YOU",
				"two_Words": "i_♥_You",
				"two-Words": "i-♥-You",
				"twowords":  "i♥you",
				"TWOWORDS":  "I♥YOU",
			},
		},
		{
			In: []string{
				"ACLs",
				"ACLs ",
				" ACLs",
				" ACLs ",
			},
			Out: map[string]string{
				"twoWords":  "acls",
				"TwoWords":  "Acls",
				"two_words": "acls",
				"Two_Words": "Acls",
				"TWO_WORDS": "ACLS",
				"two-words": "acls",
				"Two-Words": "Acls",
				"TWO-WORDS": "ACLS",
				"two words": "acls",
				"Two Words": "Acls",
				"TWO WORDS": "ACLS",
				"two_Words": "acls",
				"two-Words": "acls",
				"twowords":  "acls",
				"TWOWORDS":  "ACLS",
			},
		},
		{
			In: []string{
				"MoonsAsPlanets",
				"MoonsAsPlanets ",
				" MoonsAsPlanets",
				" MoonsAsPlanets ",
			},
			Out: map[string]string{
				"twoWords":  "moonsAsPlanets",
				"TwoWords":  "MoonsAsPlanets",
				"two_words": "moons_as_planets",
				"Two_Words": "Moons_As_Planets",
				"TWO_WORDS": "MOONS_AS_PLANETS",
				"two-words": "moons-as-planets",
				"Two-Words": "Moons-As-Planets",
				"TWO-WORDS": "MOONS-AS-PLANETS",
				"two words": "moons as planets",
				"Two Words": "Moons As Planets",
				"TWO WORDS": "MOONS AS PLANETS",
				"two_Words": "moons_As_Planets",
				"two-Words": "moons-As-Planets",
				"twowords":  "moonsasplanets",
				"TWOWORDS":  "MOONSASPLANETS",
			},
		},
		{
			In: []string{
				"DefaultTTLs",
				"DefaultTTLs ",
				" DefaultTTLs ",
				" DefaultTTLs",
			},
			Out: map[string]string{
				"twoWords":  "defaultTtls",
				"TwoWords":  "DefaultTtls",
				"two_words": "default_ttls",
			},
		},
		{
			In: []string{
				"single",
				"Single",
				"SINGLE",
				" single",
				"   single",
			},
			Out: map[string]string{
				"twoWords":  "single",
				"TwoWords":  "Single",
				"two_words": "single",
				"Two_Words": "Single",
				"TWO_WORDS": "SINGLE",
				"two-words": "single",
				"Two-Words": "Single",
				"TWO-WORDS": "SINGLE",
				"two words": "single",
				"Two Words": "Single",
				"TWO WORDS": "SINGLE",
				"two_Words": "single",
				"two-Words": "single",
				"twowords":  "single",
				"TWOWORDS":  "SINGLE",
			},
		},
		{
			In: []string{
				"",
				" ",
			},
			Out: map[string]string{
				"twoWords":  "",
				"TwoWords":  "",
				"two_words": "",
				"Two_Words": "",
				"TWO_WORDS": "",
				"two-words": "",
				"Two-Words": "",
				"TWO-WORDS": "",
				"two words": "",
				"Two Words": "",
				"TWO WORDS": "",
				"two_Words": "",
				"two-Words": "",
				"twowords":  "",
				"TWOWORDS":  "",
			},
		},
		{
			In: []string{
				"a",
				"A",
				" a",
				" A",
			},
			Out: map[string]string{
				"twoWords":  "a",
				"TwoWords":  "A",
				"two_words": "a",
				"Two_Words": "A",
				"TWO_WORDS": "A",
				"two-words": "a",
				"Two-Words": "A",
				"TWO-WORDS": "A",
				"two words": "a",
				"Two Words": "A",
				"TWO WORDS": "A",
				"two_Words": "a",
				"two-Words": "a",
				"twowords":  "a",
				"TWOWORDS":  "A",
			},
		},
	}
)

func Test(t *testing.T) {
	for _, c := range cases {
		for _, in := range c.In {
			for converter, expected := range c.Out {
				got := Find(converter)(in)
				if got != expected {
					t.Errorf("Converting %q to %s expected %q, but got %q", in, converter, expected, got)
				}
			}
		}
	}
}

func TestFindingInvalid(t *testing.T) {
	got := Find("invalid")
	if got != nil {
		t.Errorf("Expected to find nil.")
	}
}

var benchmarkPhrase = "xCAMELSnakeKebab_screaming pascal XXX"

func benchmark(b *testing.B, which string) {
	fn := Find(which)
	if fn == nil {
		return
	}
	for i := 0; i < b.N; i++ {
		fn(benchmarkPhrase)
	}
}

func BenchmarkCamel(b *testing.B) {
	benchmark(b, "camelCase")
}

func BenchmarkPascal(b *testing.B) {
	benchmark(b, "PascalCase")
}

func BenchmarkSnake(b *testing.B) {
	benchmark(b, "snake_case")
}

func BenchmarkTitleSnake(b *testing.B) {
	benchmark(b, "Title_Snake_Case")
}

func BenchmarkScreamingSnake(b *testing.B) {
	benchmark(b, "SCREAMING_SNAKE_CASE")
}

func BenchmarkKebab(b *testing.B) {
	benchmark(b, "kebab-case")
}

func BenchmarkTitleKebab(b *testing.B) {
	benchmark(b, "title-kebab-case")
}

func BenchmarkScreamingKebab(b *testing.B) {
	benchmark(b, "SCREAMING-KEBAB-CASE")
}

func BenchmarkLower(b *testing.B) {
	benchmark(b, "lower case")
}

func BenchmarkTitle(b *testing.B) {
	benchmark(b, "Title Case")
}

func BenchmarkScreaming(b *testing.B) {
	benchmark(b, "SCREAMING CASE")
}

func FuzzAll(f *testing.F) {
	f.Fuzz(func(t *testing.T, in string) {
		list := []string{
			"twoWords",
			"TwoWords",
			"two_words",
			"Two_Words",
			"TWO_WORDS",
			"two-words",
			"Two-Words",
			"TWO-WORDS",
			"two words",
			"Two Words",
			"TWO WORDS",
			"two_Words",
			"two-Words",
			"twowords",
			"TWOWORDS",
		}

		for _, item := range list {
			fn := Find(item)
			got := fn(in)

			min := strings.ReplaceAll(in, "-", "")
			min = strings.ReplaceAll(min, "_", "")
			min = strings.ReplaceAll(min, " ", "")

			if len([]rune(got)) < len([]rune(min)) {
				t.Errorf("the final %q is too short %q\n", got, in)
			}

			switch item {
			case "twoWords", "TwoWords", "twowords", "TWOWORDS":
				if strings.ContainsAny(got, "-_ ") {
					t.Errorf("the final %q is contains '-_ ' %q\n", got, in)
				}
			case "two_words", "Two_Words", "TWO_WORDS", "two_Words":
				if strings.ContainsAny(got, "- ") {
					t.Errorf("the final %q is contains '- ' %q\n", got, in)
				}
			case "two-words", "Two-Words", "TWO-WORDS", "two-Words":
				if strings.ContainsAny(got, "_ ") {
					t.Errorf("the final %q is contains '_ ' %q\n", got, in)
				}
			case "two words", "Two Words", "TWO WORDS", "two Words":
				if strings.ContainsAny(got, "-_") {
					t.Errorf("the final %q is contains '-_' %q\n", got, in)
				}
			default:
				t.Errorf("the input of %q is missed", item)
			}
		}
	})
}
