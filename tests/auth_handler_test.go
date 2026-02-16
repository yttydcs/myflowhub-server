package tests

import (
	"context"
	"encoding/json"
	"net"
	"testing"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-core/config"
	"github.com/yttydcs/myflowhub-core/connmgr"
	"github.com/yttydcs/myflowhub-core/eventbus"
	"github.com/yttydcs/myflowhub-core/header"
	permission "github.com/yttydcs/myflowhub-core/kit/permission"
	auth "github.com/yttydcs/myflowhub-server/subproto/auth"
)

func TestLoginHandlerGetPermsAndListRoles(t *testing.T) {
	cfg := config.NewMap(map[string]string{
		config.KeyAuthNodeRoles: "5:admin",
		config.KeyAuthRolePerms: "admin:p.read,p.write",
	})
	h := newLoginHandlerForTest(cfg)

	cm := connmgr.New()
	conn := newAuthConn("c1")
	conn.SetMeta("nodeID", uint32(5))
	_ = cm.Add(conn)
	srv := newAuthServer(1, cm)
	ctx := core.WithServerContext(context.Background(), srv)

	// get_perms
	req := mustJSON(map[string]any{"action": "get_perms", "data": map[string]any{"node_id": 5}})
	hdr := (&header.HeaderTcp{}).WithMajor(header.MajorCmd).WithSubProto(2)
	hdr.WithSourceID(5)
	h.OnReceive(ctx, conn, hdr, req)

	if len(conn.sent) != 1 {
		t.Fatalf("expected 1 response, got %d", len(conn.sent))
	}
	var msg struct {
		Action string          `json:"action"`
		Data   json.RawMessage `json:"data"`
	}
	_ = json.Unmarshal(conn.sent[0].payload, &msg)
	if msg.Action != "get_perms_resp" {
		t.Fatalf("unexpected action %s", msg.Action)
	}
	var data struct {
		Code  int      `json:"code"`
		Role  string   `json:"role"`
		Perms []string `json:"perms"`
	}
	_ = json.Unmarshal(msg.Data, &data)
	if data.Code != 1 || data.Role != "admin" || len(data.Perms) != 2 {
		t.Fatalf("unexpected perms resp %+v", data)
	}

	// list_roles
	conn.sent = nil
	reqList := mustJSON(map[string]any{"action": "list_roles", "data": map[string]any{}})
	h.OnReceive(ctx, conn, hdr, reqList)
	if len(conn.sent) != 1 {
		t.Fatalf("expected list_roles resp, got %d", len(conn.sent))
	}
	_ = json.Unmarshal(conn.sent[0].payload, &msg)
	if msg.Action != "list_roles_resp" {
		t.Fatalf("unexpected action %s", msg.Action)
	}
	var list struct {
		Code  int `json:"code"`
		Total int `json:"total"`
		Roles []struct {
			NodeID uint32 `json:"node_id"`
			Role   string `json:"role"`
		} `json:"roles"`
	}
	_ = json.Unmarshal(msg.Data, &list)
	if list.Code != 1 || list.Total == 0 || len(list.Roles) == 0 || list.Roles[0].Role != "admin" {
		t.Fatalf("unexpected list_roles_resp %+v", list)
	}
}

