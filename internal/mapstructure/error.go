// SPDX-FileCopyrightText: 2013-2025 Mitchell Hashimoto
// SPDX-FileCopyrightText: 2025 Weston Schmidt
// SPDX-License-Identifier: MIT
//
// This file originated from https://github.com/mitchellh/mapstructure

package mapstructure

import (
	"errors"
)

var (
	ErrDecoding = errors.New("error decoding")
)
