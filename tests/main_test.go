package tests

import (
	"fmt"
	"os"
	"testing"
	"time"
)

// 全局测试超时：30 秒内未结束直接退出，避免卡住。
func TestMain(m *testing.M) {
	timer := time.AfterFunc(30*time.Second, func() {
		fmt.Fprintln(os.Stderr, "tests aborted: exceeded 30s timeout")
		os.Exit(1)
	})
	code := m.Run()
	if timer.Stop() {
		// 正常结束，取消超时计时
	}
	os.Exit(code)
}
