package hubruntime

// 本文件承载 `hubruntime` 中与 `options` 相关的逻辑。

import (
	"log/slog"
	"os"
	"sort"
	"strconv"
	"strings"

	coreconfig "github.com/yttydcs/myflowhub-core/config"
)

const defaultAuthRolePerms = coreconfig.DefaultAuthRolePerms

// Options defines a reusable HubServer runtime configuration.
//
// Notes:
// - Keep this struct gomobile-friendly (basic types only) to simplify future binding.
// - Fields are intentionally aligned with cmd/hub_server flags/env to avoid drift.
type Options struct {
	// Listener toggles (restart required to take effect).
	//
	// TCP remains the default transport in v1.
	TCPEnable bool
	Addr      string

	// QUIC listener config (UDP-based, stream semantics).
	QUICEnable            bool
	QUICAddr              string
	QUICALPN              string
	QUICCertFile          string
	QUICKeyFile           string
	QUICDevCertAuto       bool
	QUICClientCAFile      string
	QUICRequireClientCert bool

	// Bluetooth Classic (RFCOMM/SPP-style byte stream) listener config.
	// NOTE:
	// - RFCOMM is a byte-stream transport (similar to TCP), suitable to carry MyFlowHub frames.
	// - Channel=0 means "auto/UUID-first" (platform will resolve/assign channel via SDP when supported).
	RFCOMMEnable   bool
	RFCOMMUUID     string
	RFCOMMChannel  int
	RFCOMMAdapter  string
	RFCOMMInsecure bool

	// NodeID is the local node id for this hub. If ParentEnable and SelfID are set,
	// runtime may self-register against parent and override NodeID to match parent assignment.
	NodeID uint32

	// Parent link
	// ParentEndpoint supports scheme prefixes like:
	// - tcp://127.0.0.1:9000
	// - bt+rfcomm://AA:BB:CC:DD:EE:FF?uuid=...
	// - quic://127.0.0.1:9000?server_name=...&pin_sha256=...
	ParentEndpoint     string
	ParentAddr         string
	ParentEnable       bool
	ParentJoinPermit   string
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

	// ConfigOverrideKeys records config keys explicitly supplied by env/flags/caller.
	// It is stored as a comma-separated list to keep Options gomobile-friendly.
	ConfigOverrideKeys string

	Logger *slog.Logger
}

// DefaultOptions 提供与 hub_server CLI 对齐的默认运行参数。
func DefaultOptions() Options {
	return Options{
		TCPEnable:             true,
		Addr:                  ":9000",
		QUICEnable:            false,
		QUICAddr:              ":9000",
		QUICALPN:              "myflowhub",
		QUICCertFile:          "",
		QUICKeyFile:           "",
		QUICDevCertAuto:       false,
		QUICClientCAFile:      "",
		QUICRequireClientCert: false,
		NodeID:                1,
		ParentEndpoint:        "",
		ParentAddr:            "",
		ParentEnable:          false,
		ParentJoinPermit:      "",
		ParentReconnectSec:    3,

		RFCOMMEnable:   false,
		RFCOMMUUID:     "0eef65b8-9374-42ea-b992-6ee2d0699f5c",
		RFCOMMChannel:  0,
		RFCOMMAdapter:  "hci0",
		RFCOMMInsecure: false,

		ProcChannels: 4,
		ProcWorkers:  8,
		ProcBuffer:   256,

		SendChannels:      2,
		SendWorkers:       2,
		SendChannelBuffer: 128,
		SendConnBuffer:    128,

		AuthDefaultRole:  "node",
		AuthDefaultPerms: "",
		AuthNodeRoles:    "",
		// Server runtime follows the Core exported default role mapping to avoid drift.
		AuthRolePerms: defaultAuthRolePerms,

		WorkDir:            "",
		SelfID:             "",
		ConfigOverrideKeys: "",
	}
}