func TestLoginHandlerPermsInvalidate(t *testing.T) {
	cfg := config.NewMap(map[string]string{
		config.KeyAuthNodeRoles: "5:admin;6:node",
	})
	h := newLoginHandlerForTest(cfg)
	cm := connmgr.New()
	connTarget := newAuthConn("c5")
	_ = cm.Add(connTarget)

	connOther := newAuthConn("c6")
	connOther.SetMeta("nodeID", uint32(6))
	connOther.SetMeta("role", "node")
	_ = cm.Add(connOther)

	srv := newAuthServer(1, cm)
	ctx := core.WithServerContext(context.Background(), srv)

	// 注册写入绑定，分配 nodeID
	regMsg := mustJSON(map[string]any{"action": "register", "data": map[string]any{"device_id": "dev-1"}})
	hdr := (&header.HeaderTcp{}).WithMajor(header.MajorCmd).WithSubProto(2)
	h.OnReceive(ctx, connTarget, hdr, regMsg)
	nodeIDVal, _ := connTarget.GetMeta("nodeID")
	nodeID, _ := nodeIDVal.(uint32)
	if nodeID == 0 {
		t.Fatalf("expected nodeID assigned")
	}
	connTarget.SetMeta("role", "admin")
	connTarget.SetMeta("perms", []string{"p.read"})

	// invalidate node 5
	req := mustJSON(map[string]any{"action": "perms_invalidate", "data": map[string]any{"node_ids": []uint32{nodeID}}})
	hdr.WithSourceID(nodeID)
	h.OnReceive(ctx, connTarget, hdr, req)

	// meta cleared for node 5
	if role, _ := connTarget.GetMeta("role"); role != "" {
		t.Fatalf("expected role cleared for node 5, got %v", role)
	}
	if perms, _ := connTarget.GetMeta("perms"); perms != nil {
		if v, ok := perms.([]string); !ok || len(v) != 0 {
			t.Fatalf("expected perms cleared for node 5, got %v", perms)
		}
	}
	// other node untouched
	if role, _ := connOther.GetMeta("role"); role == "" {
		t.Fatalf("unexpected role cleared for other node")
	}
}

func TestLoginHandlerPermsInvalidateRefreshToParent(t *testing.T) {
	cfg := config.NewMap(nil)
	h := newLoginHandlerForTest(cfg)
	cm := connmgr.New()

	parent := newAuthConn("parent")
	parent.SetMeta(core.MetaRoleKey, core.RoleParent)
	parent.SetMeta("nodeID", uint32(99))
	_ = cm.Add(parent)

	child := newAuthConn("child")
	child.SetMeta("nodeID", uint32(10))
	_ = cm.Add(child)

	srv := newAuthServer(1, cm)
	ctx := core.WithServerContext(context.Background(), srv)

	req := mustJSON(map[string]any{"action": "perms_invalidate", "data": map[string]any{"node_ids": []uint32{10}, "refresh": true}})
	hdr := (&header.HeaderTcp{}).WithMajor(header.MajorCmd).WithSubProto(2)
	hdr.WithSourceID(10)
	h.OnReceive(ctx, child, hdr, req)

	if len(parent.sent) == 0 {
		t.Fatalf("expected refresh get_perms sent to parent")
	}
	var msg struct {
		Action string          `json:"action"`
		Data   json.RawMessage `json:"data"`
	}
	_ = json.Unmarshal(parent.sent[0].payload, &msg)
	if msg.Action != "get_perms" {
		t.Fatalf("expected get_perms, got %s", msg.Action)
	}
}

func TestLoginHandlerPermsInvalidateRefreshSnapshotRequest(t *testing.T) {
	cfg := config.NewMap(nil)
	h := newLoginHandlerForTest(cfg)
	cm := connmgr.New()

	parent := newAuthConn("parent")
	parent.SetMeta(core.MetaRoleKey, core.RoleParent)
	parent.SetMeta("nodeID", uint32(99))
	_ = cm.Add(parent)

	child := newAuthConn("child")
	child.SetMeta("nodeID", uint32(10))
	_ = cm.Add(child)

	srv := newAuthServer(1, cm)
	ctx := core.WithServerContext(context.Background(), srv)

	req := mustJSON(map[string]any{"action": "perms_invalidate", "data": map[string]any{"refresh": true}})
	hdr := (&header.HeaderTcp{}).WithMajor(header.MajorCmd).WithSubProto(2)
	hdr.WithSourceID(10)
	h.OnReceive(ctx, child, hdr, req)

	if len(parent.sent) == 0 {
		t.Fatalf("expected snapshot request sent to parent")
	}
	var msg struct {
		Action string          `json:"action"`
		Data   json.RawMessage `json:"data"`
	}
	_ = json.Unmarshal(parent.sent[0].payload, &msg)
	if msg.Action != "perms_snapshot" {
		t.Fatalf("expected perms_snapshot, got %s", msg.Action)
	}
}

