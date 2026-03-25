//go:build !notopicbus
// +build !notopicbus

package defaultset

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-subproto/exec/runtimedeps"
	topicbushandler "github.com/yttydcs/myflowhub-subproto/topicbus"
)

func newTopicBusHandler(cfg core.IConfig, deps runtimedeps.Deps, log *slog.Logger) (core.ISubProcess, error) {
	return topicbushandler.NewTopicBusHandlerWithDeps(cfg, deps, log), nil
}
