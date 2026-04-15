//go:build !nostream
// +build !nostream

package defaultset

// 本文件承载默认模块集合中与 `stream_enabled` 相关的装配逻辑。

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
	streamhandler "github.com/yttydcs/myflowhub-subproto/stream"
)

func newStreamHandler(cfg core.IConfig, log *slog.Logger) (core.ISubProcess, error) {
	// Stream 默认接线只做装配，不把流状态管理细节泄漏到 Server 层。
	return streamhandler.NewHandlerWithConfig(cfg, log), nil
}
