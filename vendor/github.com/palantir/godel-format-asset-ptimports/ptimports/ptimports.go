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
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/kardianos/osext"
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
	self, err := osext.Executable()
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
		projectDirLocalPath, err := projectLocalPath(projectDir)
		if err != nil {
			return err
		}
		args = append(args, "--local", projectDirLocalPath)
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

func projectLocalPath(projectDir string) (string, error) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return "", errors.Errorf("GOPATH environment variable not set")
	}
	canonicalGoPath, err := filepath.EvalSymlinks(gopath)
	if err != nil {
		return "", errors.Wrapf(err, "failed to evaluate symlinks for GOPATH %q", gopath)
	}
	canonicalProjectDirPath, err := filepath.EvalSymlinks(projectDir)
	if err != nil {
		return "", errors.Wrapf(err, "failed to evaulate symlinks for project dir %q", projectDir)
	}
	gopathSrc := path.Join(canonicalGoPath, "src") + "/"
	if !strings.HasPrefix(canonicalProjectDirPath, gopathSrc) {
		return "", errors.Errorf("project dir %q is not within $GOPATH/src %q", canonicalProjectDirPath, gopathSrc)
	}
	return strings.TrimPrefix(canonicalProjectDirPath, gopathSrc) + "/", nil
}
