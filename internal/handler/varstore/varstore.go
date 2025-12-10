package varstore

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"strings"
	"sync"

	core "github.com/yttydcs/myflowhub-core"
	coreconfig "github.com/yttydcs/myflowhub-core/config"
	"github.com/yttydcs/myflowhub-core/header"
	permission "github.com/yttydcs/myflowhub-core/kit/permission"
)

type VarStoreHandler struct {
	log *slog.Logger

	mu      sync.RWMutex
	records map[string]varRecord       // key: owner:name
	pending map[pendingKey][]string    // (owner,name,kind) -> waiting connIDs
	cache   map[string]map[uint32]bool // name -> owners known

	permCfg *permission.Config
	actions map[string]core.SubProcessAction
}

func NewVarStoreHandler(log *slog.Logger) *VarStoreHandler {
	return NewVarStoreHandlerWithConfig(nil, log)
}

func NewVarStoreHandlerWithConfig(cfg core.IConfig, log *slog.Logger) *VarStoreHandler {
	if log == nil {
		log = slog.Default()
	}
	if cfg == nil {
		cfg = coreconfig.NewMap(map[string]string{
			coreconfig.KeyAuthDefaultPerms: "",
		})
	}
	h := &VarStoreHandler{
		log:     log,
		records: make(map[string]varRecord),
		pending: make(map[pendingKey][]string),
		cache:   make(map[string]map[uint32]bool),
	}
	if cfg != nil {
		h.permCfg = permission.SharedConfig(cfg)
	}
	if h.permCfg == nil {
		h.permCfg = permission.NewConfig(nil)
	}
	h.Init()
	return h
}

// AcceptCmd 声明 Cmd 帧在 target!=local 时也需要本地处理一次。
func (h *VarStoreHandler) AcceptCmd() bool { return true }

// AllowSourceMismatch varstore 必须绑定 nodeID 后才能处理。
func (h *VarStoreHandler) AllowSourceMismatch() bool { return false }

func (h *VarStoreHandler) SubProto() uint8 { return 3 }

func (h *VarStoreHandler) Init() bool {
	h.initActions()
	return true
}

func (h *VarStoreHandler) OnReceive(ctx context.Context, conn core.IConnection, hdr core.IHeader, payload []byte) {
	var msg varMessage
	if err := json.Unmarshal(payload, &msg); err != nil {
		h.log.Warn("varstore invalid payload", "err", err)
		return
	}
	act := strings.ToLower(strings.TrimSpace(msg.Action))
	entry, ok := h.actions[act]
	if !ok {
		h.log.Debug("unknown varstore action", "action", act)
		return
	}
	entry.Handle(ctx, conn, hdr, msg.Data)
}

