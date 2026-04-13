package hubruntime

// Context: This file lives in the Server assembly layer and supports runtime_test.

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync/atomic"
	"testing"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-core/config"
	"github.com/yttydcs/myflowhub-core/header"
	"github.com/yttydcs/myflowhub-core/listener/tcp_listener"
)

func TestSendRegisterOnConnIncludesDisplayNameAndJoinPermit(t *testing.T) {
	conn := &registerTestConn{id: "parent"}
	var seq atomic.Uint32

	if err := sendRegisterOnConn(context.Background(), conn, "device-1", "  Runtime Hub  ", "  permit-1  ", &seq); err != nil {
		t.Fatalf("sendRegisterOnConn: %v", err)
	}
	if len(conn.sent) != 1 {
		t.Fatalf("expected 1 sent frame, got %d", len(conn.sent))
	}

	if conn.sent[0].hdr.Major() != header.MajorCmd {
		t.Fatalf("unexpected major: got %d want %d", conn.sent[0].hdr.Major(), header.MajorCmd)
	}
	if conn.sent[0].hdr.SubProto() != 2 {
		t.Fatalf("unexpected subproto: got %d want 2", conn.sent[0].hdr.SubProto())
	}
	if conn.sent[0].hdr.GetMsgID() != 1 {
		t.Fatalf("unexpected msg id: got %d want 1", conn.sent[0].hdr.GetMsgID())
	}

	msg := decodeRegisterPayload(t, conn.sent[0].payload)
	if msg.Action != "register" {
		t.Fatalf("unexpected action: got %q want %q", msg.Action, "register")
	}
	if got := msg.Data["device_id"]; got != "device-1" {
		t.Fatalf("unexpected device_id: got %v want %q", got, "device-1")
	}
	if got := msg.Data["display_name"]; got != "Runtime Hub" {
		t.Fatalf("unexpected display_name: got %v want %q", got, "Runtime Hub")
	}
	if got := msg.Data["join_permit"]; got != "permit-1" {
		t.Fatalf("unexpected join_permit: got %v want %q", got, "permit-1")
	}
}

func TestSendRegisterOnConnOmitsBlankDisplayName(t *testing.T) {
	conn := &registerTestConn{id: "parent"}
	var seq atomic.Uint32

	if err := sendRegisterOnConn(context.Background(), conn, "device-2", " \t ", " \t ", &seq); err != nil {
		t.Fatalf("sendRegisterOnConn: %v", err)
	}
	if len(conn.sent) != 1 {
		t.Fatalf("expected 1 sent frame, got %d", len(conn.sent))
	}

	msg := decodeRegisterPayload(t, conn.sent[0].payload)
	if got := msg.Data["device_id"]; got != "device-2" {
		t.Fatalf("unexpected device_id: got %v want %q", got, "device-2")
	}
	if _, ok := msg.Data["display_name"]; ok {
		t.Fatalf("display_name should be omitted for blank input")
	}
	if _, ok := msg.Data["join_permit"]; ok {
		t.Fatalf("join_permit should be omitted for blank input")
	}
}

func TestTrimmedConfigValue(t *testing.T) {
	cfg := config.NewMap(map[string]string{
		"node.display_name": "  Persisted Hub  ",
	})

	if got := trimmedConfigValue(cfg, "node.display_name"); got != "Persisted Hub" {
		t.Fatalf("unexpected trimmed value: got %q want %q", got, "Persisted Hub")
	}
	if got := trimmedConfigValue(cfg, "missing"); got != "" {
		t.Fatalf("unexpected missing value: got %q want empty", got)
	}
}

