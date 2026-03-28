//go:build !nostream
// +build !nostream

package defaultset

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
	streamhandler "github.com/yttydcs/myflowhub-subproto/stream"
)

func newStreamHandler(cfg core.IConfig, log *slog.Logger) (core.ISubProcess, error) {
	return streamhandler.NewHandlerWithConfig(cfg, log), nil
}
