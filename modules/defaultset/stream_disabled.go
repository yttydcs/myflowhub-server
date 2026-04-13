//go:build nostream
// +build nostream

package defaultset

// Context: This file lives in the Server assembly layer and supports stream_disabled.

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
)

func newStreamHandler(core.IConfig, *slog.Logger) (core.ISubProcess, error) {
	return nil, nil
}
