package hubruntime

// Context: This file lives in the Server assembly layer and supports quic_dev_cert_test.

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureQUICDevCertIfNeeded_Disabled(t *testing.T) {
	opts := Options{
		QUICEnable:      true,
		QUICDevCertAuto: false,
	}
	if err := ensureQUICDevCertIfNeeded(&opts, nil); err != nil {
		t.Fatalf("ensureQUICDevCertIfNeeded disabled error: %v", err)
	}
	if opts.QUICCertFile != "" || opts.QUICKeyFile != "" {
		t.Fatalf("expected cert/key remain empty, got cert=%q key=%q", opts.QUICCertFile, opts.QUICKeyFile)
	}
}

func TestEnsureQUICDevCertIfNeeded_PreserveProvided(t *testing.T) {
	opts := Options{
		QUICEnable:      true,
		QUICDevCertAuto: true,
		QUICCertFile:    "cert.pem",
		QUICKeyFile:     "key.pem",
	}
	if err := ensureQUICDevCertIfNeeded(&opts, nil); err != nil {
		t.Fatalf("ensureQUICDevCertIfNeeded preserve error: %v", err)
	}
	if opts.QUICCertFile != "cert.pem" || opts.QUICKeyFile != "key.pem" {
		t.Fatalf("expected preserve cert/key, got cert=%q key=%q", opts.QUICCertFile, opts.QUICKeyFile)
	}
}

func TestEnsureQUICDevCertIfNeeded_PartialProvided(t *testing.T) {
	opts := Options{
		QUICEnable:      true,
		QUICDevCertAuto: true,
		QUICCertFile:    "cert-only.pem",
	}
	err := ensureQUICDevCertIfNeeded(&opts, nil)
	if err == nil {
		t.Fatalf("expected error when only one of cert/key is provided")
	}
	if !strings.Contains(err.Error(), "both be set or both empty") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEnsureQUICDevCertIfNeeded_Generate(t *testing.T) {
	workDir := t.TempDir()
	opts := Options{
		QUICEnable:      true,
		QUICDevCertAuto: true,
		WorkDir:         workDir,
	}
	if err := ensureQUICDevCertIfNeeded(&opts, nil); err != nil {
		t.Fatalf("ensureQUICDevCertIfNeeded generate error: %v", err)
	}
	if opts.QUICCertFile == "" || opts.QUICKeyFile == "" {
		t.Fatalf("expected generated cert/key path, got cert=%q key=%q", opts.QUICCertFile, opts.QUICKeyFile)
	}
	if !strings.HasPrefix(opts.QUICCertFile, workDir) || !strings.HasPrefix(opts.QUICKeyFile, workDir) {
		t.Fatalf("expected generated paths under workdir, got cert=%q key=%q", opts.QUICCertFile, opts.QUICKeyFile)
	}
	if _, err := os.Stat(opts.QUICCertFile); err != nil {
		t.Fatalf("cert file missing: %v", err)
	}
	if _, err := os.Stat(opts.QUICKeyFile); err != nil {
		t.Fatalf("key file missing: %v", err)
	}
	if _, err := tls.LoadX509KeyPair(opts.QUICCertFile, opts.QUICKeyFile); err != nil {
		t.Fatalf("generated cert/key invalid: %v", err)
	}

	certData, err := os.ReadFile(opts.QUICCertFile)
	if err != nil {
		t.Fatalf("read cert file failed: %v", err)
	}
	block, _ := pem.Decode(certData)
	if block == nil {
		t.Fatalf("decode cert pem failed")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("parse cert failed: %v", err)
	}
	if cert.Subject.CommonName != "MyFlowHub QUIC Dev Certificate" {
		t.Fatalf("unexpected cert common name: %q", cert.Subject.CommonName)
	}
	if cert.NotAfter.Before(cert.NotBefore) {
		t.Fatalf("invalid cert time range")
	}
	if certPath := filepath.Base(opts.QUICCertFile); certPath != quicDevCertFileName {
		t.Fatalf("unexpected cert file name: %s", certPath)
	}
	if keyPath := filepath.Base(opts.QUICKeyFile); keyPath != quicDevKeyFileName {
		t.Fatalf("unexpected key file name: %s", keyPath)
	}
}
