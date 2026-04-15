package hubruntime

// 本文件承载 `hubruntime` 中与 `runtime` 相关的逻辑。

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-core/bootstrap"
	"github.com/yttydcs/myflowhub-core/connmgr"
	"github.com/yttydcs/myflowhub-core/header"
	"github.com/yttydcs/myflowhub-core/listener/multi_listener"
	"github.com/yttydcs/myflowhub-core/listener/quic_listener"
	"github.com/yttydcs/myflowhub-core/listener/rfcomm_listener"
	"github.com/yttydcs/myflowhub-core/listener/tcp_listener"
	"github.com/yttydcs/myflowhub-core/process"
	"github.com/yttydcs/myflowhub-core/server"
	"github.com/yttydcs/myflowhub-server/modules"
)

type Status struct {
	Running bool

	Addr   string
	NodeID uint32

	ParentEnabled   bool
	ParentAddr      string
	ParentConnected bool
	ParentConnID    string

	WorkDir string

	LastError string
}

type Runtime struct {
	mu sync.Mutex

	opts Options
	log  *slog.Logger

	srv core.IServer

	startCtx    context.Context
	startCancel context.CancelFunc

	workdirPrev string

	parentWatchCancel context.CancelFunc

	lastErr atomic.Value // string

	msgSeq atomic.Uint32
}

// New 校验监听器开关并创建可嵌入的 Hub runtime 实例。
func New(opts Options) (*Runtime, error) {
	opts.Normalize()
	if !opts.TCPEnable && !opts.RFCOMMEnable && !opts.QUICEnable {
		return nil, errors.New("no listener enabled")
	}
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
	return &Runtime{opts: opts, log: opts.Logger}, nil
}

