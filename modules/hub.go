package modules

import (
	"errors"
	"fmt"
	"log/slog"

	core "github.com/yttydcs/myflowhub-core"
	authhandler "github.com/yttydcs/myflowhub-server/internal/handler/auth"
	filehandler "github.com/yttydcs/myflowhub-server/internal/handler/file"
	exechandler "github.com/yttydcs/myflowhub-server/subproto/exec"
	flowhandler "github.com/yttydcs/myflowhub-server/subproto/flow"
	"github.com/yttydcs/myflowhub-server/subproto/forward"
	"github.com/yttydcs/myflowhub-server/subproto/management"
	"github.com/yttydcs/myflowhub-server/subproto/topicbus"
	varstore "github.com/yttydcs/myflowhub-server/subproto/varstore"
)

// Dispatcher 抽象 hub_server 装配所需的最小 dispatcher 能力。
type Dispatcher interface {
	RegisterHandler(core.ISubProcess) error
	RegisterDefaultHandler(core.ISubProcess)
}

// Set 表示一组可注册的子协议 handler 集合（以及默认 fallback）。
// 注意：Set 仅负责装配与校验，不触发 handler.Init（由 Dispatcher 调用 RegisterHandler 时触发）。
type Set struct {
	Handlers []core.ISubProcess
	Default  core.ISubProcess
}

// DefaultHub 返回 hub_server 的默认启用模块集合。
func DefaultHub(cfg core.IConfig, log *slog.Logger) (Set, error) {
	set := Set{
		Handlers: []core.ISubProcess{
			management.NewHandler(log),
			authhandler.NewLoginHandlerWithConfig(cfg, log),
			varstore.NewVarStoreHandlerWithConfig(cfg, log),
			topicbus.NewTopicBusHandlerWithConfig(cfg, log),
			exechandler.NewHandlerWithConfig(cfg, log),
			flowhandler.NewHandlerWithConfig(cfg, log),
			filehandler.NewHandlerWithConfig(cfg, log),
		},
		Default: forward.NewDefaultForwardHandler(cfg, log),
	}
	if err := validateSet(set); err != nil {
		return Set{}, err
	}
	return set, nil
}

// RegisterAll 将 Set 注册到 dispatcher。
func RegisterAll(dispatcher Dispatcher, set Set) error {
	if dispatcher == nil {
		return errors.New("dispatcher nil")
	}
	if err := validateSet(set); err != nil {
		return err
	}
	for _, h := range set.Handlers {
		if err := dispatcher.RegisterHandler(h); err != nil {
			return err
		}
	}
	dispatcher.RegisterDefaultHandler(set.Default)
	return nil
}

type serverBinder interface {
	BindServer(core.IServer)
}

// BindServerHooks 对实现了 BindServer(core.IServer) 的 handler 执行启动后绑定。
// 该 hook 仅在启动期调用，不引入运行期热路径开销。
func BindServerHooks(srv core.IServer, set Set) {
	if srv == nil {
		return
	}
	for _, h := range set.Handlers {
		if b, ok := h.(serverBinder); ok {
			b.BindServer(srv)
		}
	}
}

func validateSet(set Set) error {
	if len(set.Handlers) == 0 {
		return errors.New("handlers empty")
	}
	if set.Default == nil {
		return errors.New("default handler nil")
	}
	seen := make(map[uint8]core.ISubProcess, len(set.Handlers))
	for i, h := range set.Handlers {
		if h == nil {
			return fmt.Errorf("handler[%d] nil", i)
		}
		sub := h.SubProto()
		if prev, ok := seen[sub]; ok {
			return fmt.Errorf("duplicate subproto %d between %T and %T", sub, prev, h)
		}
		seen[sub] = h
	}
	return nil
}
