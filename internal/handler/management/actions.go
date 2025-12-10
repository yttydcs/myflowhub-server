package management

import (
	"context"
	"encoding/json"
	"strings"

	core "github.com/yttydcs/myflowhub-core"
)

type mgmtMessage struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

type nodeEchoReq struct {
	Message string `json:"message"`
}

type nodeEchoResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg,omitempty"`
	Echo string `json:"echo,omitempty"`
}

// node_echo: 管理类回显指令
type nodeEchoAction struct{ h *Handler }

func (a *nodeEchoAction) Name() string      { return "node_echo" }
func (a *nodeEchoAction) RequireAuth() bool { return false }
func (a *nodeEchoAction) Handle(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage) {
	var req nodeEchoReq
	if err := json.Unmarshal(data, &req); err != nil || strings.TrimSpace(req.Message) == "" {
		a.h.sendActionResp(ctx, conn, hdr, "node_echo_resp", nodeEchoResp{Code: 400, Msg: "invalid echo data"})
		return
	}
	a.h.log.Info("management node_echo", "conn", conn.ID(), "message", req.Message)
	a.h.sendActionResp(ctx, conn, hdr, "node_echo_resp", nodeEchoResp{Code: 1, Msg: "ok", Echo: req.Message})
}

func (h *Handler) initActions() {
	h.actions = make(map[string]core.SubProcessAction)
	h.registerAction(&nodeEchoAction{h: h})
}

func (h *Handler) registerAction(a core.SubProcessAction) {
	if a == nil || a.Name() == "" {
		return
	}
	h.actions[strings.ToLower(a.Name())] = a
}