// DefaultOptionsFromEnv 读取环境变量，并同步记录哪些键属于显式覆盖。
func DefaultOptionsFromEnv() Options {
	opts := DefaultOptions()

	if v, ok := lookupEnvBool("HUB_TCP_ENABLE"); ok {
		opts.TCPEnable = v
	}
	if v, ok := lookupEnvString("HUB_ADDR"); ok {
		opts.Addr = v
		opts.AddConfigOverrideKeys("addr")
	}
	if v, ok := lookupEnvBool("HUB_QUIC_ENABLE"); ok {
		opts.QUICEnable = v
	}
	if v, ok := lookupEnvString("HUB_QUIC_ADDR"); ok {
		opts.QUICAddr = v
	}
	if v, ok := lookupEnvString("HUB_QUIC_ALPN"); ok {
		opts.QUICALPN = v
	}
	if v, ok := lookupEnvString("HUB_QUIC_CERT_FILE"); ok {
		opts.QUICCertFile = v
	}
	if v, ok := lookupEnvString("HUB_QUIC_KEY_FILE"); ok {
		opts.QUICKeyFile = v
	}
	if v, ok := lookupEnvBool("HUB_QUIC_DEV_CERT_AUTO"); ok {
		opts.QUICDevCertAuto = v
	}
	if v, ok := lookupEnvString("HUB_QUIC_CLIENT_CA_FILE"); ok {
		opts.QUICClientCAFile = v
	}
	if v, ok := lookupEnvBool("HUB_QUIC_REQUIRE_CLIENT_CERT"); ok {
		opts.QUICRequireClientCert = v
	}
	if v, ok := lookupEnvUint32("HUB_NODE_ID"); ok {
		opts.NodeID = v
	}
	if v, ok := lookupEnvString("HUB_PARENT_ENDPOINT"); ok {
		opts.ParentEndpoint = v
		opts.AddConfigOverrideKeys(coreconfig.KeyParentAddr)
	}
	if v, ok := lookupEnvString("HUB_PARENT_ADDR"); ok {
		opts.ParentAddr = v
		opts.AddConfigOverrideKeys(coreconfig.KeyParentAddr)
	}
	if v, ok := lookupEnvBool("HUB_PARENT_ENABLE"); ok {
		opts.ParentEnable = v
		opts.AddConfigOverrideKeys(coreconfig.KeyParentEnable)
	}
	if v, ok := lookupEnvString("HUB_PARENT_JOIN_PERMIT"); ok {
		opts.ParentJoinPermit = v
		opts.AddConfigOverrideKeys(coreconfig.KeyParentJoinPermit)
	}
	if v, ok := lookupEnvInt("HUB_PARENT_RECONNECT"); ok {
		opts.ParentReconnectSec = int(v)
		opts.AddConfigOverrideKeys(coreconfig.KeyParentReconnectSec)
	}

	if v, ok := lookupEnvBool("HUB_RFCOMM_ENABLE"); ok {
		opts.RFCOMMEnable = v
	}
	if v, ok := lookupEnvString("HUB_RFCOMM_UUID"); ok {
		opts.RFCOMMUUID = v
	}
	if v, ok := lookupEnvInt("HUB_RFCOMM_CHANNEL"); ok {
		opts.RFCOMMChannel = int(v)
	}
	if v, ok := lookupEnvString("HUB_RFCOMM_ADAPTER"); ok {
		opts.RFCOMMAdapter = v
	}
	if v, ok := lookupEnvBool("HUB_RFCOMM_INSECURE"); ok {
		opts.RFCOMMInsecure = v
	}

	if v, ok := lookupEnvInt("HUB_PROC_CHANNELS"); ok {
		opts.ProcChannels = int(v)
		opts.AddConfigOverrideKeys(coreconfig.KeyProcChannelCount)
	}
	if v, ok := lookupEnvInt("HUB_PROC_WORKERS"); ok {
		opts.ProcWorkers = int(v)
		opts.AddConfigOverrideKeys(coreconfig.KeyProcWorkersPerChan)
	}
	if v, ok := lookupEnvInt("HUB_PROC_BUFFER"); ok {
		opts.ProcBuffer = int(v)
		opts.AddConfigOverrideKeys(coreconfig.KeyProcChannelBuffer)
	}

	if v, ok := lookupEnvInt("HUB_SEND_CHANNELS"); ok {
		opts.SendChannels = int(v)
		opts.AddConfigOverrideKeys(coreconfig.KeySendChannelCount)
	}
	if v, ok := lookupEnvInt("HUB_SEND_WORKERS"); ok {
		opts.SendWorkers = int(v)
		opts.AddConfigOverrideKeys(coreconfig.KeySendWorkersPerChan)
	}
	if v, ok := lookupEnvInt("HUB_SEND_CHANNEL_BUFFER"); ok {
		opts.SendChannelBuffer = int(v)
		opts.AddConfigOverrideKeys(coreconfig.KeySendChannelBuffer)
	}
	if v, ok := lookupEnvInt("HUB_SEND_CONN_BUFFER"); ok {
		opts.SendConnBuffer = int(v)
		opts.AddConfigOverrideKeys(coreconfig.KeySendConnBuffer)
	}

	if v, ok := lookupEnvString("HUB_AUTH_DEFAULT_ROLE"); ok {
		opts.AuthDefaultRole = v
		opts.AddConfigOverrideKeys(coreconfig.KeyAuthDefaultRole)
	}
	if v, ok := lookupEnvString("HUB_AUTH_DEFAULT_PERMS"); ok {
		opts.AuthDefaultPerms = v
		opts.AddConfigOverrideKeys(coreconfig.KeyAuthDefaultPerms)
	}
	if v, ok := lookupEnvString("HUB_AUTH_NODE_ROLES"); ok {
		opts.AuthNodeRoles = v
		opts.AddConfigOverrideKeys(coreconfig.KeyAuthNodeRoles)
	}
	if v, ok := lookupEnvString("HUB_AUTH_ROLE_PERMS"); ok {
		opts.AuthRolePerms = v
		opts.AddConfigOverrideKeys(coreconfig.KeyAuthRolePerms)
	}

	if v, ok := lookupEnvString("HUB_WORKDIR"); ok {
		opts.WorkDir = v
	}
	if v, ok := lookupEnvString("HUB_SELF_ID"); ok {
		opts.SelfID = v
	}

	return opts
}

