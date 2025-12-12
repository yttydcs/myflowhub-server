package auth

import (
	"context"
	"encoding/json"
	"strings"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-core/subproto"
)

type registerAction struct {
	subproto.BaseAction
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

type registerRespAction struct {
	subproto.BaseAction
	h *LoginHandler
}

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
		var pubRaw []byte
		if pk := strings.TrimSpace(resp.PubKey); pk != "" {
			if _, raw, err := parseECPubKey(pk); err == nil {
				pubRaw = raw
			}
		}
		a.h.saveBinding(ctx, c, resp.DeviceID, resp.NodeID, pubRaw)
		a.h.applyRolePerms(resp.DeviceID, resp.NodeID, resp.Role, resp.Perms, c)
		a.h.applyHubID(ctx, c, resp.HubID)
		if strings.TrimSpace(resp.NodePub) != "" {
			a.h.addTrustedNode(resp.NodeID, resp.NodePub)
		}
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
	if strings.TrimSpace(req.PubKey) == "" && strings.TrimSpace(h.nodePubB64) != "" {
		req.PubKey = h.nodePubB64
	}
	req.NodePub = req.PubKey
	var pubRaw []byte
	if strings.TrimSpace(req.PubKey) != "" {
		if _, raw, err := parseECPubKey(req.PubKey); err != nil {
			h.sendResp(ctx, conn, hdr, actionRegisterResp, respData{Code: 400, Msg: "invalid pubkey"})
			return
		} else {
			pubRaw = raw
		}
	}
	if assisted {
		// being processed at authority
		nodeID := h.ensureNodeID(req.DeviceID)
		h.saveBinding(ctx, conn, req.DeviceID, nodeID, pubRaw)
		h.addRouteIndex(ctx, nodeID, conn)
		if strings.TrimSpace(req.PubKey) != "" {
			h.addTrustedNode(nodeID, req.PubKey)
		}
		h.sendResp(ctx, conn, hdr, actionAssistRegisterResp, respData{
			Code:     1,
			Msg:      "ok",
			DeviceID: req.DeviceID,
			NodeID:   nodeID,
			Role:     h.resolveRole(nodeID),
			Perms:    h.resolvePerms(nodeID),
			PubKey:   req.PubKey,
			NodePub:  req.PubKey,
		})
		h.persistState()
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
	h.saveBinding(ctx, conn, req.DeviceID, nodeID, pubRaw)
	h.applyHubID(ctx, conn, localNodeID(ctx))
	if strings.TrimSpace(req.PubKey) != "" {
		h.addTrustedNode(nodeID, req.PubKey)
	}
	h.sendResp(ctx, conn, hdr, actionRegisterResp, respData{
		Code:     1,
		Msg:      "ok",
		DeviceID: req.DeviceID,
		NodeID:   nodeID,
		HubID:    localNodeID(ctx),
		Role:     h.resolveRole(nodeID),
		Perms:    h.resolvePerms(nodeID),
		PubKey:   req.PubKey,
		NodePub:  req.PubKey,
	})
	h.persistState()
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
		var pubRaw []byte
		if pk := strings.TrimSpace(resp.PubKey); pk != "" {
			if _, raw, err := parseECPubKey(pk); err == nil {
				pubRaw = raw
			}
		}
		h.saveBinding(ctx, c, resp.DeviceID, resp.NodeID, pubRaw)
		h.applyRolePerms(resp.DeviceID, resp.NodeID, resp.Role, resp.Perms, c)
		h.applyHubID(ctx, c, resp.HubID)
		if strings.TrimSpace(resp.PubKey) != "" {
			h.addTrustedNode(resp.NodeID, resp.PubKey)
		}
		if resp.HubID == 0 {
			resp.HubID = srv.NodeID()
		}
		h.sendResp(ctx, c, nil, actionRegisterResp, resp)
		h.persistState()
	}
}
