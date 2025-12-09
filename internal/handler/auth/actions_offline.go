package auth

import (
	"context"
	"encoding/json"

	core "github.com/yttydcs/myflowhub-core"
)

type offlineAction struct {
	h        *LoginHandler
	assisted bool
}

func (a *offlineAction) Name() string {
	if a.assisted {
		return actionAssistOffline
	}
	return actionOffline
}
func (a *offlineAction) RequireAuth() bool { return true }
func (a *offlineAction) Handle(ctx context.Context, conn core.IConnection, _ core.IHeader, data json.RawMessage) {
	var req offlineData
	if err := json.Unmarshal(data, &req); err != nil || req.DeviceID == "" {
		return
	}
	a.h.removeBinding(req.DeviceID, "")
	a.h.removeIndexes(ctx, req.NodeID, conn)
	if !a.assisted {
		// forward to parent
		if parent := a.h.selectAuthorityConn(ctx); parent != nil && (conn == nil || parent.ID() != conn.ID()) {
			a.h.forward(ctx, parent, actionAssistOffline, req)
		}
	}
}

func registerOfflineActions(h *LoginHandler) []core.SubProcessAction {
	return []core.SubProcessAction{
		&offlineAction{h: h, assisted: false},
		&offlineAction{h: h, assisted: true},
	}
}
