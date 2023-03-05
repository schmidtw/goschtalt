// SPDX-FileCopyrightText: 2022-2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import "errors"

var (
	ErrAdaptFailure  = errors.New("at least one matching adapt function failed")
	ErrDecoding      = errors.New("decoding error")
	ErrEncoding      = errors.New("encoding error")
	ErrNotApplicable = errors.New("not applicable")
	ErrNotCompiled   = errors.New("the Compile() function must be called first")
	ErrCodecNotFound = errors.New("encoder/decoder not found")
	ErrInvalidInput  = errors.New("input is invalid")
	ErrFileMissing   = errors.New("required file is missing")
)
