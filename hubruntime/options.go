package hubruntime

import (
	"log/slog"
	"os"
	"strconv"
	"strings"

	core "github.com/yttydcs/myflowhub-core"
)

// Options defines a reusable HubServer runtime configuration.
//
// Notes:
// - Keep this struct gomobile-friendly (basic types only) to simplify future binding.
// - Fields are intentionally aligned with cmd/hub_server flags/env to avoid drift.
type Options struct {
	Addr string

	// NodeID is the local node id for this hub. If ParentEnable and SelfID are set,
	// runtime may self-register against parent and override NodeID to match parent assignment.
	NodeID uint32

	// Parent link
	ParentAddr         string
	ParentEnable       bool
	ParentReconnectSec int

	// Dispatcher/worker settings
	ProcChannels int
	ProcWorkers  int
	ProcBuffer   int

	// Send dispatcher settings
	SendChannels      int
	SendWorkers       int
	SendChannelBuffer int
	SendConnBuffer    int

	// Auth defaults (open registration mode by current product decision)
	AuthDefaultRole  string
	AuthDefaultPerms string
	AuthNodeRoles    string
	AuthRolePerms    string

	// WorkDir changes process working directory during Start to make relative paths (e.g. config/*)
	// resolve into an app-private directory on Android.
	WorkDir string

	// SelfID is used for parent self-register/bootstrap (auth register).
	// When empty, runtime will not perform self-register and will not bind parent conn via register.
	SelfID string

	Logger *slog.Logger
}

func DefaultOptionsFromEnv() Options {
	return Options{
		Addr:               getenv("HUB_ADDR", ":9000"),
		NodeID:             getenvUint32("HUB_NODE_ID", 1),
		ParentAddr:         getenv("HUB_PARENT_ADDR", ""),
		ParentEnable:       core.ParseBool(getenv("HUB_PARENT_ENABLE", "false"), false),
		ParentReconnectSec: int(getenvInt("HUB_PARENT_RECONNECT", 3)),

		ProcChannels: int(getenvInt("HUB_PROC_CHANNELS", 4)),
		ProcWorkers:  int(getenvInt("HUB_PROC_WORKERS", 8)),
		ProcBuffer:   int(getenvInt("HUB_PROC_BUFFER", 256)),

		SendChannels:      int(getenvInt("HUB_SEND_CHANNELS", 2)),
		SendWorkers:       int(getenvInt("HUB_SEND_WORKERS", 2)),
		SendChannelBuffer: int(getenvInt("HUB_SEND_CHANNEL_BUFFER", 128)),
		SendConnBuffer:    int(getenvInt("HUB_SEND_CONN_BUFFER", 128)),

		AuthDefaultRole:  getenv("HUB_AUTH_DEFAULT_ROLE", "node"),
		AuthDefaultPerms: getenv("HUB_AUTH_DEFAULT_PERMS", ""),
		AuthNodeRoles:    getenv("HUB_AUTH_NODE_ROLES", ""),
		// 默认给 node 角色开放 file/flow/exec 权限（可用 HUB_AUTH_ROLE_PERMS 覆盖）
		AuthRolePerms: getenv("HUB_AUTH_ROLE_PERMS", "node:file.read,file.write,flow.set,exec.call"),

		WorkDir: getenv("HUB_WORKDIR", ""),
		SelfID:  getenv("HUB_SELF_ID", ""),
	}
}

func (o *Options) Normalize() {
	if o == nil {
		return
	}
	if strings.TrimSpace(o.ParentAddr) != "" {
		o.ParentEnable = true
	}
	if strings.TrimSpace(o.Addr) == "" {
		o.Addr = ":9000"
	}
	if o.ParentReconnectSec < 0 {
		o.ParentReconnectSec = 0
	}
	o.AuthDefaultRole = strings.TrimSpace(o.AuthDefaultRole)
	o.AuthDefaultPerms = strings.TrimSpace(o.AuthDefaultPerms)
	o.AuthNodeRoles = strings.TrimSpace(o.AuthNodeRoles)
	o.AuthRolePerms = strings.TrimSpace(o.AuthRolePerms)
	o.WorkDir = strings.TrimSpace(o.WorkDir)
	o.SelfID = strings.TrimSpace(o.SelfID)
}

func getenv(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int64) int64 {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return def
}

func getenvUint32(key string, def uint32) uint32 {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if n, err := strconv.ParseUint(v, 10, 32); err == nil {
			return uint32(n)
		}
	}
	return def
}
