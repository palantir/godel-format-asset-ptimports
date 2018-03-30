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

package legacy

import (
	"reflect"
	"sort"

	"github.com/palantir/godel/pkg/versionedconfig"
	"github.com/palantir/pkg/matcher"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/palantir/godel-format-plugin/formatplugin"
	"github.com/palantir/godel-format-plugin/formatplugin/config/internal/v0"
)

type Config struct {
	versionedconfig.ConfigWithLegacy `yaml:",inline"`

	// Formatters specifies the configuration used by the formatters. The key is the name of the formatter and the
	// value is the custom configuration for that formatter.
	Formatters map[string]FormatterConfig `yaml:"formatters"`

	// Exclude specifies the files that should be excluded from formatting.
	Exclude matcher.NamesPathsCfg `yaml:"exclude"`
}

type FormatterConfig struct {
	// Args specifies the command-line arguments that are provided to the formatter.
	Args []string `yaml:"args"`
}

type AssetConfig struct {
	versionedconfig.ConfigWithLegacy `yaml:",inline"`
	Args                             []string `yaml:"args"`
}

func UpgradeConfig(cfgBytes []byte, factory formatplugin.Factory) ([]byte, error) {
	var legacyCfg Config
	if err := yaml.UnmarshalStrict(cfgBytes, &legacyCfg); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal format-plugin legacy configuration")
	}

	v0Cfg, err := upgradeLegacyConfig(legacyCfg, factory)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to upgrade format-plugin legacy configuration")
	}

	outputBytes, err := yaml.Marshal(v0Cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal format-plugin v0 configuration")
	}
	return outputBytes, nil
}

func upgradeLegacyConfig(legacyCfg Config, factory formatplugin.Factory) (v0.Config, error) {
	defaultCfg := Config{
		Formatters: map[string]FormatterConfig{
			"gofmt": {
				Args: []string{
					"-s",
				},
			},
		},
	}

	// when upgrading from V0 configuration, set "separate-project-imports" to true
	defaultV0Cfg := v0.Config{
		Formatters: map[string]v0.FormatterConfig{
			"ptimports": {
				Config: yaml.MapSlice{
					yaml.MapItem{
						Key:   "separate-project-imports",
						Value: true,
					},
				},
			},
		},
	}

	if reflect.DeepEqual(legacyCfg.Formatters, defaultCfg.Formatters) && legacyCfg.Exclude.Empty() {
		// special case: this was the default configuration that shipped with gödel. If this is all that existed, return
		// hard-coded default configuration.
		return defaultV0Cfg, nil
	}

	if reflect.DeepEqual(legacyCfg.Formatters, defaultCfg.Formatters) {
		// special case: formatter configuration matches default, but exclude configuration does not. Return hard-coded
		// default configuration with upgraded excludes.
		return v0.Config{
			Formatters: defaultV0Cfg.Formatters,
			Exclude:    legacyCfg.Exclude,
		}, nil
	}

	// configuration does not match default: delegate to asset upgraders
	upgradedCfg := v0.Config{}

	var sortedKeys []string
	for k := range legacyCfg.Formatters {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	if len(sortedKeys) > 0 {
		upgradedCfg.Formatters = make(map[string]v0.FormatterConfig)
	}

	for _, k := range sortedKeys {
		upgrader, err := factory.ConfigUpgrader(k)
		if err != nil {
			return v0.Config{}, err
		}

		assetCfgBytes, err := yaml.Marshal(AssetConfig{
			ConfigWithLegacy: versionedconfig.ConfigWithLegacy{
				Legacy: true,
			},
			Args: legacyCfg.Formatters[k].Args,
		})
		if err != nil {
			return v0.Config{}, errors.Wrapf(err, "failed to marshal formatter %q legacy configuration", k)
		}

		upgradedBytes, err := upgrader.UpgradeConfig(assetCfgBytes)
		if err != nil {
			return v0.Config{}, errors.Wrapf(err, "failed to upgrade formatter %q legacy configuration", k)
		}

		var yamlRep yaml.MapSlice
		if err := yaml.Unmarshal(upgradedBytes, &yamlRep); err != nil {
			return v0.Config{}, errors.Wrapf(err, "failed to unmarshal formatter %q configuration as yaml.MapSlice", k)
		}

		upgradedCfg.Formatters[k] = v0.FormatterConfig{
			Config: yamlRep,
		}
	}
	return upgradedCfg, nil
}
