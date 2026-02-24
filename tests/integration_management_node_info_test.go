package tests

import (
	"encoding/json"
	"net"
	"testing"
	"time"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-core/config"
	"github.com/yttydcs/myflowhub-core/connmgr"
	"github.com/yttydcs/myflowhub-core/header"
	"github.com/yttydcs/myflowhub-core/listener/tcp_listener"
	"github.com/yttydcs/myflowhub-core/server"
	"github.com/yttydcs/myflowhub-subproto/management"
)

func TestManagementNodeInfo(t *testing.T) {
	addr := freeAddr()
	cfg := config.NewMap(map[string]string{"addr": addr})

	handlers := []core.ISubProcess{management.NewHandler(nil)}
	srv := startTestServer(t, server.Options{
		Name:     "NodeInfo",
		Process:  makeProcess(t, cfg, handlers),
		Codec:    header.HeaderTcpCodec{},
		Listener: tcp_listener.New(addr),
		Config:   cfg,
		Manager:  connmgr.New(),
		NodeID:   1,
	})
	defer stopTestServer(t, srv)
	waitListen(t, addr, 2*time.Second)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// 绑定服务端对应的连接 nodeID，模拟已登录状态以满足 Source 校验。
	clientAddr := conn.LocalAddr().String()
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		found := false
		srv.ConnManager().Range(func(c core.IConnection) bool {
			if c.RemoteAddr() != nil && c.RemoteAddr().String() == clientAddr {
				c.SetMeta("nodeID", uint32(1))
				found = true
				return false
			}
			return true
		})
		if found {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	codec := header.HeaderTcpCodec{}
	payload := mustJSON(map[string]any{
		"action": "node_info",
		"data":   map[string]any{},
	})
	hdr := (&header.HeaderTcp{}).
		WithMajor(header.MajorCmd).
		WithSubProto(management.SubProtoManagement).
		WithSourceID(1).
		WithTargetID(1).
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
	if msg.Action != "node_info_resp" {
		t.Fatalf("unexpected action: %s", msg.Action)
	}
	var resp struct {
		Code  int               `json:"code"`
		Msg   string            `json:"msg"`
		Items map[string]string `json:"items"`
	}
	_ = json.Unmarshal(msg.Data, &resp)
	if resp.Code != 1 {
		t.Fatalf("unexpected resp: %+v", resp)
	}
	if resp.Items == nil || resp.Items["platform"] == "" {
		t.Fatalf("missing platform in resp: %+v", resp)
	}
	if resp.Items["node_id"] != "1" {
		t.Fatalf("unexpected node_id: %q", resp.Items["node_id"])
	}
}
