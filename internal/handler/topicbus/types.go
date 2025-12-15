package topicbus

import "encoding/json"

// 子协议：Topic 订阅/发布。
const SubProtoTopicBus uint8 = 4

// 动作常量定义。
const (
	actionSubscribe          = "subscribe"
	actionSubscribeResp      = "subscribe_resp"
	actionSubscribeBatch     = "subscribe_batch"
	actionSubscribeBatchResp = "subscribe_batch_resp"

	actionUnsubscribe          = "unsubscribe"
	actionUnsubscribeResp      = "unsubscribe_resp"
	actionUnsubscribeBatch     = "unsubscribe_batch"
	actionUnsubscribeBatchResp = "unsubscribe_batch_resp"

	actionListSubs     = "list_subs"
	actionListSubsResp = "list_subs_resp"

	actionPublish = "publish"
)

type message struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

type subscribeReq struct {
	Topic string `json:"topic"`
}

type subscribeBatchReq struct {
	Topics []string `json:"topics"`
}

type publishReq struct {
	Topic   string          `json:"topic"`
	Name    string          `json:"name"`
	TS      int64           `json:"ts"` // unix ms
	Payload json.RawMessage `json:"payload,omitempty"`
}

type resp struct {
	Code   int      `json:"code"`
	Msg    string   `json:"msg,omitempty"`
	Topic  string   `json:"topic,omitempty"`
	Topics []string `json:"topics,omitempty"`
}

type listResp struct {
	Code   int      `json:"code"`
	Msg    string   `json:"msg,omitempty"`
	Topics []string `json:"topics"`
}
