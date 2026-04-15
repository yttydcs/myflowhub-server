//go:build !notopicbus
// +build !notopicbus

package defaultset

// 本文件承载默认模块集合中与 `topicbus_enabled` 相关的装配逻辑。

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-subproto/exec/runtimedeps"
	topicbushandler "github.com/yttydcs/myflowhub-subproto/topicbus"
)

func newTopicBusHandler(cfg core.IConfig, deps runtimedeps.Deps, log *slog.Logger) (core.ISubProcess, error) {
	// TopicBus 依赖 capability registry 等共享运行时依赖，因此在默认装配阶段统一注入。
	return topicbushandler.NewTopicBusHandlerWithDeps(cfg, deps, log), nil
}
