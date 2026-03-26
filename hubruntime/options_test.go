package hubruntime

import (
	"testing"

	coreconfig "github.com/yttydcs/myflowhub-core/config"
)

func TestDefaultOptions_AuthRoleHierarchyDefaults(t *testing.T) {
	opts := DefaultOptions()

	if opts.AuthDefaultRole != "node" {
		t.Fatalf("unexpected auth default role: got %q want %q", opts.AuthDefaultRole, "node")
	}
	if opts.AuthDefaultPerms != "" {
		t.Fatalf("unexpected auth default perms: got %q want empty", opts.AuthDefaultPerms)
	}
	if opts.AuthRolePerms != coreconfig.DefaultAuthRolePerms {
		t.Fatalf("unexpected auth role perms: got %q want %q", opts.AuthRolePerms, coreconfig.DefaultAuthRolePerms)
	}
}
