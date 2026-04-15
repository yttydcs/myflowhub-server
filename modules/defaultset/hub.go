package defaultset

// 本文件承载默认模块集合中与 `hub` 相关的装配逻辑。

import (
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-subproto/forward"
	"github.com/yttydcs/myflowhub-subproto/management"
)

// DefaultHub 返回 hub_server 的默认启用模块集合（handlers + default fallback）。
//
// 注意：本包仅负责“默认集合的构造策略”，不做重复校验；校验由上层 `modules.DefaultHub` 统一完成。
// DefaultHub 按默认产品口径拼出一套可直接用于 hub_server 的 handler 集合。
func DefaultHub(cfg core.IConfig, log *slog.Logger) (handlers []core.ISubProcess, def core.ISubProcess, err error) {
	deps := newRuntimeDeps(cfg)
	handlers = make([]core.ISubProcess, 0, 8)
	handlers = append(handlers, management.NewHandlerWithDeps(deps, log))

	if h := newAuthHandler(cfg, log); h != nil {
		handlers = append(handlers, h)
	}
	if h, err := newVarStoreHandler(cfg, deps, log); err != nil {
		return nil, nil, err
	} else if h != nil {
		handlers = append(handlers, h)
	}
	if h, err := newTopicBusHandler(cfg, deps, log); err != nil {
		return nil, nil, err
	} else if h != nil {
		handlers = append(handlers, h)
	}
	if h, err := newExecHandler(cfg, deps, log); err != nil {
		return nil, nil, err
	} else if h != nil {
		handlers = append(handlers, h)
	}
	if h, err := newFlowHandler(cfg, deps, log); err != nil {
		return nil, nil, err
	} else if h != nil {
		handlers = append(handlers, h)
	}
	if h, err := newFileHandler(cfg, deps, log); err != nil {
		return nil, nil, err
	} else if h != nil {
		handlers = append(handlers, h)
	}
	if h, err := newStreamHandler(cfg, log); err != nil {
		return nil, nil, err
	} else if h != nil {
		handlers = append(handlers, h)
	}

	def = forward.NewDefaultForwardHandler(cfg, log)
	return handlers, def, nil
}
