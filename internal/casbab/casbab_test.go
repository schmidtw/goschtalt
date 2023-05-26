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

import "testing"

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
				"Camel":           "camelSnakeKebab",
				"Pascal":          "CamelSnakeKebab",
				"Snake":           "camel_snake_kebab",
				"CamelSnake":      "Camel_Snake_Kebab",
				"ScreamingSnake":  "CAMEL_SNAKE_KEBAB",
				"Kebab":           "camel-snake-kebab",
				"CamelKebab":      "Camel-Snake-Kebab",
				"ScreamingKebab":  "CAMEL-SNAKE-KEBAB",
				"Lower":           "camel snake kebab",
				"Title":           "Camel Snake Kebab",
				"Screaming":       "CAMEL SNAKE KEBAB",
				"LowerCamelSnake": "camel_Snake_Kebab",
				"LowerCamelKebab": "camel-Snake-Kebab",
				"Flat":            "camelsnakekebab",
				"Upper":           "CAMELSNAKEKEBAB",
			},
		},
		{
			In: []string{
				"__camel_snake_kebab__",
				"__camel_snakeKEBAB__",
				"__ Camel-snakeKEBAB__",
			},
			Out: map[string]string{
				"Camel":           "camelSnakeKebab",
				"Pascal":          "CamelSnakeKebab",
				"Snake":           "__camel_snake_kebab__",
				"CamelSnake":      "__Camel_Snake_Kebab__",
				"ScreamingSnake":  "__CAMEL_SNAKE_KEBAB__",
				"Kebab":           "camel-snake-kebab",
				"CamelKebab":      "Camel-Snake-Kebab",
				"ScreamingKebab":  "CAMEL-SNAKE-KEBAB",
				"Lower":           "camel snake kebab",
				"Title":           "Camel Snake Kebab",
				"Screaming":       "CAMEL SNAKE KEBAB",
				"LowerCamelSnake": "__camel_Snake_Kebab__",
				"LowerCamelKebab": "camel-Snake-Kebab",
				"Flat":            "camelsnakekebab",
				"Upper":           "CAMELSNAKEKEBAB",
			},
		},
		{
			In: []string{
				"__camel_snake_kebabs__",
				"__camel_snakeKEBABs__",
				"__ Camel-snakeKEBABs__",
			},
			Out: map[string]string{
				"Camel":           "camelSnakeKebabs",
				"Pascal":          "CamelSnakeKebabs",
				"Snake":           "__camel_snake_kebabs__",
				"CamelSnake":      "__Camel_Snake_Kebabs__",
				"ScreamingSnake":  "__CAMEL_SNAKE_KEBABS__",
				"Kebab":           "camel-snake-kebabs",
				"CamelKebab":      "Camel-Snake-Kebabs",
				"ScreamingKebab":  "CAMEL-SNAKE-KEBABS",
				"Lower":           "camel snake kebabs",
				"Title":           "Camel Snake Kebabs",
				"Screaming":       "CAMEL SNAKE KEBABS",
				"LowerCamelSnake": "__camel_Snake_Kebabs__",
				"LowerCamelKebab": "camel-Snake-Kebabs",
				"Flat":            "camelsnakekebabs",
				"Upper":           "CAMELSNAKEKEBABS",
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
				"Camel":           "camelSnakeKebab",
				"Pascal":          "CamelSnakeKebab",
				"Snake":           "__camel_snake_kebab_",
				"CamelSnake":      "__Camel_Snake_Kebab_",
				"ScreamingSnake":  "__CAMEL_SNAKE_KEBAB_",
				"Kebab":           "camel-snake-kebab",
				"CamelKebab":      "Camel-Snake-Kebab",
				"ScreamingKebab":  "CAMEL-SNAKE-KEBAB",
				"Lower":           "camel snake kebab",
				"Title":           "Camel Snake Kebab",
				"Screaming":       "CAMEL SNAKE KEBAB",
				"LowerCamelSnake": "__camel_Snake_Kebab_",
				"LowerCamelKebab": "camel-Snake-Kebab",
				"Flat":            "camelsnakekebab",
				"Upper":           "CAMELSNAKEKEBAB",
			},
		},
		{
			In: []string{
				"--camel-snake-kebab",
				"--CAMELSnake_kebab",
			},
			Out: map[string]string{
				"Camel":           "camelSnakeKebab",
				"Pascal":          "CamelSnakeKebab",
				"Snake":           "camel_snake_kebab",
				"CamelSnake":      "Camel_Snake_Kebab",
				"ScreamingSnake":  "CAMEL_SNAKE_KEBAB",
				"Kebab":           "--camel-snake-kebab",
				"CamelKebab":      "--Camel-Snake-Kebab",
				"ScreamingKebab":  "--CAMEL-SNAKE-KEBAB",
				"Lower":           "camel snake kebab",
				"Title":           "Camel Snake Kebab",
				"Screaming":       "CAMEL SNAKE KEBAB",
				"LowerCamelSnake": "camel_Snake_Kebab",
				"LowerCamelKebab": "--camel-Snake-Kebab",
				"Flat":            "camelsnakekebab",
				"Upper":           "CAMELSNAKEKEBAB",
			},
		},
		{
			In: []string{
				"-camel-snake-kebab----",
				"-CAMEL   Snake_kebab ----",
			},
			Out: map[string]string{
				"Camel":           "camelSnakeKebab",
				"Pascal":          "CamelSnakeKebab",
				"Snake":           "camel_snake_kebab",
				"CamelSnake":      "Camel_Snake_Kebab",
				"ScreamingSnake":  "CAMEL_SNAKE_KEBAB",
				"Kebab":           "-camel-snake-kebab----",
				"CamelKebab":      "-Camel-Snake-Kebab----",
				"ScreamingKebab":  "-CAMEL-SNAKE-KEBAB----",
				"Lower":           "camel snake kebab",
				"Title":           "Camel Snake Kebab",
				"Screaming":       "CAMEL SNAKE KEBAB",
				"LowerCamelSnake": "camel_Snake_Kebab",
				"LowerCamelKebab": "-camel-Snake-Kebab----",
				"Flat":            "camelsnakekebab",
				"Upper":           "CAMELSNAKEKEBAB",
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
				"Camel":           "xCamelXxSnakeXxxKebab",
				"Pascal":          "XCamelXxSnakeXxxKebab",
				"Snake":           "x_camel_xx_snake_xxx_kebab",
				"CamelSnake":      "X_Camel_Xx_Snake_Xxx_Kebab",
				"ScreamingSnake":  "X_CAMEL_XX_SNAKE_XXX_KEBAB",
				"Kebab":           "x-camel-xx-snake-xxx-kebab",
				"CamelKebab":      "X-Camel-Xx-Snake-Xxx-Kebab",
				"ScreamingKebab":  "X-CAMEL-XX-SNAKE-XXX-KEBAB",
				"Lower":           "x camel xx snake xxx kebab",
				"Title":           "X Camel Xx Snake Xxx Kebab",
				"Screaming":       "X CAMEL XX SNAKE XXX KEBAB",
				"LowerCamelSnake": "x_Camel_Xx_Snake_Xxx_Kebab",
				"LowerCamelKebab": "x-Camel-Xx-Snake-Xxx-Kebab",
				"Flat":            "xcamelxxsnakexxxkebab",
				"Upper":           "XCAMELXXSNAKEXXXKEBAB",
			},
		},
		{
			In: []string{
				"",
			},
			Out: map[string]string{
				"Camel":           "",
				"Pascal":          "",
				"Snake":           "",
				"CamelSnake":      "",
				"ScreamingSnake":  "",
				"Kebab":           "",
				"CamelKebab":      "",
				"ScreamingKebab":  "",
				"Lower":           "",
				"Title":           "",
				"Screaming":       "",
				"LowerCamelSnake": "",
				"LowerCamelKebab": "",
				"Flat":            "",
				"Upper":           "",
			},
		},
		{
			In: []string{
				"I♥You",
			},
			Out: map[string]string{
				"Camel":           "i♥You",
				"Pascal":          "I♥You",
				"Snake":           "i♥_you",
				"CamelSnake":      "I♥_You",
				"ScreamingSnake":  "I♥_YOU",
				"Kebab":           "i♥-you",
				"CamelKebab":      "I♥-You",
				"ScreamingKebab":  "I♥-YOU",
				"Lower":           "i♥ you",
				"Title":           "I♥ You",
				"Screaming":       "I♥ YOU",
				"LowerCamelSnake": "i♥_You",
				"LowerCamelKebab": "i♥-You",
				"Flat":            "i♥you",
				"Upper":           "I♥YOU",
			},
		},
		{
			In: []string{
				"I ♥ You",
			},
			Out: map[string]string{
				"Camel":           "i♥You",
				"Pascal":          "I♥You",
				"Snake":           "i_♥_you",
				"CamelSnake":      "I_♥_You",
				"ScreamingSnake":  "I_♥_YOU",
				"Kebab":           "i-♥-you",
				"CamelKebab":      "I-♥-You",
				"ScreamingKebab":  "I-♥-YOU",
				"Lower":           "i ♥ you",
				"Title":           "I ♥ You",
				"Screaming":       "I ♥ YOU",
				"LowerCamelSnake": "i_♥_You",
				"LowerCamelKebab": "i-♥-You",
				"Flat":            "i♥you",
				"Upper":           "I♥YOU",
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
				"Camel":           "acls",
				"Pascal":          "Acls",
				"Snake":           "acls",
				"CamelSnake":      "Acls",
				"ScreamingSnake":  "ACLS",
				"Kebab":           "acls",
				"CamelKebab":      "Acls",
				"ScreamingKebab":  "ACLS",
				"Lower":           "acls",
				"Title":           "Acls",
				"Screaming":       "ACLS",
				"LowerCamelSnake": "acls",
				"LowerCamelKebab": "acls",
				"Flat":            "acls",
				"Upper":           "ACLS",
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
				"Camel":           "moonsAsPlanets",
				"Pascal":          "MoonsAsPlanets",
				"Snake":           "moons_as_planets",
				"CamelSnake":      "Moons_As_Planets",
				"ScreamingSnake":  "MOONS_AS_PLANETS",
				"Kebab":           "moons-as-planets",
				"CamelKebab":      "Moons-As-Planets",
				"ScreamingKebab":  "MOONS-AS-PLANETS",
				"Lower":           "moons as planets",
				"Title":           "Moons As Planets",
				"Screaming":       "MOONS AS PLANETS",
				"LowerCamelSnake": "moons_As_Planets",
				"LowerCamelKebab": "moons-As-Planets",
				"Flat":            "moonsasplanets",
				"Upper":           "MOONSASPLANETS",
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
				"Camel":  "defaultTtls",
				"Pascal": "DefaultTtls",
				"Snake":  "default_ttls",
			},
		},
	}
	converters = map[string]func(string) string{
		"Camel":           Camel,
		"Pascal":          Pascal,
		"Snake":           Snake,
		"CamelSnake":      CamelSnake,
		"ScreamingSnake":  ScreamingSnake,
		"Kebab":           Kebab,
		"CamelKebab":      CamelKebab,
		"ScreamingKebab":  ScreamingKebab,
		"Lower":           Lower,
		"Title":           Title,
		"Screaming":       Screaming,
		"LowerCamelSnake": LowerCamelSnake,
		"LowerCamelKebab": LowerCamelKebab,
		"Flat":            Flat,
		"Upper":           Upper,
	}
)

