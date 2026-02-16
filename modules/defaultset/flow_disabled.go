//go:build noflow
// +build noflow

package defaultset

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
)

func newFlowHandler(cfg core.IConfig, log *slog.Logger) core.ISubProcess {
	return nil
}
