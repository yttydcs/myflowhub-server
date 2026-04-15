package defaultset

// 本文件承载默认模块集合中与 `state_backends` 相关的装配逻辑。

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	core "github.com/yttydcs/myflowhub-core"
	flowhandler "github.com/yttydcs/myflowhub-subproto/flow"
	"github.com/yttydcs/myflowhub-subproto/varstore"
)

const (
	cfgFlowBackend                = "flow.backend"
	cfgVarStoreBackend            = "varstore.backend"
	cfgFlowRunArchiveBackend      = "flow.run_archive.backend"
	cfgStatePGDSN                 = "state.pg.dsn"
	cfgStatePGFlowTable           = "state.pg.flow_table"
	cfgStatePGVarTable            = "state.pg.varstore_table"
	cfgStatePGFlowRunArchiveTable = "state.pg.flow_run_archive_table"

	backendJSON   = "json"
	backendMemory = "memory"
	backendPG     = "pg"
	backendOff    = "off"
	backendFile   = "file"

	defaultFlowTable           = "myflowhub_flow_definitions"
	defaultVarTable            = "myflowhub_varstore_records"
	defaultFlowRunArchiveTable = "myflowhub_flow_run_archives"
)

var pgIdentifierPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func newFlowPersistence(cfg core.IConfig) (flowhandler.Persistence, error) {
	switch backendValue(cfg, cfgFlowBackend, backendJSON) {
	case backendJSON:
		return nil, nil
	case backendPG:
		return newPGFlowPersistence(cfg)
	default:
		return nil, fmt.Errorf("unsupported %s", cfgFlowBackend)
	}
}

func newVarStorePersistence(cfg core.IConfig) (varstore.Persistence, error) {
	switch backendValue(cfg, cfgVarStoreBackend, backendMemory) {
	case backendMemory:
		return nil, nil
	case backendPG:
		return newPGVarStorePersistence(cfg)
	default:
		return nil, fmt.Errorf("unsupported %s", cfgVarStoreBackend)
	}
}

func newFlowRunArchiveStore(cfg core.IConfig) (flowhandler.RunArchiveStore, error) {
	switch flowRunArchiveBackendValue(cfg) {
	case backendOff, backendFile:
		return nil, nil
	case backendPG:
		return newPGFlowRunArchiveStore(cfg)
	default:
		return nil, fmt.Errorf("unsupported %s", cfgFlowRunArchiveBackend)
	}
}

func backendValue(cfg core.IConfig, key, def string) string {
	val := strings.ToLower(strings.TrimSpace(def))
	if cfg == nil {
		return val
	}
	if raw, ok := cfg.Get(key); ok {
		if trimmed := strings.ToLower(strings.TrimSpace(raw)); trimmed != "" {
			return trimmed
		}
	}
	return val
}

func flowRunArchiveBackendValue(cfg core.IConfig) string {
	if cfg != nil {
		if raw, ok := cfg.Get(cfgFlowRunArchiveBackend); ok {
			if trimmed := strings.ToLower(strings.TrimSpace(raw)); trimmed != "" {
				return trimmed
			}
		}
		if raw, ok := cfg.Get("flow.run_archive_enabled"); ok {
			switch strings.ToLower(strings.TrimSpace(raw)) {
			case "1", "true", "yes", "on":
				return backendFile
			}
		}
	}
	return backendOff
}

func requiredConfigValue(cfg core.IConfig, key string) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("%s required", key)
	}
	raw, ok := cfg.Get(key)
	if !ok || strings.TrimSpace(raw) == "" {
		return "", fmt.Errorf("%s required", key)
	}
	return strings.TrimSpace(raw), nil
}

func normalizedPGTableName(cfg core.IConfig, key, def string) (string, error) {
	name := strings.TrimSpace(def)
	if cfg != nil {
		if raw, ok := cfg.Get(key); ok && strings.TrimSpace(raw) != "" {
			name = strings.TrimSpace(raw)
		}
	}
	if !pgIdentifierPattern.MatchString(name) {
		return "", fmt.Errorf("invalid %s", key)
	}
	return name, nil
}

func withPGConn(ctx context.Context, dsn string, fn func(context.Context, *pgx.Conn) error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	return fn(ctx, conn)
}

type pgFlowPersistence struct {
	dsn   string
	table string
}

func newPGFlowPersistence(cfg core.IConfig) (flowhandler.Persistence, error) {
	dsn, err := requiredConfigValue(cfg, cfgStatePGDSN)
	if err != nil {
		return nil, err
	}
	if _, err := pgx.ParseConfig(dsn); err != nil {
		return nil, fmt.Errorf("invalid %s: %w", cfgStatePGDSN, err)
	}
	table, err := normalizedPGTableName(cfg, cfgStatePGFlowTable, defaultFlowTable)
	if err != nil {
		return nil, err
	}
	return &pgFlowPersistence{dsn: dsn, table: table}, nil
}

