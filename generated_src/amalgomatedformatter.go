// generated by amalgomate; DO NOT EDIT
package amalgomatedformatter

import (
	"fmt"
	"sort"

	ptimports "github.com/palantir/godel-format-asset-ptimports/generated_src/internal/github.com/palantir/go-ptimports/v2"
)

var programs = map[string]func(){"ptimports": func() {
	ptimports.AmalgomatedMain()
},
}

func Instance() Amalgomated {
	return &amalgomated{}
}

type Amalgomated interface {
	Run(cmd string)
	Cmds() []string
}

type amalgomated struct{}

func (a *amalgomated) Run(cmd string) {
	if _, ok := programs[cmd]; !ok {
		panic(fmt.Sprintf("Unknown command: \"%v\". Valid values: %v", cmd, a.Cmds()))
	}
	programs[cmd]()
}

func (a *amalgomated) Cmds() []string {
	var cmds []string
	for key := range programs {
		cmds = append(cmds, key)
	}
	sort.Strings(cmds)
	return cmds
}
