package auth

import (
	"context"
	"encoding/json"

	core "github.com/yttydcs/myflowhub-core"
)

type assistQueryAction struct{ h *LoginHandler }

func (a *assistQueryAction) Name() string      { return actionAssistQueryCred }
func (a *assistQueryAction) RequireAuth() bool { return true }
func (a *assistQueryAction) Handle(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage) {
	var req queryCredData
	if err := json.Unmarshal(data, &req); err != nil || req.DeviceID == "" {
		a.h.sendResp(ctx, conn, hdr, actionAssistQueryCredResp, respData{Code: 400, Msg: "invalid query"})
		return
	}
	if rec, ok := a.h.lookup(req.DeviceID); ok {
		a.h.sendResp(ctx, conn, hdr, actionAssistQueryCredResp, respData{Code: 1, Msg: "ok", DeviceID: req.DeviceID, NodeID: rec.NodeID, Credential: rec.Credential, Role: rec.Role, Perms: rec.Perms})
		return
	}
	a.h.sendResp(ctx, conn, hdr, actionAssistQueryCredResp, respData{Code: 4001, Msg: "not found"})
}

type assistQueryRespAction struct{ h *LoginHandler }

func (a *assistQueryRespAction) Name() string      { return actionAssistQueryCredResp }
func (a *assistQueryRespAction) RequireAuth() bool { return false }
func (a *assistQueryRespAction) Handle(ctx context.Context, _ core.IConnection, _ core.IHeader, data json.RawMessage) {
	var resp respData
	if err := json.Unmarshal(data, &resp); err != nil {
		return
	}
	if resp.DeviceID == "" {
		return
	}
	connID, ok := a.h.popPending(resp.DeviceID)
	if !ok {
		return
	}
	srv := core.ServerFromContext(ctx)
	if srv == nil {
		return
	}
	if c, found := srv.ConnManager().Get(connID); found {
		if resp.Code == 1 {
			a.h.saveBinding(ctx, c, resp.DeviceID, resp.NodeID, resp.Credential)
			a.h.applyRolePerms(resp.DeviceID, resp.NodeID, resp.Role, resp.Perms, c)
			a.h.addRouteIndex(ctx, resp.NodeID, c)
			a.h.sendResp(ctx, c, nil, actionLoginResp, respData{Code: 1, Msg: "ok", DeviceID: resp.DeviceID, NodeID: resp.NodeID, Credential: resp.Credential, Role: resp.Role, Perms: resp.Perms})
			return
		}
		a.h.sendResp(ctx, c, nil, actionLoginResp, respData{Code: resp.Code, Msg: resp.Msg})
	}
}

func registerAssistQueryActions(h *LoginHandler) []SubProcessAction {
	return []SubProcessAction{
		&assistQueryAction{h: h},
		&assistQueryRespAction{h: h},
	}
}
