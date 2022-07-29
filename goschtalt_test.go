/**
 *  Copyright (c) 2022  Weston Schmidt
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
package goschtalt

import (
	iofs "io/fs"
	"testing"

	pp "github.com/k0kubun/pp/v3"
	"github.com/psanford/memfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	config1 = `{
	"maxHeaderSize": 1024,
	"header": "X-Wing: Gold Leader",
	"servers": {
		"echo": {
			"ports": [ 80, 443 ],
			"other": {
				"street": "1 North Pole",
				"country": "Sealand"
			}
		}
	}
}`

	config2 = `{
	"servers": {
		"echo": {
			"ports": [ 8080 ],
			"max-thing": 12
		},
		"tango": {
			"ports": [ 999 ]
		}
	}
}`
)

func makeGoschtaltTestFs(t *testing.T) iofs.FS {
	require := require.New(t)
	fs := memfs.New()
	require.NoError(fs.MkdirAll("conf", 0777))
	require.NoError(fs.WriteFile("conf/80_config.json", []byte(config1), 0755))
	require.NoError(fs.WriteFile("conf/90_config.json", []byte(config2), 0755))
	return fs
}

func TestReadAll(t *testing.T) {
	tests := []struct {
		description string
		group       Group
		expectedErr error
	}{
		{
			description: "Merge them all",
			group: Group{
				Paths: []string{"conf"},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			tc.group.FS = makeGoschtaltTestFs(t)
			g, err := New(WithFileGroup(tc.group))
			require.NotNil(g)
			require.NoError(err)

			got, err := g.readAll()
			if tc.expectedErr == nil {
				assert.NoError(err)
				pp.Print(got)
				//assert.True(reflect.DeepEqual(tc.expected, got))
				return
			}
			assert.ErrorIs(err, tc.expectedErr)
		})
	}
}
