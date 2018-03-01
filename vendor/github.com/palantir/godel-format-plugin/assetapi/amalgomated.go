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

package assetapi

import (
	"bytes"
	"io"
	"os"
	"os/exec"

	"github.com/kardianos/osext"
	"github.com/palantir/pkg/cobracli"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type Amalgomated interface {
	Run(cmd string)
	Cmds() []string
}

func AmalgomatedMain(assetName string, rootCmd *cobra.Command, amalgomated Amalgomated) {
	if len(os.Args) > 1 && os.Args[1] == "--"+assetName {
		os.Args = append([]string{os.Args[0]}, os.Args[2:]...)
		amalgomated.Run(assetName)
	}
	os.Exit(cobracli.ExecuteWithDefaultParamsWithVersion(rootCmd, nil, ""))
}

func RunAmalgomatedFormatCommand(assetName string, files []string, verify bool, extraArgs []string, stdout, stderr io.Writer) error {
	var cmdArgs []string
	if verify {
		cmdArgs = append(cmdArgs, "-l")
	} else {
		cmdArgs = append(cmdArgs, "-w")
	}
	cmdArgs = append(cmdArgs, extraArgs...)
	cmdArgs = append(cmdArgs, files...)

	combinedBuf := &bytes.Buffer{}
	stdoutMultiWriter := io.MultiWriter(stdout, combinedBuf)
	stderrMultiWriter := io.MultiWriter(stderr, combinedBuf)

	pathToSelf, err := osext.Executable()
	if err != nil {
		return errors.Wrapf(err, "failed to determine path for current executable")
	}

	cmd := exec.Command(pathToSelf, append([]string{"--" + assetName}, cmdArgs...)...)
	cmd.Stdout = stdoutMultiWriter
	cmd.Stderr = stderrMultiWriter
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
