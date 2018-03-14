// Copyright 2016 Palantir Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestReadConfig(t *testing.T) {
	in := `
formatters:
  ptimports:
    config:
      no-simplify: true
`
	got, err := readFormatConfig([]byte(in))
	require.NoError(t, err)

	assert.Equal(t, formatConfig{
		Formatters: map[string]formatterConfig{
			"ptimports": {
				Config: &yaml.MapSlice{
					yaml.MapItem{
						Key:   "no-simplify",
						Value: true,
					},
				},
			},
		},
	}, got)
}
