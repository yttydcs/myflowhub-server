package tests

import (
	"context"
	"encoding/json"
	"testing"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-core/connmgr"
	"github.com/yttydcs/myflowhub-core/header"
	"github.com/yttydcs/myflowhub-server/subproto/topicbus"
)

func TestTopicBusSubscribeListUnsubscribe(t *testing.T) {
	h := newTopicBusHandlerForTest()
	cm := connmgr.New()
	conn := newRecordConn("c1")
	conn.SetMeta("nodeID", uint32(10))
	_ = cm.Add(conn)
	srv := newRecordServer(1, cm)
	ctx := core.WithServerContext(context.Background(), srv)

	// subscribe
	subReq := mustJSON(map[string]any{"action": "subscribe", "data": map[string]any{"topic": "t1"}})
	hdr := (&header.HeaderTcp{}).WithMajor(header.MajorCmd).WithSubProto(topicbus.SubProtoTopicBus).WithSourceID(10).WithTargetID(1)
	h.OnReceive(ctx, conn, hdr, subReq)
	assertActionCode(t, conn, "subscribe_resp", 1)

	// list_subs
	conn.sent = nil
	listReq := mustJSON(map[string]any{"action": "list_subs", "data": map[string]any{}})
	h.OnReceive(ctx, conn, hdr, listReq)
	var msg struct {
		Action string          `json:"action"`
		Data   json.RawMessage `json:"data"`
	}
	if len(conn.sent) != 1 {
		t.Fatalf("expected 1 response, got %d", len(conn.sent))
	}
	_ = json.Unmarshal(conn.sent[0].payload, &msg)
	if msg.Action != "list_subs_resp" {
		t.Fatalf("unexpected action %s", msg.Action)
	}
	var body struct {
		Code   int      `json:"code"`
		Topics []string `json:"topics"`
	}
	_ = json.Unmarshal(msg.Data, &body)
	if body.Code != 1 || len(body.Topics) != 1 || body.Topics[0] != "t1" {
		t.Fatalf("unexpected list_subs_resp %+v", body)
	}

	// unsubscribe (idempotent)
	conn.sent = nil
	unsubReq := mustJSON(map[string]any{"action": "unsubscribe", "data": map[string]any{"topic": "t1"}})
	h.OnReceive(ctx, conn, hdr, unsubReq)
	assertActionCode(t, conn, "unsubscribe_resp", 1)

	// unsubscribe again should still ok
	conn.sent = nil
	h.OnReceive(ctx, conn, hdr, unsubReq)
	assertActionCode(t, conn, "unsubscribe_resp", 1)

	// list_subs empty
	conn.sent = nil
	h.OnReceive(ctx, conn, hdr, listReq)
	_ = json.Unmarshal(conn.sent[0].payload, &msg)
	_ = json.Unmarshal(msg.Data, &body)
	if body.Code != 1 || len(body.Topics) != 0 {
		t.Fatalf("expected empty list_subs, got %+v", body)
	}
}

func TestTopicBusUpstreamSubscribeAndUnsubscribeBatch(t *testing.T) {
	h := newTopicBusHandlerForTest()
	cm := connmgr.New()

	parent := newRecordConn("parent")
	parent.SetMeta(core.MetaRoleKey, core.RoleParent)
	parent.SetMeta("nodeID", uint32(99))
	_ = cm.Add(parent)

	child := newRecordConn("child")
	child.SetMeta("nodeID", uint32(10))
	_ = cm.Add(child)

	srv := newRecordServer(1, cm)
	ctx := core.WithServerContext(context.Background(), srv)

	hdr := (&header.HeaderTcp{}).WithMajor(header.MajorCmd).WithSubProto(topicbus.SubProtoTopicBus).WithSourceID(10).WithTargetID(1)

	// subscribe should trigger upstream subscribe_batch
	parent.sent = nil
	subReq := mustJSON(map[string]any{"action": "subscribe", "data": map[string]any{"topic": "t-up"}})
	h.OnReceive(ctx, child, hdr, subReq)
	if len(parent.sent) == 0 {
		t.Fatalf("expected upstream subscribe_batch to parent")
	}
	assertAction(t, parent.sent[len(parent.sent)-1].payload, "subscribe_batch")

	// unsubscribe should trigger upstream unsubscribe_batch
	parent.sent = nil
	unsubReq := mustJSON(map[string]any{"action": "unsubscribe", "data": map[string]any{"topic": "t-up"}})
	h.OnReceive(ctx, child, hdr, unsubReq)
	if len(parent.sent) == 0 {
		t.Fatalf("expected upstream unsubscribe_batch to parent")
	}
	assertAction(t, parent.sent[len(parent.sent)-1].payload, "unsubscribe_batch")
}

