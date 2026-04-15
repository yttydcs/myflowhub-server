//go:build !nofile
// +build !nofile

package defaultset

// 本文件承载默认模块集合中与 `file_enabled` 相关的装配逻辑。

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-subproto/exec/runtimedeps"
	filehandler "github.com/yttydcs/myflowhub-subproto/file"
)

func newFileHandler(cfg core.IConfig, deps runtimedeps.Deps, log *slog.Logger) (core.ISubProcess, error) {
	// 默认集合只负责把共享依赖注入 File 子协议，具体传输状态机仍留在子模块内部。
	return filehandler.NewHandlerWithDeps(cfg, deps, log), nil
}
