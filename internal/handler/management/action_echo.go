package management

import (
	"context"
	"encoding/json"
	"strings"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-core/subproto"
)

// node_echo: 管理类回显指令
type nodeEchoAction struct {
	subproto.BaseAction
	h *ManagementHandler
}

func (a *nodeEchoAction) Name() string      { return actionNodeEcho }
func (a *nodeEchoAction) RequireAuth() bool { return false }
func (a *nodeEchoAction) Handle(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage) {
	var req nodeEchoReq
	if err := json.Unmarshal(data, &req); err != nil || strings.TrimSpace(req.Message) == "" {
		a.h.sendActionResp(ctx, conn, hdr, actionNodeEchoResp, nodeEchoResp{Code: 400, Msg: "invalid echo data"})
		return
	}
	a.h.log.Info("management node_echo", "conn", conn.ID(), "message", req.Message)
	a.h.sendActionResp(ctx, conn, hdr, actionNodeEchoResp, nodeEchoResp{Code: 1, Msg: "ok", Echo: req.Message})
}
