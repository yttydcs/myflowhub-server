package management

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-server/internal/handler"
)

// 子协议：管理指令，仅处理发往本节点的 action+data JSON。
const SubProtoManagement uint8 = 1

type Handler struct {
	log     *slog.Logger
	actions map[string]core.SubProcessAction
}

func NewHandler(log *slog.Logger) *Handler {
	if log == nil {
		log = slog.Default()
	}
	h := &Handler{log: log}
	h.Init()
	return h
}

func (h *Handler) SubProto() uint8 { return SubProtoManagement }

func (h *Handler) Init() bool {
	h.initActions()
	return true
}

func (h *Handler) OnReceive(ctx context.Context, conn core.IConnection, hdr core.IHeader, payload []byte) {
	var msg mgmtMessage
	if err := json.Unmarshal(payload, &msg); err != nil {
		h.log.Warn("management invalid payload", "err", err)
		return
	}
	srv := core.ServerFromContext(ctx)
	if srv == nil {
		return
	}
	if hdr != nil && hdr.TargetID() != 0 && hdr.TargetID() != srv.NodeID() {
		h.log.Debug("management target mismatch, drop", "target", hdr.TargetID(), "local", srv.NodeID())
		return
	}
	act := strings.ToLower(strings.TrimSpace(msg.Action))
	entry, ok := h.actions[act]
	if !ok {
		h.log.Debug("unknown management action", "action", act)
		return
	}
	entry.Handle(ctx, conn, hdr, msg.Data)
}

// 内部响应工具
func (h *Handler) sendActionResp(ctx context.Context, conn core.IConnection, req core.IHeader, action string, data any) {
	resp := mgmtMessage{Action: action}
	raw, _ := json.Marshal(data)
	resp.Data = raw
	body, _ := json.Marshal(resp)
	handler.SendResponse(ctx, h.log, conn, req, body, h.SubProto())
}
