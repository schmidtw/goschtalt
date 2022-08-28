#!/bin/bash
# SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
# SPDX-License-Identifier: Apache-2.0

list=(`find . -name go.mod`)

rm -r licensed
mkdir -p licensed

for i in "${list[@]}"
do
    d=`dirname $i`
    fn=$(echo "$d" | sed -e 's/[\/\.]/_/g' | sed -e 's/^__//')
    if [ "$d" != "." ]; then
        cp .licensed.yml $d/.licensed.yml
    else
        fn='goschtalt'
    fi
    pushd $d
    echo "Examining: $d"
    go get ./...
    licensed cache
    licensed status -f yaml > licensing.yml || true
    yq eval 'del(.apps[].sources[].dependencies[] | select(.allowed == "true") )' licensing.yml > disallowed.yml
    licensed status
    popd
    mv $d/licensing.yml  "licensed/${fn}_licensing.yml"
    mv $d/disallowed.yml "licensed/${fn}_disallowed.yml"
done