// set / assist_set
func (h *VarStoreHandler) handleSet(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage, assisted bool) {
	var req setReq
	if err := json.Unmarshal(data, &req); err != nil || !validVarName(req.Name) || strings.TrimSpace(req.Value) == "" {
		h.sendResp(ctx, conn, hdr, chooseSetResp(assisted), varResp{Code: 2, Msg: "invalid set", Name: req.Name, Owner: req.Owner})
		return
	}
	owner := req.Owner
	if owner == 0 {
		if owners, ok := h.cache[req.Name]; ok && len(owners) == 1 {
			for o := range owners {
				owner = o
			}
		}
	}
	owner = firstNonZero(owner, hdr.SourceID())
	if owner == 0 {
		h.sendResp(ctx, conn, hdr, chooseSetResp(assisted), varResp{Code: 2, Msg: "owner required", Name: req.Name})
		return
	}
	req.Owner = owner
	srv := core.ServerFromContext(ctx)
	if srv == nil {
		return
	}

	// 判断是否在当前子树
	if !h.ownerInSubtree(ctx, owner) {
		if parent := h.findParent(ctx); parent != nil {
			h.addPending(owner, req.Name, conn.ID(), pendingKindSet)
			h.forward(ctx, parent, varActionAssistSet, req, srv.NodeID())
			return
		}
		h.sendResp(ctx, conn, hdr, chooseSetResp(assisted), varResp{Code: 4, Msg: "not found", Name: req.Name, Owner: owner})
		return
	}

	actorID := permission.SourceNodeID(hdr, conn)
	existing, found := h.lookupOwned(owner, req.Name)

	// 创建仅 owner
	if !found && actorID != owner {
		h.sendResp(ctx, conn, hdr, chooseSetResp(assisted), varResp{Code: 3, Msg: "owner required", Name: req.Name, Owner: owner})
		return
	}
	// private 更新权限
	nextVis := strings.TrimSpace(req.Visibility)
	if nextVis == "" {
		if found {
			nextVis = existing.Visibility
		} else if existing.IsPublic {
			nextVis = visibilityPublic
		}
	}
	if strings.ToLower(nextVis) != visibilityPublic && actorID != owner && !h.hasPermission(actorID, permission.VarPrivateSet) {
		h.sendResp(ctx, conn, hdr, chooseSetResp(assisted), varResp{Code: 3, Msg: "permission denied", Name: req.Name, Owner: owner})
		return
	}

	rec := existing
	rec.Owner = owner
	rec.Value = req.Value
	if strings.TrimSpace(req.Type) != "" {
		rec.Type = req.Type
	} else if rec.Type == "" {
		rec.Type = "string"
	}
	if strings.TrimSpace(req.Visibility) != "" {
		rec.Visibility = strings.TrimSpace(req.Visibility)
		rec.IsPublic = strings.ToLower(req.Visibility) == visibilityPublic
	} else if rec.Visibility == "" {
		rec.Visibility = visibilityPrivate
		rec.IsPublic = false
	}

	h.saveRecord(req.Name, rec)

	// 向上同步缓存
	if parent := h.findParent(ctx); parent != nil {
		h.forward(ctx, parent, varActionUpSet, req, srv.NodeID())
	}

	// 响应与通知：始终回请求者；若请求者!=owner，则额外通知 owner
	h.sendResp(ctx, conn, hdr, chooseSetResp(assisted), varResp{
		Code:       1,
		Msg:        "ok",
		Name:       req.Name,
		Owner:      owner,
		Visibility: rec.Visibility,
		Type:       rec.Type,
	})
	if actorID != owner {
		h.sendNotifySet(ctx, owner, req.Name, rec)
	}
}

// get / assist_get
func (h *VarStoreHandler) handleGet(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage, assisted bool) {
	var req getReq
	if err := json.Unmarshal(data, &req); err != nil || !validVarName(req.Name) {
		h.sendResp(ctx, conn, hdr, chooseGetResp(assisted), varResp{Code: 2, Msg: "invalid get"})
		return
	}
	owner := firstNonZero(req.Owner, hdr.SourceID())
	if owner == 0 {
		h.sendResp(ctx, conn, hdr, chooseGetResp(assisted), varResp{Code: 2, Msg: "owner required"})
		return
	}
	req.Owner = owner
	srv := core.ServerFromContext(ctx)
	if srv == nil {
		return
	}

	if rec, ok := h.lookupOwned(owner, req.Name); ok {
		if rec.IsPublic || owner == hdr.SourceID() || h.hasPermission(permission.SourceNodeID(hdr, conn), permission.VarPrivateSet) {
			h.sendResp(ctx, conn, hdr, chooseGetResp(assisted), varResp{
				Code:       1,
				Msg:        "ok",
				Name:       req.Name,
				Value:      rec.Value,
				Owner:      owner,
				Visibility: rec.Visibility,
				Type:       rec.Type,
			})
			return
		}
		h.sendResp(ctx, conn, hdr, chooseGetResp(assisted), varResp{Code: 3, Msg: "forbidden"})
		return
	}

	if parent := h.findParent(ctx); parent != nil {
		h.addPending(owner, req.Name, conn.ID(), pendingKindGet)
		h.forward(ctx, parent, varActionAssistGet, req, srv.NodeID())
		return
	}
	h.sendResp(ctx, conn, hdr, chooseGetResp(assisted), varResp{Code: 4, Msg: "not found", Name: req.Name, Owner: owner})
}

