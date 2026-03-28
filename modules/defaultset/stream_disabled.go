//go:build nostream
// +build nostream

package defaultset

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
)

func newStreamHandler(core.IConfig, *slog.Logger) (core.ISubProcess, error) {
	return nil, nil
}
