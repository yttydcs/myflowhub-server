package hubruntime

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	core "github.com/yttydcs/myflowhub-core"
	coreconfig "github.com/yttydcs/myflowhub-core/config"
)

const runtimeConfigFile = "config/runtime_config.json"

type layeredConfig struct {
	mu sync.RWMutex

	path       string
	defaults   map[string]string
	explicit   map[string]string
	persistent map[string]string
	runtime    map[string]string
	effective  map[string]string
}

func buildConfig(opts Options) (core.IConfig, error) {
	return newLayeredConfig(
		runtimeConfigFile,
		configDataFromOptions(DefaultOptions()),
		explicitConfigDataFromOptions(opts),
	)
}

func newLayeredConfig(path string, defaults, explicit map[string]string) (*layeredConfig, error) {
	path = filepath.Clean(strings.TrimSpace(path))
	if path == "" {
		return nil, errors.New("config path is required")
	}
	persistent, err := loadConfigMap(path)
	if err != nil {
		return nil, err
	}
	cfg := &layeredConfig{
		path:       path,
		defaults:   cloneStringMap(defaults),
		explicit:   cloneStringMap(explicit),
		persistent: persistent,
		runtime:    make(map[string]string),
		effective:  make(map[string]string),
	}
	cfg.recomputeLocked()
	return cfg, nil
}