// list / assist_list
func (h *VarStoreHandler) handleList(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage, assisted bool) {
	var req listReq
	if err := json.Unmarshal(data, &req); err != nil {
		h.sendResp(ctx, conn, hdr, chooseListResp(assisted), varResp{Code: 2, Msg: "invalid list"})
		return
	}
	owner := firstNonZero(req.Owner, hdr.SourceID())
	if owner == 0 {
		h.sendResp(ctx, conn, hdr, chooseListResp(assisted), varResp{Code: 2, Msg: "owner required"})
		return
	}
	req.Owner = owner
	srv := core.ServerFromContext(ctx)
	if srv == nil {
		return
	}

	names := h.listPublicNames(owner)
	if len(names) > 0 {
		h.sendResp(ctx, conn, hdr, chooseListResp(assisted), varResp{Code: 1, Msg: "ok", Owner: owner, Names: names})
		return
	}

	if parent := h.findParent(ctx); parent != nil {
		h.addPending(owner, "", conn.ID(), pendingKindList)
		h.forward(ctx, parent, varActionAssistList, req, srv.NodeID())
		return
	}
	h.sendResp(ctx, conn, hdr, chooseListResp(assisted), varResp{Code: 4, Msg: "not found", Owner: owner})
}

// revoke / assist_revoke
func (h *VarStoreHandler) handleRevoke(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage, assisted bool) {
	var req getReq
	if err := json.Unmarshal(data, &req); err != nil || !validVarName(req.Name) {
		h.sendResp(ctx, conn, hdr, chooseRevokeResp(assisted), varResp{Code: 2, Msg: "invalid revoke"})
		return
	}
	owner := firstNonZero(req.Owner, hdr.SourceID())
	if owner == 0 {
		h.sendResp(ctx, conn, hdr, chooseRevokeResp(assisted), varResp{Code: 2, Msg: "owner required"})
		return
	}
	req.Owner = owner
	srv := core.ServerFromContext(ctx)
	if srv == nil {
		return
	}

	if !h.ownerInSubtree(ctx, owner) {
		if parent := h.findParent(ctx); parent != nil {
			h.addPending(owner, req.Name, conn.ID(), pendingKindRevoke)
			h.forward(ctx, parent, varActionAssistRevoke, req, srv.NodeID())
			return
		}
		h.sendResp(ctx, conn, hdr, chooseRevokeResp(assisted), varResp{Code: 4, Msg: "not found", Name: req.Name, Owner: owner})
		return
	}

	actorID := permission.SourceNodeID(hdr, conn)
	if actorID != owner && !h.hasPermission(actorID, permission.VarRevoke) {
		h.sendResp(ctx, conn, hdr, chooseRevokeResp(assisted), varResp{Code: 3, Msg: "forbidden", Name: req.Name, Owner: owner})
		return
	}

	if _, ok := h.lookupOwned(owner, req.Name); !ok {
		// 尝试继续向上查询
		if parent := h.findParent(ctx); parent != nil {
			h.addPending(owner, req.Name, conn.ID(), pendingKindRevoke)
			h.forward(ctx, parent, varActionAssistRevoke, req, srv.NodeID())
			return
		}
		h.sendResp(ctx, conn, hdr, chooseRevokeResp(assisted), varResp{Code: 4, Msg: "not found", Name: req.Name, Owner: owner})
		return
	}

	h.deleteRecord(owner, req.Name)

	if parent := h.findParent(ctx); parent != nil {
		h.forward(ctx, parent, varActionUpRevoke, req, srv.NodeID())
	}

	h.sendResp(ctx, conn, hdr, chooseRevokeResp(assisted), varResp{Code: 1, Msg: "ok", Name: req.Name, Owner: owner})
	if actorID != owner {
		h.sendNotifyRevoke(ctx, owner, req.Name)
	}
}

// responses from upstream
func (h *VarStoreHandler) handleSetResp(ctx context.Context, data json.RawMessage) {
	var resp varResp
	if err := json.Unmarshal(data, &resp); err != nil {
		return
	}
	connIDs := h.popPending(resp.Owner, resp.Name, pendingKindSet)
	if resp.Code == 1 {
		rec := varRecord{
			Value:      resp.Value,
			Owner:      resp.Owner,
			Visibility: resp.Visibility,
			Type:       resp.Type,
			IsPublic:   strings.ToLower(resp.Visibility) == visibilityPublic,
		}
		h.saveRecord(resp.Name, rec)
	}
	h.broadcastPendingResp(ctx, connIDs, varActionSetResp, resp)
}

