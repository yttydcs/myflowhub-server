package tests

// 本文件覆盖 Server 装配层中与 `test_mocks` 相关的集成或单元行为。

import (
	"net"

	core "github.com/yttydcs/myflowhub-core"
)

type mockAddr struct{}

// Network 让 mock 地址在日志和断言里表现得像一条 TCP 连接。
func (mockAddr) Network() string { return "tcp" }

// String 返回稳定的占位地址，避免测试依赖真实端口。
func (mockAddr) String() string { return "127.0.0.1:0" }

// mockConnection is a lightweight implementation of core.IConnection used in multiple tests.
type mockConnection struct {
	id   string
	meta map[string]any
}

var _ core.IConnection = (*mockConnection)(nil)

// ID 返回测试连接 ID，供 dispatcher 与路由断言使用。
func (m *mockConnection) ID() string { return m.id }

// Pipe 返回一个永不真正收发数据的空实现，满足接口要求即可。
func (m *mockConnection) Pipe() core.IPipe { return nopPipe{} }

// Close 对测试连接来说是 no-op。
func (m *mockConnection) Close() error { return nil }

// OnReceive 测试里不需要真实回调注册，因此保持空实现。
func (m *mockConnection) OnReceive(core.ReceiveHandler) {}
func (m *mockConnection) SetMeta(k string, v any) {
	if m.meta == nil {
		m.meta = make(map[string]any)
	}
	m.meta[k] = v
}

// GetMeta 读取测试期间手工塞入的连接元数据。
func (m *mockConnection) GetMeta(k string) (any, bool) {
	if m.meta == nil {
		return nil, false
	}
	v, ok := m.meta[k]
	return v, ok
}

// Metadata 直接暴露底层 map，便于测试断言。
func (m *mockConnection) Metadata() map[string]any { return m.meta }

// LocalAddr 返回固定占位地址，避免引入真实 socket。
func (m *mockConnection) LocalAddr() net.Addr { return mockAddr{} }

// RemoteAddr 返回固定占位地址，避免引入真实 socket。
func (m *mockConnection) RemoteAddr() net.Addr { return mockAddr{} }

// Reader 测试里不会真正消费流式数据。
func (m *mockConnection) Reader() core.IReader { return nil }

// SetReader 保持空实现即可满足接口。
func (m *mockConnection) SetReader(core.IReader) {}

// DispatchReceive 测试里不复现底层 frame 分发。
func (m *mockConnection) DispatchReceive(core.IHeader, []byte) {}

// Send 忽略裸 payload 发送，因为当前测试只关心是否被调用。
func (m *mockConnection) Send([]byte) error { return nil }

// SendWithHeader 忽略真实发送过程，让上层逻辑专注于装配和路由断言。
func (m *mockConnection) SendWithHeader(core.IHeader, []byte, core.IHeaderCodec) error { return nil }
