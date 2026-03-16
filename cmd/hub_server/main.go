package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yttydcs/myflowhub-server/hubruntime"
)

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

func setupLogger() *slog.Logger {
	level := new(slog.LevelVar)
	level.Set(slog.LevelInfo)
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	l := slog.New(handler)
	slog.SetDefault(l)
	return l
}

func waitSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch
}
