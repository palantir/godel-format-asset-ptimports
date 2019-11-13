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

package formattester

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"testing"

	"github.com/nmiyake/pkg/dirs"
	"github.com/nmiyake/pkg/gofiles"
	"github.com/palantir/godel/v2/framework/pluginapitester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type AssetTestCase struct {
	Name        string
	Specs       []gofiles.GoFileSpec
	ConfigFiles map[string]string
	// Verify specifies whether or not formatter should be run in verify mode.
	Verify     bool
	Wd         string
	WantError  bool
	WantOutput func(projectDir string) string
	WantFiles  func(specFiles map[string]gofiles.GoFile) map[string]string
}

// RunAssetFormatTest tests the "format" operation using the provided asset. Resolves the format plugin using the
// provided locator and resolver, provides it with the asset and invokes the "format" command for the specified asset.
func RunAssetFormatTest(t *testing.T,
	pluginProvider pluginapitester.PluginProvider,
	assetProvider pluginapitester.AssetProvider,
	testBaseDir string,
	testCases []AssetTestCase,
) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	if testBaseDir != "" && !filepath.IsAbs(testBaseDir) {
		testBaseDir = path.Join(wd, testBaseDir)
	}

	tmpDir, cleanup, err := dirs.TempDir(testBaseDir, "")
	require.NoError(t, err)
	defer cleanup()

	tmpDir, err = filepath.EvalSymlinks(tmpDir)
	require.NoError(t, err)

	for i, tc := range testCases {
		projectDir, err := ioutil.TempDir(tmpDir, "")
		require.NoError(t, err)

		var sortedKeys []string
		for k := range tc.ConfigFiles {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)

		for _, k := range sortedKeys {
			err = os.MkdirAll(path.Dir(path.Join(projectDir, k)), 0755)
			require.NoError(t, err)
			err = ioutil.WriteFile(path.Join(projectDir, k), []byte(tc.ConfigFiles[k]), 0644)
			require.NoError(t, err)
		}

		specFiles, err := gofiles.Write(projectDir, tc.Specs)
		require.NoError(t, err)

		outputBuf := &bytes.Buffer{}
		func() {
			wd, err := os.Getwd()
			require.NoError(t, err)

			wantWd := projectDir
			if tc.Wd != "" {
				wantWd = path.Join(wantWd, tc.Wd)
			}
			err = os.Chdir(wantWd)
			require.NoError(t, err)
			defer func() {
				err = os.Chdir(wd)
				require.NoError(t, err)
			}()

			var args []string
			if tc.Verify {
				args = append(args, "--verify")
			}
			runPluginCleanup, err := pluginapitester.RunPlugin(
				pluginProvider,
				[]pluginapitester.AssetProvider{assetProvider},
				"format",
				args,
				projectDir,
				false,
				outputBuf,
			)
			defer runPluginCleanup()
			if tc.WantError {
				require.EqualError(t, err, "", "Case %d: %s\nOutput: %s", i, tc.Name, outputBuf.String())
			} else {
				require.NoError(t, err, "Case %d: %s\nOutput: %s", i, tc.Name, outputBuf.String())
			}
			var wantOutput string
			if tc.WantOutput != nil {
				wantOutput = tc.WantOutput(projectDir)
			}
			assert.Equal(t, wantOutput, outputBuf.String(), "Case %d: %s", i, tc.Name)

			wantFiles := tc.WantFiles(specFiles)
			var sortedKeys []string
			for k := range wantFiles {
				sortedKeys = append(sortedKeys, k)
			}
			sort.Strings(sortedKeys)
			for _, k := range sortedKeys {
				wantContent := wantFiles[k]
				bytes, err := ioutil.ReadFile(path.Join(projectDir, k))
				require.NoError(t, err, "Case %d: %s", i, tc.Name)
				got := string(bytes)
				assert.Equal(t, wantContent, got, "Case %d: %s\nGot:\n%s", i, tc.Name, got)
			}
		}()
	}
}
