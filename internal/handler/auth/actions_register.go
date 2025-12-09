package auth

import (
	"context"
	"encoding/json"

	core "github.com/yttydcs/myflowhub-core"
)

type registerAction struct {
	h        *LoginHandler
	assisted bool
}

func (a *registerAction) Name() string {
	if a.assisted {
		return actionAssistRegister
	}
	return actionRegister
}
func (a *registerAction) RequireAuth() bool { return false }
func (a *registerAction) Handle(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage) {
	a.h.handleRegister(ctx, conn, hdr, data, a.assisted)
}

type registerRespAction struct{ h *LoginHandler }

func (a *registerRespAction) Name() string      { return actionRegisterResp }
func (a *registerRespAction) RequireAuth() bool { return false }
func (a *registerRespAction) Handle(ctx context.Context, _ core.IConnection, _ core.IHeader, data json.RawMessage) {
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
		a.h.saveBinding(ctx, c, resp.DeviceID, resp.NodeID, resp.Credential)
		a.h.applyRolePerms(resp.DeviceID, resp.NodeID, resp.Role, resp.Perms, c)
		a.h.applyHubID(ctx, c, resp.HubID)
		if resp.HubID == 0 {
			resp.HubID = srv.NodeID()
		}
		a.h.sendResp(ctx, c, nil, actionRegisterResp, resp)
	}
}

func registerRegisterActions(h *LoginHandler) []core.SubProcessAction {
	return []core.SubProcessAction{
		&registerAction{h: h, assisted: false},
		&registerAction{h: h, assisted: true},
		&registerRespAction{h: h},
	}
}

func (h *LoginHandler) handleRegister(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage, assisted bool) {
	var req registerData
	if err := json.Unmarshal(data, &req); err != nil || req.DeviceID == "" {
		h.sendResp(ctx, conn, hdr, actionRegisterResp, respData{Code: 400, Msg: "invalid register data"})
		return
	}
	if assisted {
		// being processed at authority
		nodeID := h.ensureNodeID(req.DeviceID)
		cred := h.ensureCredential(req.DeviceID)
		h.addRouteIndex(ctx, nodeID, conn)
		h.sendResp(ctx, conn, hdr, actionAssistRegisterResp, respData{
			Code:       1,
			Msg:        "ok",
			DeviceID:   req.DeviceID,
			NodeID:     nodeID,
			Credential: cred,
			Role:       h.resolveRole(nodeID),
			Perms:      h.resolvePerms(nodeID),
		})
		return
	}
	authority := h.selectAuthority(ctx)
	if authority != nil {
		h.setPending(req.DeviceID, conn.ID())
		h.forward(ctx, authority, actionAssistRegister, req)
		return
	}
	// self authority
	nodeID := h.ensureNodeID(req.DeviceID)
	cred := h.ensureCredential(req.DeviceID)
	h.saveBinding(ctx, conn, req.DeviceID, nodeID, cred)
	h.applyHubID(ctx, conn, localNodeID(ctx))
	h.sendResp(ctx, conn, hdr, actionRegisterResp, respData{
		Code:       1,
		Msg:        "ok",
		DeviceID:   req.DeviceID,
		NodeID:     nodeID,
		HubID:      localNodeID(ctx),
		Credential: cred,
		Role:       h.resolveRole(nodeID),
		Perms:      h.resolvePerms(nodeID),
	})
}

func (h *LoginHandler) handleRegisterResp(ctx context.Context, data json.RawMessage) {
	var resp respData
	if err := json.Unmarshal(data, &resp); err != nil {
		return
	}
	if resp.DeviceID == "" {
		return
	}
	connID, ok := h.popPending(resp.DeviceID)
	if !ok {
		return
	}
	srv := core.ServerFromContext(ctx)
	if srv == nil {
		return
	}
	if c, found := srv.ConnManager().Get(connID); found {
		h.saveBinding(ctx, c, resp.DeviceID, resp.NodeID, resp.Credential)
		h.applyRolePerms(resp.DeviceID, resp.NodeID, resp.Role, resp.Perms, c)
		h.applyHubID(ctx, c, resp.HubID)
		if resp.HubID == 0 {
			resp.HubID = srv.NodeID()
		}
		h.sendResp(ctx, c, nil, actionRegisterResp, resp)
	}
}
