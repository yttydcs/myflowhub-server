package hubruntime

// Context: This file lives in the Server assembly layer and supports layered_config_test.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	coreconfig "github.com/yttydcs/myflowhub-core/config"
)

func TestLayeredConfigPrecedence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime_config.json")
	if err := saveConfigMap(path, map[string]string{
		coreconfig.KeyAuthDefaultRole:  "persist-role",
		coreconfig.KeyAuthDefaultPerms: "persist-perms",
		coreconfig.KeyProcChannelCount: "9",
		coreconfig.KeyParentAddr:       "tcp://persist-parent:9000",
		"node.display_name":            "Persisted Name",
	}); err != nil {
		t.Fatalf("seed persistent config: %v", err)
	}

	opts := DefaultOptions()
	opts.AuthDefaultRole = "flag-role"
	opts.ProcChannels = 6
	opts.ParentReconnectSec = 0
	opts.AddConfigOverrideKeys(coreconfig.KeyAuthDefaultPerms, coreconfig.KeyParentReconnectSec)
	opts.Normalize()

	cfg, err := newLayeredConfig(path, configDataFromOptions(DefaultOptions()), explicitConfigDataFromOptions(opts))
	if err != nil {
		t.Fatalf("newLayeredConfig: %v", err)
	}

	assertConfigValue(t, cfg, coreconfig.KeyAuthDefaultRole, "flag-role")
	assertConfigValue(t, cfg, coreconfig.KeyAuthDefaultPerms, "")
	assertConfigValue(t, cfg, coreconfig.KeyProcChannelCount, "6")
	assertConfigValue(t, cfg, coreconfig.KeyParentAddr, "tcp://persist-parent:9000")
	assertConfigValue(t, cfg, coreconfig.KeyParentReconnectSec, "0")
	assertConfigValue(t, cfg, "node.display_name", "Persisted Name")
}

func TestLayeredConfigRuntimeOnlySet(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime_config.json")

	cfg, err := newLayeredConfig(path, configDataFromOptions(DefaultOptions()), nil)
	if err != nil {
		t.Fatalf("newLayeredConfig: %v", err)
	}
	if err := cfg.SetPersistent("node.display_name", "Persisted Hub"); err != nil {
		t.Fatalf("SetPersistent: %v", err)
	}
	cfg.Set("node.display_name", "Runtime Hub")

	assertConfigValue(t, cfg, "node.display_name", "Runtime Hub")
	assertStoredValue(t, path, "node.display_name", "Persisted Hub")

	reloaded, err := newLayeredConfig(path, configDataFromOptions(DefaultOptions()), nil)
	if err != nil {
		t.Fatalf("reload layered config: %v", err)
	}
	assertConfigValue(t, reloaded, "node.display_name", "Persisted Hub")
}

func TestLayeredConfigUsesServerDefaultAuthRolePerms(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime_config.json")

	cfg, err := newLayeredConfig(path, configDataFromOptions(DefaultOptions()), nil)
	if err != nil {
		t.Fatalf("newLayeredConfig: %v", err)
	}

	assertConfigValue(t, cfg, coreconfig.KeyAuthRolePerms, defaultAuthRolePerms)
}

func TestLayeredConfigAllowsExplicitEmptyAuthRolePermsOverride(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime_config.json")

	opts := DefaultOptions()
	opts.AuthRolePerms = ""
	opts.AddConfigOverrideKeys(coreconfig.KeyAuthRolePerms)
	opts.Normalize()

	cfg, err := newLayeredConfig(path, configDataFromOptions(DefaultOptions()), explicitConfigDataFromOptions(opts))
	if err != nil {
		t.Fatalf("newLayeredConfig: %v", err)
	}

	assertConfigValue(t, cfg, coreconfig.KeyAuthRolePerms, "")
}

func TestApplyConfigToOptions(t *testing.T) {
	opts := DefaultOptions()
	opts.Addr = ":1234"
	opts.ParentEnable = false
	opts.ParentReconnectSec = 9

	cfg := coreconfig.NewMap(map[string]string{
		"addr":                           "127.0.0.1:9001",
		coreconfig.KeyParentAddr:         "tcp://127.0.0.1:9100",
		coreconfig.KeyParentEnable:       "true",
		coreconfig.KeyParentJoinPermit:   "permit-parent-1",
		coreconfig.KeyParentReconnectSec: "0",
	})

	applied := applyConfigToOptions(opts, cfg)
	if applied.Addr != "127.0.0.1:9001" {
		t.Fatalf("addr not applied: got %q", applied.Addr)
	}
	if applied.ParentEndpoint != "tcp://127.0.0.1:9100" {
		t.Fatalf("parent endpoint not applied: got %q", applied.ParentEndpoint)
	}
	if applied.ParentAddr != "" {
		t.Fatalf("legacy parent addr should be cleared, got %q", applied.ParentAddr)
	}
	if !applied.ParentEnable {
		t.Fatalf("parent enable not applied")
	}
	if applied.ParentJoinPermit != "permit-parent-1" {
		t.Fatalf("parent join permit not applied: got %q", applied.ParentJoinPermit)
	}
	if applied.ParentReconnectSec != 0 {
		t.Fatalf("parent reconnect not applied: got %d", applied.ParentReconnectSec)
	}
}

func assertConfigValue(t *testing.T, cfg *layeredConfig, key, want string) {
	t.Helper()
	got, ok := cfg.Get(key)
	if !ok {
		t.Fatalf("missing key %q", key)
	}
	if got != want {
		t.Fatalf("unexpected value for %q: got %q want %q", key, got, want)
	}
}

func assertStoredValue(t *testing.T, path, key, want string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read stored config: %v", err)
	}
	var data map[string]string
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("unmarshal stored config: %v", err)
	}
	if got := data[key]; got != want {
		t.Fatalf("unexpected stored value for %q: got %q want %q", key, got, want)
	}
}