func TestLoginHandlerApplyPermsSnapshot(t *testing.T) {
	cfg := config.NewMap(nil)
	h := newLoginHandlerForTest(cfg)
	cm := connmgr.New()

	parent := newAuthConn("parent")
	parent.SetMeta(core.MetaRoleKey, core.RoleParent)
	parent.SetMeta("nodeID", uint32(99))
	_ = cm.Add(parent)

	child := newAuthConn("child")
	child.SetMeta("nodeID", uint32(5))
	_ = cm.Add(child)

	srv := newAuthServer(1, cm)
	ctx := core.WithServerContext(context.Background(), srv)

	snap := permission.Snapshot{
		DefaultRole: "node",
		NodeRoles:   map[uint32]string{5: "admin"},
		RolePerms:   map[string][]string{"admin": []string{"perm.a"}},
	}
	snapData, _ := json.Marshal(snap)
	msg := struct {
		Action string          `json:"action"`
		Data   json.RawMessage `json:"data"`
	}{
		Action: "perms_snapshot",
		Data:   snapData,
	}
	payload, _ := json.Marshal(msg)
	hdr := (&header.HeaderTcp{}).WithMajor(header.MajorCmd).WithSubProto(2)
	hdr.WithSourceID(99)

	h.OnReceive(ctx, parent, hdr, payload)

	if len(child.sent) != 1 {
		t.Fatalf("expected snapshot forwarded to child")
	}
	var forwardMsg struct {
		Action string `json:"action"`
	}
	_ = json.Unmarshal(child.sent[0].payload, &forwardMsg)
	if forwardMsg.Action != "perms_snapshot" {
		t.Fatalf("expected forwarded perms_snapshot, got %s", forwardMsg.Action)
	}
	if len(parent.sent) != 0 {
		t.Fatalf("expected no echo to parent")
	}
	if roleMeta, _ := child.GetMeta("role"); roleMeta != "admin" {
		t.Fatalf("expected child role meta updated, got %v", roleMeta)
	}
	child.sent = nil
	req := mustJSON(map[string]any{"action": "get_perms", "data": map[string]any{"node_id": 5}})
	hdrChild := (&header.HeaderTcp{}).WithMajor(header.MajorCmd).WithSubProto(2).WithSourceID(5)
	h.OnReceive(ctx, child, hdrChild, req)
	if len(child.sent) == 0 {
		t.Fatalf("expected get_perms response after snapshot")
	}
	var resp struct {
		Action string          `json:"action"`
		Data   json.RawMessage `json:"data"`
	}
	_ = json.Unmarshal(child.sent[len(child.sent)-1].payload, &resp)
	if resp.Action != "get_perms_resp" {
		t.Fatalf("expected get_perms_resp, got %s", resp.Action)
	}
	var body struct {
		Code  int      `json:"code"`
		Role  string   `json:"role"`
		Perms []string `json:"perms"`
	}
	_ = json.Unmarshal(resp.Data, &body)
	if body.Code != 1 || body.Role != "admin" || len(body.Perms) != 1 || body.Perms[0] != "perm.a" {
		t.Fatalf("unexpected get_perms_resp %+v", body)
	}
}

