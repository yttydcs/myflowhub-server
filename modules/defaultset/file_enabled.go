//go:build !nofile
// +build !nofile

package defaultset

// Context: This file lives in the Server assembly layer and supports file_enabled.

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-subproto/exec/runtimedeps"
	filehandler "github.com/yttydcs/myflowhub-subproto/file"
)

func newFileHandler(cfg core.IConfig, deps runtimedeps.Deps, log *slog.Logger) (core.ISubProcess, error) {
	return filehandler.NewHandlerWithDeps(cfg, deps, log), nil
}
