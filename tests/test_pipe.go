package tests

// 本文件覆盖 Server 装配层中与 `test_pipe` 相关的集成或单元行为。

import "io"

// nopPipe is a minimal core.IPipe implementation for tests that never exercise stream I/O.
type nopPipe struct{}

// Read 立即返回 EOF，模拟一个不会产生日志噪音的空管道。
func (nopPipe) Read([]byte) (int, error) { return 0, io.EOF }

// Write 假装全部写入成功，让上层接口测试不被底层 I/O 打断。
func (nopPipe) Write(p []byte) (int, error) { return len(p), nil }

// Close 对空管道来说是 no-op。
func (nopPipe) Close() error { return nil }
