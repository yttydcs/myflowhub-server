package varstore

import "encoding/json"

const (
	SubProtoVarStore uint8 = 3

	ActionSet           = "set"
	ActionAssistSet     = "assist_set"
	ActionSetResp       = "set_resp"
	ActionAssistSetResp = "assist_set_resp"
	ActionUpSet         = "up_set"
	ActionNotifySet     = "notify_set"

	ActionGet           = "get"
	ActionAssistGet     = "assist_get"
	ActionGetResp       = "get_resp"
	ActionAssistGetResp = "assist_get_resp"

	ActionList           = "list"
	ActionAssistList     = "assist_list"
	ActionListResp       = "list_resp"
	ActionAssistListResp = "assist_list_resp"

	ActionRevoke           = "revoke"
	ActionAssistRevoke     = "assist_revoke"
	ActionRevokeResp       = "revoke_resp"
	ActionAssistRevokeResp = "assist_revoke_resp"
	ActionUpRevoke         = "up_revoke"
	ActionNotifyRevoke     = "notify_revoke"

	ActionSubscribe           = "subscribe"
	ActionAssistSubscribe     = "assist_subscribe"
	ActionSubscribeResp       = "subscribe_resp"
	ActionAssistSubscribeResp = "assist_subscribe_resp"
	ActionUnsubscribe         = "unsubscribe"
	ActionAssistUnsubscribe   = "assist_unsubscribe"

	ActionVarChanged = "var_changed"
	ActionVarDeleted = "var_deleted"

	VisibilityPublic  = "public"
	VisibilityPrivate = "private"
)

type Message struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

type SetReq struct {
	Name       string `json:"name"`
	Value      string `json:"value"`
	Visibility string `json:"visibility"`
	Type       string `json:"type,omitempty"`
	Owner      uint32 `json:"owner,omitempty"`
}

type GetReq struct {
	Name  string `json:"name"`
	Owner uint32 `json:"owner,omitempty"`
}

type ListReq struct {
	Owner uint32 `json:"owner,omitempty"`
}

type SubscribeReq struct {
	Name       string `json:"name"`
	Owner      uint32 `json:"owner"`
	Subscriber uint32 `json:"subscriber,omitempty"`
}

type VarResp struct {
	Code       int      `json:"code"`
	Msg        string   `json:"msg,omitempty"`
	Name       string   `json:"name,omitempty"`
	Value      string   `json:"value,omitempty"`
	Owner      uint32   `json:"owner,omitempty"`
	Visibility string   `json:"visibility,omitempty"`
	Type       string   `json:"type,omitempty"`
	Names      []string `json:"names,omitempty"`
}
