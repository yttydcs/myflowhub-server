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
	streamproto "github.com/yttydcs/myflowhub-server/protocol/stream"
	streamhandler "github.com/yttydcs/myflowhub-subproto/stream"
)

func TestStreamRootHubConnectDisconnect(t *testing.T) {
	rootAddr := freeAddr()
	hubAddr := freeAddr()
	controllerNodeID := uint32(101)
	producerClientNodeID := uint32(201)
	rolePerms := "superadmin:*"

	roleMap := "1:superadmin;2:superadmin;101:superadmin;201:superadmin"

	rootCfg := config.NewMap(map[string]string{
		"addr":                       rootAddr,
		config.KeyAuthNodeRoles:      roleMap,
		config.KeyAuthRolePerms:      rolePerms,
		config.KeyProcWorkersPerChan: "2",
	})
	rootSrv := startTestServer(t, server.Options{
		Name:     "RootStream",
		Process:  makeProcess(t, rootCfg, []core.ISubProcess{streamhandler.NewHandlerWithConfig(rootCfg, nil)}),
		Codec:    header.HeaderTcpCodec{},
		Listener: tcp_listener.New(rootAddr),
		Config:   rootCfg,
		Manager:  connmgr.New(),
		NodeID:   1,
	})
	defer stopTestServer(t, rootSrv)
	waitListen(t, rootAddr, 2*time.Second)

	hubCfg := config.NewMap(map[string]string{
		"addr":                       hubAddr,
		config.KeyParentEnable:       "true",
		config.KeyParentAddr:         rootAddr,
		config.KeyAuthNodeRoles:      roleMap,
		config.KeyAuthRolePerms:      rolePerms,
		config.KeyProcWorkersPerChan: "2",
	})
	hubSrv := startTestServer(t, server.Options{
		Name:     "HubStream",
		Process:  makeProcess(t, hubCfg, []core.ISubProcess{streamhandler.NewHandlerWithConfig(hubCfg, nil)}),
		Codec:    header.HeaderTcpCodec{},
		Listener: tcp_listener.New(hubAddr),
		Config:   hubCfg,
		Manager:  connmgr.New(),
		NodeID:   2,
	})
	defer stopTestServer(t, hubSrv)
	waitListen(t, hubAddr, 2*time.Second)
	bindParentChildNodeIDs(t, rootSrv, hubSrv, 2, 1)

	producerConn, err := net.Dial("tcp", hubAddr)
	if err != nil {
		t.Fatalf("dial hub: %v", err)
	}
	defer producerConn.Close()
	bindConnNodeID(t, hubSrv, producerConn, producerClientNodeID)

	controllerConn, err := net.Dial("tcp", rootAddr)
	if err != nil {
		t.Fatalf("dial root: %v", err)
	}
	defer controllerConn.Close()
	bindConnNodeID(t, rootSrv, controllerConn, controllerNodeID)

	codec := header.HeaderTcpCodec{}

	var announceSourceResp streamproto.AnnounceResp
	sendStreamCtrlExpect(t, producerConn, codec, streamCtrlHeader(1, producerClientNodeID, 2), streamproto.ActionAnnounce, streamproto.AnnounceReq{
		ReqID: "announce-source-1",
		Source: streamproto.SourceDescriptor{
			SourceID: "source-text-1",
			Producer: 2,
			Name:     "Text Feed",
			Kind:     streamproto.StreamKindText,
			Mode:     streamproto.ModeLive,
			UnitMode: streamproto.UnitModeChunk,
		},
	}, streamproto.ActionAnnounceResp, &announceSourceResp)
	if announceSourceResp.Code != 1 || announceSourceResp.Source == nil || announceSourceResp.Source.SourceID != "source-text-1" {
		t.Fatalf("unexpected announce source resp: %+v", announceSourceResp)
	}

	var announceConsumerResp streamproto.AnnounceConsumerResp
	sendStreamCtrlExpect(t, controllerConn, codec, streamCtrlHeader(2, controllerNodeID, 1), streamproto.ActionAnnounceConsumer, streamproto.AnnounceConsumerReq{
		ReqID: "announce-consumer-1",
		ConsumerEndpoint: streamproto.ConsumerDescriptor{
			ConsumerID: "consumer-text-1",
			Consumer:   1,
			Name:       "Root Text Sink",
			Kind:       streamproto.StreamKindText,
		},
	}, streamproto.ActionAnnounceConsumerResp, &announceConsumerResp)
	if announceConsumerResp.Code != 1 || announceConsumerResp.ConsumerEndpoint == nil || announceConsumerResp.ConsumerEndpoint.ConsumerID != "consumer-text-1" {
		t.Fatalf("unexpected announce consumer resp: %+v", announceConsumerResp)
	}

	var listResp streamproto.ListSourcesResp
	sendStreamCtrlExpect(t, controllerConn, codec, streamCtrlHeader(3, controllerNodeID, 1), streamproto.ActionListSources, streamproto.ListSourcesReq{
		ReqID:    "list-source-1",
		Producer: 2,
		Kind:     streamproto.StreamKindText,
	}, streamproto.ActionListSourcesResp, &listResp)
	if listResp.Code != 1 || len(listResp.Sources) != 1 || listResp.Sources[0].SourceID != "source-text-1" {
		t.Fatalf("unexpected list sources resp: %+v", listResp)
	}

	var connectResp streamproto.ConnectResp
	sendStreamCtrlExpect(t, controllerConn, codec, streamCtrlHeader(4, controllerNodeID, 1), streamproto.ActionConnect, streamproto.ConnectReq{
		ReqID:      "connect-1",
		Producer:   2,
		SourceID:   "source-text-1",
		Consumer:   1,
		ConsumerID: "consumer-text-1",
	}, streamproto.ActionConnectResp, &connectResp)
	if connectResp.Code != 1 || !connectResp.Accept || connectResp.DeliveryID == "" {
		t.Fatalf("unexpected connect resp: %+v", connectResp)
	}
	if connectResp.Source == nil || connectResp.Source.Producer != 2 || connectResp.ConsumerEndpoint == nil || connectResp.ConsumerEndpoint.Consumer != 1 {
		t.Fatalf("unexpected connect descriptors: %+v", connectResp)
	}

	var disconnectResp streamproto.DisconnectResp
	sendStreamCtrlExpect(t, controllerConn, codec, streamCtrlHeader(5, controllerNodeID, 1), streamproto.ActionDisconnect, streamproto.DisconnectReq{
		ReqID:      "disconnect-1",
		DeliveryID: connectResp.DeliveryID,
		Reason:     "test cleanup",
	}, streamproto.ActionDisconnectResp, &disconnectResp)
	if disconnectResp.Code != 1 || disconnectResp.DeliveryID != connectResp.DeliveryID {
		t.Fatalf("unexpected disconnect resp: %+v", disconnectResp)
	}

	var disconnectAgainResp streamproto.DisconnectResp
	sendStreamCtrlExpect(t, controllerConn, codec, streamCtrlHeader(6, controllerNodeID, 1), streamproto.ActionDisconnect, streamproto.DisconnectReq{
		ReqID:      "disconnect-2",
		DeliveryID: connectResp.DeliveryID,
		Reason:     "already closed",
	}, streamproto.ActionDisconnectResp, &disconnectAgainResp)
	if disconnectAgainResp.Code != 404 {
		t.Fatalf("expected second disconnect to return 404, got %+v", disconnectAgainResp)
	}
}

