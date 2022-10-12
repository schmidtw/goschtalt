#!/bin/bash
# SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2014 kepkin
# SPDX-License-Identifier: BSD-2-Clause

function highlight() {
	declare -A fg_color_map
	fg_color_map[black]=30
	fg_color_map[red]=31
	fg_color_map[green]=32
	fg_color_map[yellow]=33
	fg_color_map[blue]=34
	fg_color_map[magenta]=35
	fg_color_map[cyan]=36
	 
	fg_c=$(echo -e "\e[1;${fg_color_map[$1]}m")
	c_rs=$'\e[0m'
	sed -u s"/$2/$fg_c\0$c_rs/g"
}

list=(`find . -name go.mod`)

for i in "${list[@]}"
do
    d=`dirname $i`
    pushd $d
    echo "Testing: $d" | highlight cyan '.*'
    echo "--------------------------------------------------------------------------------" | highlight cyan '.*'
    go test ./... 2>&1 | highlight red 'FAIL' | highlight yellow '^\.[^:]*'
    popd
done