// Start 负责工作目录切换、层叠配置构建、默认模块装配以及父链 bootstrap。
func (r *Runtime) Start(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	r.mu.Lock()
	if r.srv != nil {
		r.mu.Unlock()
		return errors.New("runtime already started")
	}
	opts := r.opts
	log := r.log
	r.mu.Unlock()

	normalizedWorkDir, err := r.applyWorkDir(opts.WorkDir)
	if err != nil {
		r.storeErr(err)
		return err
	}
	if normalizedWorkDir != "" {
		opts.WorkDir = normalizedWorkDir
	}

	cfg, err := buildConfig(opts)
	if err != nil {
		_ = r.restoreWorkDir()
		r.storeErr(err)
		return err
	}
	opts = applyConfigToOptions(opts, cfg)
	if opts.TCPEnable && strings.TrimSpace(opts.Addr) == "" {
		err := errors.New("tcp addr required")
		_ = r.restoreWorkDir()
		r.storeErr(err)
		return err
	}
	if opts.QUICEnable && strings.TrimSpace(opts.QUICAddr) == "" {
		err := errors.New("quic addr required")
		_ = r.restoreWorkDir()
		r.storeErr(err)
		return err
	}
	if err := ensureQUICDevCertIfNeeded(&opts, log); err != nil {
		_ = r.restoreWorkDir()
		r.storeErr(err)
		return err
	}

	parentTarget := effectiveParentTarget(opts)

	// Pre-start: if parent enabled and self id provided, self-register to obtain/confirm node id.
	if opts.ParentEnable && parentTarget != "" {
		_, _, err := parseParentEndpoint(parentTarget)
		if err != nil {
			_ = r.restoreWorkDir()
			r.storeErr(err)
			return err
		}
	}
	if opts.ParentEnable && parentTarget != "" && opts.SelfID != "" {
		nodeID, err := selfRegisterNodeID(ctx, parentTarget, opts.SelfID, opts.ParentJoinPermit, dialParentEndpoint, log)
		if err != nil {
			_ = r.restoreWorkDir()
			r.storeErr(err)
			return err
		}
		if opts.NodeID == 0 {
			opts.NodeID = nodeID
		} else if opts.NodeID != nodeID {
			log.Warn("node-id mismatch, override by parent assignment", "configured", opts.NodeID, "assigned", nodeID, "self_id", opts.SelfID)
			opts.NodeID = nodeID
		}
	}

	cm := connmgr.New()
	base := process.NewPreRoutingProcess(log).WithConfig(cfg)
	dispatcher, err := process.NewDispatcherFromConfig(cfg, base, log)
	if err != nil {
		_ = r.restoreWorkDir()
		r.storeErr(err)
		return err
	}
	set, err := modules.DefaultHub(cfg, log)
	if err != nil {
		_ = r.restoreWorkDir()
		r.storeErr(err)
		return err
	}
	if err := modules.RegisterAll(dispatcher, set); err != nil {
		_ = r.restoreWorkDir()
		r.storeErr(err)
		return err
	}

	listeners := make([]core.IListener, 0, 3)
	if opts.TCPEnable {
		listeners = append(listeners, tcp_listener.New(opts.Addr, tcp_listener.Options{
			KeepAlive:       true,
			KeepAlivePeriod: 30 * time.Second,
			Logger:          log,
		}))
	}
	if opts.QUICEnable {
		listeners = append(listeners, quic_listener.New(quic_listener.Options{
			Addr:              opts.QUICAddr,
			ALPN:              opts.QUICALPN,
			CertFile:          opts.QUICCertFile,
			KeyFile:           opts.QUICKeyFile,
			ClientCAFile:      opts.QUICClientCAFile,
			RequireClientCert: opts.QUICRequireClientCert,
			Logger:            log,
		}))
	}
	if opts.RFCOMMEnable {
		listeners = append(listeners, rfcomm_listener.New(rfcomm_listener.Options{
			UUID:     opts.RFCOMMUUID,
			Channel:  opts.RFCOMMChannel,
			Adapter:  opts.RFCOMMAdapter,
			Insecure: opts.RFCOMMInsecure,
			Logger:   log,
		}))
	}
	if len(listeners) == 0 {
		err := errors.New("no listener enabled")
		_ = r.restoreWorkDir()
		r.storeErr(err)
		return err
	}

	var lst core.IListener
	if len(listeners) == 1 {
		lst = listeners[0]
	} else {
		ml, err := multi_listener.New(listeners...)
		if err != nil {
			_ = r.restoreWorkDir()
			r.storeErr(err)
			return err
		}
		lst = ml
	}
	codec := header.HeaderTcpCodec{}

	srv, err := server.New(server.Options{
		Name:         "HubServer",
		Logger:       log,
		Process:      dispatcher,
		Codec:        codec,
		Listener:     lst,
		Config:       cfg,
		Manager:      cm,
		ParentDialer: dialParentEndpoint,
		NodeID:       opts.NodeID,
	})
	if err != nil {
		_ = r.restoreWorkDir()
		r.storeErr(err)
		return err
	}

	startCtx, startCancel := context.WithCancel(ctx)

	if err := srv.Start(startCtx); err != nil {
		startCancel()
		_ = r.restoreWorkDir()
		r.storeErr(err)
		return err
	}
	modules.BindServerHooks(srv, set)

	r.mu.Lock()
	// Re-check to avoid race with concurrent Stop (defensive).
	if r.srv != nil {
		r.mu.Unlock()
		startCancel()
		_ = srv.Stop(context.Background())
		_ = r.restoreWorkDir()
		return errors.New("runtime already started")
	}
	r.opts = opts // keep possibly overridden NodeID
	r.srv = srv
	r.startCtx = startCtx
	r.startCancel = startCancel
	r.mu.Unlock()

	// Post-start: bind parent connection (root side) by sending an auth register on the persistent parent link.
	if opts.ParentEnable && parentTarget != "" {
		r.startParentBootstrapWatcher(cfg)
	}
	log.Info("hub runtime started", "addr", opts.Addr, "node_id", opts.NodeID, "parent", parentTarget)
	return nil
}

