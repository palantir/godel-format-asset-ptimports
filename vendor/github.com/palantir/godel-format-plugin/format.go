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

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/palantir/pkg/matcher"
	"github.com/pkg/errors"

	"github.com/palantir/godel-format-plugin/assetapi"
)

func runFormat(assets []string, assetArgs map[string]formatterConfig, projectDir string, exclude matcher.Matcher, verify bool, providedFiles []string, stdout, stderr io.Writer) error {
	var files []string
	if len(providedFiles) == 0 {
		matchingFiles, err := allMatchingFilesInDir(projectDir, exclude)
		if err != nil {
			return err
		}
		for _, f := range matchingFiles {
			files = append(files, path.Join(projectDir, f))
		}
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return errors.Wrapf(err, "failed to determine working directory")
		}
		filteredFiles, err := filterFiles(projectDir, wd, providedFiles, exclude)
		if err != nil {
			return err
		}
		files = filteredFiles
	}

	// if there are no files to check, exit
	if len(files) == 0 {
		return nil
	}

	var outputBuf bytes.Buffer
	for _, currAsset := range assets {
		var args []string
		if verify {
			args = append(args, "--"+assetapi.ListFlagName)
		}

		if nameOutput, err := exec.Command(currAsset, "--"+assetapi.NameFlagName).CombinedOutput(); err == nil {
			assetName := strings.TrimSpace(string(nameOutput))
			if formatterCfg := assetArgs[assetName]; len(formatterCfg.Args) > 0 {
				args = append(args, formatterCfg.Args...)
			}
		}
		args = append(args, files...)

		cmd := exec.Command(currAsset, args...)
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		cmd.Stdin = os.Stdin

		// if in "verify" mode, collect output in buffer rather than streaming to output
		if verify {
			cmd.Stdout = &outputBuf
			cmd.Stderr = &outputBuf
		}
		if err := cmd.Run(); err != nil {
			if verify {
				// if in "verify" mode, output has not been streamed, so print to stdout
				fmt.Fprint(stdout, outputBuf.String())
			}
			return fmt.Errorf("")
		}
	}
	// if in "verify" mode, print output
	if verify {
		output := orderedFileLines(outputBuf.String(), files)
		if len(output) != 0 {
			fmt.Fprintln(stdout, output)
			return fmt.Errorf("")
		}
	}
	return nil
}

// orderedFileLines takes "in", splits it on "\n", and then records each line as a "file" entry. It then iterates over
// the provided files and returns a string consisting of only the files that match lines in "in" concatenated with a
// '\n' character.
func orderedFileLines(in string, files []string) string {
	seen := make(map[string]struct{})
	for _, currLine := range strings.Split(in, "\n") {
		seen[currLine] = struct{}{}
	}
	var outLines []string
	for _, f := range files {
		if _, ok := seen[f]; !ok {
			continue
		}
		outLines = append(outLines, f)
	}
	return strings.Join(outLines, "\n")
}

func allMatchingFilesInDir(dir string, exclude matcher.Matcher) ([]string, error) {
	// exclude entries specified by the configuration
	matchedFiles, err := matcher.ListFiles(dir, matcher.Name(`.*\.go`), exclude)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to determine Go files in %s", dir)
	}
	return matchedFiles, nil
}

func filterFiles(projectDir, wd string, providedFiles []string, exclude matcher.Matcher) ([]string, error) {
	if exclude == nil {
		return providedFiles, nil
	}

	var projectDirAbsPath string
	if projectDir != wd && !filepath.IsAbs(projectDir) {
		absPath, err := filepath.Abs(projectDir)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to convert project directory to absolute path")
		}
		projectDirAbsPath = absPath
	}

	var files []string
	// filter arguments based on exclude config
	for _, currPath := range providedFiles {
		checkPath := currPath
		if projectDirAbsPath != "" {
			fullPath, err := filepath.Abs(path.Join(wd, checkPath))
			if err != nil {
				return nil, errors.Wrapf(err, "failed to convert to absolute path")
			}
			projectDirRelPath, err := filepath.Rel(projectDirAbsPath, fullPath)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to convert to relative path")
			}
			checkPath = projectDirRelPath
		}

		if exclude.Match(checkPath) {
			continue
		}
		files = append(files, currPath)
	}
	return files, nil
}
