#!/bin/bash
# SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
# SPDX-License-Identifier: Apache-2.0

list=(`find . -name go.mod`)

for i in "${list[@]}"
do
    d=`dirname $i`
    pushd $d
    echo "Testing: $d"
    go test ./...
    popd
done
