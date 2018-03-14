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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/palantir/godel-format-plugin/formatter"
)

var (
	debugFlagVal            bool
	projectDirFlagVal       string
	godelConfigFileFlagVal  string
	formatConfigFileFlagVal string
	verifyFlagVal           bool
	assetsFlagVal           []string

	cliFormatterFactory formatter.Factory
)

var rootCmd = &cobra.Command{
	Use:   "format-plugin [flags] [files]",
	Short: "Format specified files (if no files are specified, format all project Go files)",
	RunE: func(cmd *cobra.Command, args []string) error {
		var exclude matcher.Matcher
		if godelConfigFileFlagVal != "" {
			cfg, err := godellauncher.ReadGodelConfig(path.Dir(godelConfigFileFlagVal))
			if err != nil {
				return err
			}
			exclude = cfg.Exclude.Matcher()
		}

		var formatters []formatter.Formatter
		if formatConfigFileFlagVal != "" {
			cfg, err := readFormatConfigFromFile(formatConfigFileFlagVal)
			if err != nil {
				return err
			}
			for _, currType := range cliFormatterFactory.FormatterTypes() {
				currCfg := cfg.Formatters[currType]

				var cfgBytes []byte
				if currCfg.Config != nil {
					bytes, err := yaml.Marshal(currCfg.Config)
					if err != nil {
						return errors.Wrapf(err, "failed to marshal configuration YAML for formatter %s", currType)
					}
					cfgBytes = bytes
				}
				formatter, err := cliFormatterFactory.NewFormatter(currType, cfgBytes)
				if err != nil {
					return err
				}
				formatters = append(formatters, formatter)
			}
		}

		// no formatters specified
		if len(assetsFlagVal) == 0 {
			return nil
		}
		return runFormat(formatters, projectDirFlagVal, exclude, verifyFlagVal, args, cmd.OutOrStdout())
	},
}

func init() {
	pluginapi.AddDebugPFlagPtr(rootCmd.PersistentFlags(), &debugFlagVal)
	pluginapi.AddGodelConfigPFlagPtr(rootCmd.PersistentFlags(), &godelConfigFileFlagVal)
	pluginapi.AddConfigPFlagPtr(rootCmd.PersistentFlags(), &formatConfigFileFlagVal)
	pluginapi.AddProjectDirPFlagPtr(rootCmd.PersistentFlags(), &projectDirFlagVal)
	pluginapi.AddAssetsPFlagPtr(rootCmd.PersistentFlags(), &assetsFlagVal)
	if err := rootCmd.MarkPersistentFlagRequired(pluginapi.ProjectDirFlagName); err != nil {
		panic(err)
	}
	rootCmd.PersistentFlags().BoolVar(&verifyFlagVal, "verify", false, "verify files match formatting without applying formatting")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		assetFormatters, err := formatter.AssetFormatterCreators(assetsFlagVal...)
		if err != nil {
			return err
		}
		cliFormatterFactory, err = formatter.NewFormatterFactory(assetFormatters...)
		if err != nil {
			return err
		}
		return nil
	}
}