func TestLoginHandlerRevokePermissionDenied(t *testing.T) {
	cfg := config.NewMap(map[string]string{
		config.KeyAuthDefaultPerms: "",
		config.KeyAuthNodeRoles:    "100:guest",
		config.KeyAuthRolePerms:    "admin:auth.revoke",
	})
	h := newLoginHandlerForTest(cfg)
	cm := connmgr.New()
	conn := newAuthConn("guest")
	conn.SetMeta("nodeID", uint32(100))
	_ = cm.Add(conn)
	srv := newAuthServer(1, cm)
	ctx := core.WithServerContext(context.Background(), srv)

	req := mustJSON(map[string]any{"action": "revoke", "data": map[string]any{"device_id": "dev-100"}})
	hdr := (&header.HeaderTcp{}).
		WithMajor(header.MajorCmd).
		WithSubProto(2).
		WithSourceID(100)

	h.OnReceive(ctx, conn, hdr, req)

	if len(conn.sent) == 0 {
		t.Fatalf("expected revoke_resp for permission failure")
	}
	var msg struct {
		Action string          `json:"action"`
		Data   json.RawMessage `json:"data"`
	}
	_ = json.Unmarshal(conn.sent[0].payload, &msg)
	if msg.Action != "revoke_resp" {
		t.Fatalf("unexpected action %s", msg.Action)
	}
	var resp struct {
		Code int `json:"code"`
	}
	_ = json.Unmarshal(msg.Data, &resp)
	if resp.Code != 4403 {
		t.Fatalf("expected 4403, got %d", resp.Code)
	}
}

func TestLoginHandlerAssistRegisterRespFallback(t *testing.T) {
	cfg := config.NewMap(nil)
	h := newLoginHandlerForTest(cfg)
	cm := connmgr.New()

	parent := newAuthConn("parent")
	parent.SetMeta(core.MetaRoleKey, core.RoleParent)
	parent.SetMeta("nodeID", uint32(99))
	_ = cm.Add(parent)

	device := newAuthConn("device")
	_ = cm.Add(device)

	srv := newAuthServer(1, cm)
	ctx := core.WithServerContext(context.Background(), srv)

	// device -> A: register (A should forward assist_register to parent and set pending)
	regMsg := mustJSON(map[string]any{"action": "register", "data": map[string]any{"device_id": "dev-1"}})
	hdr := (&header.HeaderTcp{}).WithMajor(header.MajorCmd).WithSubProto(2)
	h.OnReceive(ctx, device, hdr, regMsg)

	if len(device.sent) != 0 {
		t.Fatalf("expected no immediate register_resp, got %d", len(device.sent))
	}

	// parent -> A: assist_register_resp (A must consume and fallback to register_resp to device)
	respMsg := mustJSON(map[string]any{
		"action": "assist_register_resp",
		"data": map[string]any{
			"code":      1,
			"msg":       "ok",
			"device_id": "dev-1",
			"node_id":   5,
		},
	})
	h.OnReceive(ctx, parent, hdr, respMsg)

	if len(device.sent) != 1 {
		t.Fatalf("expected 1 register_resp, got %d", len(device.sent))
	}
	var msg struct {
		Action string          `json:"action"`
		Data   json.RawMessage `json:"data"`
	}
	_ = json.Unmarshal(device.sent[0].payload, &msg)
	if msg.Action != "register_resp" {
		t.Fatalf("expected register_resp, got %s", msg.Action)
	}
}

func TestLoginHandlerAssistLoginRespFallback(t *testing.T) {
	cfg := config.NewMap(nil)
	h := newLoginHandlerForTest(cfg)
	cm := connmgr.New()

	parent := newAuthConn("parent")
	parent.SetMeta(core.MetaRoleKey, core.RoleParent)
	parent.SetMeta("nodeID", uint32(99))
	_ = cm.Add(parent)

	device := newAuthConn("device")
	_ = cm.Add(device)

	srv := newAuthServer(1, cm)
	ctx := core.WithServerContext(context.Background(), srv)

	// device -> A: login (A should forward assist_login to parent and set pending)
	loginMsg := mustJSON(map[string]any{"action": "login", "data": map[string]any{"device_id": "dev-1", "ts": 1, "nonce": "n1", "sig": "s1", "alg": "ES256"}})
	hdr := (&header.HeaderTcp{}).WithMajor(header.MajorCmd).WithSubProto(2)
	h.OnReceive(ctx, device, hdr, loginMsg)

	if len(device.sent) != 0 {
		t.Fatalf("expected no immediate login_resp, got %d", len(device.sent))
	}

	// parent -> A: assist_login_resp (A must consume and fallback to login_resp to device)
	respMsg := mustJSON(map[string]any{
		"action": "assist_login_resp",
		"data": map[string]any{
			"code":      4001,
			"msg":       "invalid signature",
			"device_id": "dev-1",
		},
	})
	h.OnReceive(ctx, parent, hdr, respMsg)

	if len(device.sent) != 1 {
		t.Fatalf("expected 1 login_resp, got %d", len(device.sent))
	}
	var msg struct {
		Action string          `json:"action"`
		Data   json.RawMessage `json:"data"`
	}
	_ = json.Unmarshal(device.sent[0].payload, &msg)
	if msg.Action != "login_resp" {
		t.Fatalf("expected login_resp, got %s", msg.Action)
	}
}

