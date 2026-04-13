//go:build nofile
// +build nofile

package defaultset

// Context: This file lives in the Server assembly layer and supports file_disabled.

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-subproto/exec/runtimedeps"
)

func newFileHandler(cfg core.IConfig, deps runtimedeps.Deps, log *slog.Logger) (core.ISubProcess, error) {
	return nil, nil
}
