package auth

import protocol "github.com/yttydcs/myflowhub-proto/protocol/auth"

// 本包为兼容壳：保留原 import path `github.com/yttydcs/myflowhub-server/protocol/auth`，
// 其类型/常量全部委托到独立协议仓库 MyFlowHub-Proto（wire 不变）。

const (
	SubProtoAuth uint8 = protocol.SubProtoAuth

	ActionRegister                 = protocol.ActionRegister
	ActionAssistRegister           = protocol.ActionAssistRegister
	ActionRegisterResp             = protocol.ActionRegisterResp
	ActionAssistRegisterResp       = protocol.ActionAssistRegisterResp
	ActionLogin                    = protocol.ActionLogin
	ActionAssistLogin              = protocol.ActionAssistLogin
	ActionLoginResp                = protocol.ActionLoginResp
	ActionAssistLoginResp          = protocol.ActionAssistLoginResp
	ActionRevoke                   = protocol.ActionRevoke
	ActionRevokeResp               = protocol.ActionRevokeResp
	ActionAssistQueryCred          = protocol.ActionAssistQueryCred
	ActionAssistQueryCredResp      = protocol.ActionAssistQueryCredResp
	ActionOffline                  = protocol.ActionOffline
	ActionAssistOffline            = protocol.ActionAssistOffline
	ActionGetPerms                 = protocol.ActionGetPerms
	ActionGetPermsResp             = protocol.ActionGetPermsResp
	ActionAuthorityPolicySync      = protocol.ActionAuthorityPolicySync
	ActionListPendingRegisters     = protocol.ActionListPendingRegisters
	ActionListPendingRegistersResp = protocol.ActionListPendingRegistersResp
	ActionListRegisterPermits      = protocol.ActionListRegisterPermits
	ActionListRegisterPermitsResp  = protocol.ActionListRegisterPermitsResp
	ActionApproveRegister          = protocol.ActionApproveRegister
	ActionApproveRegisterResp      = protocol.ActionApproveRegisterResp
	ActionRejectRegister           = protocol.ActionRejectRegister
	ActionRejectRegisterResp       = protocol.ActionRejectRegisterResp
	ActionIssueRegisterPermit      = protocol.ActionIssueRegisterPermit
	ActionIssueRegisterPermitResp  = protocol.ActionIssueRegisterPermitResp
	ActionRevokeRegisterPermit     = protocol.ActionRevokeRegisterPermit
	ActionRevokeRegisterPermitResp = protocol.ActionRevokeRegisterPermitResp
	ActionListRoles                = protocol.ActionListRoles
	ActionListRolesResp            = protocol.ActionListRolesResp
	ActionPermsInvalidate          = protocol.ActionPermsInvalidate
	ActionPermsSnapshot            = protocol.ActionPermsSnapshot
	ActionUpLogin                  = protocol.ActionUpLogin
	ActionUpLoginResp              = protocol.ActionUpLoginResp
)

type Message = protocol.Message

type RegisterData = protocol.RegisterData
type LoginData = protocol.LoginData
type RevokeData = protocol.RevokeData
type QueryCredData = protocol.QueryCredData
type OfflineData = protocol.OfflineData
type RespData = protocol.RespData
type PermsQueryData = protocol.PermsQueryData
type InvalidateData = protocol.InvalidateData
type AuthorityPolicySyncData = protocol.AuthorityPolicySyncData
type RolePermEntry = protocol.RolePermEntry
type ListRolesReq = protocol.ListRolesReq
type PendingRegisterInfo = protocol.PendingRegisterInfo
type ListPendingRegistersReq = protocol.ListPendingRegistersReq
type ListPendingRegistersResp = protocol.ListPendingRegistersResp
type RegisterPermitInfo = protocol.RegisterPermitInfo
type ListRegisterPermitsReq = protocol.ListRegisterPermitsReq
type ListRegisterPermitsResp = protocol.ListRegisterPermitsResp
type ApproveRegisterReq = protocol.ApproveRegisterReq
type ApproveRegisterResp = protocol.ApproveRegisterResp
type RejectRegisterReq = protocol.RejectRegisterReq
type RejectRegisterResp = protocol.RejectRegisterResp
type IssueRegisterPermitReq = protocol.IssueRegisterPermitReq
type IssueRegisterPermitResp = protocol.IssueRegisterPermitResp
type RevokeRegisterPermitReq = protocol.RevokeRegisterPermitReq
type RevokeRegisterPermitResp = protocol.RevokeRegisterPermitResp
type UpLoginData = protocol.UpLoginData
