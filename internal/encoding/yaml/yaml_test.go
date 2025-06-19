// SPDX-FileCopyrightText: 2025 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/goschtalt/goschtalt/internal/encoding"
	"github.com/k0kubun/pp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ymlTest struct {
	desc     string
	key      *string
	val      *string
	indent   int
	headers  []string
	inline   []string
	children []ymlTest
	r        Renderer
	expect   string
}

func pstr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

var _ encoding.Encodeable = &ymlTest{}

func (t *ymlTest) Key() *string {
	return t.key
}

func (t *ymlTest) Value() *string {
	return t.val
}

func (t *ymlTest) Indent() int {
	return t.indent
}
func (t *ymlTest) Headers() []string {
	return t.headers
}
func (t *ymlTest) Inline() []string {
	return t.inline
}
func (t *ymlTest) Children() encoding.Encodeables {
	if len(t.children) == 0 {
		return nil
	}

	children := make([]encoding.Encodeable, len(t.children))
	for i, child := range t.children {
		children[i] = &child
	}
	return children
}

var basicYMLTests = []ymlTest{
	{
		desc:   "empty, no comments",
		key:    pstr("a"),
		expect: "a: ''\n",
	}, {
		desc:   "simple, no comments",
		key:    pstr("a"),
		val:    pstr("b"),
		expect: "a: b\n",
	}, {
		desc:   "simple, single comment",
		key:    pstr("a"),
		val:    pstr("b"),
		inline: []string{"comment"},
		expect: "a: b # comment\n",
	}, {
		desc:   "simple, single comment, spaced over",
		key:    pstr("a"),
		val:    pstr("b"),
		inline: []string{"comment"},
		r: Renderer{
			MaxLineLength:         20,
			TrailingCommentColumn: 20,
			SpacesPerIndent:       2,
		},
		expect: "a: b                # comment\n",
	}, {
		desc:    "simple, single comment with header comment",
		key:     pstr("a"),
		val:     pstr("b"),
		headers: []string{"header comment"},
		inline:  []string{"comment"},
		expect: "" +
			"# header comment\n" +
			"a: b # comment\n",
	}, {
		desc:    "simple, multi-line comment with header comment",
		key:     pstr("a"),
		val:     pstr("b"),
		headers: []string{"header comment"},
		inline:  []string{"comment 1", "comment 2"},
		expect: "" +
			"# header comment\n" +
			"a:\n" +
			"  # comment 1\n" +
			"  # comment 2\n" +
			"  b\n",
	}, {
		desc:    "simple, indented multi-line comment with header comment",
		key:     pstr("a"),
		val:     pstr("b"),
		indent:  1,
		headers: []string{"header comment"},
		inline:  []string{"comment 1", "comment 2"},
		expect: "" +
			"  # header comment\n" +
			"  a:\n" +
			"    # comment 1\n" +
			"    # comment 2\n" +
			"    b\n",
	}, {
		desc:    "simple, indented multi-line comment with header comment",
		val:     pstr("b"),
		indent:  1,
		headers: []string{"header comment"},
		inline:  []string{"comment 1", "comment 2"},
		expect: "" +
			"  # header comment\n" +
			"  -\n" +
			"    # comment 1\n" +
			"    # comment 2\n" +
			"    b\n",
	},
}

var quotedYMLTests = []ymlTest{
	{
		desc:   "complex val, no comments",
		key:    pstr("a"),
		val:    pstr("&b"),
		expect: "a: \"&b\"\n",
	}, {
		desc:   "complex val, single comment",
		key:    pstr("a"),
		val:    pstr("&b"),
		inline: []string{"comment"},
		expect: "a: \"&b\" # comment\n",
	}, {
		desc:    "complex val, single comment with header comment",
		key:     pstr("a"),
		val:     pstr("&b"),
		headers: []string{"header comment"},
		inline:  []string{"comment"},
		expect: "" +
			"# header comment\n" +
			"a: \"&b\" # comment\n",
	}, {
		desc:    "complex val, multiple comment with header comment",
		key:     pstr("a"),
		val:     pstr("&b"),
		headers: []string{"header comment"},
		inline:  []string{"comment 1", "comment 2"},
		expect: "" +
			"# header comment\n" +
			"a:\n" +
			"  # comment 1\n" +
			"  # comment 2\n" +
			"  \"&b\"\n",
	}, {
		desc:    "special spaces, single comment with header comment",
		key:     pstr("a"),
		val:     pstr("  b  "),
		headers: []string{"header comment"},
		inline:  []string{"comment"},
		expect: "" +
			"# header comment\n" +
			"a: \"  b  \" # comment\n",
	}, {
		desc:    "special spaces,  multiple comment with header comment",
		key:     pstr("a"),
		val:     pstr("  b  "),
		headers: []string{"header comment"},
		inline:  []string{"comment 1", "comment 2"},
		expect: "" +
			"# header comment\n" +
			"a:\n" +
			"  # comment 1\n" +
			"  # comment 2\n" +
			"  \"  b  \"\n",
	},
}

