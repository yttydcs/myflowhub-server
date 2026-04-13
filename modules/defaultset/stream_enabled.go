//go:build !nostream
// +build !nostream

package defaultset

// Context: This file lives in the Server assembly layer and supports stream_enabled.

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
	streamhandler "github.com/yttydcs/myflowhub-subproto/stream"
)

func newStreamHandler(cfg core.IConfig, log *slog.Logger) (core.ISubProcess, error) {
	return streamhandler.NewHandlerWithConfig(cfg, log), nil
}
