// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import "fmt"

func stdCfgLayout(appName string, files []string) Option {
	return WithError(fmt.Errorf("%v: StdCfgLayout() on windows", ErrUnsupported))
}
