package management

import (
	"context"
	"encoding/json"
	"strings"

	core "github.com/yttydcs/myflowhub-core"
	coreconfig "github.com/yttydcs/myflowhub-core/config"
	"github.com/yttydcs/myflowhub-core/subproto"
)

// config_get: 读取配置项
type configGetAction struct {
	subproto.BaseAction
	h *ManagementHandler
}

func (a *configGetAction) Name() string      { return "config_get" }
func (a *configGetAction) RequireAuth() bool { return false }
func (a *configGetAction) Handle(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage) {
	var req configGetReq
	if err := json.Unmarshal(data, &req); err != nil || strings.TrimSpace(req.Key) == "" {
		a.h.sendActionResp(ctx, conn, hdr, "config_get_resp", configResp{Code: 400, Msg: "invalid key"})
		return
	}
	srv := core.ServerFromContext(ctx)
	if srv == nil || srv.Config() == nil {
		a.h.sendActionResp(ctx, conn, hdr, "config_get_resp", configResp{Code: 500, Msg: "config unavailable"})
		return
	}
	val, ok := srv.Config().Get(strings.TrimSpace(req.Key))
	if !ok {
		a.h.sendActionResp(ctx, conn, hdr, "config_get_resp", configResp{Code: 404, Msg: "not found", Key: req.Key})
		return
	}
	a.h.sendActionResp(ctx, conn, hdr, "config_get_resp", configResp{Code: 1, Msg: "ok", Key: req.Key, Value: val})
}

// config_set: 更新配置项（仅支持可写 MapConfig）
type configSetAction struct {
	subproto.BaseAction
	h *ManagementHandler
}

func (a *configSetAction) Name() string      { return "config_set" }
func (a *configSetAction) RequireAuth() bool { return false }
func (a *configSetAction) Handle(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage) {
	var req configSetReq
	if err := json.Unmarshal(data, &req); err != nil || strings.TrimSpace(req.Key) == "" {
		a.h.sendActionResp(ctx, conn, hdr, "config_set_resp", configResp{Code: 400, Msg: "invalid key"})
		return
	}
	key := strings.TrimSpace(req.Key)
	srv := core.ServerFromContext(ctx)
	if srv == nil || srv.Config() == nil {
		a.h.sendActionResp(ctx, conn, hdr, "config_set_resp", configResp{Code: 500, Msg: "config unavailable"})
		return
	}
	cfg := srv.Config()
	if mc, ok := cfg.(*coreconfig.MapConfig); ok && mc != nil {
		mc.Set(key, req.Value)
		a.h.sendActionResp(ctx, conn, hdr, "config_set_resp", configResp{Code: 1, Msg: "ok", Key: key, Value: req.Value})
		return
	}
	// fallback: try interface with Set
	if setter, ok := cfg.(interface{ Set(string, string) }); ok {
		setter.Set(key, req.Value)
		a.h.sendActionResp(ctx, conn, hdr, "config_set_resp", configResp{Code: 1, Msg: "ok", Key: key, Value: req.Value})
		return
	}
	a.h.sendActionResp(ctx, conn, hdr, "config_set_resp", configResp{Code: 501, Msg: "config not writable"})
}

// config_list: 列出全部配置键（仅支持可枚举配置）
type configListAction struct {
	subproto.BaseAction
	h *ManagementHandler
}

func (a *configListAction) Name() string      { return "config_list" }
func (a *configListAction) RequireAuth() bool { return false }
func (a *configListAction) Handle(ctx context.Context, conn core.IConnection, hdr core.IHeader, _ json.RawMessage) {
	srv := core.ServerFromContext(ctx)
	if srv == nil || srv.Config() == nil {
		a.h.sendActionResp(ctx, conn, hdr, "config_list_resp", configListResp{Code: 500, Msg: "config unavailable"})
		return
	}
	cfg := srv.Config()
	if lister, ok := cfg.(interface{ Keys() []string }); ok && lister != nil {
		keys := lister.Keys()
		a.h.sendActionResp(ctx, conn, hdr, "config_list_resp", configListResp{Code: 1, Msg: "ok", Keys: keys})
		return
	}
	a.h.sendActionResp(ctx, conn, hdr, "config_list_resp", configListResp{Code: 501, Msg: "config not listable"})
}
