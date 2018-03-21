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

package cmd

import (
	"io/ioutil"

	"github.com/palantir/godel/framework/pluginapi"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/palantir/godel-format-plugin/formatplugin"
	"github.com/palantir/godel-format-plugin/formatplugin/config"
	"github.com/palantir/godel-format-plugin/formatter"
	"github.com/palantir/godel-format-plugin/formatter/formatterfactory"
)

var (
	DebugFlagVal bool

	projectDirFlagVal       string
	godelConfigFileFlagVal  string
	formatConfigFileFlagVal string
	verifyFlagVal           bool
	assetsFlagVal           []string

	cliFormatterFactory formatplugin.Factory
)

var RootCmd = &cobra.Command{
	Use:   "format-plugin [flags] [files]",
	Short: "Format specified files (if no files are specified, format all project Go files)",
}

func init() {
	pluginapi.AddDebugPFlagPtr(RootCmd.PersistentFlags(), &DebugFlagVal)
	pluginapi.AddGodelConfigPFlagPtr(RootCmd.PersistentFlags(), &godelConfigFileFlagVal)
	pluginapi.AddConfigPFlagPtr(RootCmd.PersistentFlags(), &formatConfigFileFlagVal)
	pluginapi.AddProjectDirPFlagPtr(RootCmd.PersistentFlags(), &projectDirFlagVal)
	pluginapi.AddAssetsPFlagPtr(RootCmd.PersistentFlags(), &assetsFlagVal)
	if err := RootCmd.MarkPersistentFlagRequired(pluginapi.ProjectDirFlagName); err != nil {
		panic(err)
	}

	RootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		assetFormatters, assetConfigUpgraders, err := formatter.AssetFormatterCreators(assetsFlagVal...)
		if err != nil {
			return err
		}
		cliFormatterFactory, err = formatterfactory.New(assetFormatters, assetConfigUpgraders)
		if err != nil {
			return err
		}
		return nil
	}
}

func readFormatConfigFromFile(cfgFile string) (config.Format, error) {
	bytes, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return config.Format{}, errors.Wrapf(err, "failed to read config file")
	}

	upgradedCfg, err := config.UpgradeConfig(bytes, cliFormatterFactory)
	if err != nil {
		return config.Format{}, err
	}

	var formatCfg config.Format
	if err := yaml.Unmarshal(upgradedCfg, &formatCfg); err != nil {
		return config.Format{}, errors.Wrapf(err, "failed to unmarshal YAML")
	}
	return formatCfg, nil
}
