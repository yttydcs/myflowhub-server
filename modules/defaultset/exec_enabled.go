//go:build !noexec
// +build !noexec

package defaultset

// Context: This file lives in the Server assembly layer and supports exec_enabled.

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
	exechandler "github.com/yttydcs/myflowhub-subproto/exec"
	"github.com/yttydcs/myflowhub-subproto/exec/runtimedeps"
)

func newExecHandler(cfg core.IConfig, deps runtimedeps.Deps, log *slog.Logger) (core.ISubProcess, error) {
	return exechandler.NewHandlerWithDeps(cfg, deps, log), nil
}
