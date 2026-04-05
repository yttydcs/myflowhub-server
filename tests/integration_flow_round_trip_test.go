package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-core/header"
	protocolflow "github.com/yttydcs/myflowhub-proto/protocol/flow"
	"github.com/yttydcs/myflowhub-server/hubruntime"
)

func TestIntegrationFlowRoundTrip(t *testing.T) {
	addr := freeAddr()

	rt, err := hubruntime.New(hubruntime.Options{
		TCPEnable:        true,
		Addr:             addr,
		NodeID:           1,
		WorkDir:          t.TempDir(),
		AuthDefaultRole:  "superadmin",
		AuthDefaultPerms: "*",
		AuthRolePerms:    "superadmin:*",
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
	waitListen(t, addr, 2*time.Second)

	executorNodeID := rt.Status().NodeID
	if executorNodeID == 0 {
		t.Fatalf("runtime node id 0")
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial runtime: %v", err)
	}
	defer conn.Close()

	codec := header.HeaderTcpCodec{}
	clientNodeID := registerOnConn(t, conn, codec, "flow-roundtrip-client")

	msgID := uint32(10)
	nextHdr := func() core.IHeader {
		hdr := flowCtrlHeader(msgID, clientNodeID, executorNodeID)
		msgID++
		return hdr
	}

	flowID := "123e4567-e89b-12d3-a456-426614174201"
	composeSpec := json.RawMessage(mustJSON(map[string]any{
		"template": map[string]any{
			"payload": map[string]any{
				"source": "integration",
			},
			"meta": map[string]any{},
		},
		"inputs": []map[string]any{
			{
				"to": "/meta/flow_id",
				"source": map[string]any{
					"kind":  "flow_meta",
					"field": "flow_id",
				},
				"required": true,
			},
			{
				"to": "/meta/run_id",
				"source": map[string]any{
					"kind":  "run_meta",
					"field": "run_id",
				},
				"required": true,
			},
		},
	}))
	setReq := protocolflow.SetReq{
		ReqID:  "set-flow-roundtrip",
		FlowID: flowID,
		Name:   "compose-roundtrip",
		Trigger: protocolflow.Trigger{
			Type:    "interval",
			EveryMs: 86400000,
		},
		Graph: protocolflow.Graph{
			Nodes: []protocolflow.Node{
				{
					ID:   "compose",
					Kind: "compose",
					Spec: composeSpec,
				},
			},
		},
	}

	var setResp protocolflow.SetResp
	sendFlowCtrlExpect(t, conn, codec, nextHdr(), protocolflow.ActionSet, setReq, protocolflow.ActionSetResp, &setResp)
	if setResp.Code != 1 || setResp.FlowID != flowID {
		t.Fatalf("unexpected set resp: %+v", setResp)
	}

	run1, status1 := runFlowAndWaitTerminal(t, conn, codec, nextHdr, flowID, "run-1")
	if status1.Status != "succeeded" {
		t.Fatalf("expected run1 succeeded, got %+v", status1)
	}
	assertNodeStatus(t, status1.Nodes, "compose", "succeeded")

	run2, status2 := runFlowAndWaitTerminal(t, conn, codec, nextHdr, flowID, "run-2")
	if status2.Status != "succeeded" {
		t.Fatalf("expected run2 succeeded, got %+v", status2)
	}
	assertNodeStatus(t, status2.Nodes, "compose", "succeeded")

	var latestStatus protocolflow.StatusResp
	sendFlowCtrlExpect(t, conn, codec, nextHdr(), protocolflow.ActionStatus, protocolflow.StatusReq{
		ReqID:  "status-latest",
		FlowID: flowID,
	}, protocolflow.ActionStatusResp, &latestStatus)
	if latestStatus.Code != 1 || latestStatus.RunID != run2.RunID || latestStatus.Status != "succeeded" {
		t.Fatalf("unexpected latest status resp: %+v", latestStatus)
	}
	assertNodeStatus(t, latestStatus.Nodes, "compose", "succeeded")

	var detailResp protocolflow.DetailResp
	sendFlowCtrlExpect(t, conn, codec, nextHdr(), protocolflow.ActionDetail, protocolflow.DetailReq{
		ReqID:  "detail-run-2",
		FlowID: flowID,
		RunID:  run2.RunID,
		NodeID: "compose",
	}, protocolflow.ActionDetailResp, &detailResp)
	if detailResp.Code != 1 || detailResp.RunID != run2.RunID || detailResp.Node == nil || detailResp.Node.ID != "compose" {
		t.Fatalf("unexpected detail resp: %+v", detailResp)
	}
	assertJSONSemanticEqual(t, detailResp.Result, map[string]any{
		"payload": map[string]any{
			"source": "integration",
		},
		"meta": map[string]any{
			"flow_id": flowID,
			"run_id":  run2.RunID,
		},
	})

	var listRunsResp protocolflow.ListRunsResp
	sendFlowCtrlExpect(t, conn, codec, nextHdr(), protocolflow.ActionListRuns, protocolflow.ListRunsReq{
		ReqID:  "list-runs-roundtrip",
		FlowID: flowID,
		Limit:  2,
	}, protocolflow.ActionListRunsResp, &listRunsResp)
	if listRunsResp.Code != 1 || len(listRunsResp.Runs) != 2 {
		t.Fatalf("unexpected list_runs resp: %+v", listRunsResp)
	}
	if listRunsResp.Runs[0].RunID != run2.RunID || listRunsResp.Runs[1].RunID != run1.RunID {
		t.Fatalf("unexpected run order: %+v", listRunsResp.Runs)
	}
	if listRunsResp.Runs[0].Status != "succeeded" || listRunsResp.Runs[1].Status != "succeeded" {
		t.Fatalf("expected retained runs to succeed, got %+v", listRunsResp.Runs)
	}
	if listRunsResp.Runs[0].StartedAtMs == 0 || listRunsResp.Runs[0].EndedAtMs == 0 {
		t.Fatalf("expected latest run timestamps, got %+v", listRunsResp.Runs[0])
	}
	if listRunsResp.Runs[1].StartedAtMs == 0 || listRunsResp.Runs[1].EndedAtMs == 0 {
		t.Fatalf("expected previous run timestamps, got %+v", listRunsResp.Runs[1])
	}

	var getResp protocolflow.GetResp
	sendFlowCtrlExpect(t, conn, codec, nextHdr(), protocolflow.ActionGet, protocolflow.GetReq{
		ReqID:  "get-roundtrip",
		FlowID: flowID,
	}, protocolflow.ActionGetResp, &getResp)
	if getResp.Code != 1 || getResp.FlowID != flowID || getResp.Name != setReq.Name {
		t.Fatalf("unexpected get resp: %+v", getResp)
	}
	if getResp.Trigger.Type != setReq.Trigger.Type || getResp.Trigger.EveryMs != setReq.Trigger.EveryMs {
		t.Fatalf("unexpected get trigger: %+v", getResp.Trigger)
	}
	if len(getResp.Graph.Nodes) != 1 || getResp.Graph.Nodes[0].ID != "compose" || getResp.Graph.Nodes[0].Kind != "compose" {
		t.Fatalf("unexpected get graph nodes: %+v", getResp.Graph.Nodes)
	}
	assertJSONSemanticEqual(t, getResp.Graph.Nodes[0].Spec, map[string]any{
		"template": map[string]any{
			"payload": map[string]any{
				"source": "integration",
			},
			"meta": map[string]any{},
		},
		"inputs": []map[string]any{
			{
				"to": "/meta/flow_id",
				"source": map[string]any{
					"kind":  "flow_meta",
					"field": "flow_id",
				},
				"required": true,
			},
			{
				"to": "/meta/run_id",
				"source": map[string]any{
					"kind":  "run_meta",
					"field": "run_id",
				},
				"required": true,
			},
		},
	})

	var listResp protocolflow.ListResp
	sendFlowCtrlExpect(t, conn, codec, nextHdr(), protocolflow.ActionList, protocolflow.ListReq{
		ReqID: "list-roundtrip",
	}, protocolflow.ActionListResp, &listResp)
	if listResp.Code != 1 {
		t.Fatalf("unexpected list resp: %+v", listResp)
	}
	summary := findFlowSummary(t, listResp.Flows, flowID)
	if summary.Name != setReq.Name || summary.LastRunID != run2.RunID || summary.LastStatus != "succeeded" {
		t.Fatalf("unexpected flow summary: %+v", summary)
	}
}

func flowCtrlHeader(msgID, sourceID, targetID uint32) core.IHeader {
	return (&header.HeaderTcp{}).
		WithMajor(header.MajorCmd).
		WithSubProto(protocolflow.SubProtoFlow).
		WithSourceID(sourceID).
		WithTargetID(targetID).
		WithMsgID(msgID)
}

func sendFlowCtrlExpect(t *testing.T, conn net.Conn, codec header.HeaderTcpCodec, hdr core.IHeader, action string, data any, wantAction string, out any) core.IHeader {
	t.Helper()

	var raw json.RawMessage
	if data != nil {
		raw = mustJSON(data)
	}
	payload := mustJSON(protocolflow.Message{
		Action: action,
		Data:   raw,
	})
	frame, err := codec.Encode(hdr.WithPayloadLength(uint32(len(payload))), payload)
	if err != nil {
		t.Fatalf("encode %s: %v", action, err)
	}
	if _, err := conn.Write(frame); err != nil {
		t.Fatalf("write %s: %v", action, err)
	}

	respHdr, respPayload, err := codec.Decode(conn)
	if err != nil {
		t.Fatalf("decode %s: %v", action, err)
	}
	if respHdr.Major() != header.MajorOKResp {
		t.Fatalf("unexpected %s resp major=%d", action, respHdr.Major())
	}
	if respHdr.SubProto() != protocolflow.SubProtoFlow {
		t.Fatalf("unexpected %s resp subproto=%d", action, respHdr.SubProto())
	}
	if respHdr.TargetID() != hdr.SourceID() {
		t.Fatalf("unexpected %s resp target=%d want=%d", action, respHdr.TargetID(), hdr.SourceID())
	}

	var env protocolflow.Message
	if err := json.Unmarshal(respPayload, &env); err != nil {
		t.Fatalf("unmarshal %s envelope: %v", action, err)
	}
	if env.Action != wantAction {
		t.Fatalf("unexpected action for %s: got %s want %s", action, env.Action, wantAction)
	}
	if out != nil {
		if err := json.Unmarshal(env.Data, out); err != nil {
			t.Fatalf("unmarshal %s resp: %v", action, err)
		}
	}
	return respHdr
}

func runFlowAndWaitTerminal(t *testing.T, conn net.Conn, codec header.HeaderTcpCodec, nextHdr func() core.IHeader, flowID, reqPrefix string) (protocolflow.RunResp, protocolflow.StatusResp) {
	t.Helper()

	var runResp protocolflow.RunResp
	sendFlowCtrlExpect(t, conn, codec, nextHdr(), protocolflow.ActionRun, protocolflow.RunReq{
		ReqID:  reqPrefix,
		FlowID: flowID,
	}, protocolflow.ActionRunResp, &runResp)
	if runResp.Code != 1 || runResp.FlowID != flowID || runResp.RunID == "" {
		t.Fatalf("unexpected run resp: %+v", runResp)
	}

	statusResp := waitFlowRunTerminal(t, conn, codec, nextHdr, flowID, runResp.RunID, reqPrefix)
	if statusResp.Code != 1 || statusResp.FlowID != flowID || statusResp.RunID != runResp.RunID {
		t.Fatalf("unexpected terminal status resp: %+v", statusResp)
	}
	return runResp, statusResp
}

func waitFlowRunTerminal(t *testing.T, conn net.Conn, codec header.HeaderTcpCodec, nextHdr func() core.IHeader, flowID, runID, reqPrefix string) protocolflow.StatusResp {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	attempt := 0
	for time.Now().Before(deadline) {
		attempt++
		var statusResp protocolflow.StatusResp
		sendFlowCtrlExpect(t, conn, codec, nextHdr(), protocolflow.ActionStatus, protocolflow.StatusReq{
			ReqID:  fmt.Sprintf("%s-status-%d", reqPrefix, attempt),
			FlowID: flowID,
			RunID:  runID,
		}, protocolflow.ActionStatusResp, &statusResp)

		if statusResp.Code == 1 {
			switch statusResp.Status {
			case "queued", "running":
				time.Sleep(20 * time.Millisecond)
				continue
			case "succeeded", "failed", "cancelled":
				return statusResp
			default:
				t.Fatalf("unexpected run status %q for %s", statusResp.Status, runID)
			}
		}
		if statusResp.Code != 404 {
			t.Fatalf("unexpected status resp while waiting terminal: %+v", statusResp)
		}
		time.Sleep(20 * time.Millisecond)
	}

	t.Fatalf("run %s did not reach terminal status", runID)
	return protocolflow.StatusResp{}
}

func assertNodeStatus(t *testing.T, nodes []protocolflow.NodeStatus, nodeID, wantStatus string) {
	t.Helper()

	for _, node := range nodes {
		if node.ID == nodeID {
			if node.Status != wantStatus {
				t.Fatalf("unexpected node status for %s: %+v", nodeID, node)
			}
			return
		}
	}
	t.Fatalf("node %s not found in status %+v", nodeID, nodes)
}

func assertJSONSemanticEqual(t *testing.T, raw json.RawMessage, want any) {
	t.Helper()

	var gotDoc any
	if err := json.Unmarshal(raw, &gotDoc); err != nil {
		t.Fatalf("unmarshal got json: %v", err)
	}
	wantBytes, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal want json: %v", err)
	}
	var wantDoc any
	if err := json.Unmarshal(wantBytes, &wantDoc); err != nil {
		t.Fatalf("unmarshal want json: %v", err)
	}
	if reflect.DeepEqual(gotDoc, wantDoc) {
		return
	}
	gotNorm, _ := json.Marshal(gotDoc)
	wantNorm, _ := json.Marshal(wantDoc)
	t.Fatalf("json mismatch want=%s got=%s", wantNorm, gotNorm)
}

func findFlowSummary(t *testing.T, flows []protocolflow.FlowSummary, flowID string) protocolflow.FlowSummary {
	t.Helper()

	for _, flow := range flows {
		if flow.FlowID == flowID {
			return flow
		}
	}
	t.Fatalf("flow %s not found in summaries %+v", flowID, flows)
	return protocolflow.FlowSummary{}
}
