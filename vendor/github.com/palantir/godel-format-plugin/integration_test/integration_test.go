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
	"io/ioutil"
	"os/exec"
	"path"
	"strings"
	"testing"

	"github.com/nmiyake/pkg/dirs"
	"github.com/palantir/godel/pkg/products"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatTask(t *testing.T) {
	for i, tc := range []struct {
		name        string
		args        func(assetPath string) []string
		assetScript string
		wantError   bool
		want        func(assetPath string) string
	}{
		{
			"regular run",
			func(assetPath string) []string {
				return []string{
					"--project-dir", ".",
					"--assets", assetPath,
					"input-file-1", "input-file-2",
				}
			},
			`#!/bin/sh
echo $@
`,
			false,
			func(assetPath string) string {
				return `input-file-1 input-file-2
`
			},
		},
		{
			"verify run",
			func(assetPath string) []string {
				return []string{
					"--project-dir", ".",
					"--assets", assetPath,
					"--verify",
					"input-file-1", "input-file-2",
				}
			},
			`#!/bin/sh
echo input-file-1
`,
			true,
			func(assetPath string) string {
				return `input-file-1
`
			},
		},
		{
			"verify run empty returns success and no output",
			func(assetPath string) []string {
				return []string{
					"--project-dir", ".",
					"--assets", assetPath,
					"--verify",
					"input-file-1", "input-file-2",
				}
			},
			`#!/bin/sh
`,
			false,
			func(assetPath string) string {
				return ``
			},
		},
	} {
		func() {
			tmpDir, cleanup, err := dirs.TempDir(".", "")
			require.NoError(t, err, "Case %d: %s", i, tc.name)
			defer cleanup()

			echoScriptPath := path.Join(tmpDir, "echo.sh")
			err = ioutil.WriteFile(echoScriptPath, []byte(tc.assetScript), 0755)
			require.NoError(t, err, "Case %d: %s", i, tc.name)

			cli, err := products.Bin("format-plugin")
			require.NoError(t, err, "Case %d: %s", i, tc.name)

			cmd := exec.Command(cli, tc.args(echoScriptPath)...)
			output, err := cmd.CombinedOutput()

			if tc.wantError {
				require.Error(t, err, fmt.Sprintf("Case %d: %s", i, tc.name))
			} else {
				require.NoError(t, err, "Case %d: %s", i, tc.name)
			}
			assert.Equal(t, tc.want(echoScriptPath), string(output), "Case %d: %s", i, tc.name)
		}()
	}
}

func TestFormatTaskWithConfig(t *testing.T) {
	tmpDir, cleanup, err := dirs.TempDir(".", "")
	require.NoError(t, err)
	defer cleanup()

	const echoBash = `#!/bin/sh
if [ "$1" = "--name" ]; then
    echo 'echo'
    exit 0
fi
echo $@
`
	echoScriptPath := path.Join(tmpDir, "echo.sh")
	err = ioutil.WriteFile(echoScriptPath, []byte(echoBash), 0755)
	require.NoError(t, err)

	const cfgFileContent = `
formatters:
  echo:
    args:
      - "-cfgArg"
`
	cfgFilePath := path.Join(tmpDir, "format.yml")
	err = ioutil.WriteFile(cfgFilePath, []byte(cfgFileContent), 0644)
	require.NoError(t, err)

	cli, err := products.Bin("format-plugin")
	require.NoError(t, err)

	cmd := exec.Command(cli,
		"--project-dir", ".",
		"--config", cfgFilePath,
		"--assets", echoScriptPath,
		"input-file-1", "input-file-2")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err)

	assert.Equal(t, `-cfgArg input-file-1 input-file-2
`, string(output))
}

func TestFormatVerifyUniquifiesOutput(t *testing.T) {
	tmpDir, cleanup, err := dirs.TempDir(".", "")
	require.NoError(t, err)
	defer cleanup()

	const oneBash = `#!/bin/sh
echo 'bar.go'
`
	oneScriptPath := path.Join(tmpDir, "one.sh")
	err = ioutil.WriteFile(oneScriptPath, []byte(oneBash), 0755)
	require.NoError(t, err)

	const twoBash = `#!/bin/sh
echo 'foo.go'
echo 'baz.go'
echo 'bar.go'
`
	twoScriptPath := path.Join(tmpDir, "two.sh")
	err = ioutil.WriteFile(twoScriptPath, []byte(twoBash), 0755)
	require.NoError(t, err)

	cli, err := products.Bin("format-plugin")
	require.NoError(t, err)

	cmd := exec.Command(cli,
		"--project-dir", ".",
		"--verify",
		"--assets", strings.Join([]string{oneScriptPath, twoScriptPath}, ","),
		"foo.go", "bar.go", "baz.go")
	output, err := cmd.CombinedOutput()
	require.Error(t, err)

	assert.Equal(t, `foo.go
bar.go
baz.go
`, string(output))
}