// Stop 终止 watcher、停止 server，并恢复启动前的工作目录。
func (r *Runtime) Stop(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	r.mu.Lock()
	srv := r.srv
	cancel := r.startCancel
	r.srv = nil
	r.startCtx = nil
	r.startCancel = nil
	parentCancel := r.parentWatchCancel
	r.parentWatchCancel = nil
	r.mu.Unlock()

	if parentCancel != nil {
		parentCancel()
	}
	if cancel != nil {
		cancel()
	}
	var stopErr error
	if srv != nil {
		stopErr = srv.Stop(ctx)
	}
	if err := r.restoreWorkDir(); err != nil && stopErr == nil {
		stopErr = err
	}
	return stopErr
}

// Status 返回面向宿主或 CLI 的当前 runtime 快照。
func (r *Runtime) Status() Status {
	r.mu.Lock()
	opts := r.opts
	srv := r.srv
	r.mu.Unlock()

	st := Status{
		Running:       srv != nil,
		Addr:          opts.Addr,
		NodeID:        opts.NodeID,
		ParentEnabled: opts.ParentEnable,
		ParentAddr:    effectiveParentTarget(opts),
		WorkDir:       opts.WorkDir,
		LastError:     r.loadErr(),
	}
	if srv == nil {
		return st
	}
	if conn, ok := findParentConn(srv.ConnManager()); ok {
		st.ParentConnected = true
		st.ParentConnID = conn.ID()
	}
	return st
}

// applyWorkDir 在启动前切到应用私有目录，使相对路径配置能稳定落盘。
func (r *Runtime) applyWorkDir(dir string) (string, error) {
	if strings.TrimSpace(dir) == "" {
		return "", nil
	}
	abs := dir
	if !filepath.IsAbs(abs) {
		if wd, err := os.Getwd(); err == nil && strings.TrimSpace(wd) != "" {
			abs = filepath.Join(wd, dir)
		}
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return "", fmt.Errorf("mkdir workdir: %w", err)
	}
	prev, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
	}
	if err := os.Chdir(abs); err != nil {
		return "", fmt.Errorf("chdir: %w", err)
	}
	r.mu.Lock()
	r.workdirPrev = prev
	r.mu.Unlock()
	return abs, nil
}

// restoreWorkDir 把进程 cwd 恢复到启动前的目录。
func (r *Runtime) restoreWorkDir() error {
	r.mu.Lock()
	prev := r.workdirPrev
	r.workdirPrev = ""
	r.mu.Unlock()
	if strings.TrimSpace(prev) == "" {
		return nil
	}
	return os.Chdir(prev)
}

// startParentBootstrapWatcher 监听持久父连接建立，并在其上补发 auth register。
func (r *Runtime) startParentBootstrapWatcher(cfg core.IConfig) {
	r.mu.Lock()
	if r.srv == nil || r.parentWatchCancel != nil {
		r.mu.Unlock()
		return
	}
	srv := r.srv
	opts := r.opts
	log := r.log
	ctx := r.startCtx
	watchCtx, cancel := context.WithCancel(ctx)
	r.parentWatchCancel = cancel
	r.mu.Unlock()

	go func() {
		ticker := time.NewTicker(300 * time.Millisecond)
		defer ticker.Stop()
		var lastConnID string
		var lastMetaInitConnID string
		var lastRegisterConnID string
		for {
			select {
			case <-watchCtx.Done():
				return
			case <-ticker.C:
			}
			conn, ok := findParentConn(srv.ConnManager())
			if !ok || conn == nil {
				lastConnID = ""
				lastMetaInitConnID = ""
				lastRegisterConnID = ""
				continue
			}
			lastConnID = conn.ID()

			// Local side: mark parent conn as "logged-in" to pass sourceMismatch gating.
			// We don't need exact parent node id here; any non-zero value is sufficient for the
			// later "parent conn exemption" branch. Use 1 as a stable default.
			if lastMetaInitConnID != lastConnID {
				if ensured := ensureConnNodeIDNonZero(conn, 1); ensured {
					log.Info("parent conn meta node_id initialized", "conn", conn.ID(), "node_id", uint32(1))
				}
				lastMetaInitConnID = lastConnID
			}

			// Root side: bind meta(nodeID) on the persistent parent link via auth register.
			if strings.TrimSpace(opts.SelfID) == "" {
				continue
			}
			if lastRegisterConnID == lastConnID {
				continue
			}
			displayName := trimmedConfigValue(cfg, "node.display_name")
			if err := sendRegisterOnConn(watchCtx, conn, opts.SelfID, displayName, opts.ParentJoinPermit, &r.msgSeq); err != nil {
				log.Warn("parent bootstrap register failed", "err", err, "conn", conn.ID())
				r.storeErr(err)
				continue
			}
			lastRegisterConnID = lastConnID
			log.Info("parent bootstrap register sent", "conn", conn.ID(), "self_id", opts.SelfID)
		}
	}()
}

