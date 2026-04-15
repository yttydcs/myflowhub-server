package hubruntime

// 本文件承载 `hubruntime` 中与 `quic_dev_cert` 相关的逻辑。

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	quicDevCertFileName = "quic-dev-cert.pem"
	quicDevKeyFileName  = "quic-dev-key.pem"
)

// ensureQUICDevCertIfNeeded 在开发模式下自动补齐 QUIC 所需的自签名证书。
func ensureQUICDevCertIfNeeded(opts *Options, log *slog.Logger) error {
	if opts == nil || !opts.QUICEnable || !opts.QUICDevCertAuto {
		return nil
	}

	certFile := strings.TrimSpace(opts.QUICCertFile)
	keyFile := strings.TrimSpace(opts.QUICKeyFile)
	certSet := certFile != ""
	keySet := keyFile != ""
	if certSet != keySet {
		return errors.New("quic cert_file and key_file must both be set or both empty when quic-dev-cert-auto is enabled")
	}
	if certSet {
		return nil
	}

	certPEM, keyPEM, err := generateSelfSignedQUICDevCert()
	if err != nil {
		return err
	}

	dir := strings.TrimSpace(opts.WorkDir)
	if dir == "" {
		dir = filepath.Join(os.TempDir(), "myflowhub")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir quic dev cert dir failed: %w", err)
	}

	certPath := filepath.Join(dir, quicDevCertFileName)
	keyPath := filepath.Join(dir, quicDevKeyFileName)
	if err := os.WriteFile(certPath, certPEM, 0o644); err != nil {
		return fmt.Errorf("write quic dev cert failed: %w", err)
	}
	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		return fmt.Errorf("write quic dev key failed: %w", err)
	}

	opts.QUICCertFile = certPath
	opts.QUICKeyFile = keyPath
	if log != nil {
		log.Warn("quic dev certificate auto-generated (development only)", "cert_file", certPath, "key_file", keyPath)
	}
	return nil
}

// generateSelfSignedQUICDevCert 生成一次性的本地开发证书和私钥 PEM。
func generateSelfSignedQUICDevCert() ([]byte, []byte, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate quic dev private key failed: %w", err)
	}
	serialLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("generate quic dev serial failed: %w", err)
	}

	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "MyFlowHub QUIC Dev Certificate",
		},
		NotBefore:             now.Add(-5 * time.Minute),
		NotAfter:              now.Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, fmt.Errorf("create quic dev certificate failed: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	if len(certPEM) == 0 {
		return nil, nil, errors.New("encode quic dev certificate pem failed")
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal quic dev private key failed: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})
	if len(keyPEM) == 0 {
		return nil, nil, errors.New("encode quic dev private key pem failed")
	}
	return certPEM, keyPEM, nil
}
