package modules

import (
	"context"
	"testing"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-core/config"
	"github.com/yttydcs/myflowhub-core/eventbus"
)

type stubHandler struct {
	sub uint8
}

func (h *stubHandler) SubProto() uint8 { return h.sub }
func (h *stubHandler) OnReceive(context.Context, core.IConnection, core.IHeader, []byte) {
}
func (h *stubHandler) Init() bool                { return true }
func (h *stubHandler) AcceptCmd() bool           { return false }
func (h *stubHandler) AllowSourceMismatch() bool { return false }

type bindHandler struct {
	stubHandler
	calls int
	last  core.IServer
}

func (h *bindHandler) BindServer(srv core.IServer) {
	h.calls++
	h.last = srv
}

type dummyServer struct{}

func (s dummyServer) Start(context.Context) error { return nil }
func (s dummyServer) Stop(context.Context) error  { return nil }
func (s dummyServer) Config() core.IConfig        { return nil }
func (s dummyServer) ConnManager() core.IConnectionManager {
	return nil
}
func (s dummyServer) Process() core.IProcess         { return nil }
func (s dummyServer) HeaderCodec() core.IHeaderCodec { return nil }
func (s dummyServer) NodeID() uint32                 { return 0 }
func (s dummyServer) UpdateNodeID(uint32)            {}
func (s dummyServer) EventBus() eventbus.IBus        { return nil }
func (s dummyServer) Send(context.Context, string, core.IHeader, []byte) error {
	return nil
}

func TestDefaultHub_UniqueSubProtoAndNotEmpty(t *testing.T) {
	cfg := config.NewMap(map[string]string{})
	set, err := DefaultHub(cfg, nil)
	if err != nil {
		t.Fatalf("DefaultHub() err: %v", err)
	}
	if len(set.Handlers) == 0 {
		t.Fatalf("DefaultHub() handlers empty")
	}
	if set.Default == nil {
		t.Fatalf("DefaultHub() default nil")
	}
	seen := map[uint8]struct{}{}
	for _, h := range set.Handlers {
		if h == nil {
			t.Fatalf("DefaultHub() handler nil")
		}
		sub := h.SubProto()
		if _, ok := seen[sub]; ok {
			t.Fatalf("duplicate subproto: %d", sub)
		}
		seen[sub] = struct{}{}
	}
}

func TestBindServerHooks_OnlyBindableCalled(t *testing.T) {
	bind := &bindHandler{stubHandler: stubHandler{sub: 2}}
	set := Set{
		Handlers: []core.ISubProcess{
			&stubHandler{sub: 1},
			bind,
		},
		Default: &stubHandler{sub: 0},
	}

	srv := dummyServer{}
	BindServerHooks(srv, set)

	if bind.calls != 1 {
		t.Fatalf("BindServerHooks() calls=%d, want=1", bind.calls)
	}
	if bind.last != srv {
		t.Fatalf("BindServerHooks() last != srv")
	}
}
