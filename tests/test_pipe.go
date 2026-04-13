package tests

// Context: This file lives in the Server assembly layer and supports test_pipe.

import "io"

// nopPipe is a minimal core.IPipe implementation for tests that never exercise stream I/O.
type nopPipe struct{}

func (nopPipe) Read([]byte) (int, error)    { return 0, io.EOF }
func (nopPipe) Write(p []byte) (int, error) { return len(p), nil }
func (nopPipe) Close() error                { return nil }
