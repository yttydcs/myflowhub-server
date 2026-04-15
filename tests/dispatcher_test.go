package tests

// 本文件覆盖 Server 装配层中与 `dispatcher` 相关的集成或单元行为。

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-core/config"
	"github.com/yttydcs/myflowhub-core/header"
	"github.com/yttydcs/myflowhub-core/process"
)

const (
	testSubProtoEcho = 1
)

// TestDispatcherRoutesSubProtocols 验证 dispatcher 能按 SubProto 把消息送到对应 handler。
func TestDispatcherRoutesSubProtocols(t *testing.T) {
	cfg := config.NewMap(map[string]string{
		config.KeyProcChannelCount:   "1",
		config.KeyProcWorkersPerChan: "1",
		config.KeyProcChannelBuffer:  "8",
	})
	base := &spyBaseProcess{}
	dispatcher, err := process.NewDispatcherFromConfig(cfg, base, slog.Default())
	if err != nil {
		t.Fatalf("failed to create dispatcher: %v", err)
	}
	defer dispatcher.Shutdown()

	chEcho := make(chan string, 1)
	if err := dispatcher.RegisterHandler(&recordHandler{sub: testSubProtoEcho, ch: chEcho}); err != nil {
		t.Fatalf("register echo handler: %v", err)
	}

	conn := &mockConnection{id: "test-conn"}
	hdrEcho := &header.HeaderTcp{}
	hdrEcho.WithSubProto(testSubProtoEcho)

	dispatcher.OnReceive(context.Background(), conn, hdrEcho, []byte("hello"))

	expectMessage(t, chEcho, "test-conn|hello")

	if got := base.receives.Load(); got != 1 {
		t.Fatalf("base OnReceive called %d times, want 1", got)
	}
}

// TestDispatcherConfigSnapshot 验证 dispatcher 暴露的并发配置快照和输入一致。
func TestDispatcherConfigSnapshot(t *testing.T) {
	cfg := config.NewMap(map[string]string{
		config.KeyProcChannelCount:   "3",
		config.KeyProcWorkersPerChan: "2",
		config.KeyProcChannelBuffer:  "32",
	})
	dispatcher, err := process.NewDispatcherFromConfig(cfg, nil, slog.Default())
	if err != nil {
		t.Fatalf("failed to create dispatcher: %v", err)
	}
	defer dispatcher.Shutdown()

	channels, workers, buffer := dispatcher.ConfigSnapshot()
	if channels != 3 || workers != 2 || buffer != 32 {
		t.Fatalf("snapshot mismatch: got channels=%d workers=%d buffer=%d", channels, workers, buffer)
	}
}

type spyBaseProcess struct {
	listens  atomic.Int64
	receives atomic.Int64
	sends    atomic.Int64
	closes   atomic.Int64
}

// OnListen 统计底层 listen 钩子是否被调用。
func (s *spyBaseProcess) OnListen(core.IConnection) { s.listens.Add(1) }

// OnReceive 统计 dispatcher 是否仍会先经过 base process。
func (s *spyBaseProcess) OnReceive(context.Context, core.IConnection, core.IHeader, []byte) {
	s.receives.Add(1)
}

// OnSend 统计发送钩子调用次数。
func (s *spyBaseProcess) OnSend(context.Context, core.IConnection, core.IHeader, []byte) error {
	s.sends.Add(1)
	return nil
}

// OnClose 统计关闭钩子调用次数。
func (s *spyBaseProcess) OnClose(core.IConnection) { s.closes.Add(1) }

type recordHandler struct {
	sub uint8
	ch  chan<- string
}

// SubProto 返回这个测试 handler 绑定的子协议号。
func (h *recordHandler) SubProto() uint8 { return h.sub }

// AllowSourceMismatch 避免测试被登录态校验打断。
func (h *recordHandler) AllowSourceMismatch() bool { return true }

// AcceptCmd 保持默认 false，避免掺入命令链路逻辑。
func (h *recordHandler) AcceptCmd() bool { return false }

// Init 这个测试 handler 不需要额外初始化。
func (h *recordHandler) Init() bool { return true }

// OnReceive 把观测到的连接和 payload 发到通道，供测试断言。
func (h *recordHandler) OnReceive(_ context.Context, conn core.IConnection, _ core.IHeader, payload []byte) {
	h.ch <- fmt.Sprintf("%s|%s", conn.ID(), string(payload))
}

// expectMessage 统一处理带超时的通道断言，避免单测在失败时挂死。
func expectMessage(t *testing.T, ch <-chan string, want string) {
	t.Helper()
	select {
	case got := <-ch:
		if got != want {
			t.Fatalf("unexpected message: got %q want %q", got, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for message %q", want)
	}
}
