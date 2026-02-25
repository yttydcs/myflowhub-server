package hubruntime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-core/bootstrap"
	coreconfig "github.com/yttydcs/myflowhub-core/config"
	"github.com/yttydcs/myflowhub-core/connmgr"
	"github.com/yttydcs/myflowhub-core/header"
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

func New(opts Options) (*Runtime, error) {
	opts.Normalize()
	if strings.TrimSpace(opts.Addr) == "" {
		return nil, errors.New("addr required")
	}
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
	return &Runtime{opts: opts, log: opts.Logger}, nil
}

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

	// Pre-start: if parent enabled and self id provided, self-register to obtain/confirm node id.
	if opts.ParentEnable && opts.ParentAddr != "" && opts.SelfID != "" {
		nodeID, err := selfRegisterNodeID(ctx, opts.ParentAddr, opts.SelfID, log)
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

	cfg := buildConfig(opts)

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

	lst := tcp_listener.New(opts.Addr, tcp_listener.Options{
		KeepAlive:       true,
		KeepAlivePeriod: 30 * time.Second,
		Logger:          log,
	})
	codec := header.HeaderTcpCodec{}

	srv, err := server.New(server.Options{
		Name:     "HubServer",
		Logger:   log,
		Process:  dispatcher,
		Codec:    codec,
		Listener: lst,
		Config:   cfg,
		Manager:  cm,
		NodeID:   opts.NodeID,
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
	if opts.ParentEnable && opts.ParentAddr != "" {
		r.startParentBootstrapWatcher()
	}
	log.Info("hub runtime started", "addr", opts.Addr, "node_id", opts.NodeID, "parent", opts.ParentAddr)
	return nil
}

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
		ParentAddr:    opts.ParentAddr,
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

func (r *Runtime) startParentBootstrapWatcher() {
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
			if err := sendRegisterOnConn(watchCtx, conn, opts.SelfID, &r.msgSeq); err != nil {
				log.Warn("parent bootstrap register failed", "err", err, "conn", conn.ID())
				r.storeErr(err)
				continue
			}
			lastRegisterConnID = lastConnID
			log.Info("parent bootstrap register sent", "conn", conn.ID(), "self_id", opts.SelfID)
		}
	}()
}

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

func selfRegisterNodeID(ctx context.Context, parentAddr, selfID string, log *slog.Logger) (uint32, error) {
	// Use a short timeout to avoid blocking embedded runtimes on flaky networks.
	cctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	nodeID, _, err := bootstrap.SelfRegister(cctx, bootstrap.SelfRegisterOptions{
		ParentAddr: parentAddr,
		SelfID:     selfID,
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

func sendRegisterOnConn(ctx context.Context, conn core.IConnection, selfID string, seq *atomic.Uint32) error {
	if ctx == nil {
		return errors.New("ctx nil")
	}
	if conn == nil {
		return errors.New("conn nil")
	}
	if strings.TrimSpace(selfID) == "" {
		return errors.New("self id required for parent bootstrap")
	}
	payload, _ := json.Marshal(map[string]any{
		"action": "register",
		"data":   map[string]any{"device_id": selfID},
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

func buildConfig(opts Options) core.IConfig {
	reconnect := "3"
	if opts.ParentReconnectSec > 0 {
		reconnect = fmt.Sprintf("%d", opts.ParentReconnectSec)
	}
	data := map[string]string{
		"addr":                           opts.Addr,
		coreconfig.KeyParentAddr:         opts.ParentAddr,
		coreconfig.KeyParentEnable:       boolString(opts.ParentEnable),
		coreconfig.KeyParentReconnectSec: reconnect,

		coreconfig.KeyProcChannelCount:   fmt.Sprintf("%d", opts.ProcChannels),
		coreconfig.KeyProcWorkersPerChan: fmt.Sprintf("%d", opts.ProcWorkers),
		coreconfig.KeyProcChannelBuffer:  fmt.Sprintf("%d", opts.ProcBuffer),

		coreconfig.KeySendChannelCount:   fmt.Sprintf("%d", opts.SendChannels),
		coreconfig.KeySendWorkersPerChan: fmt.Sprintf("%d", opts.SendWorkers),
		coreconfig.KeySendChannelBuffer:  fmt.Sprintf("%d", opts.SendChannelBuffer),
		coreconfig.KeySendConnBuffer:     fmt.Sprintf("%d", opts.SendConnBuffer),

		coreconfig.KeyRoutingForwardRemote: "true",

		coreconfig.KeyAuthDefaultRole:  opts.AuthDefaultRole,
		coreconfig.KeyAuthDefaultPerms: opts.AuthDefaultPerms,
		coreconfig.KeyAuthNodeRoles:    opts.AuthNodeRoles,
		coreconfig.KeyAuthRolePerms:    opts.AuthRolePerms,
	}
	return coreconfig.NewMap(data)
}

func boolString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func (r *Runtime) storeErr(err error) {
	if err == nil {
		return
	}
	r.lastErr.Store(err.Error())
}

func (r *Runtime) loadErr() string {
	if v := r.lastErr.Load(); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