// --- helpers ---

type authConn struct {
	id   string
	meta map[string]any
	sent []sentFrame
}

func newAuthConn(id string) *authConn {
	return &authConn{id: id, meta: make(map[string]any)}
}

func (c *authConn) ID() string                           { return c.id }
func (c *authConn) Close() error                         { return nil }
func (c *authConn) OnReceive(core.ReceiveHandler)        {}
func (c *authConn) SetMeta(key string, val any)          { c.meta[key] = val }
func (c *authConn) GetMeta(key string) (any, bool)       { v, ok := c.meta[key]; return v, ok }
func (c *authConn) Metadata() map[string]any             { return c.meta }
func (c *authConn) LocalAddr() net.Addr                  { return mockAddr{} }
func (c *authConn) RemoteAddr() net.Addr                 { return mockAddr{} }
func (c *authConn) Reader() core.IReader                 { return nil }
func (c *authConn) SetReader(core.IReader)               {}
func (c *authConn) DispatchReceive(core.IHeader, []byte) {}
func (c *authConn) RawConn() net.Conn                    { return nil }
func (c *authConn) Send([]byte) error                    { return nil }
func (c *authConn) SendWithHeader(h core.IHeader, payload []byte, _ core.IHeaderCodec) error {
	c.sent = append(c.sent, sentFrame{hdr: h, payload: payload})
	return nil
}

type authServer struct {
	nodeID uint32
	cm     core.IConnectionManager
	bus    eventbus.IBus
}

func newAuthServer(nodeID uint32, cm core.IConnectionManager) *authServer {
	return &authServer{nodeID: nodeID, cm: cm}
}

func (s *authServer) Start(context.Context) error          { return nil }
func (s *authServer) Stop(context.Context) error           { return nil }
func (s *authServer) Config() core.IConfig                 { return config.NewMap(nil) }
func (s *authServer) ConnManager() core.IConnectionManager { return s.cm }
func (s *authServer) Process() core.IProcess               { return nil }
func (s *authServer) HeaderCodec() core.IHeaderCodec       { return header.HeaderTcpCodec{} }
func (s *authServer) NodeID() uint32                       { return s.nodeID }
func (s *authServer) UpdateNodeID(id uint32)               { s.nodeID = id }
func (s *authServer) EventBus() eventbus.IBus {
	if s.bus == nil {
		s.bus = eventbus.New(eventbus.Options{})
	}
	return s.bus
}
func (s *authServer) Send(_ context.Context, connID string, hdr core.IHeader, payload []byte) error {
	if c, ok := s.cm.Get(connID); ok {
		return c.SendWithHeader(hdr, payload, header.HeaderTcpCodec{})
	}
	return nil
}

func newLoginHandlerForTest(cfg core.IConfig) *auth.LoginHandler {
	if cfg != nil {
		cfg.Set("auth.disable_persist", "true")
	}
	h := auth.NewLoginHandlerWithConfig(cfg, nil)
	h.Init()
	return h
}