func streamCtrlHeader(msgID uint32, sourceID, targetID uint32) core.IHeader {
	return (&header.HeaderTcp{}).
		WithMajor(header.MajorCmd).
		WithSubProto(streamproto.SubProtoStream).
		WithSourceID(sourceID).
		WithTargetID(targetID).
		WithMsgID(msgID)
}

func sendStreamCtrlExpect(t *testing.T, conn net.Conn, codec header.HeaderTcpCodec, hdr core.IHeader, action string, data any, wantAction string, out any) {
	t.Helper()

	payload := mustStreamCtrlPayload(t, action, data)
	frame, err := codec.Encode(hdr.WithPayloadLength(uint32(len(payload))), payload)
	if err != nil {
		t.Fatalf("encode %s: %v", action, err)
	}
	if _, err := conn.Write(frame); err != nil {
		t.Fatalf("write %s: %v", action, err)
	}

	_, respPayload, err := codec.Decode(conn)
	if err != nil {
		t.Fatalf("decode %s: %v", action, err)
	}
	gotAction, raw := decodeStreamCtrlPayload(t, respPayload)
	if gotAction != wantAction {
		t.Fatalf("unexpected action for %s: got %s want %s", action, gotAction, wantAction)
	}
	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			t.Fatalf("unmarshal %s resp: %v", action, err)
		}
	}
}

func mustStreamCtrlPayload(t *testing.T, action string, data any) []byte {
	t.Helper()

	var raw json.RawMessage
	if data != nil {
		raw = mustJSON(data)
	}
	body, err := json.Marshal(streamproto.Message{Action: action, Data: raw})
	if err != nil {
		t.Fatalf("marshal ctrl payload %s: %v", action, err)
	}

	payload := make([]byte, 1+len(body))
	payload[0] = streamproto.KindCtrl
	copy(payload[1:], body)
	return payload
}

func decodeStreamCtrlPayload(t *testing.T, payload []byte) (string, json.RawMessage) {
	t.Helper()

	if len(payload) == 0 || payload[0] != streamproto.KindCtrl {
		t.Fatalf("unexpected stream ctrl payload prefix: %v", payload)
	}

	var msg streamproto.Message
	if err := json.Unmarshal(payload[1:], &msg); err != nil {
		t.Fatalf("unmarshal stream ctrl payload: %v", err)
	}
	return msg.Action, msg.Data
}
