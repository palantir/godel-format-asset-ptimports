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
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	ListFlagName = "list"
	NameFlagName = "name"
)

var (
	ListFlagVal bool
	nameFlagVal bool
)

func RootCommand(assetName string, runE func(cmd *cobra.Command, args []string) error) *cobra.Command {
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s [flags] [files]", assetName),
		Short: fmt.Sprintf("Runs %s", assetName),
		RunE: func(cmd *cobra.Command, args []string) error {
			if nameFlagVal {
				cmd.Println(assetName)
				return nil
			}
			if len(args) == 0 {
				return cmd.Help()
			}
			return runE(cmd, args)
		},
	}
	AddFlags(cmd.Flags(), &ListFlagVal, &nameFlagVal)
	return cmd
}

func AddFlags(fset *pflag.FlagSet, listFlag, nameFlag *bool) {
	fset.BoolVarP(listFlag, ListFlagName, "l", false, "run format in verify mode")
	fset.BoolVar(nameFlag, NameFlagName, false, "print the name of the formatter")
}
