//go:build !notopicbus
// +build !notopicbus

package defaultset

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
	topicbushandler "github.com/yttydcs/myflowhub-subproto/topicbus"
)

func newTopicBusHandler(cfg core.IConfig, log *slog.Logger) core.ISubProcess {
	return topicbushandler.NewTopicBusHandlerWithConfig(cfg, log)
}
