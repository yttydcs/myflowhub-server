//go:build noauth
// +build noauth

package defaultset

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
)

func newAuthHandler(cfg core.IConfig, log *slog.Logger) core.ISubProcess {
	return nil
}
