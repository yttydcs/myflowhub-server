package hubruntime

import (
	"strings"
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
	if opts.AuthRolePerms != defaultAuthRolePerms {
		t.Fatalf("unexpected auth role perms: got %q want %q", opts.AuthRolePerms, defaultAuthRolePerms)
	}

	cfg := configDataFromOptions(opts)
	if cfg[coreconfig.KeyAuthRolePerms] != defaultAuthRolePerms {
		t.Fatalf("unexpected transport auth role perms: got %q want %q", cfg[coreconfig.KeyAuthRolePerms], defaultAuthRolePerms)
	}
	for _, perm := range []string{"flow.run", "flow.read"} {
		if !strings.Contains(opts.AuthRolePerms, perm) {
			t.Fatalf("auth role perms should include %q: %q", perm, opts.AuthRolePerms)
		}
	}
}
