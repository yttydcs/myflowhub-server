package main

// 本文件提供 Server 中与 `main` 相关的命令入口。

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	coreconfig "github.com/yttydcs/myflowhub-core/config"
	"github.com/yttydcs/myflowhub-server/hubruntime"
)

// main 负责把 env/flag 配置归一化后交给 hubruntime 启停。
func main() {
	opts := hubruntime.DefaultOptionsFromEnv()
	nodeID := uint(opts.NodeID)

	flag.BoolVar(&opts.TCPEnable, "tcp-enable", opts.TCPEnable, "enable tcp listener")
	flag.StringVar(&opts.Addr, "addr", opts.Addr, "listen address")
	flag.BoolVar(&opts.QUICEnable, "quic-enable", opts.QUICEnable, "enable quic listener")
	flag.StringVar(&opts.QUICAddr, "quic-addr", opts.QUICAddr, "quic listen address")
	flag.StringVar(&opts.QUICALPN, "quic-alpn", opts.QUICALPN, "quic ALPN protocol")
	flag.StringVar(&opts.QUICCertFile, "quic-cert-file", opts.QUICCertFile, "quic tls cert file path")
	flag.StringVar(&opts.QUICKeyFile, "quic-key-file", opts.QUICKeyFile, "quic tls key file path")
	flag.BoolVar(&opts.QUICDevCertAuto, "quic-dev-cert-auto", opts.QUICDevCertAuto, "auto-generate self-signed quic cert/key for development when cert/key are missing")
	flag.StringVar(&opts.QUICClientCAFile, "quic-client-ca-file", opts.QUICClientCAFile, "quic client CA file path")
	flag.BoolVar(&opts.QUICRequireClientCert, "quic-require-client-cert", opts.QUICRequireClientCert, "require and verify quic client cert")
	flag.UintVar(&nodeID, "node-id", nodeID, "node id for this hub (0 means auto when parent+self-id enabled)")
	flag.StringVar(&opts.ParentEndpoint, "parent-endpoint", opts.ParentEndpoint, "parent endpoint, e.g. tcp://127.0.0.1:9000 or bt+rfcomm://... or quic://127.0.0.1:9000?server_name=...")
	flag.StringVar(&opts.ParentAddr, "parent", opts.ParentAddr, "parent address")
	flag.BoolVar(&opts.ParentEnable, "parent-enable", opts.ParentEnable, "enable parent link")
	flag.IntVar(&opts.ParentReconnectSec, "parent-reconnect", opts.ParentReconnectSec, "parent reconnect seconds")
	flag.BoolVar(&opts.RFCOMMEnable, "rfcomm-enable", opts.RFCOMMEnable, "enable bluetooth rfcomm listener")
	flag.StringVar(&opts.RFCOMMUUID, "rfcomm-uuid", opts.RFCOMMUUID, "rfcomm service uuid (default: MyFlowHub)")
	flag.IntVar(&opts.RFCOMMChannel, "rfcomm-channel", opts.RFCOMMChannel, "rfcomm channel (1-30, 0 means auto/uuid-first)")
	flag.StringVar(&opts.RFCOMMAdapter, "rfcomm-adapter", opts.RFCOMMAdapter, "rfcomm adapter (linux, default: hci0)")
	flag.BoolVar(&opts.RFCOMMInsecure, "rfcomm-insecure", opts.RFCOMMInsecure, "use insecure rfcomm (android, default: false)")
	flag.IntVar(&opts.ProcChannels, "proc-channels", opts.ProcChannels, "process channel count")
	flag.IntVar(&opts.ProcWorkers, "proc-workers", opts.ProcWorkers, "process workers per channel")
	flag.IntVar(&opts.ProcBuffer, "proc-buffer", opts.ProcBuffer, "process channel buffer")
	flag.IntVar(&opts.SendChannels, "send-channels", opts.SendChannels, "send dispatcher channels")
	flag.IntVar(&opts.SendWorkers, "send-workers", opts.SendWorkers, "send dispatcher workers per channel")
	flag.IntVar(&opts.SendChannelBuffer, "send-channel-buffer", opts.SendChannelBuffer, "send dispatcher channel buffer")
	flag.IntVar(&opts.SendConnBuffer, "send-conn-buffer", opts.SendConnBuffer, "per-connection send buffer")
	flag.StringVar(&opts.AuthDefaultRole, "auth-default-role", opts.AuthDefaultRole, "default role for nodes")
	flag.StringVar(&opts.AuthDefaultPerms, "auth-default-perms", opts.AuthDefaultPerms, "default perms (comma separated)")
	flag.StringVar(&opts.AuthNodeRoles, "auth-node-roles", opts.AuthNodeRoles, "node roles mapping, e.g. 1:admin;2:node")
	flag.StringVar(&opts.AuthRolePerms, "auth-role-perms", opts.AuthRolePerms, "role perms mapping, e.g. admin:p1,p2;node:p3")
	flag.StringVar(&opts.WorkDir, "workdir", opts.WorkDir, "working directory for relative paths (optional)")
	flag.StringVar(&opts.SelfID, "self-id", opts.SelfID, "self device id (for parent self-register/bootstrap)")
	flag.Parse()

	opts.NodeID = uint32(nodeID)
	captureFlagOverrides(&opts)
	opts.Logger = setupLogger()
	opts.Normalize()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rt, err := hubruntime.New(opts)
	if err != nil {
		slog.Error("init runtime failed", "err", err)
		os.Exit(1)
	}
	if err := rt.Start(ctx); err != nil {
		slog.Error("start runtime failed", "err", err)
		os.Exit(1)
	}
	st := rt.Status()
	slog.Info("hub server started", "addr", st.Addr, "node_id", st.NodeID, "parent", st.ParentAddr)

	waitSignal()

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer stopCancel()
	if err := rt.Stop(stopCtx); err != nil {
		slog.Error("stop runtime failed", "err", err)
		os.Exit(1)
	}
	slog.Info("hub server stopped")
}

