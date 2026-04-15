//go:build !noauth
// +build !noauth

package defaultset

// 本文件承载默认模块集合中与 `auth_enabled` 相关的装配逻辑。

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
	authhandler "github.com/yttydcs/myflowhub-subproto/auth"
)

// newAuthHandler 在启用 auth build tag 时构造默认 auth handler。
func newAuthHandler(cfg core.IConfig, log *slog.Logger) core.ISubProcess {
	return authhandler.NewLoginHandlerWithConfig(cfg, log)
}
