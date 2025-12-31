package exec

import "testing"

func TestBrokerDeliver(t *testing.T) {
	b := SharedBroker()
	ch, cancel := b.Register("req1")
	defer cancel()
	ok := b.Deliver(CallResp{ReqID: "req1", Code: 1, Msg: "ok"})
	if !ok {
		t.Fatalf("expected deliver ok")
	}
	resp, ok := <-ch
	if !ok || resp.Code != 1 {
		t.Fatalf("unexpected resp: ok=%v resp=%#v", ok, resp)
	}
}
