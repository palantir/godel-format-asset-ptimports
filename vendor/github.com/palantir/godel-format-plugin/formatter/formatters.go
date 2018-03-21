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
	"sort"

	"github.com/pkg/errors"

	"github.com/palantir/godel-format-plugin/formatplugin"
)

type CreatorFunction func(cfgYML []byte) (formatplugin.Formatter, error)

type Creator interface {
	TypeName() string
	Creator() CreatorFunction
}

type creatorStruct struct {
	typeName string
	creator  CreatorFunction
}

func (c *creatorStruct) TypeName() string {
	return c.typeName
}

func (c *creatorStruct) Creator() CreatorFunction {
	return c.creator
}

func NewCreator(typeName string, creatorFn CreatorFunction) Creator {
	return &creatorStruct{
		typeName: typeName,
		creator:  creatorFn,
	}
}

func AssetFormatterCreators(assetPaths ...string) ([]Creator, []formatplugin.ConfigUpgrader, error) {
	var formatterCreators []Creator
	var configUpgraders []formatplugin.ConfigUpgrader
	formatterNameToAssets := make(map[string][]string)
	for _, currAssetPath := range assetPaths {
		currFormatter := assetFormatter{
			assetPath: currAssetPath,
		}
		formatterName, err := currFormatter.TypeName()
		if err != nil {
			return nil, nil, errors.Wrapf(err, "failed to determine formatter type name for asset %s", currAssetPath)
		}
		formatterNameToAssets[formatterName] = append(formatterNameToAssets[formatterName], currAssetPath)
		formatterCreators = append(formatterCreators, NewCreator(formatterName,
			func(cfgYML []byte) (formatplugin.Formatter, error) {
				currFormatter.cfgYML = string(cfgYML)
				if err := currFormatter.VerifyConfig(); err != nil {
					return nil, err
				}
				return &currFormatter, nil
			}))
		configUpgraders = append(configUpgraders, &assetConfigUpgrader{
			typeName:  formatterName,
			assetPath: currAssetPath,
		})
	}
	var sortedKeys []string
	for k := range formatterNameToAssets {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	for _, k := range sortedKeys {
		if len(formatterNameToAssets[k]) <= 1 {
			continue
		}
		sort.Strings(formatterNameToAssets[k])
		return nil, nil, errors.Errorf("formatter type %s provided by multiple assets: %v", k, formatterNameToAssets[k])
	}
	return formatterCreators, configUpgraders, nil
}