func (h *VarStoreHandler) handleGetResp(ctx context.Context, data json.RawMessage) {
	var resp varResp
	if err := json.Unmarshal(data, &resp); err != nil {
		return
	}
	connIDs := h.popPending(resp.Owner, resp.Name, pendingKindGet)
	if resp.Code == 1 {
		rec := varRecord{
			Value:      resp.Value,
			Owner:      resp.Owner,
			Visibility: resp.Visibility,
			Type:       resp.Type,
			IsPublic:   strings.ToLower(resp.Visibility) == visibilityPublic,
		}
		h.saveRecord(resp.Name, rec)
	}
	h.broadcastPendingResp(ctx, connIDs, varActionGetResp, resp)
}

func (h *VarStoreHandler) handleListResp(ctx context.Context, data json.RawMessage) {
	var resp varResp
	if err := json.Unmarshal(data, &resp); err != nil {
		return
	}
	connIDs := h.popPending(resp.Owner, "", pendingKindList)
	h.broadcastPendingResp(ctx, connIDs, varActionListResp, resp)
}

func (h *VarStoreHandler) handleRevokeResp(ctx context.Context, data json.RawMessage) {
	var resp varResp
	if err := json.Unmarshal(data, &resp); err != nil {
		return
	}
	connIDs := h.popPending(resp.Owner, resp.Name, pendingKindRevoke)
	if resp.Code == 1 {
		h.deleteRecord(resp.Owner, resp.Name)
	}
	h.broadcastPendingResp(ctx, connIDs, varActionRevokeResp, resp)
}

// up_* handlers
func (h *VarStoreHandler) handleUpSet(ctx context.Context, data json.RawMessage) {
	var req setReq
	if err := json.Unmarshal(data, &req); err != nil || req.Owner == 0 || req.Name == "" {
		return
	}
	rec := varRecord{
		Value:      req.Value,
		Owner:      req.Owner,
		Visibility: strings.TrimSpace(req.Visibility),
		Type:       req.Type,
		IsPublic:   strings.ToLower(req.Visibility) == visibilityPublic,
	}
	h.saveRecord(req.Name, rec)
	if parent := h.findParent(ctx); parent != nil {
		srv := core.ServerFromContext(ctx)
		if srv != nil {
			h.forward(ctx, parent, varActionUpSet, req, srv.NodeID())
		}
	}
}

func (h *VarStoreHandler) handleUpRevoke(ctx context.Context, data json.RawMessage) {
	var req getReq
	if err := json.Unmarshal(data, &req); err != nil || req.Owner == 0 || req.Name == "" {
		return
	}
	h.deleteRecord(req.Owner, req.Name)
	if parent := h.findParent(ctx); parent != nil {
		srv := core.ServerFromContext(ctx)
		if srv != nil {
			h.forward(ctx, parent, varActionUpRevoke, req, srv.NodeID())
		}
	}
}

// notify handlers
func (h *VarStoreHandler) handleNotifySet(ctx context.Context, data json.RawMessage) {
	var resp varResp
	if err := json.Unmarshal(data, &resp); err != nil || resp.Name == "" || resp.Owner == 0 {
		return
	}
	rec := varRecord{
		Value:      resp.Value,
		Owner:      resp.Owner,
		Visibility: resp.Visibility,
		Type:       resp.Type,
		IsPublic:   strings.ToLower(resp.Visibility) == visibilityPublic,
	}
	h.saveRecord(resp.Name, rec)
}

func (h *VarStoreHandler) handleNotifyRevoke(ctx context.Context, data json.RawMessage) {
	var resp varResp
	if err := json.Unmarshal(data, &resp); err != nil || resp.Owner == 0 || resp.Name == "" {
		return
	}
	h.deleteRecord(resp.Owner, resp.Name)
}

// helpers
func (h *VarStoreHandler) ownerInSubtree(ctx context.Context, owner uint32) bool {
	if owner == 0 {
		return false
	}
	srv := core.ServerFromContext(ctx)
	if srv == nil {
		return false
	}
	if srv.NodeID() == owner {
		return true
	}
	if _, ok := srv.ConnManager().GetByNode(owner); ok {
		return true
	}
	return false
}

