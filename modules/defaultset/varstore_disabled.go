//go:build novarstore
// +build novarstore

package defaultset

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
)

func newVarStoreHandler(cfg core.IConfig, log *slog.Logger) core.ISubProcess {
	return nil
}
