package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-core/config"
	"github.com/yttydcs/myflowhub-core/connmgr"
	"github.com/yttydcs/myflowhub-core/header"
	"github.com/yttydcs/myflowhub-core/listener/tcp_listener"
	"github.com/yttydcs/myflowhub-core/process"
	"github.com/yttydcs/myflowhub-core/server"
	"github.com/yttydcs/myflowhub-server/internal/handler"
	authhandler "github.com/yttydcs/myflowhub-server/internal/handler/auth"
)

type options struct {
	addr               string
	nodeID             uint
	parentAddr         string
	parentEnable       bool
	parentReconnectSec int
	procChannels       int
	procWorkers        int
	procBuffer         int
	sendChannels       int
	sendWorkers        int
	sendChannelBuffer  int
	sendConnBuffer     int
}

func main() {
	opts := defaultOptions()
	flag.StringVar(&opts.addr, "addr", opts.addr, "listen address")
	flag.UintVar(&opts.nodeID, "node-id", opts.nodeID, "node id for this hub")
	flag.StringVar(&opts.parentAddr, "parent", opts.parentAddr, "parent address")
	flag.BoolVar(&opts.parentEnable, "parent-enable", opts.parentEnable, "enable parent link")
	flag.IntVar(&opts.parentReconnectSec, "parent-reconnect", opts.parentReconnectSec, "parent reconnect seconds")
	flag.IntVar(&opts.procChannels, "proc-channels", opts.procChannels, "process channel count")
	flag.IntVar(&opts.procWorkers, "proc-workers", opts.procWorkers, "process workers per channel")
	flag.IntVar(&opts.procBuffer, "proc-buffer", opts.procBuffer, "process channel buffer")
	flag.IntVar(&opts.sendChannels, "send-channels", opts.sendChannels, "send dispatcher channels")
	flag.IntVar(&opts.sendWorkers, "send-workers", opts.sendWorkers, "send dispatcher workers per channel")
	flag.IntVar(&opts.sendChannelBuffer, "send-channel-buffer", opts.sendChannelBuffer, "send dispatcher channel buffer")
	flag.IntVar(&opts.sendConnBuffer, "send-conn-buffer", opts.sendConnBuffer, "per-connection send buffer")
	flag.Parse()

	if opts.parentAddr != "" {
		opts.parentEnable = true
	}

	log := setupLogger()
	cfg := buildConfig(opts)

	cm := connmgr.New()
	base := process.NewPreRoutingProcess(log).WithConfig(cfg)
	dispatcher, err := process.NewDispatcherFromConfig(cfg, base, log)
	if err != nil {
		log.Error("build dispatcher failed", "err", err)
		os.Exit(1)
	}
	if err := dispatcher.RegisterHandler(handler.NewEchoHandler(log)); err != nil {
		log.Error("register echo handler failed", "err", err)
		os.Exit(1)
	}
	if err := dispatcher.RegisterHandler(authhandler.NewLoginHandlerWithConfig(cfg, log)); err != nil {
		log.Error("register login handler failed", "err", err)
		os.Exit(1)
	}
	if err := dispatcher.RegisterHandler(handler.NewVarStoreHandlerWithConfig(cfg, log)); err != nil {
		log.Error("register varstore handler failed", "err", err)
		os.Exit(1)
	}
	dispatcher.RegisterDefaultHandler(handler.NewDefaultForwardHandler(cfg, log))

	lst := tcp_listener.New(opts.addr, tcp_listener.Options{
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
		NodeID:   uint32(opts.nodeID),
	})
	if err != nil {
		log.Error("init server failed", "err", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := srv.Start(ctx); err != nil {
		log.Error("start server failed", "err", err)
		os.Exit(1)
	}
	log.Info("hub server started", "addr", opts.addr, "node_id", opts.nodeID, "parent", opts.parentAddr)

	waitSignal()

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer stopCancel()
	if err := srv.Stop(stopCtx); err != nil {
		log.Error("stop server failed", "err", err)
		os.Exit(1)
	}
	log.Info("hub server stopped")
}

func buildConfig(opts options) core.IConfig {
	data := map[string]string{
		"addr":                         opts.addr,
		config.KeyParentAddr:           opts.parentAddr,
		config.KeyParentEnable:         strconv.FormatBool(opts.parentEnable),
		config.KeyParentReconnectSec:   strconv.Itoa(opts.parentReconnectSec),
		config.KeyProcChannelCount:     strconv.Itoa(opts.procChannels),
		config.KeyProcWorkersPerChan:   strconv.Itoa(opts.procWorkers),
		config.KeyProcChannelBuffer:    strconv.Itoa(opts.procBuffer),
		config.KeySendChannelCount:     strconv.Itoa(opts.sendChannels),
		config.KeySendWorkersPerChan:   strconv.Itoa(opts.sendWorkers),
		config.KeySendChannelBuffer:    strconv.Itoa(opts.sendChannelBuffer),
		config.KeySendConnBuffer:       strconv.Itoa(opts.sendConnBuffer),
		config.KeyRoutingForwardRemote: "true",
	}
	return config.NewMap(data)
}

func defaultOptions() options {
	return options{
		addr:               getenv("HUB_ADDR", ":9000"),
		nodeID:             getenvUint("HUB_NODE_ID", 1),
		parentAddr:         getenv("HUB_PARENT_ADDR", ""),
		parentEnable:       core.ParseBool(getenv("HUB_PARENT_ENABLE", "false"), false),
		parentReconnectSec: int(getenvInt("HUB_PARENT_RECONNECT", 3)),
		procChannels:       int(getenvInt("HUB_PROC_CHANNELS", 2)),
		procWorkers:        int(getenvInt("HUB_PROC_WORKERS", 2)),
		procBuffer:         int(getenvInt("HUB_PROC_BUFFER", 128)),
		sendChannels:       int(getenvInt("HUB_SEND_CHANNELS", 1)),
		sendWorkers:        int(getenvInt("HUB_SEND_WORKERS", 1)),
		sendChannelBuffer:  int(getenvInt("HUB_SEND_CHANNEL_BUFFER", 64)),
		sendConnBuffer:     int(getenvInt("HUB_SEND_CONN_BUFFER", 64)),
	}
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

func getenvUint(key string, def uint) uint {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if n, err := strconv.ParseUint(v, 10, 32); err == nil {
			return uint(n)
		}
	}
	return def
}
