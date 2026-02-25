package tests

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"testing"
	"time"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-core/config"
	"github.com/yttydcs/myflowhub-core/connmgr"
	"github.com/yttydcs/myflowhub-core/header"
	"github.com/yttydcs/myflowhub-core/listener/tcp_listener"
	"github.com/yttydcs/myflowhub-core/process"
	"github.com/yttydcs/myflowhub-core/server"
	"github.com/yttydcs/myflowhub-server/hubruntime"
	auth "github.com/yttydcs/myflowhub-subproto/auth"
	"github.com/yttydcs/myflowhub-subproto/forward"
	"github.com/yttydcs/myflowhub-subproto/management"
)

// Integration: Client -> Root (forward cmd) -> Hub (resp) -> Root (route back) -> Client.
//
// This test covers:
// - hubruntime self-register (pre-start) to obtain node_id
// - parent bootstrap (post-start) register on persistent parent link to bind root-side meta(nodeID)
// - management cmd forwarding by target id across hub tree
func TestRootHubPing(t *testing.T) {
	oldWD, _ := os.Getwd()
	tmp := t.TempDir()
	_ = os.Chdir(tmp)
	t.Cleanup(func() {
		_ = os.Chdir(oldWD)
	})

	rootAddr := freeAddr()
	hubAddr := freeAddr()

	// Root handles login locally and can forward management cmd by target id.
	rootCfg := config.NewMap(map[string]string{
		"addr": rootAddr,
	})
	rootHandlers := []core.ISubProcess{
		management.NewHandler(nil),
		auth.NewLoginHandlerWithConfig(rootCfg, nil),
	}
	rootSrv := startTestServer(t, server.Options{
		Name:     "Root",
		Process:  makeProcess(t, rootCfg, rootHandlers),
		Codec:    header.HeaderTcpCodec{},
		Listener: tcp_listener.New(rootAddr),
		Config:   rootCfg,
		Manager:  connmgr.New(),
		NodeID:   1,
	})
	defer stopTestServer(t, rootSrv)

	waitListen(t, rootAddr, 2*time.Second)

	// Hub: start via hubruntime (self-register + parent bootstrap).
	rt, err := hubruntime.New(hubruntime.Options{
		Addr:         hubAddr,
		NodeID:       0,
		ParentAddr:   rootAddr,
		ParentEnable: true,
		SelfID:       "hub-test",
	})
	if err != nil {
		t.Fatalf("init hub runtime: %v", err)
	}
	if err := rt.Start(context.Background()); err != nil {
		t.Fatalf("start hub runtime: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = rt.Stop(ctx)
	}()
	hubNodeID := rt.Status().NodeID
	if hubNodeID == 0 {
		t.Fatalf("hub runtime node id 0")
	}
	waitListen(t, hubAddr, 2*time.Second)

	// Wait for root to bind meta(nodeID) for the persistent parent link (bootstrap must kick in).
	waitNodeIndex(t, rootSrv.ConnManager(), hubNodeID, 3*time.Second)

	// Client registers to root, then sends management echo (target=hub).
	conn, err := net.Dial("tcp", rootAddr)
	if err != nil {
		t.Fatalf("dial root: %v", err)
	}
	defer conn.Close()

	codec := header.HeaderTcpCodec{}
	clientNodeID := registerOnConn(t, conn, codec, "client-test")
	payload := mustJSON(map[string]any{
		"action": "node_echo",
		"data":   map[string]any{"message": "ping"},
	})
	hdr := (&header.HeaderTcp{}).
		WithMajor(header.MajorCmd).
		WithSubProto(management.SubProtoManagement).
		WithSourceID(clientNodeID).
		WithTargetID(hubNodeID).
		WithMsgID(1).
		WithPayloadLength(uint32(len(payload)))
	frame, _ := codec.Encode(hdr, payload)
	if _, err := conn.Write(frame); err != nil {
		t.Fatalf("write: %v", err)
	}
	_, respPayload, err := codec.Decode(conn)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	var msg struct {
		Action string          `json:"action"`
		Data   json.RawMessage `json:"data"`
	}
	_ = json.Unmarshal(respPayload, &msg)
	if msg.Action != "node_echo_resp" {
		t.Fatalf("unexpected action: %s", msg.Action)
	}
	var resp struct {
		Code int    `json:"code"`
		Echo string `json:"echo"`
	}
	_ = json.Unmarshal(msg.Data, &resp)
	if resp.Code != 1 || resp.Echo != "ping" {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

func registerOnConn(t *testing.T, conn net.Conn, codec header.HeaderTcpCodec, deviceID string) uint32 {
	t.Helper()
	payload := mustJSON(map[string]any{
		"action": "register",
		"data":   map[string]any{"device_id": deviceID},
	})
	hdr := (&header.HeaderTcp{}).
		WithMajor(header.MajorCmd).
		WithSubProto(2).
		WithSourceID(0).
		WithTargetID(0).
		WithMsgID(1).
		WithPayloadLength(uint32(len(payload)))
	frame, _ := codec.Encode(hdr, payload)
	if _, err := conn.Write(frame); err != nil {
		t.Fatalf("register write: %v", err)
	}
	_, respPayload, err := codec.Decode(conn)
	if err != nil {
		t.Fatalf("register decode: %v", err)
	}
	var msg struct {
		Action string          `json:"action"`
		Data   json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(respPayload, &msg); err != nil {
		t.Fatalf("register unmarshal msg: %v", err)
	}
	var resp struct {
		Code   int    `json:"code"`
		NodeID uint32 `json:"node_id"`
		Msg    string `json:"msg"`
	}
	if err := json.Unmarshal(msg.Data, &resp); err != nil {
		t.Fatalf("register unmarshal resp: %v", err)
	}
	if resp.Code != 1 || resp.NodeID == 0 {
		t.Fatalf("register failed: code=%d node_id=%d msg=%s", resp.Code, resp.NodeID, resp.Msg)
	}
	return resp.NodeID
}

func waitNodeIndex(t *testing.T, cm core.IConnectionManager, nodeID uint32, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if c, ok := cm.GetByNode(nodeID); ok && c != nil {
			return
		}
		time.Sleep(30 * time.Millisecond)
	}
	t.Fatalf("waitNodeIndex timeout for node_id=%d", nodeID)
}

func makeProcess(t *testing.T, cfg core.IConfig, handlers []core.ISubProcess) core.IProcess {
	base := process.NewPreRoutingProcess(nil).WithConfig(cfg)
	dp, err := process.NewDispatcherFromConfig(cfg, base, nil)
	if err != nil {
		t.Fatalf("dispatcher: %v", err)
	}
	for _, h := range handlers {
		if err := dp.RegisterHandler(h); err != nil {
			t.Fatalf("register handler: %v", err)
		}
	}
	dp.RegisterDefaultHandler(forward.NewDefaultForwardHandler(cfg, nil))
	return dp
}

func startTestServer(t *testing.T, opts server.Options) core.IServer {
	srv, err := server.New(opts)
	if err != nil {
		t.Fatalf("server new: %v", err)
	}
	ctx := context.Background()
	if err := srv.Start(ctx); err != nil {
		t.Fatalf("server start: %v", err)
	}
	return srv
}

func stopTestServer(t *testing.T, srv core.IServer) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := srv.Stop(ctx); err != nil {
		t.Logf("stop server: %v", err)
	}
}

func freeAddr() string {
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	addr := ln.Addr().String()
	_ = ln.Close()
	return addr
}

func waitListen(t *testing.T, addr string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		c, err := net.Dial("tcp", addr)
		if err == nil {
			_ = c.Close()
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("waitListen timeout for %s", addr)
}
