package tests

import (
	"context"
	"testing"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-core/connmgr"
	"github.com/yttydcs/myflowhub-core/header"
	protocolexec "github.com/yttydcs/myflowhub-proto/protocol/exec"
	serverexec "github.com/yttydcs/myflowhub-subproto/exec"
)

func TestExecCallRespMajorOKResp(t *testing.T) {
	cm := connmgr.New()
	executor := newStubConn("executor-1")
	executor.SetMeta("nodeID", uint32(2))
	if err := cm.Add(executor); err != nil {
		t.Fatalf("add executor: %v", err)
	}

	srv := &stubServer{nodeID: 1, cm: cm}
	ctx := core.WithServerContext(context.Background(), srv)

	h := serverexec.NewHandler(nil)
	h.Init()

	req := protocolexec.CallReq{
		ReqID:        "r1",
		ExecutorNode: 2,
		TargetNode:   1,
		Method:       "debug::echo",
	}
	payload := mustJSON(map[string]any{
		"action": protocolexec.ActionCall,
		"data":   req,
	})
	hdr := (&header.HeaderTcp{}).
		WithMajor(header.MajorCmd).
		WithSubProto(serverexec.SubProtoExec).
		WithSourceID(2)

	h.OnReceive(ctx, executor, hdr, payload)

	if len(srv.sends) != 1 {
		t.Fatalf("expected 1 call_resp send, got %d", len(srv.sends))
	}
	if srv.sends[0].major != header.MajorOKResp {
		t.Fatalf("expected call_resp major=%d, got %d", header.MajorOKResp, srv.sends[0].major)
	}
}

