//go:build !noauth
// +build !noauth

package defaultset

// Context: This file lives in the Server assembly layer and supports auth_enabled.

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
	authhandler "github.com/yttydcs/myflowhub-subproto/auth"
)

func newAuthHandler(cfg core.IConfig, log *slog.Logger) core.ISubProcess {
	return authhandler.NewLoginHandlerWithConfig(cfg, log)
}
