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

// Package casbab is a Go library for converting
// representation style of compound words or phrases.
// Different writing styles of compound words are used
// for different purposes in computer code and variables
// to easily distinguish type, properties or meaning.
//
// Functions in this package are separating words from
// input string and constructing an appropriate phrase
// representation.
//
// Examples:
//
//	kebab("camel_snake_kebab") == "camel-snake-kebab"
//	screamingSnake("camel_snake_kebab") == "CAMEL_SNAKE_KEBAB"
//	camel("camel_snake_kebab") == "camelSnakeKebab"
//	pascal("camel_snake_kebab") == "CamelSnakeKebab"
//	snake("camelSNAKEKebab") == "camel_snake_kebab"
//
// Word separation works by detecting delimiters hyphen (-),
// underscore (_), space ( ) and letter case change.
//
// Note: Leading and trailing separators will be preserved
// only within the Snake family or within the Kebab family
// and not across them. This restriction is based on different
// semantics between different writings.
//
// Examples:
//
//	titleSnake("__camel_snake_kebab__") == "__Camel_Snake_Kebab__"
//	kebab("__camel_snake_kebab") == "camel-snake-kebab"
//	screaming("__camel_snake_kebab") == "CAMEL SNAKE KEBAB"
//	titleKebab("--camel-snake-kebab") == "--Camel-Snake-Kebab"
//	snake("--camel-snake-kebab") == "camel_snake_kebab"
//	screaming("--camel-snake-kebab") == "CAMEL SNAKE KEBAB"
package casbab

import (
	"strings"
)

// Find returns the function based on the capitalization scheme desired, or
// nil if it doesn't recognize the scheme.
//
// The complete list:
//
//   - "two words" aliases: "lower case"
//   - "two-words" aliases: "kebab-case"
//   - "two-Words" aliases: "camel-Kebab-Case"
//   - "two_words" aliases: "snake_case"
//   - "two_Words" aliases: "camel_Snake_Case"
//   - "twowords"  aliases: "flatcase"
//   - "twoWords"  aliases: "camelCase"
//   - "Two Words" aliases: "Title Case"
//   - "Two-Words" aliases: "Pascal-Kebab-Case", "Title-Kebab-Case"
//   - "Two_Words" aliases: "Pascal_Snake_Case", "Title_Snake_Case"
//   - "TwoWords"  aliases: "PascalCase"
//   - "TWO WORDS" aliases: "SCREAMING CASE"
//   - "TWO-WORDS" aliases: "SCREAMING-KEBAB-CASE"
//   - "TWO_WORDS" aliases: "SCREAMING_SNAKE_CASE"
//   - "TWOWORDS"  aliases: "UPPERCASE"
func Find(s string) func(string) string {
	switch s {
	case "two words", "lower case":
		return lower
	case "two-words", "kebab-case":
		return kebab
	case "two-Words", "camel-kebab-case":
		return camelKebab
	case "two_words", "snake_case":
		return snake
	case "two_Words", "camel_Snake_Case":
		return camelSnake
	case "twowords", "flatcase":
		return flat
	case "twoWords", "camelCase":
		return camel
	case "Two Words", "Title Case":
		return title
	case "Two-Words", "Pascal-Kebab-Case", "Title-Kebab-Case":
		return titleKebab
	case "Two_Words", "Pascal_Snake_Case", "Title_Snake_Case":
		return titleSnake
	case "TwoWords", "PascalCase":
		return pascal
	case "TWO WORDS", "SCREAMING CASE":
		return screaming
	case "TWO-WORDS", "SCREAMING-KEBAB-CASE":
		return screamingKebab
	case "TWO_WORDS", "SCREAMING_SNAKE_CASE":
		return screamingSnake
	case "TWOWORDS", "UPPERCASE":
		return upper
	}

	return nil
}

// camel case is the practice of writing compound words
// or phrases such that each word or abbreviation in the
// middle of the phrase begins with a capital letter,
// with no spaces or hyphens.
//
// Example: "camelSnakeKebab".
func camel(s string) string {
	return strings.Join(capitalize(words(s), 1), "")
}

// pascal case is a variant of Camel case writing where
// the first letter of the first word is always capitalized.
//
// Example: "CamelSnakeKebab".
func pascal(s string) string {
	return strings.Join(capitalize(words(s), 0), "")
}

// snake case is the practice of writing compound words
// or phrases in which the elements are separated with
// one underscore character (_) and no spaces, with all
// element letters lowercased within the compound.
//
// Example: "camel_snake_kebab".
func snake(s string) string {
	head, tail := headTailCount(s, '_')
	return strings.Repeat("_", head) + strings.Join(words(s), "_") + strings.Repeat("_", tail)
}

// titleSnake case is a variant of Camel case with
// each element's first letter uppercased.
//
// Example: "Camel_Snake_Kebab".
func titleSnake(s string) string {
	head, tail := headTailCount(s, '_')
	return strings.Repeat("_", head) + strings.Join(capitalize(words(s), 0), "_") + strings.Repeat("_", tail)
}

// camelSnake case is a variant of Camel case with
// each element's first letter uppercased, except the first.
//
// Example: "camel_Snake_Kebab".
func camelSnake(s string) string {
	head, tail := headTailCount(s, '_')
	return strings.Repeat("_", head) + strings.Join(capitalize(words(s), 1), "_") + strings.Repeat("_", tail)
}

