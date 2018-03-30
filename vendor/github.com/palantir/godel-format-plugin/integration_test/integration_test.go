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

package integration_test

import (
	"testing"

	"github.com/palantir/godel/framework/pluginapitester"
	"github.com/palantir/godel/pkg/products"
	"github.com/stretchr/testify/require"
)

const (
	ptimportsAssetLocator  = "com.palantir.godel-format-asset-ptimports:ptimports-asset:1.0.0-rc5"
	ptimportsAssetResolver = "https://palantir.bintray.com/releases/{{GroupPath}}/{{Product}}/{{Version}}/{{Product}}-{{Version}}-{{OS}}-{{Arch}}.tgz"
)

func TestUpgradeConfig(t *testing.T) {
	pluginPath, err := products.Bin("format-plugin")
	require.NoError(t, err)
	pluginProvider := pluginapitester.NewPluginProvider(pluginPath)

	assetProvider, err := pluginapitester.NewAssetProviderFromLocator(ptimportsAssetLocator, ptimportsAssetResolver)
	require.NoError(t, err)

	pluginapitester.RunUpgradeConfigTest(t,
		pluginProvider,
		[]pluginapitester.AssetProvider{
			assetProvider,
		},
		[]pluginapitester.UpgradeConfigTestCase{
			{
				Name: "default legacy format config is upgraded to blank",
				ConfigFiles: map[string]string{
					"godel/config/format.yml": `formatters:
  gofmt:
    args:
      - "-s"
`,
				},
				Legacy:     true,
				WantOutput: "Upgraded configuration for format-plugin.yml\n",
				WantFiles: map[string]string{
					"godel/config/format-plugin.yml": `formatters:
  ptimports:
    config:
      separate-project-imports: true
exclude:
  names: []
  paths: []
`,
				},
			},
			{
				Name: "legacy format config excludes are upgraded",
				ConfigFiles: map[string]string{
					"godel/config/format-plugin.yml": `
legacy-config: true
formatters:
  gofmt:
    args:
      - "-s"
exclude:
  names:
    - "foo.go"
  paths:
    - "integration_test"
`,
				},
				WantOutput: "Upgraded configuration for format-plugin.yml\n",
				WantFiles: map[string]string{
					"godel/config/format-plugin.yml": `formatters:
  ptimports:
    config:
      separate-project-imports: true
exclude:
  names:
  - foo.go
  paths:
  - integration_test
`,
				},
			},
			{
				Name: "current config is unmodified",
				ConfigFiles: map[string]string{
					"godel/config/format-plugin.yml": `
# comment
exclude:
  names:
    - "foo.go"
  paths:
    - "integration_test"
`,
				},
				WantOutput: "",
				WantFiles: map[string]string{
					"godel/config/format-plugin.yml": `
# comment
exclude:
  names:
    - "foo.go"
  paths:
    - "integration_test"
`,
				},
			},
		},
	)
}