// Normalize 补齐缺省值并清洗输入，保证后续构建配置时语义稳定。
func (o *Options) Normalize() {
	if o == nil {
		return
	}
	defaults := DefaultOptions()
	overrideKeys := o.configOverrideKeySet()
	if strings.TrimSpace(o.ParentEndpoint) != "" || strings.TrimSpace(o.ParentAddr) != "" {
		o.ParentEnable = true
	}
	o.Addr = strings.TrimSpace(o.Addr)
	o.QUICAddr = strings.TrimSpace(o.QUICAddr)
	o.QUICALPN = strings.TrimSpace(o.QUICALPN)
	o.QUICCertFile = strings.TrimSpace(o.QUICCertFile)
	o.QUICKeyFile = strings.TrimSpace(o.QUICKeyFile)
	o.QUICClientCAFile = strings.TrimSpace(o.QUICClientCAFile)
	o.ParentEndpoint = strings.TrimSpace(o.ParentEndpoint)
	o.ParentAddr = strings.TrimSpace(o.ParentAddr)
	o.ParentJoinPermit = strings.TrimSpace(o.ParentJoinPermit)
	if o.TCPEnable && o.Addr == "" {
		if _, ok := overrideKeys["addr"]; !ok {
			o.Addr = defaults.Addr
		}
	}
	if o.QUICEnable && o.QUICAddr == "" {
		o.QUICAddr = defaults.QUICAddr
	}
	if o.QUICEnable && o.QUICALPN == "" {
		o.QUICALPN = defaults.QUICALPN
	}
	if o.ParentReconnectSec < 0 {
		o.ParentReconnectSec = 0
	} else if o.ParentReconnectSec == 0 {
		if _, ok := overrideKeys[coreconfig.KeyParentReconnectSec]; !ok {
			o.ParentReconnectSec = defaults.ParentReconnectSec
		}
	}
	o.RFCOMMUUID = strings.TrimSpace(o.RFCOMMUUID)
	o.RFCOMMAdapter = strings.TrimSpace(o.RFCOMMAdapter)
	if o.RFCOMMUUID == "" {
		o.RFCOMMUUID = defaults.RFCOMMUUID
	}
	if o.RFCOMMChannel < 0 {
		o.RFCOMMChannel = 0
	}
	if o.RFCOMMEnable && o.RFCOMMAdapter == "" {
		o.RFCOMMAdapter = defaults.RFCOMMAdapter
	}
	if o.ProcChannels == 0 {
		if _, ok := overrideKeys[coreconfig.KeyProcChannelCount]; !ok {
			o.ProcChannels = defaults.ProcChannels
		}
	}
	if o.ProcWorkers == 0 {
		if _, ok := overrideKeys[coreconfig.KeyProcWorkersPerChan]; !ok {
			o.ProcWorkers = defaults.ProcWorkers
		}
	}
	if o.ProcBuffer == 0 {
		if _, ok := overrideKeys[coreconfig.KeyProcChannelBuffer]; !ok {
			o.ProcBuffer = defaults.ProcBuffer
		}
	}
	if o.SendChannels == 0 {
		if _, ok := overrideKeys[coreconfig.KeySendChannelCount]; !ok {
			o.SendChannels = defaults.SendChannels
		}
	}
	if o.SendWorkers == 0 {
		if _, ok := overrideKeys[coreconfig.KeySendWorkersPerChan]; !ok {
			o.SendWorkers = defaults.SendWorkers
		}
	}
	if o.SendChannelBuffer == 0 {
		if _, ok := overrideKeys[coreconfig.KeySendChannelBuffer]; !ok {
			o.SendChannelBuffer = defaults.SendChannelBuffer
		}
	}
	if o.SendConnBuffer == 0 {
		if _, ok := overrideKeys[coreconfig.KeySendConnBuffer]; !ok {
			o.SendConnBuffer = defaults.SendConnBuffer
		}
	}
	o.AuthDefaultRole = strings.TrimSpace(o.AuthDefaultRole)
	o.AuthDefaultPerms = strings.TrimSpace(o.AuthDefaultPerms)
	o.AuthNodeRoles = strings.TrimSpace(o.AuthNodeRoles)
	o.AuthRolePerms = strings.TrimSpace(o.AuthRolePerms)
	if o.AuthDefaultRole == "" {
		if _, ok := overrideKeys[coreconfig.KeyAuthDefaultRole]; !ok {
			o.AuthDefaultRole = defaults.AuthDefaultRole
		}
	}
	if o.AuthRolePerms == "" {
		if _, ok := overrideKeys[coreconfig.KeyAuthRolePerms]; !ok {
			o.AuthRolePerms = defaults.AuthRolePerms
		}
	}
	o.WorkDir = strings.TrimSpace(o.WorkDir)
	o.SelfID = strings.TrimSpace(o.SelfID)
	o.ConfigOverrideKeys = joinOverrideKeys(splitOverrideKeys(o.ConfigOverrideKeys))
}

