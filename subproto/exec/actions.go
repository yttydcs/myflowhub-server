package exec

import (
	"context"
	"encoding/json"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-core/subproto"
)

type execAction struct {
	subproto.BaseAction
	name string
	fn   func(context.Context, core.IConnection, core.IHeader, json.RawMessage)
}

func (a execAction) Name() string { return a.name }

func (a execAction) Handle(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage) {
	a.fn(ctx, conn, hdr, data)
}

func registerActions(h *Handler) []core.SubProcessAction {
	return []core.SubProcessAction{
		execAction{name: actionCall, fn: h.handleCall},
		execAction{name: actionCallResp, fn: h.handleCallResp},
	}
}
