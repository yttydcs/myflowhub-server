package auth

import (
	"context"
	"encoding/json"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-core/header"
	permission "github.com/yttydcs/myflowhub-core/kit/permission"
)

type getPermsAction struct{ h *LoginHandler }

func (a *getPermsAction) Name() string      { return actionGetPerms }
func (a *getPermsAction) RequireAuth() bool { return true }
func (a *getPermsAction) Handle(ctx context.Context, conn core.IConnection, _ core.IHeader, data json.RawMessage) {
	var req permsQueryData
	if err := json.Unmarshal(data, &req); err != nil || req.NodeID == 0 {
		a.h.sendResp(ctx, conn, nil, actionGetPermsResp, respData{Code: 400, Msg: "invalid node id"})
		return
	}
	role, perms, ok := a.h.lookupByNode(req.NodeID)
	if !ok {
		a.h.sendResp(ctx, conn, nil, actionGetPermsResp, respData{Code: 4404, Msg: "not found", NodeID: req.NodeID})
		return
	}
	a.h.sendResp(ctx, conn, nil, actionGetPermsResp, respData{Code: 1, Msg: "ok", NodeID: req.NodeID, Role: role, Perms: perms})
}

type listRolesAction struct{ h *LoginHandler }

func (a *listRolesAction) Name() string      { return actionListRoles }
func (a *listRolesAction) RequireAuth() bool { return true }
func (a *listRolesAction) Handle(ctx context.Context, conn core.IConnection, _ core.IHeader, data json.RawMessage) {
	var req listRolesReq
	_ = json.Unmarshal(data, &req)
	snapshot := a.h.listRolePerms()
	filtered, total := filterRolePerms(snapshot, req)
	resp := struct {
		Code  int             `json:"code"`
		Msg   string          `json:"msg,omitempty"`
		Total int             `json:"total"`
		Roles []rolePermEntry `json:"roles,omitempty"`
	}{
		Code:  1,
		Msg:   "ok",
		Total: total,
		Roles: filtered,
	}
	raw, _ := json.Marshal(resp)
	msg := message{Action: actionListRolesResp, Data: raw}
	body, _ := json.Marshal(msg)
	hdr := a.h.buildHeader(ctx, nil)
	if srv := core.ServerFromContext(ctx); srv != nil && conn != nil {
		_ = srv.Send(ctx, conn.ID(), hdr, body)
		return
	}
	if conn != nil {
		_ = conn.SendWithHeader(hdr, body, header.HeaderTcpCodec{})
	}
}

type permsInvalidateAction struct{ h *LoginHandler }

func (a *permsInvalidateAction) Name() string      { return actionPermsInvalidate }
func (a *permsInvalidateAction) RequireAuth() bool { return true }
func (a *permsInvalidateAction) Handle(ctx context.Context, _ core.IConnection, _ core.IHeader, data json.RawMessage) {
	var req invalidateData
	_ = json.Unmarshal(data, &req)
	a.h.invalidateCache(req.NodeIDs)
	if req.Refresh {
		a.h.refreshPerms(ctx, req.NodeIDs)
	}
	// 清理当前连接的 meta
	if srv := core.ServerFromContext(ctx); srv != nil {
		if cm := srv.ConnManager(); cm != nil {
			targets := make(map[uint32]bool)
			for _, id := range req.NodeIDs {
				if id != 0 {
					targets[id] = true
				}
			}
			cm.Range(func(c core.IConnection) bool {
				if len(targets) == 0 {
					c.SetMeta("role", "")
					c.SetMeta("perms", []string(nil))
					return true
				}
				if nid, ok := c.GetMeta("nodeID"); ok {
					if v, ok2 := nid.(uint32); ok2 && targets[v] {
						c.SetMeta("role", "")
						c.SetMeta("perms", []string(nil))
					}
				}
				return true
			})
		}
		// 广播给子节点（不回父）
		srv.ConnManager().Range(func(c core.IConnection) bool {
			if role, ok := c.GetMeta(core.MetaRoleKey); ok {
				if s, ok2 := role.(string); ok2 && s == core.RoleParent {
					return true
				}
			}
			msg := message{Action: actionPermsInvalidate, Data: data}
			body, _ := json.Marshal(msg)
			hdr := (&header.HeaderTcp{}).WithMajor(header.MajorCmd).WithSubProto(2).WithSourceID(srv.NodeID()).WithTargetID(0)
			_ = srv.Send(ctx, c.ID(), hdr, body)
			return true
		})
	}
}

type permsSnapshotAction struct{ h *LoginHandler }

func (a *permsSnapshotAction) Name() string      { return actionPermsSnapshot }
func (a *permsSnapshotAction) RequireAuth() bool { return true }
func (a *permsSnapshotAction) Handle(ctx context.Context, conn core.IConnection, _ core.IHeader, data json.RawMessage) {
	if len(data) == 0 {
		return
	}
	var snap permission.Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		a.h.log.Warn("invalid perms snapshot", "err", err)
		return
	}
	a.h.applyPermSnapshot(ctx, snap)
	// forward downstream except parent
	a.h.broadcastPermsSnapshot(ctx, conn, data)
}

func registerPermActions(h *LoginHandler) []core.SubProcessAction {
	return []core.SubProcessAction{
		&getPermsAction{h: h},
		&listRolesAction{h: h},
		&permsInvalidateAction{h: h},
		&permsSnapshotAction{h: h},
	}
}
