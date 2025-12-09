package auth

import (
	"context"
	"encoding/json"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-core/header"
)

func (h *LoginHandler) setPending(deviceID, connID string) {
	h.mu.Lock()
	h.pendingConn[deviceID] = connID
	h.mu.Unlock()
}

func (h *LoginHandler) popPending(deviceID string) (string, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	id, ok := h.pendingConn[deviceID]
	if ok {
		delete(h.pendingConn, deviceID)
	}
	return id, ok
}

func (h *LoginHandler) sendResp(ctx context.Context, conn core.IConnection, reqHdr core.IHeader, action string, data respData) {
	msg := message{Action: action}
	raw, _ := json.Marshal(data)
	msg.Data = raw
	payload, _ := json.Marshal(msg)
	hdr := h.buildHeader(ctx, reqHdr)
	if srv := core.ServerFromContext(ctx); srv != nil {
		if data.HubID == 0 {
			data.HubID = srv.NodeID()
			raw, _ = json.Marshal(data)
			msg.Data = raw
			payload, _ = json.Marshal(msg)
		}
		if conn != nil {
			if err := srv.Send(ctx, conn.ID(), hdr, payload); err != nil {
				h.log.Warn("send resp failed", "err", err)
			}
			return
		}
	}
	if conn != nil {
		codec := header.HeaderTcpCodec{}
		_ = conn.SendWithHeader(hdr, payload, codec)
	}
}

func (h *LoginHandler) buildHeader(ctx context.Context, reqHdr core.IHeader) core.IHeader {
	if reqHdr != nil {
		return reqHdr.Clone()
	}
	base := &header.HeaderTcp{}
	src := uint32(0)
	if srv := core.ServerFromContext(ctx); srv != nil {
		src = srv.NodeID()
	}
	return base.WithMajor(header.MajorOKResp).WithSubProto(2).WithSourceID(src).WithTargetID(0)
}

func (h *LoginHandler) forward(ctx context.Context, targetConn core.IConnection, action string, data any) {
	if targetConn == nil {
		return
	}
	payloadData, _ := json.Marshal(data)
	msg := message{Action: action, Data: payloadData}
	payload, _ := json.Marshal(msg)
	hdr := (&header.HeaderTcp{}).WithMajor(header.MajorCmd).WithSubProto(2)
	if srv := core.ServerFromContext(ctx); srv != nil {
		hdr.WithSourceID(srv.NodeID())
	}
	if nid, ok := targetConn.GetMeta("nodeID"); ok {
		if v, ok2 := nid.(uint32); ok2 {
			hdr.WithTargetID(v)
		}
	}
	if srv := core.ServerFromContext(ctx); srv != nil {
		_ = srv.Send(ctx, targetConn.ID(), hdr, payload)
		return
	}
	codec := header.HeaderTcpCodec{}
	_ = targetConn.SendWithHeader(hdr, payload, codec)
}

// route index helpers: allow mapping child nodeIDs to the connection carrying them.
func (h *LoginHandler) addRouteIndex(ctx context.Context, nodeID uint32, conn core.IConnection) {
	if nodeID == 0 || conn == nil {
		return
	}
	if srv := core.ServerFromContext(ctx); srv != nil {
		if cm := srv.ConnManager(); cm != nil {
			cm.AddNodeIndex(nodeID, conn)
		}
	}
}

func (h *LoginHandler) removeRouteIndex(ctx context.Context, nodeID uint32) {
	if nodeID == 0 {
		return
	}
	if srv := core.ServerFromContext(ctx); srv != nil {
		if cm := srv.ConnManager(); cm != nil {
			cm.RemoveNodeIndex(nodeID)
		}
	}
}
