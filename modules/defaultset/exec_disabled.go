//go:build noexec
// +build noexec

package defaultset

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
)

func newExecHandler(cfg core.IConfig, log *slog.Logger) core.ISubProcess {
	return nil
}
