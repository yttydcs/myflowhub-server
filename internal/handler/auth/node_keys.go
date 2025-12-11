package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	core "github.com/yttydcs/myflowhub-core"
	coreconfig "github.com/yttydcs/myflowhub-core/config"
)

const (
	nodeKeysFile        = "config/node_keys.json"
	trustedNodesFile    = "config/trusted_nodes.json"
	confNodePrivKey     = coreconfig.KeyAuthNodePrivKey
	confNodePubKey      = coreconfig.KeyAuthNodePubKey
	confTrustedNodesKey = coreconfig.KeyAuthTrustedNodes
)

type nodeKeys struct {
	PrivKey string `json:"privkey"` // base64 DER
	PubKey  string `json:"pubkey"`  // base64 DER
}

// loadOrCreateNodeKeys 加载节点密钥，若不存在则生成并写入文件与配置。
func loadOrCreateNodeKeys(cfg core.IConfig) (*ecdsa.PrivateKey, string, error) {
	privStr, _ := cfg.Get(confNodePrivKey)
	pubStr, _ := cfg.Get(confNodePubKey)
	if strings.TrimSpace(privStr) != "" && strings.TrimSpace(pubStr) != "" {
		priv, err := parsePrivKey(privStr)
		if err == nil {
			return priv, pubStr, nil
		}
	}
	// 尝试从文件加载
	if k, err := readNodeKeysFile(); err == nil && k.PrivKey != "" && k.PubKey != "" {
		if priv, err := parsePrivKey(k.PrivKey); err == nil {
			cfg.Set(confNodePrivKey, k.PrivKey)
			cfg.Set(confNodePubKey, k.PubKey)
			return priv, k.PubKey, nil
		}
	}
	// 生成新的
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, "", err
	}
	privDER, _ := x509.MarshalECPrivateKey(priv)
	pubDER, _ := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	privB64 := base64.StdEncoding.EncodeToString(privDER)
	pubB64 := base64.StdEncoding.EncodeToString(pubDER)
	_ = writeNodeKeysFile(nodeKeys{PrivKey: privB64, PubKey: pubB64})
	cfg.Set(confNodePrivKey, privB64)
	cfg.Set(confNodePubKey, pubB64)
	return priv, pubB64, nil
}

func parsePrivKey(b64 string) (*ecdsa.PrivateKey, error) {
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(b64))
	if err != nil {
		return nil, err
	}
	priv, err := x509.ParseECPrivateKey(raw)
	if err != nil {
		return nil, err
	}
	if priv == nil || priv.Curve != elliptic.P256() {
		return nil, errors.New("not p256")
	}
	return priv, nil
}

func readNodeKeysFile() (nodeKeys, error) {
	var k nodeKeys
	path := filepath.Clean(nodeKeysFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return k, err
	}
	err = json.Unmarshal(data, &k)
	return k, err
}

func writeNodeKeysFile(k nodeKeys) error {
	path := filepath.Clean(nodeKeysFile)
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	data, _ := json.MarshalIndent(k, "", "  ")
	return os.WriteFile(path, data, 0o600)
}

// loadTrustedNodes 读取 config/trusted_nodes.json 并写入 cfg。
func loadTrustedNodes(cfg core.IConfig) map[uint32][]byte {
	path := filepath.Clean(trustedNodesFile)
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return nil
	}
	tmp := make(map[string]string)
	_ = json.Unmarshal(data, &tmp)
	out := make(map[uint32][]byte)
	for k, v := range tmp {
		if id, err := parseUint32(k); err == nil {
			if pk, err := base64.StdEncoding.DecodeString(strings.TrimSpace(v)); err == nil {
				out[id] = pk
			}
		}
	}
	// flatten to cfg as json string for sharing
	if len(out) > 0 {
		strMap := make(map[uint32]string)
		for id, raw := range out {
			strMap[id] = base64.StdEncoding.EncodeToString(raw)
		}
		buf, _ := json.Marshal(strMap)
		cfg.Set(confTrustedNodesKey, string(buf))
	}
	return out
}

func saveTrustedNodesFile(m map[uint32][]byte) {
	path := filepath.Clean(trustedNodesFile)
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	tmp := make(map[string]string)
	for id, raw := range m {
		if id == 0 || len(raw) == 0 {
			continue
		}
		tmp[strconv.FormatUint(uint64(id), 10)] = base64.StdEncoding.EncodeToString(raw)
	}
	data, _ := json.MarshalIndent(tmp, "", "  ")
	_ = os.WriteFile(path, data, 0o600)
}
