//go:build !noflow
// +build !noflow

package defaultset

// Context: This file lives in the Server assembly layer and supports flow_enabled.

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-subproto/exec/runtimedeps"
	flowhandler "github.com/yttydcs/myflowhub-subproto/flow"
)

func newFlowHandler(cfg core.IConfig, deps runtimedeps.Deps, log *slog.Logger) (core.ISubProcess, error) {
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