func TestTopicBusPublishFanoutNoEchoAndForwardUp(t *testing.T) {
	h := newTopicBusHandlerForTest()
	cm := connmgr.New()

	parent := newRecordConn("parent")
	parent.SetMeta(core.MetaRoleKey, core.RoleParent)
	parent.SetMeta("nodeID", uint32(99))
	_ = cm.Add(parent)

	sub := newRecordConn("sub")
	sub.SetMeta("nodeID", uint32(20))
	_ = cm.Add(sub)

	pub := newRecordConn("pub")
	pub.SetMeta("nodeID", uint32(10))
	_ = cm.Add(pub)

	srv := newRecordServer(1, cm)
	ctx := core.WithServerContext(context.Background(), srv)

	hdrSub := (&header.HeaderTcp{}).WithMajor(header.MajorCmd).WithSubProto(topicbus.SubProtoTopicBus).WithSourceID(20).WithTargetID(1)
	hdrPub := (&header.HeaderTcp{}).WithMajor(header.MajorCmd).WithSubProto(topicbus.SubProtoTopicBus).WithSourceID(10).WithTargetID(1)

	// sub + pub subscribe to topic t
	sub.sent = nil
	pub.sent = nil
	h.OnReceive(ctx, sub, hdrSub, mustJSON(map[string]any{"action": "subscribe", "data": map[string]any{"topic": "t"}}))
	h.OnReceive(ctx, pub, hdrPub, mustJSON(map[string]any{"action": "subscribe", "data": map[string]any{"topic": "t"}}))

	// publish from pub
	sub.sent = nil
	pub.sent = nil
	parent.sent = nil
	pubReq := mustJSON(map[string]any{
		"action": "publish",
		"data": map[string]any{
			"topic":   "t",
			"name":    "evt",
			"ts":      int64(1700000000000),
			"payload": map[string]any{"k": 1},
		},
	})
	h.OnReceive(ctx, pub, hdrPub, pubReq)

	// sub should receive one publish
	if len(sub.sent) == 0 {
		t.Fatalf("expected sub to receive publish")
	}
	assertAction(t, sub.sent[len(sub.sent)-1].payload, "publish")

	// pub should NOT receive publish (no echo)
	for _, f := range pub.sent {
		if isAction(f.payload, "publish") {
			t.Fatalf("expected no echo to publisher")
		}
	}

	// parent should receive publish forwarded upstream
	if len(parent.sent) == 0 {
		t.Fatalf("expected publish forwarded to parent")
	}
	assertAction(t, parent.sent[len(parent.sent)-1].payload, "publish")
}

func newTopicBusHandlerForTest() *topicbus.TopicBusHandler {
	h := topicbus.NewTopicBusHandlerWithConfig(nil, nil)
	h.Init()
	return h
}

func assertActionCode(t *testing.T, c *recordConn, wantAction string, wantCode int) {
	t.Helper()
	if len(c.sent) == 0 {
		t.Fatalf("expected response %s", wantAction)
	}
	var msg struct {
		Action string          `json:"action"`
		Data   json.RawMessage `json:"data"`
	}
	_ = json.Unmarshal(c.sent[len(c.sent)-1].payload, &msg)
	if msg.Action != wantAction {
		t.Fatalf("unexpected action %s", msg.Action)
	}
	var body struct {
		Code int `json:"code"`
	}
	_ = json.Unmarshal(msg.Data, &body)
	if body.Code != wantCode {
		t.Fatalf("expected code %d, got %d", wantCode, body.Code)
	}
}

func assertAction(t *testing.T, payload []byte, want string) {
	t.Helper()
	if !isAction(payload, want) {
		var msg struct {
			Action string `json:"action"`
		}
		_ = json.Unmarshal(payload, &msg)
		t.Fatalf("expected action %s, got %s", want, msg.Action)
	}
}

func isAction(payload []byte, want string) bool {
	var msg struct {
		Action string `json:"action"`
	}
	if err := json.Unmarshal(payload, &msg); err != nil {
		return false
	}
	return msg.Action == want
}
