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
	"fmt"
	"testing"

	"github.com/nmiyake/pkg/gofiles"
	"github.com/palantir/godel-format-plugin/formattester"
	"github.com/palantir/godel/framework/pluginapitester"
	"github.com/palantir/godel/pkg/products"
	"github.com/stretchr/testify/require"
)

const (
	formatPluginLocator  = "com.palantir.godel-format-plugin:format-plugin:1.0.0-rc6"
	formatPluginResolver = "https://palantir.bintray.com/releases/{{GroupPath}}/{{Product}}/{{Version}}/{{Product}}-{{Version}}-{{OS}}-{{Arch}}.tgz"

	godelYML = `exclude:
  names:
    - "\\..+"
    - "vendor"
  paths:
    - "godel"
`
)

func TestFormat(t *testing.T) {
	pluginProvider, err := pluginapitester.NewPluginProviderFromLocator(formatPluginLocator, formatPluginResolver)
	require.NoError(t, err)

	assetPath, err := products.Bin("ptimports-asset")
	require.NoError(t, err)

	configFiles := map[string]string{
		"godel/config/godel.yml":         godelYML,
		"godel/config/format-plugin.yml": "",
	}
	formattester.RunAssetFormatTest(t,
		pluginProvider,
		pluginapitester.NewAssetProvider(assetPath),
		".",
		[]formattester.AssetTestCase{
			{
				Name: "formats file",
				Specs: []gofiles.GoFileSpec{
					{
						RelPath: "foo.go",
						Src: `package foo

import _ "fmt"
`,
					},
				},
				ConfigFiles: configFiles,
				WantFiles: func(specFiles map[string]gofiles.GoFile) map[string]string {
					return map[string]string{
						"foo.go": `package foo

import (
	_ "fmt"
)
`,
					}
				},
			},
			{
				Name: "groups imports",
				Specs: []gofiles.GoFileSpec{
					{
						RelPath: "foo.go",
						Src: `package foo

import _ "fmt"
import _ "github.com/org/repo"
import _ "{{index . "bar/bar.go"}}"
import _ "{{index . "baz/baz.go"}}"
`,
					},
					{
						RelPath: "bar/bar.go",
						Src: `package bar
`,
					},
					{
						RelPath: "baz/baz.go",
						Src: `package baz
`,
					},
				},
				ConfigFiles: configFiles,
				WantFiles: func(specFiles map[string]gofiles.GoFile) map[string]string {
					return map[string]string{
						"foo.go": fmt.Sprintf(`package foo

import (
	_ "fmt"

	_ "github.com/org/repo"
	_ "%s"
	_ "%s"
)
`, specFiles["bar/bar.go"].ImportPath, specFiles["baz/baz.go"].ImportPath),
					}
				},
			},
			{
				Name: "removes unused imports",
				Specs: []gofiles.GoFileSpec{
					{
						RelPath: "foo.go",
						Src: `package foo

import "fmt"
`,
					},
				},
				ConfigFiles: configFiles,
				WantFiles: func(specFiles map[string]gofiles.GoFile) map[string]string {
					return map[string]string{
						"foo.go": `package foo
`,
					}
				},
			},
			{
				Name: "adds required imports",
				Specs: []gofiles.GoFileSpec{
					{
						RelPath: "foo.go",
						Src: `package foo

func Foo() {
	fmt.Println("foo")
}
`,
					},
				},
				ConfigFiles: configFiles,
				WantFiles: func(specFiles map[string]gofiles.GoFile) map[string]string {
					return map[string]string{
						"foo.go": `package foo

import (
	"fmt"
)

func Foo() {
	fmt.Println("foo")
}
`,
					}
				},
			},
			{
				Name: "simplifies code",
				Specs: []gofiles.GoFileSpec{
					{
						RelPath: "foo.go",
						Src: `package foo

func Foo() {
	for i, _ := range []string{} {
		_ = i
	}
}
`,
					},
				},
				ConfigFiles: configFiles,
				WantFiles: func(specFiles map[string]gofiles.GoFile) map[string]string {
					return map[string]string{
						"foo.go": `package foo

func Foo() {
	for i := range []string{} {
		_ = i
	}
}
`,
					}
				},
			},
			{
				Name: "does not simplify code based on configuration",
				Specs: []gofiles.GoFileSpec{
					{
						RelPath: "foo.go",
						Src: `package foo

func Foo() {
	for i, _ := range []string{} {
		_ = i
	}
}
`,
					},
				},
				ConfigFiles: map[string]string{
					"godel/config/godel.yml": godelYML,
					"godel/config/format-plugin.yml": `
formatters:
  ptimports:
    config:
      skip-simplify: true
`,
				},
				WantFiles: func(specFiles map[string]gofiles.GoFile) map[string]string {
					return map[string]string{
						"foo.go": `package foo

func Foo() {
	for i, _ := range []string{} {
		_ = i
	}
}
`,
					}
				},
			},
			{
				Name: "does not refactor imports to block style based on configuration",
				Specs: []gofiles.GoFileSpec{
					{
						RelPath: "foo.go",
						Src: `package foo

import _ "fmt"
`,
					},
				},
				ConfigFiles: map[string]string{
					"godel/config/godel.yml": godelYML,
					"godel/config/format-plugin.yml": `
formatters:
  ptimports:
    config:
      skip-refactor: true
`,
				},
				WantFiles: func(specFiles map[string]gofiles.GoFile) map[string]string {
					return map[string]string{
						"foo.go": `package foo

import _ "fmt"
`,
					}
				},
			},
			{
				Name: "verify does not modify files and prints unformatted files",
				Specs: []gofiles.GoFileSpec{
					{
						RelPath: "foo.go",
						Src: `package foo

import _ "fmt"
`,
					},
				},
				ConfigFiles: configFiles,
				Verify:      true,
				WantError:   true,
				WantOutput: func(projectDir string) string {
					return fmt.Sprintf(`%s/foo.go
`, projectDir)
				},
				WantFiles: func(specFiles map[string]gofiles.GoFile) map[string]string {
					return map[string]string{
						"foo.go": `package foo

import _ "fmt"
`,
					}
				},
			},
		},
	)
}

func TestUpgradeConfig(t *testing.T) {
	pluginProvider, err := pluginapitester.NewPluginProviderFromLocator(formatPluginLocator, formatPluginResolver)
	require.NoError(t, err)

	assetPath, err := products.Bin("ptimports-asset")
	require.NoError(t, err)
	assetProvider := pluginapitester.NewAssetProvider(assetPath)

	pluginapitester.RunUpgradeConfigTest(t,
		pluginProvider,
		[]pluginapitester.AssetProvider{assetProvider},
		[]pluginapitester.UpgradeConfigTestCase{
			{
				Name: "current configuration is not upgraded",
				ConfigFiles: map[string]string{
					"godel/config/godel.yml": godelYML,
					"godel/config/format-plugin.yml": `
# comment
formatters:
  ptimports:
    config:
      # an inner comment
      skip-refactor: true
      # inner comment
      skip-simplify: true
`,
				},
				WantOutput: "",
				WantFiles: map[string]string{
					"godel/config/format-plugin.yml": `
# comment
formatters:
  ptimports:
    config:
      # an inner comment
      skip-refactor: true
      # inner comment
      skip-simplify: true
`,
				},
			},
		},
	)
}