func TestSelfRegisterNodeIDUsesInjectedDialerForGenericTarget(t *testing.T) {
	client, server := net.Pipe()
	defer server.Close()

	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		defer server.Close()

		codec := header.HeaderTcpCodec{}
		reqHdr, reqBody, err := codec.Decode(server)
		if err != nil {
			errCh <- err
			return
		}
		var msg registerPayload
		if err := json.Unmarshal(reqBody, &msg); err != nil {
			errCh <- err
			return
		}
		if got := msg.Action; got != "register" {
			errCh <- fmt.Errorf("unexpected action: got %q want %q", got, "register")
			return
		}
		if got := msg.Data["device_id"]; got != "hub-quic" {
			errCh <- fmt.Errorf("unexpected device_id: got %v want %q", got, "hub-quic")
			return
		}

		respPayload, err := json.Marshal(map[string]any{
			"action": "register_resp",
			"data": map[string]any{
				"code":    1,
				"status":  "approved",
				"node_id": 12,
			},
		})
		if err != nil {
			errCh <- err
			return
		}
		respHdr := (&header.HeaderTcp{}).
			WithMajor(header.MajorCmd).
			WithSubProto(2).
			WithTargetID(reqHdr.SourceID()).
			WithMsgID(reqHdr.GetMsgID())
		frame, err := codec.Encode(respHdr, respPayload)
		if err != nil {
			errCh <- err
			return
		}
		_, err = server.Write(frame)
		errCh <- err
	}()

	const target = "quic://parent.example:9000?insecure=true"
	var gotTarget string
	nodeID, err := selfRegisterNodeID(
		context.Background(),
		target,
		"hub-quic",
		"",
		func(_ context.Context, actualTarget string) (core.IConnection, error) {
			gotTarget = actualTarget
			return tcp_listener.NewTCPConnection(client), nil
		},
		nil,
	)
	if err != nil {
		t.Fatalf("selfRegisterNodeID: %v", err)
	}
	if gotTarget != target {
		t.Fatalf("unexpected dial target: got %q want %q", gotTarget, target)
	}
	if nodeID != 12 {
		t.Fatalf("unexpected node id: got %d want %d", nodeID, 12)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("bootstrap server: %v", err)
	}
}

type registerPayload struct {
	Action string         `json:"action"`
	Data   map[string]any `json:"data"`
}

func decodeRegisterPayload(t *testing.T, payload []byte) registerPayload {
	t.Helper()

	var msg registerPayload
	if err := json.Unmarshal(payload, &msg); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	return msg
}

type registerTestPipe struct{}

func (registerTestPipe) Read([]byte) (int, error)    { return 0, nil }
func (registerTestPipe) Write(p []byte) (int, error) { return len(p), nil }
func (registerTestPipe) Close() error                { return nil }

type registerTestAddr struct{}

func (registerTestAddr) Network() string { return "tcp" }
func (registerTestAddr) String() string  { return "127.0.0.1:0" }

type registerSentFrame struct {
	hdr     core.IHeader
	payload []byte
}

type registerTestConn struct {
	id   string
	meta map[string]any
	sent []registerSentFrame
}

var _ core.IConnection = (*registerTestConn)(nil)

func (c *registerTestConn) ID() string                    { return c.id }
func (c *registerTestConn) Pipe() core.IPipe              { return registerTestPipe{} }
func (c *registerTestConn) Close() error                  { return nil }
func (c *registerTestConn) OnReceive(core.ReceiveHandler) {}
func (c *registerTestConn) SetMeta(key string, val any) {
	if c.meta == nil {
		c.meta = make(map[string]any)
	}
	c.meta[key] = val
}
func (c *registerTestConn) GetMeta(key string) (any, bool)       { v, ok := c.meta[key]; return v, ok }
func (c *registerTestConn) Metadata() map[string]any             { return c.meta }
func (c *registerTestConn) LocalAddr() net.Addr                  { return registerTestAddr{} }
func (c *registerTestConn) RemoteAddr() net.Addr                 { return registerTestAddr{} }
func (c *registerTestConn) Reader() core.IReader                 { return nil }
func (c *registerTestConn) SetReader(core.IReader)               {}
func (c *registerTestConn) DispatchReceive(core.IHeader, []byte) {}
func (c *registerTestConn) Send([]byte) error                    { return nil }
func (c *registerTestConn) SendWithHeader(h core.IHeader, payload []byte, _ core.IHeaderCodec) error {
	c.sent = append(c.sent, registerSentFrame{hdr: h, payload: payload})
	return nil
}