// AddConfigOverrideKeys 记录调用方显式给出的配置键，避免 Normalize 覆盖它们。
func (o *Options) AddConfigOverrideKeys(keys ...string) {
	if o == nil {
		return
	}
	set := splitOverrideKeys(o.ConfigOverrideKeys)
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		set[key] = struct{}{}
	}
	o.ConfigOverrideKeys = joinOverrideKeys(set)
}

// configOverrideKeySet 把逗号分隔字符串转成便于查询的 set。
func (o Options) configOverrideKeySet() map[string]struct{} {
	return splitOverrideKeys(o.ConfigOverrideKeys)
}

// splitOverrideKeys 兼容逗号、分号和空白分隔的覆盖键列表。
func splitOverrideKeys(raw string) map[string]struct{} {
	out := make(map[string]struct{})
	for _, part := range strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\r' || r == '\t'
	}) {
		key := strings.TrimSpace(part)
		if key == "" {
			continue
		}
		out[key] = struct{}{}
	}
	return out
}

// joinOverrideKeys 生成稳定排序后的覆盖键字符串，便于持久化与比较。
func joinOverrideKeys(set map[string]struct{}) string {
	if len(set) == 0 {
		return ""
	}
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return strings.Join(keys, ",")
}

// lookupEnvString 读取并裁剪字符串环境变量。
func lookupEnvString(key string) (string, bool) {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return "", false
	}
	return strings.TrimSpace(raw), true
}

// lookupEnvBool 只接受明确的 true/false 字面值，避免脏输入进入配置。
func lookupEnvBool(key string) (bool, bool) {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return false, false
	}
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "y", "on":
		return true, true
	case "0", "false", "no", "n", "off":
		return false, true
	default:
		return false, false
	}
}

// lookupEnvInt 读取十进制整数字段。
func lookupEnvInt(key string) (int64, bool) {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return 0, false
	}
	n, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

// lookupEnvUint32 读取 uint32 范围内的节点号等配置。
func lookupEnvUint32(key string) (uint32, bool) {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return 0, false
	}
	n, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 32)
	if err != nil {
		return 0, false
	}
	return uint32(n), true
}
