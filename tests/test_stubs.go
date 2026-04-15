package tests

// 本文件覆盖 Server 装配层中与 `test_stubs` 相关的集成或单元行为。

import (
	"context"
	"net"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-core/config"
	"github.com/yttydcs/myflowhub-core/eventbus"
)

type stubServer struct {
	nodeID uint32
	cm     core.IConnectionManager
	sends  []sendCall
	bus    eventbus.IBus
}

type sendCall struct {
	connID  string
	target  uint32
	major   uint8
	msgID   uint32
	traceID uint32
}

// Start 让 stubServer 满足 core.IServer 接口，但不真正启动任何组件。
func (s *stubServer) Start(context.Context) error { return nil }

// Stop 对测试桩来说没有实际资源要释放。
func (s *stubServer) Stop(context.Context) error { return nil }

// Config 返回一个最小空配置，供装配逻辑读取。
func (s *stubServer) Config() core.IConfig { return config.NewMap(nil) }

// ConnManager 暴露测试自行注入的连接表。
func (s *stubServer) ConnManager() core.IConnectionManager { return s.cm }

// Process 当前这组测试不关心 process 细节，因此返回 nil。
func (s *stubServer) Process() core.IProcess { return nil }

// HeaderCodec 当前测试只观察 Send 记录，不需要真实 codec。
func (s *stubServer) HeaderCodec() core.IHeaderCodec { return nil }

// NodeID 返回测试中的本地节点号。
func (s *stubServer) NodeID() uint32 { return s.nodeID }

// UpdateNodeID 允许测试动态调整本地节点号。
func (s *stubServer) UpdateNodeID(id uint32) { s.nodeID = id }

// EventBus 惰性创建总线，避免未使用的测试提前初始化。
func (s *stubServer) EventBus() eventbus.IBus {
	if s.bus == nil {
		s.bus = eventbus.New(eventbus.Options{})
	}
	return s.bus
}

// Send 不做真实网络发送，只把 header 关键信息记录到 sends 切片。
func (s *stubServer) Send(_ context.Context, connID string, hdr core.IHeader, _ []byte) error {
	s.sends = append(s.sends, sendCall{
		connID:  connID,
		target:  hdr.TargetID(),
		major:   hdr.Major(),
		msgID:   hdr.GetMsgID(),
		traceID: hdr.GetTraceID(),
	})
	return nil
}

type stubConn struct {
	id   string
	meta map[string]any
}

// newStubConn 创建带独立 metadata map 的测试连接。
func newStubConn(id string) *stubConn {
	return &stubConn{id: id, meta: make(map[string]any)}
}

// ID 返回测试连接 ID。
func (c *stubConn) ID() string { return c.id }

// Pipe 复用 nopPipe，避免为装配测试引入真实管道。
func (c *stubConn) Pipe() core.IPipe { return nopPipe{} }

// Close 在测试里不需要真正关闭任何底层资源。
func (c *stubConn) Close() error {
	return nil
}

// OnReceive 保持空实现即可。
func (c *stubConn) OnReceive(core.ReceiveHandler) {}

// SetMeta 写入测试断言所需的连接元数据。
func (c *stubConn) SetMeta(key string, val any) { c.meta[key] = val }

// GetMeta 读取测试断言所需的连接元数据。
func (c *stubConn) GetMeta(key string) (any, bool) {
	v, ok := c.meta[key]
	return v, ok
}

// Metadata 暴露 metadata map，便于测试直接检查。
func (c *stubConn) Metadata() map[string]any { return c.meta }

// LocalAddr 返回固定占位地址。
func (c *stubConn) LocalAddr() net.Addr { return mockAddr{} }

// RemoteAddr 返回固定占位地址。
func (c *stubConn) RemoteAddr() net.Addr { return mockAddr{} }

// Reader 在这组测试里从未被真正消费。
func (c *stubConn) Reader() core.IReader { return nil }

// SetReader 保持空实现即可。
func (c *stubConn) SetReader(core.IReader) {}

// DispatchReceive 不复现底层收包流程。
func (c *stubConn) DispatchReceive(core.IHeader, []byte) {}

// Send 忽略真实发送过程。
func (c *stubConn) Send([]byte) error { return nil }

// SendWithHeader 忽略真实发送过程，只满足接口。
func (c *stubConn) SendWithHeader(core.IHeader, []byte, core.IHeaderCodec) error {
	return nil
}
