// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package meta

import (
	"regexp"
	"strings"
)

var (
	outerRe    = regexp.MustCompile(`^(.*\b)\s*\(\((.*)\)\)\s*$`)
	limitChars = regexp.MustCompile(`^[a-zA-Z0-9 _-]*$`)
)

type command struct {
	full   string
	cmd    string
	secret bool
	final  string
}

// getCmd processes the input string and extracts the commands that my be
// present.
func getCmd(s string) (command, error) {
	cmd := command{
		full:  s,
		final: s,
	}

	// Split out the 'foobar(( commands, secret ))'
	sub := outerRe.FindStringSubmatch(s)

	if len(sub) == 0 {
		return cmd, nil
	}

	cmd.final = strings.TrimSpace(sub[1])

	// Get rid of ',' and take the interior of the (()) and split into fields.
	inner := strings.ReplaceAll(sub[2], ",", " ")
	list := strings.Fields(inner)

	// Make sure commands are only a limited character set.
	for _, val := range list {
		if !limitChars.MatchString(val) {
			return command{}, ErrInvalidCommand
		}
	}

	var cmds []string
	for _, val := range list {
		if cmdSecret == val {
			// 'secret' can only show up once.
			if cmd.secret {
				return command{}, ErrInvalidCommand
			}
			cmd.secret = true
			continue
		}

		cmds = append(cmds, val)
	}
	if len(cmds) > 1 {
		return command{}, ErrInvalidCommand
	}
	if len(cmds) == 1 {
		cmd.cmd = cmds[0]
	}

	return cmd, nil
}
