//go:build !noexec
// +build !noexec

package defaultset

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
	exechandler "github.com/yttydcs/myflowhub-server/subproto/exec"
)

func newExecHandler(cfg core.IConfig, log *slog.Logger) core.ISubProcess {
	return exechandler.NewHandlerWithConfig(cfg, log)
}
