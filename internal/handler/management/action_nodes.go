package management

import (
	"context"
	"encoding/json"

	core "github.com/yttydcs/myflowhub-core"
	"github.com/yttydcs/myflowhub-core/subproto"
)

// list_nodes: 列出直接连接的节点，并标记是否为父节点（视为有子节点）
type listNodesAction struct {
	subproto.BaseAction
	h *ManagementHandler
}

func (a *listNodesAction) Name() string      { return "list_nodes" }
func (a *listNodesAction) RequireAuth() bool { return false }
func (a *listNodesAction) Handle(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage) {
	srv := core.ServerFromContext(ctx)
	if srv == nil {
		return
	}
	nodes := enumerateDirectNodes(srv.ConnManager())
	a.h.sendActionResp(ctx, conn, hdr, "list_nodes_resp", listNodesResp{Code: 1, Msg: "ok", Nodes: nodes})
}

// list_subtree: 返回本节点 + 直接连接的节点（最佳努力子树）
type listSubtreeAction struct {
	subproto.BaseAction
	h *ManagementHandler
}

func (a *listSubtreeAction) Name() string      { return "list_subtree" }
func (a *listSubtreeAction) RequireAuth() bool { return false }
func (a *listSubtreeAction) Handle(ctx context.Context, conn core.IConnection, hdr core.IHeader, _ json.RawMessage) {
	srv := core.ServerFromContext(ctx)
	if srv == nil {
		return
	}
	nodes := enumerateDirectNodes(srv.ConnManager())
	// 包含自身
	nodes = append(nodes, nodeInfo{NodeID: srv.NodeID(), HasChildren: len(nodes) > 0})
	a.h.sendActionResp(ctx, conn, hdr, "list_subtree_resp", listSubtreeResp{Code: 1, Msg: "ok", Nodes: nodes})
}

func enumerateDirectNodes(cm core.IConnectionManager) []nodeInfo {
	if cm == nil {
		return nil
	}
	seen := make(map[uint32]bool)
	nodes := make([]nodeInfo, 0)
	cm.Range(func(c core.IConnection) bool {
		if nidVal, ok := c.GetMeta("nodeID"); ok {
			if nid, ok2 := nidVal.(uint32); ok2 && nid != 0 && !seen[nid] {
				hasChildren := false
				if role, ok := c.GetMeta(core.MetaRoleKey); ok {
					if s, ok2 := role.(string); ok2 && s == core.RoleParent {
						hasChildren = true
					}
				}
				nodes = append(nodes, nodeInfo{NodeID: nid, HasChildren: hasChildren})
				seen[nid] = true
			}
		}
		return true
	})
	return nodes
}
