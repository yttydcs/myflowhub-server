package varstore

import "encoding/json"

const (
	varActionSet           = "set"
	varActionAssistSet     = "assist_set"
	varActionSetResp       = "set_resp"
	varActionAssistSetResp = "assist_set_resp"
	varActionUpSet         = "up_set"
	varActionNotifySet     = "notify_set"

	varActionGet           = "get"
	varActionAssistGet     = "assist_get"
	varActionGetResp       = "get_resp"
	varActionAssistGetResp = "assist_get_resp"

	varActionList           = "list"
	varActionAssistList     = "assist_list"
	varActionListResp       = "list_resp"
	varActionAssistListResp = "assist_list_resp"

	varActionRevoke           = "revoke"
	varActionAssistRevoke     = "assist_revoke"
	varActionRevokeResp       = "revoke_resp"
	varActionAssistRevokeResp = "assist_revoke_resp"
	varActionUpRevoke         = "up_revoke"
	varActionNotifyRevoke     = "notify_revoke"

	visibilityPublic  = "public"
	visibilityPrivate = "private"
)

type varMessage struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

type setReq struct {
	Name       string `json:"name"`
	Value      string `json:"value"`
	Visibility string `json:"visibility"`
	Type       string `json:"type,omitempty"`
	Owner      uint32 `json:"owner,omitempty"`
}

type getReq struct {
	Name  string `json:"name"`
	Owner uint32 `json:"owner,omitempty"`
}

type listReq struct {
	Owner uint32 `json:"owner,omitempty"`
}

type varResp struct {
	Code       int      `json:"code"`
	Msg        string   `json:"msg,omitempty"`
	Name       string   `json:"name,omitempty"`
	Value      string   `json:"value,omitempty"`
	Owner      uint32   `json:"owner,omitempty"`
	Visibility string   `json:"visibility,omitempty"`
	Type       string   `json:"type,omitempty"`
	Names      []string `json:"names,omitempty"`
}

type varRecord struct {
	Value      string
	Owner      uint32
	IsPublic   bool
	Visibility string
	Type       string
}

type pendingKey struct {
	owner uint32
	name  string
	kind  string
}

const (
	pendingKindGet    = "get"
	pendingKindList   = "list"
	pendingKindSet    = "set"
	pendingKindRevoke = "revoke"
)
