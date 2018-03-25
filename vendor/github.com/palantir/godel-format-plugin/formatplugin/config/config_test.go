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

package config_test

import (
	"testing"

	"github.com/palantir/pkg/matcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/palantir/godel-format-plugin/formatplugin/config"
)

func TestReadConfig(t *testing.T) {
	in := `
formatters:
  ptimports:
    config:
      no-simplify: true
exclude:
  names:
    - "*.pb.go"
`

	var got config.Format
	err := yaml.Unmarshal([]byte(in), &got)
	require.NoError(t, err)

	want := config.Format{
		Formatters: config.ToFormatters(
			map[string]config.Formatter{
				"ptimports": {
					Config: yaml.MapSlice{
						yaml.MapItem{
							Key:   "no-simplify",
							Value: true,
						},
					},
				},
			},
		),
		Exclude: matcher.NamesPathsCfg{
			Names: []string{
				"*.pb.go",
			},
		},
	}

	assert.Equal(t, want, got)
}