func (c *layeredConfig) Get(key string) (string, bool) {
	if c == nil {
		return "", false
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return "", false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.effective[key]
	return val, ok
}

func (c *layeredConfig) Keys() []string {
	if c == nil {
		return nil
	}
	c.mu.RLock()
	keys := make([]string, 0, len(c.effective))
	for key := range c.effective {
		keys = append(keys, key)
	}
	c.mu.RUnlock()
	sort.Strings(keys)
	return keys
}

// Set keeps current behavior: runtime-only overlay without touching persisted storage.
func (c *layeredConfig) Set(key, val string) {
	if c == nil {
		return
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return
	}
	c.mu.Lock()
	c.runtime[key] = val
	c.recomputeLocked()
	c.mu.Unlock()
}

func (c *layeredConfig) SetPersistent(key, val string) error {
	if c == nil {
		return errors.New("config not initialized")
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return errors.New("key is required")
	}

	c.mu.RLock()
	nextPersistent := cloneStringMap(c.persistent)
	path := c.path
	c.mu.RUnlock()

	nextPersistent[key] = val
	if err := saveConfigMap(path, nextPersistent); err != nil {
		return err
	}

	c.mu.Lock()
	c.persistent = nextPersistent
	c.recomputeLocked()
	c.mu.Unlock()
	return nil
}

func (c *layeredConfig) Merge(other core.IConfig) core.IConfig {
	if c == nil || other == nil {
		return c
	}
	overlay := make(map[string]string)
	for _, key := range other.Keys() {
		if val, ok := other.Get(key); ok {
			overlay[key] = val
		}
	}
	c.mu.Lock()
	mergeStringMap(c.runtime, overlay)
	c.recomputeLocked()
	c.mu.Unlock()
	return c
}

func (c *layeredConfig) recomputeLocked() {
	merged := cloneStringMap(c.defaults)
	mergeStringMap(merged, c.persistent)
	mergeStringMap(merged, c.explicit)
	mergeStringMap(merged, c.runtime)
	c.effective = snapshotConfig(coreconfig.NewMap(merged))
}

func configDataFromOptions(opts Options) map[string]string {
	return map[string]string{
		"addr":                           strings.TrimSpace(opts.Addr),
		coreconfig.KeyParentAddr:         effectiveParentTarget(opts),
		coreconfig.KeyParentEnable:       boolString(opts.ParentEnable),
		coreconfig.KeyParentReconnectSec: strconv.Itoa(opts.ParentReconnectSec),

		coreconfig.KeyProcChannelCount:   strconv.Itoa(opts.ProcChannels),
		coreconfig.KeyProcWorkersPerChan: strconv.Itoa(opts.ProcWorkers),
		coreconfig.KeyProcChannelBuffer:  strconv.Itoa(opts.ProcBuffer),

		coreconfig.KeySendChannelCount:   strconv.Itoa(opts.SendChannels),
		coreconfig.KeySendWorkersPerChan: strconv.Itoa(opts.SendWorkers),
		coreconfig.KeySendChannelBuffer:  strconv.Itoa(opts.SendChannelBuffer),
		coreconfig.KeySendConnBuffer:     strconv.Itoa(opts.SendConnBuffer),

		coreconfig.KeyRoutingForwardRemote: "true",

		coreconfig.KeyAuthDefaultRole:  strings.TrimSpace(opts.AuthDefaultRole),
		coreconfig.KeyAuthDefaultPerms: strings.TrimSpace(opts.AuthDefaultPerms),
		coreconfig.KeyAuthNodeRoles:    strings.TrimSpace(opts.AuthNodeRoles),
		coreconfig.KeyAuthRolePerms:    strings.TrimSpace(opts.AuthRolePerms),
	}
}

func explicitConfigDataFromOptions(opts Options) map[string]string {
	current := configDataFromOptions(opts)
	defaults := configDataFromOptions(DefaultOptions())
	explicit := make(map[string]string)
	overrideKeys := opts.configOverrideKeySet()
	for key, val := range current {
		if _, ok := overrideKeys[key]; ok || defaults[key] != val {
			explicit[key] = val
		}
	}
	return explicit
}

func applyConfigToOptions(opts Options, cfg core.IConfig) Options {
	if cfg == nil {
		return opts
	}
	if val, ok := cfg.Get("addr"); ok {
		opts.Addr = strings.TrimSpace(val)
	}
	if val, ok := cfg.Get(coreconfig.KeyParentAddr); ok {
		target := strings.TrimSpace(val)
		if strings.Contains(target, "://") {
			opts.ParentEndpoint = target
			opts.ParentAddr = ""
		} else {
			opts.ParentEndpoint = ""
			opts.ParentAddr = target
		}
	}
	if val, ok := cfg.Get(coreconfig.KeyParentEnable); ok {
		opts.ParentEnable = parseBoolValue(val, opts.ParentEnable)
	}
	if val, ok := cfg.Get(coreconfig.KeyParentReconnectSec); ok {
		opts.ParentReconnectSec = parseIntValue(val, opts.ParentReconnectSec)
		if opts.ParentReconnectSec < 0 {
			opts.ParentReconnectSec = 0
		}
	}
	if val, ok := cfg.Get(coreconfig.KeyProcChannelCount); ok {
		opts.ProcChannels = parseIntValue(val, opts.ProcChannels)
	}
	if val, ok := cfg.Get(coreconfig.KeyProcWorkersPerChan); ok {
		opts.ProcWorkers = parseIntValue(val, opts.ProcWorkers)
	}
	if val, ok := cfg.Get(coreconfig.KeyProcChannelBuffer); ok {
		opts.ProcBuffer = parseIntValue(val, opts.ProcBuffer)
	}
	if val, ok := cfg.Get(coreconfig.KeySendChannelCount); ok {
		opts.SendChannels = parseIntValue(val, opts.SendChannels)
	}
	if val, ok := cfg.Get(coreconfig.KeySendWorkersPerChan); ok {
		opts.SendWorkers = parseIntValue(val, opts.SendWorkers)
	}
	if val, ok := cfg.Get(coreconfig.KeySendChannelBuffer); ok {
		opts.SendChannelBuffer = parseIntValue(val, opts.SendChannelBuffer)
	}
	if val, ok := cfg.Get(coreconfig.KeySendConnBuffer); ok {
		opts.SendConnBuffer = parseIntValue(val, opts.SendConnBuffer)
	}
	if val, ok := cfg.Get(coreconfig.KeyAuthDefaultRole); ok {
		opts.AuthDefaultRole = strings.TrimSpace(val)
	}
	if val, ok := cfg.Get(coreconfig.KeyAuthDefaultPerms); ok {
		opts.AuthDefaultPerms = strings.TrimSpace(val)
	}
	if val, ok := cfg.Get(coreconfig.KeyAuthNodeRoles); ok {
		opts.AuthNodeRoles = strings.TrimSpace(val)
	}
	if val, ok := cfg.Get(coreconfig.KeyAuthRolePerms); ok {
		opts.AuthRolePerms = strings.TrimSpace(val)
	}
	if effectiveParentTarget(opts) != "" {
		opts.ParentEnable = true
	}
	return opts
}

func parseBoolValue(raw string, def bool) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return def
	}
}

func parseIntValue(raw string, def int) int {
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return def
	}
	return n
}

func loadConfigMap(path string) (map[string]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return make(map[string]string), nil
		}
		return nil, err
	}
	if len(strings.TrimSpace(string(raw))) == 0 {
		return make(map[string]string), nil
	}
	var data map[string]string
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	if data == nil {
		data = make(map[string]string)
	}
	return data, nil
}

func saveConfigMap(path string, data map[string]string) error {
	dir := filepath.Dir(path)
	if dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(path, raw, 0o600)
}

func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err == nil {
		return nil
	}
	_ = os.Remove(path)
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func snapshotConfig(cfg core.IConfig) map[string]string {
	out := make(map[string]string)
	if cfg == nil {
		return out
	}
	for _, key := range cfg.Keys() {
		if val, ok := cfg.Get(key); ok {
			out[key] = val
		}
	}
	return out
}

func cloneStringMap(src map[string]string) map[string]string {
	out := make(map[string]string, len(src))
	for key, val := range src {
		out[key] = val
	}
	return out
}

func mergeStringMap(dst, src map[string]string) {
	for key, val := range src {
		dst[key] = val
	}
}
