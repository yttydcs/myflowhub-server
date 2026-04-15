//go:build noauth
// +build noauth

package defaultset

// 本文件承载默认模块集合中与 `auth_disabled` 相关的装配逻辑。

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
)

// newAuthHandler 在 noauth 变体下返回 nil，让上层集合自动跳过 auth。
func newAuthHandler(cfg core.IConfig, log *slog.Logger) core.ISubProcess {
	return nil
}
