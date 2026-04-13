//go:build notopicbus
// +build notopicbus

package defaultset

// Context: This file lives in the Server assembly layer and supports topicbus_disabled.

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-subproto/exec/runtimedeps"
)

func newTopicBusHandler(cfg core.IConfig, deps runtimedeps.Deps, log *slog.Logger) (core.ISubProcess, error) {
	return nil, nil
}
