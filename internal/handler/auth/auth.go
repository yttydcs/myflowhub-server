package auth

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"sync/atomic"

	core "github.com/yttydcs/myflowhub-core"
	permission "github.com/yttydcs/myflowhub-core/kit/permission"
	"github.com/yttydcs/myflowhub-core/subproto"
)

// LoginHandler implements register/login/revoke/offline flows with action+data payload.
type LoginHandler struct {
	subproto.ActionBaseSubProcess
	log *slog.Logger

	nextID atomic.Uint32

	mu          sync.RWMutex
	whitelist   map[string]bindingRecord // deviceID -> record
	pendingConn map[string]string        // deviceID -> connID (in-flight assist)

	authNode uint32

	permCfg *permission.Config
}

func NewLoginHandler(log *slog.Logger) *LoginHandler {
	return NewLoginHandlerWithConfig(nil, log)
}

func NewLoginHandlerWithConfig(cfg core.IConfig, log *slog.Logger) *LoginHandler {
	if log == nil {
		log = slog.Default()
	}
	h := &LoginHandler{
		log:         log,
		whitelist:   make(map[string]bindingRecord),
		pendingConn: make(map[string]string),
	}
	h.loadAuthConfig(cfg)
	h.nextID.Store(2)
	return h
}

func (h *LoginHandler) SubProto() uint8 { return 2 }

func (h *LoginHandler) Init() bool {
	h.initActions()
	return true
}

// AllowSourceMismatch 登录阶段允许 SourceID 与连接元数据不一致（尚未绑定 nodeID）。
func (h *LoginHandler) AllowSourceMismatch() bool { return true }

func (h *LoginHandler) OnReceive(ctx context.Context, conn core.IConnection, hdr core.IHeader, payload []byte) {
	var msg message
	if err := json.Unmarshal(payload, &msg); err != nil {
		h.log.Warn("invalid login payload", "err", err)
		return
	}
	entry, ok := h.LookupAction(msg.Action)
	if !ok {
		h.log.Debug("unknown login action", "action", msg.Action)
		return
	}
	if entry.RequireAuth() && !h.sourceMatches(conn, hdr) {
		h.log.Warn("drop login action due to source mismatch", "action", msg.Action, "hdr_source", hdr.SourceID())
		return
	}
	entry.Handle(ctx, conn, hdr, msg.Data)
}

func (h *LoginHandler) initActions() {
	h.ResetActions()
	for _, act := range registerRegisterActions(h) {
		h.RegisterAction(act)
	}
	for _, act := range registerLoginActions(h) {
		h.RegisterAction(act)
	}
	for _, act := range registerRevokeActions(h) {
		h.RegisterAction(act)
	}
	for _, act := range registerAssistQueryActions(h) {
		h.RegisterAction(act)
	}
	for _, act := range registerOfflineActions(h) {
		h.RegisterAction(act)
	}
	for _, act := range registerPermActions(h) {
		h.RegisterAction(act)
	}
}
