package defaultset

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
	authhandler "github.com/yttydcs/myflowhub-server/subproto/auth"
	exechandler "github.com/yttydcs/myflowhub-server/subproto/exec"
	filehandler "github.com/yttydcs/myflowhub-server/subproto/file"
	flowhandler "github.com/yttydcs/myflowhub-server/subproto/flow"
	"github.com/yttydcs/myflowhub-server/subproto/forward"
	"github.com/yttydcs/myflowhub-server/subproto/management"
	"github.com/yttydcs/myflowhub-server/subproto/topicbus"
	varstore "github.com/yttydcs/myflowhub-server/subproto/varstore"
)

// DefaultHub 返回 hub_server 的默认启用模块集合（handlers + default fallback）。
//
// 注意：本包仅负责“默认集合的构造策略”，不做重复校验；校验由上层 `modules.DefaultHub` 统一完成。
func DefaultHub(cfg core.IConfig, log *slog.Logger) (handlers []core.ISubProcess, def core.ISubProcess) {
	return []core.ISubProcess{
			management.NewHandler(log),
			authhandler.NewLoginHandlerWithConfig(cfg, log),
			varstore.NewVarStoreHandlerWithConfig(cfg, log),
			topicbus.NewTopicBusHandlerWithConfig(cfg, log),
			exechandler.NewHandlerWithConfig(cfg, log),
			flowhandler.NewHandlerWithConfig(cfg, log),
			filehandler.NewHandlerWithConfig(cfg, log),
		},
		forward.NewDefaultForwardHandler(cfg, log)
}

