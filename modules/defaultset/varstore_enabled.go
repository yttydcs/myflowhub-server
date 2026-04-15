//go:build !novarstore
// +build !novarstore

package defaultset

// 本文件承载默认模块集合中与 `varstore_enabled` 相关的装配逻辑。

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-subproto/exec/runtimedeps"
	varstorehandler "github.com/yttydcs/myflowhub-subproto/varstore"
)

func newVarStoreHandler(cfg core.IConfig, deps runtimedeps.Deps, log *slog.Logger) (core.ISubProcess, error) {
	// VarStore 先在装配层选定持久化后端，再把统一的运行时依赖交给子协议实现。
	store, err := newVarStorePersistence(cfg)
	if err != nil {
		return nil, err
	}
	return varstorehandler.NewVarStoreHandlerWithOptions(cfg, varstorehandler.HandlerOptions{
		RuntimeDeps: deps,
		Persistence: store,
	}, log), nil
}
