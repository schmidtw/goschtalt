<!--
SPDX-FileCopyrightText: 2024 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
SPDX-License-Identifier: Apache-2.0
-->

# Why this directory?

By having this directory as a sub-directory that only imports goschtalt and is
a subdirectory, it's possible to run a couple of end to end tests that need to
access files that are not in the current working directory safely.  It also
keeps circular import loops from being a problem.
