package auth

import core "github.com/yttydcs/myflowhub-core"

func registerUpLoginActions(h *LoginHandler) []core.SubProcessAction {
	return []core.SubProcessAction{
		&upLoginAction{h: h},
		&upLoginRespAction{h: h},
	}
}
