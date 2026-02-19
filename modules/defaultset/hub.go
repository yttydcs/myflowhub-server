package defaultset

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-subproto/management"
	"github.com/yttydcs/myflowhub-server/subproto/forward"
)

// DefaultHub 返回 hub_server 的默认启用模块集合（handlers + default fallback）。
//
// 注意：本包仅负责“默认集合的构造策略”，不做重复校验；校验由上层 `modules.DefaultHub` 统一完成。
func DefaultHub(cfg core.IConfig, log *slog.Logger) (handlers []core.ISubProcess, def core.ISubProcess) {
	handlers = make([]core.ISubProcess, 0, 7)
	handlers = append(handlers, management.NewHandler(log))

	if h := newAuthHandler(cfg, log); h != nil {
		handlers = append(handlers, h)
	}
	if h := newVarStoreHandler(cfg, log); h != nil {
		handlers = append(handlers, h)
	}
	if h := newTopicBusHandler(cfg, log); h != nil {
		handlers = append(handlers, h)
	}
	if h := newExecHandler(cfg, log); h != nil {
		handlers = append(handlers, h)
	}
	if h := newFlowHandler(cfg, log); h != nil {
		handlers = append(handlers, h)
	}
	if h := newFileHandler(cfg, log); h != nil {
		handlers = append(handlers, h)
	}

	def = forward.NewDefaultForwardHandler(cfg, log)
	return handlers, def
}
