//go:build nostream
// +build nostream

package defaultset

// 本文件承载默认模块集合中与 `stream_disabled` 相关的装配逻辑。

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
)

func newStreamHandler(core.IConfig, *slog.Logger) (core.ISubProcess, error) {
	return nil, nil
}