func (h *VarStoreHandler) lookupOwned(owner uint32, name string) (varRecord, bool) {
	if owner == 0 || name == "" {
		return varRecord{}, false
	}
	h.mu.RLock()
	rec, ok := h.records[h.key(owner, name)]
	h.mu.RUnlock()
	return rec, ok
}

func (h *VarStoreHandler) saveRecord(name string, rec varRecord) {
	if name == "" || rec.Owner == 0 {
		return
	}
	key := h.key(rec.Owner, name)
	h.mu.Lock()
	h.records[key] = rec
	h.addOwnerCache(name, rec.Owner)
	h.mu.Unlock()
}

func (h *VarStoreHandler) listPublicNames(owner uint32) []string {
	if owner == 0 {
		return nil
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	var names []string
	prefix := strconv.FormatUint(uint64(owner), 10) + ":"
	for key, rec := range h.records {
		if !rec.IsPublic {
			continue
		}
		if strings.HasPrefix(key, prefix) {
			if parts := strings.SplitN(key, ":", 2); len(parts) == 2 && parts[1] != "" {
				names = append(names, parts[1])
			}
		}
	}
	return names
}

func (h *VarStoreHandler) key(owner uint32, name string) string {
	return strconv.FormatUint(uint64(owner), 10) + ":" + name
}

func (h *VarStoreHandler) addOwnerCache(name string, owner uint32) {
	if owner == 0 || name == "" {
		return
	}
	if _, ok := h.cache[name]; !ok {
		h.cache[name] = make(map[uint32]bool)
	}
	h.cache[name][owner] = true
}

func (h *VarStoreHandler) deleteRecord(owner uint32, name string) {
	if owner == 0 || name == "" {
		return
	}
	k := h.key(owner, name)
	h.mu.Lock()
	delete(h.records, k)
	if owners, ok := h.cache[name]; ok {
		delete(owners, owner)
		if len(owners) == 0 {
			delete(h.cache, name)
		} else {
			h.cache[name] = owners
		}
	}
	h.mu.Unlock()
}

func (h *VarStoreHandler) addPending(owner uint32, name, connID string, kind string) {
	h.mu.Lock()
	k := pendingKey{owner: owner, name: name, kind: kind}
	h.pending[k] = append(h.pending[k], connID)
	h.mu.Unlock()
}

func (h *VarStoreHandler) popPending(owner uint32, name string, kind string) []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	k := pendingKey{owner: owner, name: name, kind: kind}
	conns := h.pending[k]
	delete(h.pending, k)
	return conns
}

func (h *VarStoreHandler) broadcastPendingResp(ctx context.Context, connIDs []string, action string, resp varResp) {
	if len(connIDs) == 0 {
		return
	}
	for _, id := range connIDs {
		h.sendResp(ctx, h.lookupConn(ctx, id), nil, action, resp)
	}
}

func (h *VarStoreHandler) lookupConn(ctx context.Context, id string) core.IConnection {
	srv := core.ServerFromContext(ctx)
	if srv == nil {
		return nil
	}
	if c, ok := srv.ConnManager().Get(id); ok {
		return c
	}
	return nil
}

func (h *VarStoreHandler) sendResp(ctx context.Context, conn core.IConnection, reqHdr core.IHeader, action string, data varResp) {
	msg := varMessage{Action: action}
	raw, _ := json.Marshal(data)
	msg.Data = raw
	payload, _ := json.Marshal(msg)
	hdr := h.buildRespHeader(ctx, reqHdr, data.Owner)
	if srv := core.ServerFromContext(ctx); srv != nil && conn != nil {
		_ = srv.Send(ctx, conn.ID(), hdr, payload)
		return
	}
	if conn != nil {
		codec := header.HeaderTcpCodec{}
		_ = conn.SendWithHeader(hdr, payload, codec)
	}
}

func (h *VarStoreHandler) buildRespHeader(ctx context.Context, reqHdr core.IHeader, target uint32) core.IHeader {
	var base core.IHeader = &header.HeaderTcp{}
	if reqHdr != nil {
		base = reqHdr.Clone()
	}
	src := uint32(0)
	if srv := core.ServerFromContext(ctx); srv != nil {
		src = srv.NodeID()
	}
	if target == 0 && reqHdr != nil && reqHdr.SourceID() != 0 {
		target = reqHdr.SourceID()
	}
	return base.WithMajor(header.MajorOKResp).WithSubProto(3).WithSourceID(src).WithTargetID(target)
}

