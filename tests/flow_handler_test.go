package tests

import (
	"context"
	"testing"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-core/config"
	"github.com/yttydcs/myflowhub-core/connmgr"
	"github.com/yttydcs/myflowhub-core/header"
	protocolflow "github.com/yttydcs/myflowhub-proto/protocol/flow"
	serverflow "github.com/yttydcs/myflowhub-subproto/flow"
)

func TestFlowSetRespMajorOKRespOnInvalidSet(t *testing.T) {
	cm := connmgr.New()
	requester := newStubConn("requester-1")
	requester.SetMeta("nodeID", uint32(5))
	if err := cm.Add(requester); err != nil {
		t.Fatalf("add requester: %v", err)
	}

	srv := &stubServer{nodeID: 1, cm: cm}
	ctx := core.WithServerContext(context.Background(), srv)

	cfg := config.NewMap(map[string]string{
		"flow.base_dir": t.TempDir(),
	})
	h := serverflow.NewHandlerWithConfig(cfg, nil)
	h.Init()

	payload := mustJSON(map[string]any{
		"action": protocolflow.ActionSet,
		"data":   map[string]any{},
	})
	hdr := (&header.HeaderTcp{}).
		WithMajor(header.MajorCmd).
		WithSubProto(serverflow.SubProtoFlow).
		WithSourceID(5)

	h.OnReceive(ctx, requester, hdr, payload)

	if len(srv.sends) != 1 {
		t.Fatalf("expected 1 set_resp send, got %d", len(srv.sends))
	}
	if srv.sends[0].major != header.MajorOKResp {
		t.Fatalf("expected set_resp major=%d, got %d", header.MajorOKResp, srv.sends[0].major)
	}
}

