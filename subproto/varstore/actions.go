package varstore

import (
	"context"
	"encoding/json"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-core/subproto"
)

type varAction struct {
	subproto.BaseAction
	name string
	fn   func(context.Context, core.IConnection, core.IHeader, json.RawMessage)
}

func (a varAction) Name() string { return a.name }
func (a varAction) Handle(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage) {
	a.fn(ctx, conn, hdr, data)
}

func registerVarActions(h *VarStoreHandler) []core.SubProcessAction {
	return []core.SubProcessAction{
		varAction{name: varActionSet, fn: func(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage) {
			h.handleSet(ctx, conn, hdr, data, false)
		}},
		varAction{name: varActionAssistSet, fn: func(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage) {
			h.handleSet(ctx, conn, hdr, data, true)
		}},
		varAction{name: varActionSetResp, fn: func(ctx context.Context, _ core.IConnection, _ core.IHeader, data json.RawMessage) {
			h.handleSetResp(ctx, data)
		}},
		varAction{name: varActionAssistSetResp, fn: func(ctx context.Context, _ core.IConnection, _ core.IHeader, data json.RawMessage) {
			h.handleSetResp(ctx, data)
		}},
		varAction{name: varActionUpSet, fn: func(ctx context.Context, _ core.IConnection, hdr core.IHeader, data json.RawMessage) {
			h.handleUpSet(ctx, hdr, data)
		}},
		varAction{name: varActionNotifySet, fn: func(ctx context.Context, _ core.IConnection, hdr core.IHeader, data json.RawMessage) {
			h.handleNotifySet(ctx, hdr, data)
		}},

		varAction{name: varActionGet, fn: func(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage) {
			h.handleGet(ctx, conn, hdr, data, false)
		}},
		varAction{name: varActionAssistGet, fn: func(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage) {
			h.handleGet(ctx, conn, hdr, data, true)
		}},
		varAction{name: varActionGetResp, fn: func(ctx context.Context, _ core.IConnection, _ core.IHeader, data json.RawMessage) {
			h.handleGetResp(ctx, data)
		}},
		varAction{name: varActionAssistGetResp, fn: func(ctx context.Context, _ core.IConnection, _ core.IHeader, data json.RawMessage) {
			h.handleGetResp(ctx, data)
		}},

		varAction{name: varActionList, fn: func(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage) {
			h.handleList(ctx, conn, hdr, data, false)
		}},
		varAction{name: varActionAssistList, fn: func(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage) {
			h.handleList(ctx, conn, hdr, data, true)
		}},
		varAction{name: varActionListResp, fn: func(ctx context.Context, _ core.IConnection, _ core.IHeader, data json.RawMessage) {
			h.handleListResp(ctx, data)
		}},
		varAction{name: varActionAssistListResp, fn: func(ctx context.Context, _ core.IConnection, _ core.IHeader, data json.RawMessage) {
			h.handleListResp(ctx, data)
		}},

		varAction{name: varActionRevoke, fn: func(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage) {
			h.handleRevoke(ctx, conn, hdr, data, false)
		}},
		varAction{name: varActionAssistRevoke, fn: func(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage) {
			h.handleRevoke(ctx, conn, hdr, data, true)
		}},
		varAction{name: varActionRevokeResp, fn: func(ctx context.Context, _ core.IConnection, _ core.IHeader, data json.RawMessage) {
			h.handleRevokeResp(ctx, data)
		}},
		varAction{name: varActionAssistRevokeResp, fn: func(ctx context.Context, _ core.IConnection, _ core.IHeader, data json.RawMessage) {
			h.handleRevokeResp(ctx, data)
		}},
		varAction{name: varActionUpRevoke, fn: func(ctx context.Context, _ core.IConnection, hdr core.IHeader, data json.RawMessage) {
			h.handleUpRevoke(ctx, hdr, data)
		}},
		varAction{name: varActionNotifyRevoke, fn: func(ctx context.Context, _ core.IConnection, hdr core.IHeader, data json.RawMessage) {
			h.handleNotifyRevoke(ctx, hdr, data)
		}},

		varAction{name: varActionSubscribe, fn: func(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage) {
			h.handleSubscribe(ctx, conn, hdr, data, false)
		}},
		varAction{name: varActionAssistSubscribe, fn: func(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage) {
			h.handleSubscribe(ctx, conn, hdr, data, true)
		}},
		varAction{name: varActionSubscribeResp, fn: func(ctx context.Context, _ core.IConnection, hdr core.IHeader, data json.RawMessage) {
			h.handleSubscribeResp(ctx, hdr, data)
		}},
		varAction{name: varActionAssistSubscribeResp, fn: func(ctx context.Context, _ core.IConnection, hdr core.IHeader, data json.RawMessage) {
			h.handleSubscribeResp(ctx, hdr, data)
		}},
		varAction{name: varActionUnsubscribe, fn: func(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage) {
			h.handleUnsubscribe(ctx, conn, hdr, data, false)
		}},
		varAction{name: varActionAssistUnsubscribe, fn: func(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage) {
			h.handleUnsubscribe(ctx, conn, hdr, data, true)
		}},

		varAction{name: varActionVarChanged, fn: func(ctx context.Context, _ core.IConnection, hdr core.IHeader, data json.RawMessage) {
			h.handleVarChanged(ctx, hdr, data)
		}},
		varAction{name: varActionVarDeleted, fn: func(ctx context.Context, _ core.IConnection, hdr core.IHeader, data json.RawMessage) {
			h.handleVarDeleted(ctx, hdr, data)
		}},
	}
}