// ensureConnNodeIDNonZero 给 parent 连接补一个非零 nodeID，以通过早期 source 校验。
func ensureConnNodeIDNonZero(conn core.IConnection, fallback uint32) (ensured bool) {
	if conn == nil || fallback == 0 {
		return false
	}
	if v, ok := conn.GetMeta("nodeID"); ok {
		switch vv := v.(type) {
		case uint32:
			if vv != 0 {
				return false
			}
		case uint64:
			if vv != 0 {
				return false
			}
		case int:
			if vv > 0 {
				return false
			}
		case int64:
			if vv > 0 {
				return false
			}
		}
	}
	conn.SetMeta("nodeID", fallback)
	return true
}

// selfRegisterNodeID 在 runtime 启动前短路完成一次自注册，拿到父节点分配的 nodeID。
func selfRegisterNodeID(
	ctx context.Context,
	parentTarget, selfID, joinPermit string,
	dialer func(context.Context, string) (core.IConnection, error),
	log *slog.Logger,
) (uint32, error) {
	// Use a short timeout to avoid blocking embedded runtimes on flaky networks.
	cctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	nodeID, _, err := bootstrap.SelfRegister(cctx, bootstrap.SelfRegisterOptions{
		ParentAddr: parentTarget,
		Dial: func(ctx context.Context) (core.IConnection, error) {
			if dialer == nil {
				return nil, errors.New("parent dialer required")
			}
			return dialer(ctx, parentTarget)
		},
		SelfID:     selfID,
		JoinPermit: joinPermit,
		Timeout:    8 * time.Second,
		DoLogin:    false,
		Logger:     log,
	})
	if err != nil {
		return 0, err
	}
	if nodeID == 0 {
		return 0, errors.New("self register returned node id 0")
	}
	return nodeID, nil
}

// trimmedConfigValue 读取配置键并返回去空白后的结果。
func trimmedConfigValue(cfg core.IConfig, key string) string {
	if cfg == nil {
		return ""
	}
	val, ok := cfg.Get(key)
	if !ok {
		return ""
	}
	return strings.TrimSpace(val)
}

// sendRegisterOnConn 在已建立的父连接上发送一次 auth register，绑定该持久链路的 meta(nodeID)。
func sendRegisterOnConn(ctx context.Context, conn core.IConnection, selfID, displayName, joinPermit string, seq *atomic.Uint32) error {
	if ctx == nil {
		return errors.New("ctx nil")
	}
	if conn == nil {
		return errors.New("conn nil")
	}
	if strings.TrimSpace(selfID) == "" {
		return errors.New("self id required for parent bootstrap")
	}
	data := map[string]any{
		"device_id": strings.TrimSpace(selfID),
	}
	if trimmedDisplayName := strings.TrimSpace(displayName); trimmedDisplayName != "" {
		data["display_name"] = trimmedDisplayName
	}
	if trimmedPermit := strings.TrimSpace(joinPermit); trimmedPermit != "" {
		data["join_permit"] = trimmedPermit
	}
	payload, _ := json.Marshal(map[string]any{
		"action": "register",
		"data":   data,
	})
	msgID := seq.Add(1)
	hdr := (&header.HeaderTcp{}).
		WithMajor(header.MajorCmd).
		WithSubProto(2). // auth
		WithSourceID(0).
		WithTargetID(0).
		WithMsgID(msgID).
		WithTimestamp(uint32(time.Now().Unix()))

	return conn.SendWithHeader(hdr, payload, header.HeaderTcpCodec{})
}

