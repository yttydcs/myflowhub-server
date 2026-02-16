package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-core/subproto"
)

type loginAction struct {
	subproto.BaseAction
	h        *LoginHandler
	assisted bool
}

func (a *loginAction) Name() string {
	if a.assisted {
		return actionAssistLogin
	}
	return actionLogin
}
func (a *loginAction) Handle(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage) {
	send := a.h.sendDirectResp
	if a.assisted {
		send = a.h.sendResp
	}
	var req loginData
	if err := json.Unmarshal(data, &req); err != nil || req.DeviceID == "" {
		send(ctx, conn, hdr, actionLoginResp, respData{Code: 400, Msg: "invalid login data"})
		return
	}
	if a.assisted {
		rec, ok := a.h.lookup(req.DeviceID)
		if (!ok || len(rec.PubKey) == 0) && a.h.selectAuthority(ctx) != nil {
			// 向上查询公钥
			a.h.setPending(req.DeviceID, conn.ID(), hdr)
			a.h.forward(ctx, a.h.selectAuthority(ctx), actionAssistQueryCred, queryCredData{DeviceID: req.DeviceID, NodeID: req.NodeID})
			return
		}
		valid := false
		if ok && len(rec.PubKey) > 0 && strings.EqualFold(strings.TrimSpace(req.Alg), defaultAlgES256) && strings.TrimSpace(req.Sig) != "" {
			if pub, err := parseECPubKeyRaw(rec.PubKey); err == nil {
				valid = verifyEcdsaSig(pub, loginSignBytes(req), req.Sig)
			}
		}
		if !ok || !valid {
			a.h.sendResp(ctx, conn, hdr, actionAssistLoginResp, respData{Code: 4001, Msg: "invalid signature"})
			return
		}
		if len(rec.PubKey) > 0 {
			conn.SetMeta("pubkey", rec.PubKey)
		}
		a.h.addRouteIndex(ctx, rec.NodeID, conn)
		a.h.sendResp(ctx, conn, hdr, actionAssistLoginResp, respData{
			Code:     1,
			Msg:      "ok",
			DeviceID: req.DeviceID,
			NodeID:   rec.NodeID,
			HubID:    localNodeID(ctx),
			PubKey:   base64.StdEncoding.EncodeToString(rec.PubKey),
			NodePub:  base64.StdEncoding.EncodeToString(rec.PubKey),
		})
		go a.h.sendUpLogin(ctx, conn, req.DeviceID, rec.NodeID, rec.PubKey, req.Sig, req.Alg, req.TS, req.Nonce)
		return
	}
	// local check
	if rec, ok := a.h.lookup(req.DeviceID); ok {
		if len(rec.PubKey) == 0 && a.h.selectAuthority(ctx) != nil {
			a.h.setPending(req.DeviceID, conn.ID(), hdr)
			a.h.forward(ctx, a.h.selectAuthority(ctx), actionAssistQueryCred, queryCredData{DeviceID: req.DeviceID, NodeID: req.NodeID})
			return
		}
		valid := false
		if len(rec.PubKey) > 0 && strings.EqualFold(strings.TrimSpace(req.Alg), defaultAlgES256) && strings.TrimSpace(req.Sig) != "" {
			if pub, err := parseECPubKeyRaw(rec.PubKey); err == nil {
				valid = verifyEcdsaSig(pub, loginSignBytes(req), req.Sig)
			}
		}
		if valid {
			a.h.saveBinding(ctx, conn, req.DeviceID, rec.NodeID, rec.PubKey)
			a.h.applyHubID(ctx, conn, localNodeID(ctx))
			send(ctx, conn, hdr, actionLoginResp, respData{Code: 1, Msg: "ok", DeviceID: req.DeviceID, NodeID: rec.NodeID, HubID: localNodeID(ctx), PubKey: base64.StdEncoding.EncodeToString(rec.PubKey), NodePub: base64.StdEncoding.EncodeToString(rec.PubKey)})
			go a.h.sendUpLogin(ctx, conn, req.DeviceID, rec.NodeID, rec.PubKey, req.Sig, req.Alg, req.TS, req.Nonce)
			return
		}
		send(ctx, conn, hdr, actionLoginResp, respData{Code: 4001, Msg: "invalid signature"})
		return
	}
	// not found locally, try authority
	authority := a.h.selectAuthority(ctx)
	if authority != nil {
		a.h.setPending(req.DeviceID, conn.ID(), hdr)
		a.h.forward(ctx, authority, actionAssistLogin, req)
		return
	}
	send(ctx, conn, hdr, actionLoginResp, respData{Code: 4001, Msg: "invalid signature"})
}

type loginRespAction struct {
	subproto.BaseAction
	h *LoginHandler
}

func (a *loginRespAction) Name() string { return actionLoginResp }
func (a *loginRespAction) Handle(ctx context.Context, _ core.IConnection, _ core.IHeader, data json.RawMessage) {
	a.h.handleLoginResp(ctx, data)
}

type assistLoginRespAction struct {
	subproto.BaseAction
	h *LoginHandler
}

func (a *assistLoginRespAction) Name() string { return actionAssistLoginResp }
func (a *assistLoginRespAction) Handle(ctx context.Context, _ core.IConnection, _ core.IHeader, data json.RawMessage) {
	a.h.handleLoginResp(ctx, data)
}

func (h *LoginHandler) handleLoginResp(ctx context.Context, data json.RawMessage) {
	var resp respData
	if err := json.Unmarshal(data, &resp); err != nil {
		return
	}
	if resp.DeviceID == "" {
		return
	}
	pending, ok := h.popPending(resp.DeviceID)
	if !ok {
		return
	}
	srv := core.ServerFromContext(ctx)
	if srv == nil {
		return
	}
	if c, found := srv.ConnManager().Get(pending.connID); found {
		if resp.Code == 1 {
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
			// 此分支没有原始 device 签名，避免上行；由实际验证节点负责上报
		}
		if resp.HubID == 0 {
			resp.HubID = srv.NodeID()
		}
		h.sendResp(ctx, c, h.buildPendingRespHeader(ctx, pending), actionLoginResp, resp)
	}
}

func registerLoginActions(h *LoginHandler) []core.SubProcessAction {
	return []core.SubProcessAction{
		&loginAction{h: h, assisted: false},
		&loginAction{h: h, assisted: true},
		&loginRespAction{h: h},
		&assistLoginRespAction{h: h},
	}
}
