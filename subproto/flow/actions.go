package flow

import (
	"context"
	"encoding/json"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-core/subproto"
)

type flowAction struct {
	subproto.BaseAction
	name string
	fn   func(context.Context, core.IConnection, core.IHeader, json.RawMessage)
}

func (a flowAction) Name() string { return a.name }

func (a flowAction) Handle(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage) {
	a.fn(ctx, conn, hdr, data)
}

func registerActions(h *Handler) []core.SubProcessAction {
	return []core.SubProcessAction{
		flowAction{name: actionSet, fn: h.handleSet},
		flowAction{name: actionRun, fn: h.handleRun},
		flowAction{name: actionStatus, fn: h.handleStatus},
		flowAction{name: actionList, fn: h.handleList},
		flowAction{name: actionGet, fn: h.handleGet},
	}
}
