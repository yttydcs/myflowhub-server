package defaultset

// Context: This file lives in the Server assembly layer and supports state_backends_test.

import (
	"context"
	"testing"

	"github.com/yttydcs/myflowhub-core/config"
	flowproto "github.com/yttydcs/myflowhub-proto/protocol/flow"
)

func TestDefaultHubRejectsUnsupportedFlowBackend(t *testing.T) {
	cfg := config.NewMap(map[string]string{
		cfgFlowBackend: "sqlite",
	})

	if _, _, err := DefaultHub(cfg, nil); err == nil {
		t.Fatalf("expected unsupported flow backend error")
	}
}

func TestDefaultHubRejectsUnsupportedVarStoreBackend(t *testing.T) {
	cfg := config.NewMap(map[string]string{
		cfgVarStoreBackend: "sqlite",
	})

	if _, _, err := DefaultHub(cfg, nil); err == nil {
		t.Fatalf("expected unsupported varstore backend error")
	}
}

func TestDefaultHubRequiresPGDSNWhenBackendIsPG(t *testing.T) {
	cfg := config.NewMap(map[string]string{
		cfgFlowBackend: backendPG,
	})

	if _, _, err := DefaultHub(cfg, nil); err == nil {
		t.Fatalf("expected missing pg dsn error")
	}
}

func TestDefaultHubRejectsUnsupportedFlowRunArchiveBackend(t *testing.T) {
	cfg := config.NewMap(map[string]string{
		cfgFlowRunArchiveBackend: "sqlite",
	})

	if _, _, err := DefaultHub(cfg, nil); err == nil {
		t.Fatalf("expected unsupported flow run archive backend error")
	}
}

func TestDefaultHubRequiresPGDSNWhenFlowRunArchiveBackendIsPG(t *testing.T) {
	cfg := config.NewMap(map[string]string{
		cfgFlowRunArchiveBackend: backendPG,
	})

	if _, _, err := DefaultHub(cfg, nil); err == nil {
		t.Fatalf("expected missing pg dsn error")
	}
}

func TestPGFlowPersistenceLoadAllReturnsConnectionError(t *testing.T) {
	cfg := config.NewMap(map[string]string{
		cfgStatePGDSN: "postgres://127.0.0.1:1/myflowhub?sslmode=disable&connect_timeout=1",
	})

	store, err := newPGFlowPersistence(cfg)
	if err != nil {
		t.Fatalf("newPGFlowPersistence err=%v", err)
	}
	if _, err := store.LoadAll(context.Background()); err == nil {
		t.Fatalf("expected pg connection error")
	}
}

func TestPGFlowRunArchiveStoreLoadAllReturnsConnectionError(t *testing.T) {
	cfg := config.NewMap(map[string]string{
		cfgStatePGDSN: "postgres://127.0.0.1:1/myflowhub?sslmode=disable&connect_timeout=1",
	})

	store, err := newPGFlowRunArchiveStore(cfg)
	if err != nil {
		t.Fatalf("newPGFlowRunArchiveStore err=%v", err)
	}
	if _, err := store.LoadAll(context.Background()); err == nil {
		t.Fatalf("expected pg connection error")
	}
}

func TestPGVarStorePersistenceLoadAllReturnsConnectionError(t *testing.T) {
	cfg := config.NewMap(map[string]string{
		cfgStatePGDSN: "postgres://127.0.0.1:1/myflowhub?sslmode=disable&connect_timeout=1",
	})

	store, err := newPGVarStorePersistence(cfg)
	if err != nil {
		t.Fatalf("newPGVarStorePersistence err=%v", err)
	}
	if _, err := store.LoadAll(context.Background()); err == nil {
		t.Fatalf("expected pg connection error")
	}
}

func TestDefaultHubWithPGFlowBackendIncludesFlowHandler(t *testing.T) {
	cfg := config.NewMap(map[string]string{
		cfgFlowBackend:      backendPG,
		cfgStatePGDSN:       "postgres://postgres:postgres@127.0.0.1:5432/myflowhub?sslmode=disable",
		cfgStatePGFlowTable: "myflowhub_flow_definitions",
	})

	handlers, _, err := DefaultHub(cfg, nil)
	if err != nil {
		t.Fatalf("DefaultHub() err=%v", err)
	}
	for _, h := range handlers {
		if h != nil && h.SubProto() == flowproto.SubProtoFlow {
			return
		}
	}
	t.Fatalf("DefaultHub() missing flow handler for subproto=%d", flowproto.SubProtoFlow)
}

func TestDefaultHubWithPGFlowRunArchiveBackendIncludesFlowHandler(t *testing.T) {
	cfg := config.NewMap(map[string]string{
		cfgFlowRunArchiveBackend:      backendPG,
		cfgStatePGDSN:                 "postgres://postgres:postgres@127.0.0.1:5432/myflowhub?sslmode=disable",
		cfgStatePGFlowRunArchiveTable: "myflowhub_flow_run_archives",
	})

	handlers, _, err := DefaultHub(cfg, nil)
	if err != nil {
		t.Fatalf("DefaultHub() err=%v", err)
	}
	for _, h := range handlers {
		if h != nil && h.SubProto() == flowproto.SubProtoFlow {
			return
		}
	}
	t.Fatalf("DefaultHub() missing flow handler for subproto=%d", flowproto.SubProtoFlow)
}
