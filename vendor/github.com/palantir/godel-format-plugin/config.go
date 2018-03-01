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
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

func readFormatConfigFromFile(cfg string) (formatConfig, error) {
	bytes, err := ioutil.ReadFile(cfg)
	if err != nil {
		return formatConfig{}, errors.Wrapf(err, "failed to read config file")
	}
	return readFormatConfig(bytes)
}

func readFormatConfig(cfg []byte) (formatConfig, error) {
	var formatCfg formatConfig
	if err := yaml.Unmarshal(cfg, &formatCfg); err != nil {
		return formatConfig{}, errors.Wrapf(err, "failed to unmarshal YAML")
	}
	return formatCfg, nil
}

type formatConfig struct {
	Formatters map[string]formatterConfig `yaml:"formatters"`
}

type formatterConfig struct {
	Args []string `yaml:"args"`
}
