//go:build !noauth
// +build !noauth

package defaultset

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
	authhandler "github.com/yttydcs/myflowhub-subproto/auth"
)

func newAuthHandler(cfg core.IConfig, log *slog.Logger) core.ISubProcess {
	return authhandler.NewLoginHandlerWithConfig(cfg, log)
}
