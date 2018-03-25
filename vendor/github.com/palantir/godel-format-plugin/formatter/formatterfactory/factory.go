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

package formatterfactory

import (
	"github.com/pkg/errors"

	"github.com/palantir/godel-format-plugin/formatplugin"
	"github.com/palantir/godel-format-plugin/formatter"
)

func New(providedFormatterCreators []formatter.Creator, providedConfigUpgraders []formatplugin.ConfigUpgrader) (formatplugin.Factory, error) {
	var formatterTypes []string
	formatterCreators := make(map[string]formatter.CreatorFunction)
	for _, currCreator := range providedFormatterCreators {
		formatterTypes = append(formatterTypes, currCreator.TypeName())
		formatterCreators[currCreator.TypeName()] = currCreator.Creator()
	}
	configUpgraders := make(map[string]formatplugin.ConfigUpgrader)
	for _, currUpgrader := range providedConfigUpgraders {
		currUpgrader := currUpgrader
		configUpgraders[currUpgrader.TypeName()] = currUpgrader
	}
	return &formatterFactoryImpl{
		types:                    formatterTypes,
		formatterCreators:        formatterCreators,
		formatterConfigUpgraders: configUpgraders,
	}, nil
}

type formatterFactoryImpl struct {
	types                    []string
	formatterCreators        map[string]formatter.CreatorFunction
	formatterConfigUpgraders map[string]formatplugin.ConfigUpgrader
}

func (f *formatterFactoryImpl) NewFormatter(typeName string, cfgYMLBytes []byte) (formatplugin.Formatter, error) {
	creatorFn, ok := f.formatterCreators[typeName]
	if !ok {
		return nil, errors.Errorf("no formatters registered for formatter type %q (registered formatters: %v)", typeName, f.types)
	}
	return creatorFn(cfgYMLBytes)
}

func (f *formatterFactoryImpl) Types() []string {
	return f.types
}

func (f *formatterFactoryImpl) ConfigUpgrader(typeName string) (formatplugin.ConfigUpgrader, error) {
	if _, ok := f.formatterCreators[typeName]; !ok {
		return nil, errors.Errorf("no formatters registered for formatter type %q (registered formatters: %v)", typeName, f.types)
	}
	upgrader, ok := f.formatterConfigUpgraders[typeName]
	if !ok {
		return nil, errors.Errorf("%s is a valid formatter but does not have a config upgrader", typeName)
	}
	return upgrader, nil
}
