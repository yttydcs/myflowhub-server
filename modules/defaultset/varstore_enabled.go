//go:build !novarstore
// +build !novarstore

package defaultset

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
	varstorehandler "github.com/yttydcs/myflowhub-subproto/varstore"
)

func newVarStoreHandler(cfg core.IConfig, log *slog.Logger) core.ISubProcess {
	return varstorehandler.NewVarStoreHandlerWithConfig(cfg, log)
}
