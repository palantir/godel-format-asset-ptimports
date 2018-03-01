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
	"path"

	"github.com/palantir/godel/framework/godellauncher"
	"github.com/palantir/godel/framework/pluginapi"
	"github.com/palantir/pkg/matcher"
	"github.com/spf13/cobra"
)

var (
	projectDirFlagVal       string
	godelConfigFileFlagVal  string
	formatConfigFileFlagVal string
	verifyFlagVal           bool
	assetsFlagVal           []string
)

var rootCmd = &cobra.Command{
	Use:   "format-plugin [flags] [files]",
	Short: "Run format on all project files",
	RunE: func(cmd *cobra.Command, args []string) error {
		var exclude matcher.Matcher
		if godelConfigFileFlagVal != "" {
			cfg, err := godellauncher.ReadGodelConfig(path.Dir(godelConfigFileFlagVal))
			if err != nil {
				return err
			}
			exclude = cfg.Exclude.Matcher()
		}

		var assetArgs map[string]formatterConfig
		if formatConfigFileFlagVal != "" {
			cfg, err := readFormatConfigFromFile(formatConfigFileFlagVal)
			if err != nil {
				return err
			}
			assetArgs = cfg.Formatters
		}

		// no formatters specified
		if len(assetsFlagVal) == 0 {
			return nil
		}
		return runFormat(assetsFlagVal, assetArgs, projectDirFlagVal, exclude, verifyFlagVal, args, cmd.OutOrStdout(), cmd.OutOrStderr())
	},
}

func init() {
	pluginapi.AddGodelConfigPFlagPtr(rootCmd.PersistentFlags(), &godelConfigFileFlagVal)
	pluginapi.AddConfigPFlagPtr(rootCmd.PersistentFlags(), &formatConfigFileFlagVal)
	pluginapi.AddProjectDirPFlagPtr(rootCmd.PersistentFlags(), &projectDirFlagVal)
	pluginapi.AddAssetsPFlagPtr(rootCmd.PersistentFlags(), &assetsFlagVal)
	if err := rootCmd.MarkPersistentFlagRequired(pluginapi.ProjectDirFlagName); err != nil {
		panic(err)
	}
	rootCmd.PersistentFlags().BoolVar(&verifyFlagVal, "verify", false, "verify files match formatting without applying formatting")
}
