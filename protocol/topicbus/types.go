package topicbus

import "encoding/json"

const SubProtoTopicBus uint8 = 4

const (
	ActionSubscribe          = "subscribe"
	ActionSubscribeResp      = "subscribe_resp"
	ActionSubscribeBatch     = "subscribe_batch"
	ActionSubscribeBatchResp = "subscribe_batch_resp"

	ActionUnsubscribe          = "unsubscribe"
	ActionUnsubscribeResp      = "unsubscribe_resp"
	ActionUnsubscribeBatch     = "unsubscribe_batch"
	ActionUnsubscribeBatchResp = "unsubscribe_batch_resp"

	ActionListSubs     = "list_subs"
	ActionListSubsResp = "list_subs_resp"

	ActionPublish = "publish"
)

type Message struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

type SubscribeReq struct {
	Topic string `json:"topic"`
}

type SubscribeBatchReq struct {
	Topics []string `json:"topics"`
}

type PublishReq struct {
	Topic   string          `json:"topic"`
	Name    string          `json:"name"`
	TS      int64           `json:"ts"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type Resp struct {
	Code   int      `json:"code"`
	Msg    string   `json:"msg,omitempty"`
	Topic  string   `json:"topic,omitempty"`
	Topics []string `json:"topics,omitempty"`
}

type ListResp struct {
	Code   int      `json:"code"`
	Msg    string   `json:"msg,omitempty"`
	Topics []string `json:"topics"`
}
