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

package formatplugin

import (
	"sort"

	"github.com/palantir/pkg/matcher"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/palantir/godel-format-plugin/formatter"
)

type Param struct {
	Formatters []formatter.Formatter
	Exclude    matcher.Matcher
}

type Config struct {
	Formatters map[string]FormatterConfig `yaml:"formatters"`
	Exclude    matcher.NamesPathsCfg      `yaml:"exclude"`
}

func (cfg *Config) ToParam(factory formatter.Factory) (Param, error) {
	knownTypes := make(map[string]struct{})
	for _, formatterType := range factory.FormatterTypes() {
		knownTypes[formatterType] = struct{}{}
	}

	var sortedKeys []string
	for k := range cfg.Formatters {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	var unknownFormatters []string
	for _, k := range sortedKeys {
		if _, ok := knownTypes[k]; ok {
			continue
		}
		unknownFormatters = append(unknownFormatters, k)
	}
	if len(unknownFormatters) > 0 {
		return Param{}, errors.Errorf("formatters %v not recognized -- known formatters are %v", unknownFormatters, factory.FormatterTypes())
	}

	var formatters []formatter.Formatter
	for _, formatterName := range factory.FormatterTypes() {
		var cfgBytes []byte
		if formatterCfg, ok := cfg.Formatters[formatterName]; ok {
			if cfgMapSlice := formatterCfg.Config; cfgMapSlice != nil {
				outBytes, err := yaml.Marshal(*cfgMapSlice)
				if err != nil {
					return Param{}, errors.Wrapf(err, "failed to marshal configuration for %s", formatterName)
				}
				cfgBytes = outBytes
			}
		}
		formatter, err := factory.NewFormatter(formatterName, cfgBytes)
		if err != nil {
			return Param{}, err
		}
		formatters = append(formatters, formatter)
	}

	return Param{
		Formatters: formatters,
		Exclude:    cfg.Exclude.Matcher(),
	}, nil
}

type FormatterConfig struct {
	Config *yaml.MapSlice `yaml:"config"`
}
