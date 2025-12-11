package auth

import "encoding/json"

// 动作常量定义
const (
	actionRegister            = "register"
	actionAssistRegister      = "assist_register"
	actionRegisterResp        = "register_resp"
	actionAssistRegisterResp  = "assist_register_resp"
	actionLogin               = "login"
	actionAssistLogin         = "assist_login"
	actionLoginResp           = "login_resp"
	actionAssistLoginResp     = "assist_login_resp"
	actionRevoke              = "revoke"
	actionRevokeResp          = "revoke_resp"
	actionAssistQueryCred     = "assist_query_credential"
	actionAssistQueryCredResp = "assist_query_credential_resp"
	actionOffline             = "offline"
	actionAssistOffline       = "assist_offline"
	actionGetPerms            = "get_perms"
	actionGetPermsResp        = "get_perms_resp"
	actionListRoles           = "list_roles"
	actionListRolesResp       = "list_roles_resp"
	actionPermsInvalidate     = "perms_invalidate"
	actionPermsSnapshot       = "perms_snapshot"
	actionUpLogin             = "up_login"
	actionUpLoginResp         = "up_login_resp"
)

type message struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

type registerData struct {
	DeviceID string `json:"device_id"`
	NodeID   uint32 `json:"node_id,omitempty"`
	PubKey   string `json:"pubkey,omitempty"`
	NodePub  string `json:"node_pub,omitempty"`
	TS       int64  `json:"ts,omitempty"`
	Nonce    string `json:"nonce,omitempty"`
}

type loginData struct {
	DeviceID string `json:"device_id"`
	NodeID   uint32 `json:"node_id,omitempty"`
	TS       int64  `json:"ts,omitempty"`
	Nonce    string `json:"nonce,omitempty"`
	Sig      string `json:"sig,omitempty"`
	Alg      string `json:"alg,omitempty"`
}

type revokeData struct {
	DeviceID string `json:"device_id"`
	NodeID   uint32 `json:"node_id,omitempty"`
}

type queryCredData struct {
	DeviceID string `json:"device_id"`
	NodeID   uint32 `json:"node_id,omitempty"`
}

type offlineData struct {
	DeviceID string `json:"device_id"`
	NodeID   uint32 `json:"node_id,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

type respData struct {
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

type bindingRecord struct {
	NodeID uint32
	Role   string
	Perms  []string
	PubKey []byte
}

type permsQueryData struct {
	NodeID uint32 `json:"node_id"`
}

type invalidateData struct {
	NodeIDs []uint32 `json:"node_ids,omitempty"`
	Reason  string   `json:"reason,omitempty"`
	Refresh bool     `json:"refresh,omitempty"`
}

type rolePermEntry struct {
	NodeID uint32   `json:"node_id,omitempty"`
	Role   string   `json:"role,omitempty"`
	Perms  []string `json:"perms,omitempty"`
}

type listRolesReq struct {
	Offset  int      `json:"offset,omitempty"`
	Limit   int      `json:"limit,omitempty"`
	Role    string   `json:"role,omitempty"`
	NodeIDs []uint32 `json:"node_ids,omitempty"`
}

type upLoginData struct {
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
