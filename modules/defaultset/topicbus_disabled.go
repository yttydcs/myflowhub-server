//go:build notopicbus
// +build notopicbus

package defaultset

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
)

func newTopicBusHandler(cfg core.IConfig, log *slog.Logger) core.ISubProcess {
	return nil
}