func (p *pgFlowPersistence) ensureSchema(ctx context.Context, conn *pgx.Conn) error {
	query := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	flow_id TEXT PRIMARY KEY,
	doc JSONB NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`, p.table)
	_, err := conn.Exec(ctx, query)
	return err
}

func (p *pgFlowPersistence) LoadAll(ctx context.Context) ([]flowhandler.FlowDocument, error) {
	var docs []flowhandler.FlowDocument
	err := withPGConn(ctx, p.dsn, func(ctx context.Context, conn *pgx.Conn) error {
		if err := p.ensureSchema(ctx, conn); err != nil {
			return err
		}
		rows, err := conn.Query(ctx, fmt.Sprintf(`SELECT flow_id, doc FROM %s ORDER BY flow_id`, p.table))
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var flowID string
			var raw []byte
			if err := rows.Scan(&flowID, &raw); err != nil {
				return err
			}
			var doc flowhandler.FlowDocument
			if err := json.Unmarshal(raw, &doc); err != nil {
				return err
			}
			if strings.TrimSpace(doc.FlowID) == "" {
				doc.FlowID = flowID
			}
			docs = append(docs, doc)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return docs, nil
}

func (p *pgFlowPersistence) Save(ctx context.Context, doc flowhandler.FlowDocument) error {
	raw, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	return withPGConn(ctx, p.dsn, func(ctx context.Context, conn *pgx.Conn) error {
		if err := p.ensureSchema(ctx, conn); err != nil {
			return err
		}
		_, err := conn.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s (flow_id, doc, updated_at)
VALUES ($1, $2::jsonb, NOW())
ON CONFLICT (flow_id) DO UPDATE
SET doc = EXCLUDED.doc,
	updated_at = NOW()`, p.table), strings.TrimSpace(doc.FlowID), string(raw))
		return err
	})
}

func (p *pgFlowPersistence) Delete(ctx context.Context, flowID string) error {
	return withPGConn(ctx, p.dsn, func(ctx context.Context, conn *pgx.Conn) error {
		if err := p.ensureSchema(ctx, conn); err != nil {
			return err
		}
		_, err := conn.Exec(ctx, fmt.Sprintf(`DELETE FROM %s WHERE flow_id = $1`, p.table), strings.TrimSpace(flowID))
		return err
	})
}

type pgVarStorePersistence struct {
	dsn   string
	table string
}

func newPGVarStorePersistence(cfg core.IConfig) (varstore.Persistence, error) {
	dsn, err := requiredConfigValue(cfg, cfgStatePGDSN)
	if err != nil {
		return nil, err
	}
	if _, err := pgx.ParseConfig(dsn); err != nil {
		return nil, fmt.Errorf("invalid %s: %w", cfgStatePGDSN, err)
	}
	table, err := normalizedPGTableName(cfg, cfgStatePGVarTable, defaultVarTable)
	if err != nil {
		return nil, err
	}
	return &pgVarStorePersistence{dsn: dsn, table: table}, nil
}

func (p *pgVarStorePersistence) ensureSchema(ctx context.Context, conn *pgx.Conn) error {
	query := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	owner BIGINT NOT NULL,
	name TEXT NOT NULL,
	value TEXT NOT NULL,
	value_type TEXT NOT NULL,
	visibility TEXT NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY (owner, name)
)`, p.table)
	_, err := conn.Exec(ctx, query)
	return err
}

func (p *pgVarStorePersistence) LoadAll(ctx context.Context) ([]varstore.VarDocument, error) {
	var docs []varstore.VarDocument
	err := withPGConn(ctx, p.dsn, func(ctx context.Context, conn *pgx.Conn) error {
		if err := p.ensureSchema(ctx, conn); err != nil {
			return err
		}
		rows, err := conn.Query(ctx, fmt.Sprintf(`
