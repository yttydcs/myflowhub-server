package defaultset

// 本文件承载默认模块集合中与 `runtime_deps` 相关的装配逻辑。

import (
	core "github.com/yttydcs/myflowhub-core"
	permission "github.com/yttydcs/myflowhub-core/kit/permission"
	execcap "github.com/yttydcs/myflowhub-subproto/exec/capability"
	"github.com/yttydcs/myflowhub-subproto/exec/runtimedeps"
)

// newRuntimeDeps 为默认模块集合提供共享的权限配置和 capability registry。
func newRuntimeDeps(cfg core.IConfig) runtimedeps.Deps {
	return runtimedeps.Deps{
		CapRegistry: execcap.NewRegistry(),
		PermConfig:  permission.NewConfig(cfg),
	}
}
