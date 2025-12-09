package auth

import (
	"context"
	"encoding/json"

	core "github.com/yttydcs/myflowhub-core"
)

type loginAction struct {
	h        *LoginHandler
	assisted bool
}

func (a *loginAction) Name() string {
	if a.assisted {
		return actionAssistLogin
	}
	return actionLogin
}
func (a *loginAction) RequireAuth() bool { return false }
func (a *loginAction) Handle(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage) {
	var req loginData
	if err := json.Unmarshal(data, &req); err != nil || req.DeviceID == "" {
		a.h.sendResp(ctx, conn, hdr, actionLoginResp, respData{Code: 400, Msg: "invalid login data"})
		return
	}
	if a.assisted {
		rec, ok := a.h.lookup(req.DeviceID)
		if !ok || rec.Credential != req.Credential {
			a.h.sendResp(ctx, conn, hdr, actionAssistLoginResp, respData{Code: 4001, Msg: "invalid credential"})
			return
		}
		a.h.addRouteIndex(ctx, rec.NodeID, conn)
		a.h.sendResp(ctx, conn, hdr, actionAssistLoginResp, respData{
			Code:       1,
			Msg:        "ok",
			DeviceID:   req.DeviceID,
			NodeID:     rec.NodeID,
			HubID:      localNodeID(ctx),
			Credential: rec.Credential,
		})
		return
	}
	// local check
	if rec, ok := a.h.lookup(req.DeviceID); ok {
		if rec.Credential == req.Credential {
			a.h.saveBinding(ctx, conn, req.DeviceID, rec.NodeID, rec.Credential)
			a.h.applyHubID(ctx, conn, localNodeID(ctx))
			a.h.sendResp(ctx, conn, hdr, actionLoginResp, respData{Code: 1, Msg: "ok", DeviceID: req.DeviceID, NodeID: rec.NodeID, HubID: localNodeID(ctx), Credential: rec.Credential})
			return
		}
		a.h.sendResp(ctx, conn, hdr, actionLoginResp, respData{Code: 4001, Msg: "invalid credential"})
		return
	}
	// not found locally, try authority
	authority := a.h.selectAuthority(ctx)
	if authority != nil {
		a.h.setPending(req.DeviceID, conn.ID())
		a.h.forward(ctx, authority, actionAssistLogin, req)
		return
	}
	a.h.sendResp(ctx, conn, hdr, actionLoginResp, respData{Code: 4001, Msg: "invalid credential"})
}

type loginRespAction struct{ h *LoginHandler }

func (a *loginRespAction) Name() string      { return actionLoginResp }
func (a *loginRespAction) RequireAuth() bool { return false }
func (a *loginRespAction) Handle(ctx context.Context, _ core.IConnection, _ core.IHeader, data json.RawMessage) {
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
			a.h.applyHubID(ctx, c, resp.HubID)
		}
		if resp.HubID == 0 {
			resp.HubID = srv.NodeID()
		}
		a.h.sendResp(ctx, c, nil, actionLoginResp, resp)
	}
}

func registerLoginActions(h *LoginHandler) []SubProcessAction {
	return []SubProcessAction{
		&loginAction{h: h, assisted: false},
		&loginAction{h: h, assisted: true},
		&loginRespAction{h: h},
	}
}
