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
	"os/exec"

	"github.com/kardianos/osext"
	"github.com/palantir/amalgomate/amalgomated"
	"github.com/pkg/errors"
)

const TypeName = "ptimports"

type Formatter struct {
	SkipRefactor bool
	SkipSimplify bool
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
