package defaultset

import (
	"context"
	"testing"

	"github.com/yttydcs/myflowhub-core/config"
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