var keywordYMLTests = []ymlTest{
	{
		desc:   "keyword val, no comments",
		key:    pstr("a"),
		val:    pstr("null"),
		expect: "a: 'null'\n",
	},
}

var multiLineYMLTests = []ymlTest{
	{
		desc: "multi-line, no comments",
		key:  pstr("a"),
		val:  pstr("a\nb\nc"),
		expect: "" +
			"a: |-\n" +
			"  a\n" +
			"  b\n" +
			"  c\n",
	}, {
		desc:    "multi-line, single comments",
		key:     pstr("a"),
		val:     pstr("a\nb\nc"),
		headers: []string{"header comment"},
		inline:  []string{"comment"},
		expect: "" +
			"# header comment\n" +
			"a: # comment\n" +
			"  |-\n" +
			"  a\n" +
			"  b\n" +
			"  c\n",
	}, {
		desc:    "multi-line, multiple comments",
		key:     pstr("a"),
		val:     pstr("a\nb\nc"),
		headers: []string{"header comment"},
		inline:  []string{"comment 1", "comment 2"},
		expect: "" +
			"# header comment\n" +
			"a:\n" +
			"  # comment 1\n" +
			"  # comment 2\n" +
			"  |-\n" +
			"  a\n" +
			"  b\n" +
			"  c\n",
	}, {
		desc:    "multi-line, trailing newline, multiple comments",
		key:     pstr("a"),
		val:     pstr("a\nb\nc\n"),
		headers: []string{"header comment"},
		inline:  []string{"comment 1", "comment 2"},
		expect: "" +
			"# header comment\n" +
			"a:\n" +
			"  # comment 1\n" +
			"  # comment 2\n" +
			"  |+\n" +
			"  a\n" +
			"  b\n" +
			"  c\n",
	},
}

var maxLineYMLTests = []ymlTest{
	{
		desc:   "max line length, no comments",
		indent: 1,
		key:    pstr("a"),
		val: pstr("" +
			"This is a really long string that should " +
			"definitely exceed the limit of 40 characters. " +
			"Indeed, it has enough text to go beyond " +
			"that boundary."),
		expect: "" +
			"  a:\n" +
			"    \"This is a really long string that \\\n" +
			"    should definitely exceed the limit \\\n" +
			"    of 40 characters. Indeed, it has en\\\n" +
			"    ough text to go beyond that boundar\\\n" +
			"    y.\"\n",
		r: Renderer{
			MaxLineLength:         40,
			TrailingCommentColumn: 0,
			SpacesPerIndent:       2,
		},
	},
}

func TestRendererFormat(t *testing.T) {
	tests := basicYMLTests
	tests = append(tests, quotedYMLTests...)
	tests = append(tests, keywordYMLTests...)
	tests = append(tests, multiLineYMLTests...)
	tests = append(tests, maxLineYMLTests...)

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			var buf strings.Builder
			err := tc.r.Encode(&buf, &tc)
			assert.NoError(t, err, "failed to encode item")
			//got := tc.r.Format(tc.key, tc.val, tc.indent, tc.comments)
			expect := "---\n" + tc.expect
			assert.Equal(t, expect, buf.String(), "formatted output does not match expected")
		})
	}
}

func TestYAMLFormatRoundTrip(t *testing.T) {
	_, err := exec.LookPath("yq")
	if err != nil {
		t.Skip("yq command not found, skipping YAML format round-trip test")
	}

	tests := basicYMLTests
	tests = append(tests, quotedYMLTests...)
	tests = append(tests, keywordYMLTests...)
	tests = append(tests, multiLineYMLTests...)
	tests = append(tests, maxLineYMLTests...)

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			var buf strings.Builder
			err := tc.r.Encode(&buf, &tc)
			require.NoError(t, err)
			k := tc.Key()
			if k == nil {
				t.Skip("Skipping round-trip test for item without key")
				return
			}

			key := "." + *k

			cmd := exec.Command("yq", "-oj", key, "-")
			cmd.Stdin = strings.NewReader(buf.String())
			out, err := cmd.Output()
			require.NoError(t, err, "yq command failed")

			roundtrip := strings.TrimRight(string(out), "\n")
			roundtrip = strings.TrimLeft(roundtrip, "\\\"")
			roundtrip = strings.TrimRight(roundtrip, "\\\"")
			roundtrip = strings.ReplaceAll(roundtrip, "\\n", "\n")
			var want string
			if tc.Value() != nil {
				want = *tc.Value()
			}

			if !assert.Equal(t, want, roundtrip) {
				fmt.Println("YAML format round-trip test failed:", tc.desc)
				pp.Println("Original value:", want)
				pp.Println("Formatted value:", buf.String())
				pp.Println("Original value from yq:", string(out))
				pp.Println("final roundtrip value:", roundtrip)
			}
		})
	}
}