SELECT owner, name, value, value_type, visibility
FROM %s
ORDER BY owner, name`, p.table))
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var owner int64
			var doc varstore.VarDocument
			if err := rows.Scan(&owner, &doc.Name, &doc.Value, &doc.Type, &doc.Visibility); err != nil {
				return err
			}
			if owner < 0 || owner > int64(^uint32(0)) {
				return errors.New("owner out of range")
			}
			doc.Owner = uint32(owner)
			docs = append(docs, doc)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return docs, nil
}

func (p *pgVarStorePersistence) Save(ctx context.Context, doc varstore.VarDocument) error {
	return withPGConn(ctx, p.dsn, func(ctx context.Context, conn *pgx.Conn) error {
		if err := p.ensureSchema(ctx, conn); err != nil {
			return err
		}
		_, err := conn.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s (owner, name, value, value_type, visibility, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW())
ON CONFLICT (owner, name) DO UPDATE
SET value = EXCLUDED.value,
	value_type = EXCLUDED.value_type,
	visibility = EXCLUDED.visibility,
	updated_at = NOW()`, p.table), int64(doc.Owner), strings.TrimSpace(doc.Name), doc.Value, doc.Type, strings.TrimSpace(doc.Visibility))
		return err
	})
}

func (p *pgVarStorePersistence) Delete(ctx context.Context, owner uint32, name string) error {
	return withPGConn(ctx, p.dsn, func(ctx context.Context, conn *pgx.Conn) error {
		if err := p.ensureSchema(ctx, conn); err != nil {
			return err
		}
		_, err := conn.Exec(ctx, fmt.Sprintf(`DELETE FROM %s WHERE owner = $1 AND name = $2`, p.table), int64(owner), strings.TrimSpace(name))
		return err
	})
}

type pgFlowRunArchiveStore struct {
	dsn   string
	table string
}

func newPGFlowRunArchiveStore(cfg core.IConfig) (flowhandler.RunArchiveStore, error) {
	dsn, err := requiredConfigValue(cfg, cfgStatePGDSN)
	if err != nil {
		return nil, err
	}
	if _, err := pgx.ParseConfig(dsn); err != nil {
		return nil, fmt.Errorf("invalid %s: %w", cfgStatePGDSN, err)
	}
	table, err := normalizedPGTableName(cfg, cfgStatePGFlowRunArchiveTable, defaultFlowRunArchiveTable)
	if err != nil {
		return nil, err
	}
	return &pgFlowRunArchiveStore{dsn: dsn, table: table}, nil
}

func (p *pgFlowRunArchiveStore) ensureSchema(ctx context.Context, conn *pgx.Conn) error {
	query := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	flow_id TEXT NOT NULL,
	run_id TEXT NOT NULL,
	record JSONB NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY (flow_id, run_id)
)`, p.table)
	_, err := conn.Exec(ctx, query)
	return err
}

func (p *pgFlowRunArchiveStore) LoadAll(ctx context.Context) ([]flowhandler.ArchivedRunRecord, error) {
	var records []flowhandler.ArchivedRunRecord
	err := withPGConn(ctx, p.dsn, func(ctx context.Context, conn *pgx.Conn) error {
		if err := p.ensureSchema(ctx, conn); err != nil {
			return err
		}
		rows, err := conn.Query(ctx, fmt.Sprintf(`
SELECT flow_id, run_id, record
FROM %s
ORDER BY flow_id, run_id`, p.table))
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var flowID string
			var runID string
			var raw []byte
			if err := rows.Scan(&flowID, &runID, &raw); err != nil {
				return err
			}
			var record flowhandler.ArchivedRunRecord
			if err := json.Unmarshal(raw, &record); err != nil {
				return err
			}
			if strings.TrimSpace(record.FlowID) == "" {
				record.FlowID = flowID
			}
			if strings.TrimSpace(record.RunID) == "" {
				record.RunID = runID
			}
			records = append(records, record)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return records, nil
}

func (p *pgFlowRunArchiveStore) Save(ctx context.Context, record flowhandler.ArchivedRunRecord) error {
	raw, err := json.Marshal(record)
	if err != nil {
		return err
	}
	return withPGConn(ctx, p.dsn, func(ctx context.Context, conn *pgx.Conn) error {
		if err := p.ensureSchema(ctx, conn); err != nil {
			return err
		}
		_, err := conn.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s (flow_id, run_id, record, updated_at)
VALUES ($1, $2, $3::jsonb, NOW())
ON CONFLICT (flow_id, run_id) DO UPDATE
SET record = EXCLUDED.record,
	updated_at = NOW()`, p.table), strings.TrimSpace(record.FlowID), strings.TrimSpace(record.RunID), string(raw))
		return err
	})
}

func (p *pgFlowRunArchiveStore) Delete(ctx context.Context, flowID, runID string) error {
	return withPGConn(ctx, p.dsn, func(ctx context.Context, conn *pgx.Conn) error {
		if err := p.ensureSchema(ctx, conn); err != nil {
			return err
		}
		_, err := conn.Exec(ctx, fmt.Sprintf(`DELETE FROM %s WHERE flow_id = $1 AND run_id = $2`, p.table), strings.TrimSpace(flowID), strings.TrimSpace(runID))
		return err
	})
}
