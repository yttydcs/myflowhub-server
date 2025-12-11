package auth

import (
	"context"
	"encoding/json"
	"strings"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-core/subproto"
)

type assistQueryAction struct {
	subproto.BaseAction
	h *LoginHandler
}

func (a *assistQueryAction) Name() string      { return actionAssistQueryCred }
func (a *assistQueryAction) RequireAuth() bool { return true }
func (a *assistQueryAction) Handle(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage) {
	var req queryCredData
	if err := json.Unmarshal(data, &req); err != nil || req.DeviceID == "" {
		a.h.sendResp(ctx, conn, hdr, actionAssistQueryCredResp, respData{Code: 400, Msg: "invalid query"})
		return
	}
	if rec, ok := a.h.lookup(req.DeviceID); ok {
		nodePub := ""
		nodePub = encodePubKey(rec.PubKey)
		a.h.sendResp(ctx, conn, hdr, actionAssistQueryCredResp, respData{
			Code:     1,
			Msg:      "ok",
			DeviceID: req.DeviceID,
			NodeID:   rec.NodeID,
			Role:     rec.Role,
			Perms:    rec.Perms,
			PubKey:   encodePubKey(rec.PubKey),
			NodePub:  nodePub,
		})
		return
	}
	a.h.sendResp(ctx, conn, hdr, actionAssistQueryCredResp, respData{Code: 4001, Msg: "not found"})
}

type assistQueryRespAction struct {
	subproto.BaseAction
	h *LoginHandler
}

func (a *assistQueryRespAction) Name() string { return actionAssistQueryCredResp }
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
			var pubRaw []byte
			if pk := strings.TrimSpace(resp.PubKey); pk != "" {
				if _, raw, err := parseECPubKey(pk); err == nil {
					pubRaw = raw
				}
			}
			a.h.saveBinding(ctx, c, resp.DeviceID, resp.NodeID, pubRaw)
			a.h.applyRolePerms(resp.DeviceID, resp.NodeID, resp.Role, resp.Perms, c)
			a.h.addRouteIndex(ctx, resp.NodeID, c)
			if strings.TrimSpace(resp.PubKey) != "" {
				a.h.addTrustedNode(resp.NodeID, resp.PubKey)
			}
			a.h.sendResp(ctx, c, nil, actionLoginResp, respData{Code: 1, Msg: "ok", DeviceID: resp.DeviceID, NodeID: resp.NodeID, Role: resp.Role, Perms: resp.Perms, PubKey: encodePubKey(pubRaw), NodePub: resp.PubKey})
			return
		}
		a.h.sendResp(ctx, c, nil, actionLoginResp, respData{Code: resp.Code, Msg: resp.Msg})
	}
}

func registerAssistQueryActions(h *LoginHandler) []core.SubProcessAction {
	return []core.SubProcessAction{
		&assistQueryAction{h: h},
		&assistQueryRespAction{h: h},
	}
}
