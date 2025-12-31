package exec

import (
	"sync"
)

// Broker 用于在同一进程内将 call_resp 投递给等待方（例如 flow 调度器）。
// 注意：这不是网络 pending；响应仍通过 core 路由回到执行者节点，只是 handler 内部需要把 resp 交给等待逻辑。
type Broker struct {
	mu      sync.Mutex
	waiters map[string]chan CallResp
}

var (
	brokerOnce sync.Once
	brokerInst *Broker
)

func SharedBroker() *Broker {
	brokerOnce.Do(func() {
		brokerInst = &Broker{waiters: make(map[string]chan CallResp)}
	})
	return brokerInst
}

func (b *Broker) Register(reqID string) (ch <-chan CallResp, cancel func()) {
	reqID = stringsTrim(reqID)
	out := make(chan CallResp, 1)
	if reqID == "" {
		close(out)
		return out, func() {}
	}
	b.mu.Lock()
	b.waiters[reqID] = out
	b.mu.Unlock()
	return out, func() {
		b.mu.Lock()
		if c, ok := b.waiters[reqID]; ok {
			delete(b.waiters, reqID)
			close(c)
		}
		b.mu.Unlock()
	}
}

func (b *Broker) Deliver(resp CallResp) bool {
	reqID := stringsTrim(resp.ReqID)
	if reqID == "" {
		return false
	}
	b.mu.Lock()
	ch, ok := b.waiters[reqID]
	if ok {
		delete(b.waiters, reqID)
	}
	b.mu.Unlock()
	if !ok {
		return false
	}
	ch <- resp
	close(ch)
	return true
}

func stringsTrim(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\n' || s[0] == '\r' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 {
		last := s[len(s)-1]
		if last == ' ' || last == '\n' || last == '\r' || last == '\t' {
			s = s[:len(s)-1]
			continue
		}
		break
	}
	return s
}
