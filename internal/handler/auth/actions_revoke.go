package auth

import (
	"context"
	"encoding/json"

	core "github.com/yttydcs/myflowhub-core"
	permission "github.com/yttydcs/myflowhub-core/kit/permission"
	"github.com/yttydcs/myflowhub-core/subproto"
)

type revokeAction struct {
	subproto.BaseAction
	h *LoginHandler
}

func (a *revokeAction) Name() string      { return actionRevoke }
func (a *revokeAction) RequireAuth() bool { return true }
func (a *revokeAction) Handle(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage) {
	a.h.handleRevoke(ctx, conn, hdr, data)
}

func registerRevokeActions(h *LoginHandler) []core.SubProcessAction {
	return []core.SubProcessAction{&revokeAction{h: h}}
}

// revoke handling: broadcast; respond only if deleted or credential mismatch
func (h *LoginHandler) handleRevoke(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage) {
	var req revokeData
	if err := json.Unmarshal(data, &req); err != nil || req.DeviceID == "" {
		return
	}
	actorID := permission.SourceNodeID(hdr, conn)
	if !h.hasPermission(actorID, permission.AuthRevoke) {
		h.sendResp(ctx, conn, hdr, actionRevokeResp, respData{Code: 4403, Msg: "permission denied", DeviceID: req.DeviceID, NodeID: req.NodeID})
		return
	}
	removed, mismatch := h.removeBinding(req.DeviceID, req.Credential)
	if removed || mismatch {
		// respond only when changed/mismatch
		if mismatch {
			h.sendResp(ctx, conn, hdr, actionRevokeResp, respData{Code: 4402, Msg: "credential mismatch", DeviceID: req.DeviceID, NodeID: req.NodeID})
		} else {
			h.sendResp(ctx, conn, hdr, actionRevokeResp, respData{Code: 1, Msg: "ok", DeviceID: req.DeviceID, NodeID: req.NodeID})
		}
	}
	// broadcast downstream and upstream except source
	h.broadcast(ctx, conn, actionRevoke, req)
}
