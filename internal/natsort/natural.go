// SPDX-FileCopyrightText: 2015 Vincent Batoufflet and Marc Falzon
// SPDX-FileCopyrightText: 2022 Mark Karpelès
// SPDX-License-Identifier: BSD-3-Clause
//
// This file originated from https://github.com/facette/natsort/pull/2/files

package natsort

import "strings"

func Compare(a, b string) bool {
	ln_a := len(a)
	ln_b := len(b)
	posa := 0
	posb := 0

	for {
		if ln_a <= posa {
			if ln_b <= posb {
				// eof on both at the same time (equal)
				return false
			}
			return true
		} else if ln_b <= posb {
			// eof on b
			return false
		}

		av, bv := a[posa], b[posb]

		if av >= '0' && av <= '9' && bv >= '0' && bv <= '9' {
			// go into numeric mode
			intlna := 1
			intlnb := 1
			for {
				if posa+intlna >= ln_a {
					break
				}
				x := a[posa+intlna]
				if av == '0' {
					posa += 1
					av = x
					continue
				}
				if x >= '0' && x <= '9' {
					intlna += 1
				} else {
					break
				}
			}
			for {
				if posb+intlnb >= ln_b {
					break
				}
				x := b[posb+intlnb]
				if bv == '0' {
					posb += 1
					bv = x
					continue
				}
				if x >= '0' && x <= '9' {
					intlnb += 1
				} else {
					break
				}
			}
			if intlnb > intlna {
				// length of a value is longer, means it's a bigger number
				return true
			} else if intlna > intlnb {
				return false
			}
			// both have same length, let's compare as string
			v := strings.Compare(a[posa:posa+intlna], b[posb:posb+intlnb])
			if v < 0 {
				return true
			} else if v > 0 {
				return false
			}
			// equale
			posa += intlna
			posb += intlnb
			continue
		}

		if av == bv {
			posa += 1
			posb += 1
			continue
		}

		return av < bv
	}
	return false
}
