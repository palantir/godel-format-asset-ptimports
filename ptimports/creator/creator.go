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

package creator

import (
	"github.com/palantir/godel-format-asset-ptimports/ptimports"
	"github.com/palantir/godel-format-asset-ptimports/ptimports/config"
	"github.com/palantir/godel-format-plugin/formatplugin"
	"github.com/palantir/godel-format-plugin/formatter"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

func Ptimports() formatter.Creator {
	return formatter.NewCreator(
		ptimports.TypeName,
		func(cfgYML []byte) (formatplugin.Formatter, error) {
			var formatCfg config.Ptimports
			if err := yaml.Unmarshal(cfgYML, &formatCfg); err != nil {
				return nil, errors.Wrapf(err, "failed to unmarshal YAML")
			}
			return formatCfg.ToFormatter(), nil
		},
	)
}
