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

	"github.com/palantir/godel/framework/pluginapi"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func AssetRootCmd(creator Creator, upgradeConfigFn pluginapi.UpgradeConfigFn, short string) *cobra.Command {
	name := creator.TypeName()
	rootCmd := &cobra.Command{
		Use:   name,
		Short: short,
	}

	creatorFn := creator.Creator()
	rootCmd.AddCommand(newNameCmd(name))
	rootCmd.AddCommand(newVerifyConfigCmd(creatorFn))
	rootCmd.AddCommand(newRunFormatCmd(creatorFn))
	rootCmd.AddCommand(pluginapi.CobraUpgradeConfigCmd(upgradeConfigFn))

	return rootCmd
}

const nameCmdName = "name"

func newNameCmd(name string) *cobra.Command {
	return &cobra.Command{
		Use:   nameCmdName,
		Short: "Print the name of the formatter",
		RunE: func(cmd *cobra.Command, args []string) error {
			outputJSON, err := json.Marshal(name)
			if err != nil {
				return errors.Wrapf(err, "failed to marshal output as JSON")
			}
			cmd.Print(string(outputJSON))
			return nil
		},
	}
}

const commonCmdConfigYMLFlagName = "config-yml"

const (
	verifyConfigCmdName = "verify-config"
)

func newVerifyConfigCmd(creatorFn CreatorFunction) *cobra.Command {
	var configYMLFlagVal string
	verifyConfigCmd := &cobra.Command{
		Use:   verifyConfigCmdName,
		Short: "Verify that the provided input is valid configuration YML for this formatter",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := creatorFn([]byte(configYMLFlagVal))
			return err
		},
	}
	verifyConfigCmd.Flags().StringVar(&configYMLFlagVal, commonCmdConfigYMLFlagName, "", "configuration YML to verify")
	mustMarkFlagsRequired(verifyConfigCmd, commonCmdConfigYMLFlagName)
	return verifyConfigCmd
}

const (
	runFormatCmdName         = "run-format"
	runFormatCmdListFlagName = "list"
)

func newRunFormatCmd(creatorFn CreatorFunction) *cobra.Command {
	var (
		configYMLFlagVal string
		listFlagVal      bool
	)
	runFormatCmd := &cobra.Command{
		Use:   runFormatCmdName,
		Short: "Runs the format operation",
		RunE: func(cmd *cobra.Command, args []string) error {
			formatter, err := creatorFn([]byte(configYMLFlagVal))
			if err != nil {
				return err
			}
			if err := formatter.Format(args, listFlagVal, cmd.OutOrStdout()); err != nil {
				return err
			}
			return nil
		},
	}
	runFormatCmd.Flags().StringVar(&configYMLFlagVal, commonCmdConfigYMLFlagName, "", "YML of formatter configuration")
	runFormatCmd.Flags().BoolVar(&listFlagVal, runFormatCmdListFlagName, false, "list the files that would be modified by the operation")
	mustMarkFlagsRequired(runFormatCmd, commonCmdConfigYMLFlagName)
	return runFormatCmd
}

func mustMarkFlagsRequired(cmd *cobra.Command, flagNames ...string) {
	for _, currFlagName := range flagNames {
		if err := cmd.MarkFlagRequired(currFlagName); err != nil {
			panic(err)
		}
	}
}
