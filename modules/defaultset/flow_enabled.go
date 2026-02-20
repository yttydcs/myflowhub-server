//go:build !noflow
// +build !noflow

package defaultset

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
	flowhandler "github.com/yttydcs/myflowhub-subproto/flow"
)

func newFlowHandler(cfg core.IConfig, log *slog.Logger) core.ISubProcess {
	return flowhandler.NewHandlerWithConfig(cfg, log)
}
