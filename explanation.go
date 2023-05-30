// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"fmt"
	"strings"
	"time"

	"github.com/goschtalt/goschtalt/pkg/debug"
)

// Explanation is the structure that represents what happened last time the
// Config object was altered, compiled or used.
//
// Calling New() or With() will clear out all the values present.
type Explanation struct {
	// -- Configuration based information --------------------------------------

	// Options is the ordered list of Options in effect.
	Options []string

	// FileExtensions is the list of file extensions supported in the
	// configuration provided.
	FileExtensions []string

	// -- Compilation based information ----------------------------------------

	// CompileStartedAt is the start time when the most recent compile was
	// performed.
	CompileStartedAt time.Time

	// CompileFinishedAt is the completion time when the most recent compile was
	// performed.
	CompileFinishedAt time.Time

	// Records is the ordered list of records processed.
	Records []ExplanationRecord

	// VariableExpansions is the ordered list of variable expansion instructions
	// applied.
	VariableExpansions []string

	// CompileErrors is the ordered list of compilation errors encountered during
	// the last compilation.
	CompileErrors []error

	// -- Marshaling/Unmarshaling based information ----------------------------

	// Keyremapping collects the key remapping that happens when values are
	// added or used.  This helps address problems when a name may not be
	// converted the way you would expect or like.
	//
	// Not available if DisableDefaultPackageOptions() is specified.
	Keyremapping debug.Collect
}

// ExplanationRecord is the structure that represents a specific record that was
// processed.
type ExplanationRecord struct {
	Name     string        // The name of the record.
	Default  bool          // If the record was marked as a 'default' record.
	Duration time.Duration // The time needed to process the record.
}

func (er ExplanationRecord) String() string {
	empty := ExplanationRecord{}
	if er == empty {
		return ""
	}

	user := "user"
	if er.Default {
		user = "default"
	}
	return fmt.Sprintf("'%s' <%s> (%s)", er.Name, user, er.Duration)
}

func (e *Explanation) reset() {
	e.Options = []string{}
	e.FileExtensions = []string{}
	e.compileReset()
	e.Keyremapping.Reset()
}

func (e *Explanation) optionInEffect(s string) {
	e.Options = append(e.Options, s)
}

func (e *Explanation) extsSupported(s []string) {
	e.FileExtensions = append(e.FileExtensions, s...)
}

func (e *Explanation) compileReset() {
	e.CompileStartedAt = time.Time{}
	e.Records = []ExplanationRecord{}
	e.VariableExpansions = []string{}
	e.CompileErrors = []error{}
}

func (e *Explanation) compileStartedAt(t time.Time) {
	e.CompileStartedAt = t
	e.Records = []ExplanationRecord{}
	e.VariableExpansions = []string{}
	e.CompileErrors = []error{}
}

func (e *Explanation) compileRecord(name string, isDefault bool, now time.Time) {
	elapsed := now.Sub(e.CompileStartedAt)
	for _, record := range e.Records {
		elapsed -= record.Duration
	}

	e.Records = append(e.Records,
		ExplanationRecord{
			Name:     name,
			Default:  isDefault,
			Duration: elapsed,
		})
}

func (e *Explanation) compileExpansions(details string) {
	e.VariableExpansions = append(e.VariableExpansions, details)
}

func (e *Explanation) recordError(err error) {
	if err != nil {
		e.CompileErrors = append(e.CompileErrors, err)
	}
}

func (e Explanation) String() string {
	var b strings.Builder

	fmt.Fprintln(&b, "# Options")
	fmt.Fprintln(&b, "")
	fmt.Fprintln(&b, "## Options in effect")
	fmt.Fprintln(&b, "")
	for i, opt := range e.Options {
		fmt.Fprintf(&b, "  %d. %s\n", i+1, opt)
	}
	fmt.Fprintln(&b, "")
	fmt.Fprintln(&b, "## File extensions supported:")
	fmt.Fprintln(&b, "")
	if len(e.FileExtensions) == 0 {
		fmt.Fprintln(&b, "  <none>")
	} else {
		for _, ext := range e.FileExtensions {
			fmt.Fprintf(&b, "  - %s\n", ext)
		}
	}

	fmt.Fprintln(&b, "")
	fmt.Fprintln(&b, "# Compilation")
	fmt.Fprintln(&b, "")
	fmt.Fprintf(&b, "  - Started at:  %s\n", e.CompileStartedAt.Format(time.RFC3339))
	fmt.Fprintf(&b, "  - Duration: %s\n", e.CompileFinishedAt.Sub(e.CompileStartedAt))
	if len(e.CompileErrors) == 0 {
		fmt.Fprintln(&b, "  - Errors: none")
	} else {
		fmt.Fprintln(&b, "  - Errors:")
		for _, err := range e.CompileErrors {
			fmt.Fprintf(&b, "    - %s\n", err)
		}
	}
	fmt.Fprintln(&b, "")
	fmt.Fprintln(&b, "## Records processed in order.")
	fmt.Fprintln(&b, "")
	if len(e.Records) == 0 {
		fmt.Fprintln(&b, "  <none>")
	} else {
		max := 0
		for _, record := range e.Records {
			if len(record.Name) > max {
				max = len(record.Name)
			}
		}

		for i, record := range e.Records {
			fmt.Fprintf(&b, "  %d. %s\n", i+1, record.String())
		}
	}
	fmt.Fprintln(&b, "")
	fmt.Fprintln(&b, "## Variable expansions processed in order.")
	fmt.Fprintln(&b, "")
	if len(e.VariableExpansions) == 0 {
		fmt.Fprintln(&b, "  <none>")
	} else {
		for i, expansion := range e.VariableExpansions {
			fmt.Fprintf(&b, "  %d. %s\n", i+1, expansion)
		}
	}

	fmt.Fprintln(&b, "")
	fmt.Fprintln(&b, "# Structure name remapping")
	fmt.Fprintln(&b, "")
	rm := e.Keyremapping.String()
	if rm != "" {
		rm = "  - " + rm
		rm = strings.TrimSuffix(rm, "\n")
		rm = strings.ReplaceAll(rm, "\n", "\n  - ")
	}
	fmt.Fprintln(&b, rm)
	fmt.Fprintln(&b, "")

	return b.String()
}
