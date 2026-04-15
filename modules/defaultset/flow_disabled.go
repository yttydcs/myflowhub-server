//go:build noflow
// +build noflow

package defaultset

// 本文件承载默认模块集合中与 `flow_disabled` 相关的装配逻辑。

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-subproto/exec/runtimedeps"
)

func newFlowHandler(cfg core.IConfig, deps runtimedeps.Deps, log *slog.Logger) (core.ISubProcess, error) {
	return nil, nil
}