// setupLogger 初始化命令行版本使用的全局文本 logger。
func setupLogger() *slog.Logger {
	level := new(slog.LevelVar)
	level.Set(slog.LevelInfo)
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	l := slog.New(handler)
	slog.SetDefault(l)
	return l
}

// waitSignal 阻塞等待中断信号，作为 CLI 版 runtime 的退出钩子。
func waitSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch
}

// captureFlagOverrides 把命令行明确传入的 key 标记成显式覆盖项。
func captureFlagOverrides(opts *hubruntime.Options) {
	if opts == nil {
		return
	}
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "addr":
			opts.AddConfigOverrideKeys("addr")
		case "parent", "parent-endpoint":
			opts.AddConfigOverrideKeys(coreconfig.KeyParentAddr)
		case "parent-enable":
			opts.AddConfigOverrideKeys(coreconfig.KeyParentEnable)
		case "parent-reconnect":
			opts.AddConfigOverrideKeys(coreconfig.KeyParentReconnectSec)
		case "proc-channels":
			opts.AddConfigOverrideKeys(coreconfig.KeyProcChannelCount)
		case "proc-workers":
			opts.AddConfigOverrideKeys(coreconfig.KeyProcWorkersPerChan)
		case "proc-buffer":
			opts.AddConfigOverrideKeys(coreconfig.KeyProcChannelBuffer)
		case "send-channels":
			opts.AddConfigOverrideKeys(coreconfig.KeySendChannelCount)
		case "send-workers":
			opts.AddConfigOverrideKeys(coreconfig.KeySendWorkersPerChan)
		case "send-channel-buffer":
			opts.AddConfigOverrideKeys(coreconfig.KeySendChannelBuffer)
		case "send-conn-buffer":
			opts.AddConfigOverrideKeys(coreconfig.KeySendConnBuffer)
		case "auth-default-role":
			opts.AddConfigOverrideKeys(coreconfig.KeyAuthDefaultRole)
		case "auth-default-perms":
			opts.AddConfigOverrideKeys(coreconfig.KeyAuthDefaultPerms)
		case "auth-node-roles":
			opts.AddConfigOverrideKeys(coreconfig.KeyAuthNodeRoles)
		case "auth-role-perms":
			opts.AddConfigOverrideKeys(coreconfig.KeyAuthRolePerms)
		}
	})
}
