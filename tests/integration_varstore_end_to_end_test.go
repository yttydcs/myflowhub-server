package tests

import (
	"context"
	"net"
	"testing"
	"time"

	"encoding/json"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-core/bootstrap"
	"github.com/yttydcs/myflowhub-core/config"
	"github.com/yttydcs/myflowhub-core/connmgr"
	"github.com/yttydcs/myflowhub-core/header"
	"github.com/yttydcs/myflowhub-core/listener/tcp_listener"
	"github.com/yttydcs/myflowhub-core/server"
	"github.com/yttydcs/myflowhub-server/internal/handler"
)

// Integration: Root (node=1) + Hub (self-register) + set at hub, get at root.
func TestIntegrationVarStoreSetGetAcrossHub(t *testing.T) {
	rootAddr := freeAddr()
	hubAddr := freeAddr()

	rootCfg := config.NewMap(map[string]string{"addr": rootAddr})
	rootSrv := startTestServer(t, server.Options{
		Name:     "Root",
		Process:  makeProcess(t, rootCfg, []core.ISubProcess{handler.NewLoginHandlerWithConfig(rootCfg, nil), handler.NewVarStoreHandlerWithConfig(rootCfg, nil)}),
		Codec:    header.HeaderTcpCodec{},
		Listener: tcp_listener.New(rootAddr),
		Config:   rootCfg,
		Manager:  connmgr.New(),
		NodeID:   1,
	})
	defer stopTestServer(t, rootSrv)
	waitListen(t, rootAddr, 2*time.Second)

	// hub self-register to root
	hubNodeID, _, err := bootstrap.SelfRegister(context.Background(), bootstrap.SelfRegisterOptions{
		ParentAddr: rootAddr,
		SelfID:     "hub-varstore",
		Timeout:    5 * time.Second,
		DoLogin:    false,
	})
	if err != nil || hubNodeID == 0 {
		t.Fatalf("self register hub: %v id=%d", err, hubNodeID)
	}

	hubCfg := config.NewMap(map[string]string{
		"addr":                 hubAddr,
		config.KeyParentEnable: "true",
		config.KeyParentAddr:   rootAddr,
	})
	hubSrv := startTestServer(t, server.Options{
		Name:     "Hub",
		Process:  makeProcess(t, hubCfg, []core.ISubProcess{handler.NewLoginHandlerWithConfig(hubCfg, nil), handler.NewVarStoreHandlerWithConfig(hubCfg, nil)}),
		Codec:    header.HeaderTcpCodec{},
		Listener: tcp_listener.New(hubAddr),
		Config:   hubCfg,
		Manager:  connmgr.New(),
		NodeID:   hubNodeID,
	})
	defer stopTestServer(t, hubSrv)
	waitListen(t, hubAddr, 2*time.Second)

	// set from hub connection（请求方直连父节点=hub）
	setConn, err := net.Dial("tcp", hubAddr)
	if err != nil {
		t.Fatalf("dial hub for set: %v", err)
	}
	defer setConn.Close()
	setCodec := header.HeaderTcpCodec{}
	setPayload := mustJSON(map[string]any{
		"action": "set",
		"data": map[string]any{
			"name":       "temp",
			"value":      "22.5",
			"visibility": "public",
			"type":       "string",
		},
	})
	setHdr := (&header.HeaderTcp{}).
		WithMajor(header.MajorCmd).
		WithSubProto(3).
		WithSourceID(hubNodeID).
		WithTargetID(hubNodeID).
		WithPayloadLength(uint32(len(setPayload)))
	frame, _ := setCodec.Encode(setHdr, setPayload)
	if _, err := setConn.Write(frame); err != nil {
		t.Fatalf("send set: %v", err)
	}
	// read set_resp
	if _, _, err := setCodec.Decode(setConn); err != nil {
		t.Fatalf("read set_resp: %v", err)
	}

	// allow propagation
	time.Sleep(200 * time.Millisecond)

	// get from root
	getConn, err := net.Dial("tcp", rootAddr)
	if err != nil {
		t.Fatalf("dial root: %v", err)
	}
	defer getConn.Close()
	getCodec := header.HeaderTcpCodec{}
	getPayload := mustJSON(map[string]any{
		"action": "get",
		"data":   map[string]any{"name": "temp"},
	})
	getHdr := (&header.HeaderTcp{}).
		WithMajor(header.MajorCmd).
		WithSubProto(3).
		WithSourceID(hubNodeID).
		WithTargetID(1).
		WithPayloadLength(uint32(len(getPayload)))
	gframe, _ := getCodec.Encode(getHdr, getPayload)
	if _, err := getConn.Write(gframe); err != nil {
		t.Fatalf("send get: %v", err)
	}
	_, respPayload, err := getCodec.Decode(getConn)
	if err != nil {
		t.Fatalf("decode get_resp: %v", err)
	}
	var msg struct {
		Action string          `json:"action"`
		Data   json.RawMessage `json:"data"`
	}
	_ = json.Unmarshal(respPayload, &msg)
	if msg.Action != "get_resp" {
		t.Fatalf("unexpected action %s", msg.Action)
	}
	var data struct {
		Code  int    `json:"code"`
		Value string `json:"value"`
	}
	_ = json.Unmarshal(msg.Data, &data)
	if data.Code != 1 || data.Value != "22.5" {
		t.Fatalf("unexpected resp %+v", data)
	}
}
