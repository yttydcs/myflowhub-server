//go:build nofile
// +build nofile

package defaultset

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
)

func newFileHandler(cfg core.IConfig, log *slog.Logger) core.ISubProcess {
	return nil
}
