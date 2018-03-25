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

package v0

import (
	"bytes"
	"sort"

	"github.com/palantir/pkg/matcher"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/palantir/godel-format-plugin/formatplugin"
)

type Config struct {
	Formatters map[string]FormatterConfig `yaml:"formatters"`
	Exclude    matcher.NamesPathsCfg      `yaml:"exclude"`
}

type FormatterConfig struct {
	Config yaml.MapSlice `yaml:"config"`
}

func UpgradeConfig(cfgBytes []byte, factory formatplugin.Factory) ([]byte, error) {
	var cfg Config
	if err := yaml.UnmarshalStrict(cfgBytes, &cfg); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal format-plugin v0 configuration")
	}
	changed, err := upgradeAssets(&cfg, factory)
	if err != nil {
		return nil, err
	}
	if !changed {
		return cfgBytes, nil
	}
	upgradedBytes, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal format-plugin v0 configuration")
	}
	return upgradedBytes, nil
}

// upgradeAssets upgrades the assets for the provided configuration. Returns true if any upgrade operations were
// performed. If any upgrade operations were performed, the provided configuration is modified directly.
func upgradeAssets(cfg *Config, factory formatplugin.Factory) (changed bool, rErr error) {
	var sortedKeys []string
	for k := range cfg.Formatters {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for _, k := range sortedKeys {
		upgrader, err := factory.ConfigUpgrader(k)
		if err != nil {
			return false, err
		}

		assetCfgBytes, err := yaml.Marshal(cfg.Formatters[k].Config)
		if err != nil {
			return false, errors.Wrapf(err, "failed to marshal asset configuration for formatter %q", k)
		}

		upgradedBytes, err := upgrader.UpgradeConfig(assetCfgBytes)
		if err != nil {
			return false, errors.Wrapf(err, "failed to upgrade asset configuration for formatter %q", k)
		}

		if bytes.Equal(assetCfgBytes, upgradedBytes) {
			// upgrade was a no-op: do not modify configuration and continue
			continue
		}
		changed = true

		var yamlRep yaml.MapSlice
		if err := yaml.Unmarshal(upgradedBytes, &yamlRep); err != nil {
			return false, errors.Wrapf(err, "failed to unmarshal upgraded configuration for formatter %q", k)
		}

		// update configuration for asset in original configuration
		assetFormatCfg := cfg.Formatters[k]
		assetFormatCfg.Config = yamlRep
		cfg.Formatters[k] = assetFormatCfg
	}
	return changed, nil
}
