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
	"os"

	"github.com/palantir/amalgomate/amalgomated"
	"github.com/palantir/godel-format-plugin/formatter"
	"github.com/palantir/pkg/cobracli"

	"github.com/palantir/godel-format-asset-ptimports/generated_src"
	"github.com/palantir/godel-format-asset-ptimports/ptimports"
)

const assetName = "ptimports"

var debugFlagVal bool

func main() {
	if len(os.Args) >= 2 && os.Args[1] == amalgomated.ProxyCmdPrefix+assetName {
		os.Args = append(os.Args[:1], os.Args[2:]...)
		amalgomatedformatter.Instance().Run(assetName)
		os.Exit(0)
	}

	rootCmd := formatter.AssetRootCmd(ptimports.Creator(), "")
	os.Exit(cobracli.ExecuteWithDefaultParamsWithVersion(rootCmd, &debugFlagVal, ""))
}