// screamingSnake case is a variant of Camel case with
// all letters uppercased.
//
// Example: "CAMEL_SNAKE_KEBAB".
func screamingSnake(s string) string {
	head, tail := headTailCount(s, '_')
	return strings.Repeat("_", head) + strings.Join(scream(words(s)), "_") + strings.Repeat("_", tail)
}

// kebab case is the practice of writing compound words
// or phrases in which the elements are separated with
// one hyphen character (-) and no spaces, with all
// element letters lowercased within the compound.
//
// Example: "camel-snake-kebab".
func kebab(s string) string {
	head, tail := headTailCount(s, '-')
	return strings.Repeat("-", head) + strings.Join(words(s), "-") + strings.Repeat("-", tail)
}

// titleKebab case is a variant of Kebab case with
// each element's first letter uppercased.
//
// Example: "Camel-Snake-Kebab".
func titleKebab(s string) string {
	head, tail := headTailCount(s, '-')
	return strings.Repeat("-", head) + strings.Join(capitalize(words(s), 0), "-") + strings.Repeat("-", tail)
}

// camelKebab case is a variant of Kebab case with
// each element's first letter uppercased, except the first.
//
// Example: "camel-Snake-Kebab".
func camelKebab(s string) string {
	head, tail := headTailCount(s, '-')
	return strings.Repeat("-", head) + strings.Join(capitalize(words(s), 1), "-") + strings.Repeat("-", tail)
}

// screamingKebab case is a variant of Kebab case with
// all letters uppercased.
//
// Example: "CAMEL-SNAKE-KEBAB".
func screamingKebab(s string) string {
	head, tail := headTailCount(s, '-')
	return strings.Repeat("-", head) + strings.Join(scream(words(s)), "-") + strings.Repeat("-", tail)
}

// lower is returning detected words, not in a compound
// form, but separated by one space character with all
// letters in lower case.
//
// Example: "camel snake kebab".
func lower(s string) string {
	return strings.Join(words(s), " ")
}

// flat is returning detected words, as a compond form with no separation
// character and all letters in lower case.
//
// Example: "camelsnakekebab".
func flat(s string) string {
	return strings.Join(words(s), "")
}

// title is returning detected words, not in a compound
// form, but separated by one space character with first
// character in all letters in upper case and all other
// letters in lower case.
//
// Example: "Camel Snake Kebab".
func title(s string) string {
	return strings.Join(capitalize(words(s), 0), " ")
}

// screaming is returning detected words, not in a compound
// form, but separated by one space character with all
// letters in upper case.
//
// Example: "CAMEL SNAKE KEBAB".
func screaming(s string) string {
	return strings.Join(scream(words(s)), " ")
}

// upper is returning detected words, as a compond form with no separation
// character with all letters in upper case.
//
// Example: "CAMELSNAKEKEBAB".
func upper(s string) string {
	return strings.Join(scream(words(s)), "")
}

func words(s string) (w []string) { //nolint:gocognit
	runes := []rune(s)
	start := 0
	l := len(runes)
	var prevLower, prevUpper bool

Loop:
	for i, c := range runes {
		switch c {
		case '-', '_', ' ':
			if start < i {
				w = append(w, toLower(runes[start:i]))
			}
			start = i + 1
			prevLower = false
			prevUpper = false
			continue Loop
		}
		if isUpper(c) {
			prevUpper = true
			if prevLower {
				if start < i {
					w = append(w, toLower(runes[start:i]))
				}
				start = i
				prevLower = false
			}
		} else {
			prevLower = true
			if prevUpper {
				// If the last letter is 's' and the previous character is an
				// uppercase letter, join the 's' to the previous word.  This
				// helps eliminate words like 'MyURLs' becoming 'my' 'ur' 'ls'
				// and instead makes them 'my' 'urls' as you would hope.
				if !isFinalS(c, runes[i+1:]) {
					if start < i-1 {
						w = append(w, toLower(runes[start:i-1]))
					}
					start = i - 1
					prevUpper = false
				}
			}
		}
		if i == l-1 {
			w = append(w, toLower(runes[start:]))
		}
	}
	return w
}

func isFinalS(c rune, rest []rune) bool {
	if c != 's' {
		return false
	}
	for _, r := range rest {
		switch r {
		case '-', '_', ' ':
		default:
			return false
		}
	}
	return true
}

func scream(s []string) []string {
	for i := 0; i < len(s); i++ {
		s[i] = strings.ToUpper(s[i])
	}
	return s
}

func capitalize(list []string, start int) []string {
	for i := start; i < len(list); i++ {
		if len([]rune(list[i])) < 2 {
			list[i] = strings.ToUpper(list[i])
		} else {
			runes := []rune(list[i])
			list[i] = toUpper(runes[0:1]) + string(runes[1:])
		}
	}
	return list
}

func headTailCount(s string, sub rune) (head, tail int) {
	r := []rune(s)
	for i := 0; i < len(r); i++ {
		if r[i] != sub {
			head = i
			break
		}
	}
	for i := len(r) - 1; i >= 0; i-- {
		if r[i] != sub {
			tail = len(r) - i - 1
			break
		}
	}
	return
}

func isUpper(r rune) bool {
	return strings.ToUpper(string(r)) == string(r)
}

func toLower(runes []rune) string {
	return strings.ToLower(string(runes))
}

func toUpper(runes []rune) string {
	return strings.ToUpper(string(runes))
}
