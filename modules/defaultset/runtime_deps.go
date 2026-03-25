package defaultset

import (
	core "github.com/yttydcs/myflowhub-core"
	permission "github.com/yttydcs/myflowhub-core/kit/permission"
	execcap "github.com/yttydcs/myflowhub-subproto/exec/capability"
	"github.com/yttydcs/myflowhub-subproto/exec/runtimedeps"
)

func newRuntimeDeps(cfg core.IConfig) runtimedeps.Deps {
	return runtimedeps.Deps{
		CapRegistry: execcap.NewRegistry(),
		PermConfig:  permission.NewConfig(cfg),
	}
}
