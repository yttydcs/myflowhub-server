package topicbus

import (
	"context"
	"encoding/json"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-core/subproto"
)

type topicAction struct {
	subproto.BaseAction
	name string
	fn   func(context.Context, core.IConnection, core.IHeader, json.RawMessage)
}

func (a topicAction) Name() string { return a.name }

func (a topicAction) Handle(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage) {
	a.fn(ctx, conn, hdr, data)
}

func registerActions(h *TopicBusHandler) []core.SubProcessAction {
	return []core.SubProcessAction{
		topicAction{name: actionSubscribe, fn: h.handleSubscribe},
		topicAction{name: actionSubscribeBatch, fn: h.handleSubscribeBatch},
		topicAction{name: actionUnsubscribe, fn: h.handleUnsubscribe},
		topicAction{name: actionUnsubscribeBatch, fn: h.handleUnsubscribeBatch},
		topicAction{name: actionListSubs, fn: h.handleListSubs},
		topicAction{name: actionPublish, fn: h.handlePublish},
	}
}
