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

package formatter

import (
	"encoding/json"
	"io"
	"os/exec"
	"strings"

	"github.com/palantir/godel/framework/pluginapi"
	"github.com/pkg/errors"
)

type assetFormatter struct {
	assetPath string
	cfgYML    string
}

func (f *assetFormatter) TypeName() (string, error) {
	nameCmd := exec.Command(f.assetPath, nameCmdName)
	outputBytes, err := runCommand(nameCmd)
	if err != nil {
		return "", err
	}
	var typeName string
	if err := json.Unmarshal(outputBytes, &typeName); err != nil {
		return "", errors.Wrapf(err, "failed to unmarshal JSON")
	}
	return typeName, nil
}

func (f *assetFormatter) VerifyConfig() error {
	verifyConfigCmd := exec.Command(f.assetPath, verifyConfigCmdName,
		"--"+commonCmdConfigYMLFlagName, f.cfgYML,
	)
	if _, err := runCommand(verifyConfigCmd); err != nil {
		return err
	}
	return nil
}

func (f *assetFormatter) Format(files []string, list bool, projectDir string, stdout io.Writer) error {
	args := []string{
		runFormatCmdName,
		"--" + commonCmdConfigYMLFlagName, f.cfgYML,
	}
	if list {
		args = append(args, "--"+runFormatCmdListFlagName)
	}
	if projectDir != "" {
		args = append(args, "--"+pluginapi.ProjectDirFlagName, projectDir)
	}
	args = append(args, files...)

	runFormatCmd := exec.Command(f.assetPath, args...)
	runFormatCmd.Stdout = stdout
	runFormatCmd.Stderr = stdout
	if err := runFormatCmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return errors.Wrapf(err, "failed to execute command %v", runFormatCmd.Args)
		}
	}
	return nil
}

func runCommand(cmd *exec.Cmd) ([]byte, error) {
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return outputBytes, errors.New(strings.TrimSpace(strings.TrimPrefix(string(outputBytes), "Error: ")))
	}
	return outputBytes, nil
}