// findParentConn 在连接管理器里查找当前父链连接。
func findParentConn(cm core.IConnectionManager) (core.IConnection, bool) {
	if cm == nil {
		return nil, false
	}
	var parent core.IConnection
	cm.Range(func(c core.IConnection) bool {
		if c == nil {
			return true
		}
		if role, ok := c.GetMeta(core.MetaRoleKey); ok {
			if s, ok2 := role.(string); ok2 && s == core.RoleParent {
				parent = c
				return false
			}
		}
		return true
	})
	return parent, parent != nil
}

// boolString 把布尔值映射成配置层使用的字符串字面量。
func boolString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

// effectiveParentTarget 优先使用带 scheme 的 ParentEndpoint，再退化到旧的 ParentAddr。
func effectiveParentTarget(opts Options) string {
	if strings.TrimSpace(opts.ParentEndpoint) != "" {
		return strings.TrimSpace(opts.ParentEndpoint)
	}
	return strings.TrimSpace(opts.ParentAddr)
}

// parseParentEndpoint 统一验证 parent endpoint，并兼容旧的裸 tcp 地址写法。
func parseParentEndpoint(target string) (scheme string, tcpAddr string, err error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "", "", errors.New("parent target empty")
	}
	// Backward compatible: host:port implies TCP.
	if !strings.Contains(target, "://") {
		return "tcp", target, nil
	}
	u, err := url.Parse(target)
	if err != nil {
		return "", "", fmt.Errorf("parse parent endpoint: %w", err)
	}
	scheme = strings.ToLower(strings.TrimSpace(u.Scheme))
	switch scheme {
	case "tcp":
		host := strings.TrimSpace(u.Host)
		if host == "" {
			return "", "", errors.New("tcp endpoint host is empty")
		}
		return "tcp", host, nil
	case rfcomm_listener.EndpointSchemeRFCOMM:
		if _, err := rfcomm_listener.ParseEndpoint(target); err != nil {
			return "", "", err
		}
		return scheme, "", nil
	case quic_listener.EndpointSchemeQUIC:
		if _, err := quic_listener.ParseEndpoint(target); err != nil {
			return "", "", err
		}
		return scheme, "", nil
	default:
		return "", "", fmt.Errorf("unsupported parent endpoint scheme: %s", scheme)
	}
}

// dialParentEndpoint 根据 endpoint scheme 选择对应 transport 的拨号器。
func dialParentEndpoint(ctx context.Context, target string) (core.IConnection, error) {
	scheme, tcpAddr, err := parseParentEndpoint(target)
	if err != nil {
		return nil, err
	}
	switch scheme {
	case "tcp":
		var d net.Dialer
		raw, err := d.DialContext(ctx, "tcp", tcpAddr)
		if err != nil {
			return nil, err
		}
		return tcp_listener.NewTCPConnection(raw), nil
	case rfcomm_listener.EndpointSchemeRFCOMM:
		return rfcomm_listener.DialEndpoint(ctx, target)
	case quic_listener.EndpointSchemeQUIC:
		return quic_listener.DialEndpoint(ctx, target)
	default:
		return nil, fmt.Errorf("unsupported parent endpoint scheme: %s", scheme)
	}
}

// storeErr 保存最近一次可见错误，供宿主通过 Status 读取。
func (r *Runtime) storeErr(err error) {
	if err == nil {
		return
	}
	r.lastErr.Store(err.Error())
}

// loadErr 读取最近一次记录的错误文本。
func (r *Runtime) loadErr() string {
	if v := r.lastErr.Load(); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
