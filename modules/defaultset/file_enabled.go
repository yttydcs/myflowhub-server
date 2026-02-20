//go:build !nofile
// +build !nofile

package defaultset

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
	filehandler "github.com/yttydcs/myflowhub-subproto/file"
)

func newFileHandler(cfg core.IConfig, log *slog.Logger) core.ISubProcess {
	return filehandler.NewHandlerWithConfig(cfg, log)
}