func Test(t *testing.T) {
	for _, c := range cases {
		for _, in := range c.In {
			for converter, expected := range c.Out {
				got := converters[converter](in)
				if got != expected {
					t.Errorf("Converting %q to %s expected %q, but got %q", in, converter, expected, got)
				}
			}
		}
	}
}

var benchmarkPhrase = "xCAMELSnakeKebab_screaming pascal XXX"

func BenchmarkCamel(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Camel(benchmarkPhrase)
	}
}

func BenchmarkPascal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Pascal(benchmarkPhrase)
	}
}

func BenchmarkSnake(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Snake(benchmarkPhrase)
	}
}
func BenchmarkCamelSnake(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CamelSnake(benchmarkPhrase)
	}
}

func BenchmarkScreamingSnake(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ScreamingSnake(benchmarkPhrase)
	}
}

func BenchmarkKebab(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Kebab(benchmarkPhrase)
	}
}

func BenchmarkCamelKebab(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CamelKebab(benchmarkPhrase)
	}
}

func BenchmarkScreamingKebab(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ScreamingKebab(benchmarkPhrase)
	}
}

func BenchmarkLower(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Lower(benchmarkPhrase)
	}
}

func BenchmarkTitle(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Title(benchmarkPhrase)
	}
}

func BenchmarkScreaming(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Screaming(benchmarkPhrase)
	}
}
