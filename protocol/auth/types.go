package auth

import "encoding/json"

const (
	SubProtoAuth uint8 = 2

	ActionRegister            = "register"
	ActionAssistRegister      = "assist_register"
	ActionRegisterResp        = "register_resp"
	ActionAssistRegisterResp  = "assist_register_resp"
	ActionLogin               = "login"
	ActionAssistLogin         = "assist_login"
	ActionLoginResp           = "login_resp"
	ActionAssistLoginResp     = "assist_login_resp"
	ActionRevoke              = "revoke"
	ActionRevokeResp          = "revoke_resp"
	ActionAssistQueryCred     = "assist_query_credential"
	ActionAssistQueryCredResp = "assist_query_credential_resp"
	ActionOffline             = "offline"
	ActionAssistOffline       = "assist_offline"
	ActionGetPerms            = "get_perms"
	ActionGetPermsResp        = "get_perms_resp"
	ActionListRoles           = "list_roles"
	ActionListRolesResp       = "list_roles_resp"
	ActionPermsInvalidate     = "perms_invalidate"
	ActionPermsSnapshot       = "perms_snapshot"
	ActionUpLogin             = "up_login"
	ActionUpLoginResp         = "up_login_resp"
)

type Message struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

type RegisterData struct {
	DeviceID string `json:"device_id"`
	NodeID   uint32 `json:"node_id,omitempty"`
	PubKey   string `json:"pubkey,omitempty"`
	NodePub  string `json:"node_pub,omitempty"`
	TS       int64  `json:"ts,omitempty"`
	Nonce    string `json:"nonce,omitempty"`
}

type LoginData struct {
	DeviceID string `json:"device_id"`
	NodeID   uint32 `json:"node_id,omitempty"`
	TS       int64  `json:"ts,omitempty"`
	Nonce    string `json:"nonce,omitempty"`
	Sig      string `json:"sig,omitempty"`
	Alg      string `json:"alg,omitempty"`
}

type RevokeData struct {
	DeviceID string `json:"device_id"`
	NodeID   uint32 `json:"node_id,omitempty"`
}

type QueryCredData struct {
	DeviceID string `json:"device_id"`
	NodeID   uint32 `json:"node_id,omitempty"`
}

type OfflineData struct {
	DeviceID string `json:"device_id"`
	NodeID   uint32 `json:"node_id,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

type RespData struct {
	Code     int      `json:"code"`
	Msg      string   `json:"msg,omitempty"`
	DeviceID string   `json:"device_id,omitempty"`
	NodeID   uint32   `json:"node_id,omitempty"`
	HubID    uint32   `json:"hub_id,omitempty"`
	Role     string   `json:"role,omitempty"`
	Perms    []string `json:"perms,omitempty"`
	PubKey   string   `json:"pubkey,omitempty"`
	NodePub  string   `json:"node_pub,omitempty"`
	TS       int64    `json:"ts,omitempty"`
	Nonce    string   `json:"nonce,omitempty"`
}

type PermsQueryData struct {
	NodeID uint32 `json:"node_id"`
}

type InvalidateData struct {
	NodeIDs []uint32 `json:"node_ids,omitempty"`
	Reason  string   `json:"reason,omitempty"`
	Refresh bool     `json:"refresh,omitempty"`
}

type RolePermEntry struct {
	NodeID uint32   `json:"node_id,omitempty"`
	Role   string   `json:"role,omitempty"`
	Perms  []string `json:"perms,omitempty"`
}

type ListRolesReq struct {
	Offset  int      `json:"offset,omitempty"`
	Limit   int      `json:"limit,omitempty"`
	Role    string   `json:"role,omitempty"`
	NodeIDs []uint32 `json:"node_ids,omitempty"`
}

type UpLoginData struct {
	NodeID      uint32 `json:"node_id"`
	DeviceID    string `json:"device_id,omitempty"`
	HubID       uint32 `json:"hub_id,omitempty"`
	PubKey      string `json:"pubkey,omitempty"`
	TS          int64  `json:"ts,omitempty"`
	Nonce       string `json:"nonce,omitempty"`
	DeviceTS    int64  `json:"device_ts,omitempty"`
	DeviceNonce string `json:"device_nonce,omitempty"`
	DeviceSig   string `json:"device_sig,omitempty"`
	DeviceAlg   string `json:"device_alg,omitempty"`
	SenderID    uint32 `json:"sender_id,omitempty"`
	SenderTS    int64  `json:"sender_ts,omitempty"`
	SenderNonce string `json:"sender_nonce,omitempty"`
	SenderSig   string `json:"sender_sig,omitempty"`
	SenderAlg   string `json:"sender_alg,omitempty"`
	SenderPub   string `json:"sender_pub,omitempty"`
	Sig         string `json:"sig,omitempty"`
	Alg         string `json:"alg,omitempty"`
}
