//go:build !noflow
// +build !noflow

package defaultset

// 本文件承载默认模块集合中与 `flow_enabled` 相关的装配逻辑。

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-subproto/exec/runtimedeps"
	flowhandler "github.com/yttydcs/myflowhub-subproto/flow"
)

func newFlowHandler(cfg core.IConfig, deps runtimedeps.Deps, log *slog.Logger) (core.ISubProcess, error) {
	// Flow 需要先解析定义持久化与 run archive 后端，避免 handler 内部硬编码存储实现。
	store, err := newFlowPersistence(cfg)
	if err != nil {
		return nil, err
	}
	archiveStore, err := newFlowRunArchiveStore(cfg)
	if err != nil {
		return nil, err
	}
	return flowhandler.NewHandlerWithOptions(cfg, flowhandler.HandlerOptions{
		RuntimeDeps:     deps,
		Persistence:     store,
		RunArchiveStore: archiveStore,
	}, log), nil
}