func (h *VarStoreHandler) forward(ctx context.Context, target core.IConnection, action string, data any, srcID uint32) {
	if target == nil {
		return
	}
	payloadData, _ := json.Marshal(data)
	msg := varMessage{Action: action, Data: payloadData}
	payload, _ := json.Marshal(msg)
	hdr := (&header.HeaderTcp{}).WithMajor(header.MajorCmd).WithSubProto(3)
	if srv := core.ServerFromContext(ctx); srv != nil {
		if srcID != 0 {
			hdr.WithSourceID(srcID)
		} else {
			hdr.WithSourceID(srv.NodeID())
		}
	}
	if nid, ok := target.GetMeta("nodeID"); ok {
		if v, ok2 := nid.(uint32); ok2 {
			hdr.WithTargetID(v)
		}
	}
	if srv := core.ServerFromContext(ctx); srv != nil {
		_ = srv.Send(ctx, target.ID(), hdr, payload)
		return
	}
	codec := header.HeaderTcpCodec{}
	_ = target.SendWithHeader(hdr, payload, codec)
}

func (h *VarStoreHandler) sendNotifySet(ctx context.Context, owner uint32, name string, rec varRecord) {
	if owner == 0 || name == "" {
		return
	}
	srv := core.ServerFromContext(ctx)
	if srv == nil {
		return
	}
	resp := varResp{
		Code:       1,
		Name:       name,
		Owner:      owner,
		Value:      rec.Value,
		Visibility: rec.Visibility,
		Type:       rec.Type,
	}
	raw, _ := json.Marshal(resp)
	msg := varMessage{Action: varActionNotifySet, Data: raw}
	payload, _ := json.Marshal(msg)
	hdr := (&header.HeaderTcp{}).
		WithMajor(header.MajorCmd).
		WithSubProto(3).
		WithSourceID(srv.NodeID()).
		WithTargetID(owner)
	_ = srv.Send(ctx, ownerConnID(ctx, srv, owner), hdr, payload)
}

func (h *VarStoreHandler) sendNotifyRevoke(ctx context.Context, owner uint32, name string) {
	if owner == 0 || name == "" {
		return
	}
	srv := core.ServerFromContext(ctx)
	if srv == nil {
		return
	}
	resp := varResp{
		Code:  1,
		Name:  name,
		Owner: owner,
	}
	raw, _ := json.Marshal(resp)
	msg := varMessage{Action: varActionNotifyRevoke, Data: raw}
	payload, _ := json.Marshal(msg)
	hdr := (&header.HeaderTcp{}).
		WithMajor(header.MajorCmd).
		WithSubProto(3).
		WithSourceID(srv.NodeID()).
		WithTargetID(owner)
	_ = srv.Send(ctx, ownerConnID(ctx, srv, owner), hdr, payload)
}

func (h *VarStoreHandler) findParent(ctx context.Context) core.IConnection {
	srv := core.ServerFromContext(ctx)
	if srv == nil {
		return nil
	}
	if c, ok := findParentConnVar(srv.ConnManager()); ok {
		return c
	}
	return nil
}

func findParentConnVar(cm core.IConnectionManager) (core.IConnection, bool) {
	var parent core.IConnection
	cm.Range(func(c core.IConnection) bool {
		if role, ok := c.GetMeta(core.MetaRoleKey); ok {
			if s, ok2 := role.(string); ok2 && s == core.RoleParent {
				parent = c
				return false
			}
		}
		return true
	})
	return parent, parent != nil
}

func (h *VarStoreHandler) hasPermission(nodeID uint32, perm string) bool {
	if h.permCfg == nil {
		return false
	}
	return h.permCfg.Has(nodeID, perm)
}

func (h *VarStoreHandler) initActions() {
	h.actions = make(map[string]core.SubProcessAction)
	for _, act := range registerVarActions(h) {
		h.registerAction(act)
	}
}

func (h *VarStoreHandler) registerAction(a core.SubProcessAction) {
	if a == nil || a.Name() == "" {
		return
	}
	h.actions[strings.ToLower(a.Name())] = a
}
