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

package ptimports

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"

	"github.com/palantir/amalgomate/amalgomated"
	"github.com/pkg/errors"
)

const TypeName = "ptimports"

type Formatter struct {
	SkipRefactor           bool
	SkipSimplify           bool
	SeparateProjectImports bool
}

func (f *Formatter) TypeName() (string, error) {
	return TypeName, nil
}

func (f *Formatter) Format(files []string, list bool, projectDir string, stdout io.Writer) error {
	self, err := os.Executable()
	if err != nil {
		return errors.Wrapf(err, "failed to determine executable")
	}
	args := []string{
		amalgomated.ProxyCmdPrefix + TypeName,
	}
	if list {
		args = append(args, "-l")
	} else {
		args = append(args, "-w")
	}
	if !f.SkipSimplify {
		args = append(args, "-s")
	}
	if !f.SkipRefactor {
		args = append(args, "-r")
	}
	if f.SeparateProjectImports {
		projectPkgPath, err := moduleImportPath(projectDir)
		if err != nil {
			return err
		}
		args = append(args, "--local", projectPkgPath+"/")
	}
	args = append(args, files...)

	cmd := exec.Command(self, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stdout
	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return errors.Wrapf(err, "failed to run %v", cmd.Args)
		}
	}
	return nil
}

func moduleImportPath(projectDir string) (string, error) {
	cmd := exec.Command("go", "list", "-m", "-mod=readonly", "-json")
	cmd.Dir = projectDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrapf(err, "%v failed with output: %q", cmd.Args, string(output))
	}
	modJSON := struct {
		Path string
		Dir  string
	}{}
	if err := json.Unmarshal(output, &modJSON); err != nil {
		return "", errors.Wrapf(err, "failed to unmarshal output of %v as JSON", cmd.Args)
	}
	return modJSON.Path, nil
}
