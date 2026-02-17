package tests

import (
	"context"
	"testing"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-core/config"
	"github.com/yttydcs/myflowhub-core/connmgr"
	"github.com/yttydcs/myflowhub-core/header"
	protocolfile "github.com/yttydcs/myflowhub-proto/protocol/file"
	serverfile "github.com/yttydcs/myflowhub-server/subproto/file"
)

func TestFileReadRespMajorOKRespOnInvalidRead(t *testing.T) {
	cm := connmgr.New()
	requester := newStubConn("requester-1")
	requester.SetMeta("nodeID", uint32(5))
	if err := cm.Add(requester); err != nil {
		t.Fatalf("add requester: %v", err)
	}
	srv := &stubServer{nodeID: 1, cm: cm}
	ctx := core.WithServerContext(context.Background(), srv)

	cfg := config.NewMap(map[string]string{
		"file.base_dir": t.TempDir(),
	})
	h := serverfile.NewHandlerWithConfig(cfg, nil)
	h.Init()

	body := mustJSON(map[string]any{
		"action": protocolfile.ActionRead,
		"data":   map[string]any{}, // op 缺失 => invalid op => read_resp
	})
	payload := append([]byte{protocolfile.KindCtrl}, body...)
	hdr := (&header.HeaderTcp{}).
		WithMajor(header.MajorCmd).
		WithSubProto(serverfile.SubProtoFile).
		WithSourceID(5).
		WithTargetID(1)

	h.OnReceive(ctx, requester, hdr, payload)

	if len(srv.sends) != 1 {
		t.Fatalf("expected 1 read_resp send, got %d", len(srv.sends))
	}
	if srv.sends[0].major != header.MajorOKResp {
		t.Fatalf("expected read_resp major=%d, got %d", header.MajorOKResp, srv.sends[0].major)
	}
}

func TestFileWriteRespMajorOKRespOnInvalidWrite(t *testing.T) {
	cm := connmgr.New()
	requester := newStubConn("requester-1")
	requester.SetMeta("nodeID", uint32(5))
	if err := cm.Add(requester); err != nil {
		t.Fatalf("add requester: %v", err)
	}
	srv := &stubServer{nodeID: 1, cm: cm}
	ctx := core.WithServerContext(context.Background(), srv)

	cfg := config.NewMap(map[string]string{
		"file.base_dir": t.TempDir(),
	})
	h := serverfile.NewHandlerWithConfig(cfg, nil)
	h.Init()

	body := mustJSON(map[string]any{
		"action": protocolfile.ActionWrite,
		"data":   map[string]any{}, // op 缺失 => invalid op => write_resp
	})
	payload := append([]byte{protocolfile.KindCtrl}, body...)
	hdr := (&header.HeaderTcp{}).
		WithMajor(header.MajorCmd).
		WithSubProto(serverfile.SubProtoFile).
		WithSourceID(5).
		WithTargetID(1)

	h.OnReceive(ctx, requester, hdr, payload)

	if len(srv.sends) != 1 {
		t.Fatalf("expected 1 write_resp send, got %d", len(srv.sends))
	}
	if srv.sends[0].major != header.MajorOKResp {
		t.Fatalf("expected write_resp major=%d, got %d", header.MajorOKResp, srv.sends[0].major)
	}
}

