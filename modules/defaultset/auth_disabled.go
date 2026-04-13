//go:build noauth
// +build noauth

package defaultset

// Context: This file lives in the Server assembly layer and supports auth_disabled.

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
)

func newAuthHandler(cfg core.IConfig, log *slog.Logger) core.ISubProcess {
	return nil
}
