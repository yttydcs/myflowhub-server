//go:build !noexec
// +build !noexec

package defaultset

// 本文件承载默认模块集合中与 `exec_enabled` 相关的装配逻辑。

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
	exechandler "github.com/yttydcs/myflowhub-subproto/exec"
	"github.com/yttydcs/myflowhub-subproto/exec/runtimedeps"
)

func newExecHandler(cfg core.IConfig, deps runtimedeps.Deps, log *slog.Logger) (core.ISubProcess, error) {
	return exechandler.NewHandlerWithDeps(cfg, deps, log), nil
}
