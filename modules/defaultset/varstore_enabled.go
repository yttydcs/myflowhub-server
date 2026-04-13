//go:build !novarstore
// +build !novarstore

package defaultset

// Context: This file lives in the Server assembly layer and supports varstore_enabled.

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-subproto/exec/runtimedeps"
	varstorehandler "github.com/yttydcs/myflowhub-subproto/varstore"
)

func newVarStoreHandler(cfg core.IConfig, deps runtimedeps.Deps, log *slog.Logger) (core.ISubProcess, error) {
	store, err := newVarStorePersistence(cfg)
	if err != nil {
		return nil, err
	}
	return varstorehandler.NewVarStoreHandlerWithOptions(cfg, varstorehandler.HandlerOptions{
		RuntimeDeps: deps,
		Persistence: store,
	}, log), nil
}
