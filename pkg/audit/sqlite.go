package audit

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/glebarez/go-sqlite"
	"github.com/parikhrahil/gcurl/pkg/config"
)

type HistoryRepository struct {
	db *sql.DB
}

func NewHistoryRepository(dbPath string) (*HistoryRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open embedded database layer: %w", err)
	}

	repository := &HistoryRepository{db: db}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(5 * time.Second)

	if err := repository.tunePerformance(); err != nil {
		db.Close()
		return nil, err
	}

	if err := repository.migrateSchema(); err != nil {
		db.Close()
		return nil, err
	}
	return repository, nil
}

// tunePerformance injects strict engine settings to optimize local disk writes.
func (r *HistoryRepository) tunePerformance() error {
	// WAL (Write-Ahead Logging) mode allows concurrent reads while a write transaction is executing
	// Synchronous=NORMAL balances write safety with split-second terminal execution speed
	pragmas := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA synchronous = NORMAL;",
		"PRAGMA foreign_keys = ON;",
	}

	for _, pragma := range pragmas {
		if _, err := r.db.Exec(pragma); err != nil {
			return fmt.Errorf("failed to apply performance optimization pragma (%s): %w", pragma, err)
		}
	}

	return nil
}

// Close gracefully releases the underlying file handle descriptors back to the OS pool.
func (r *HistoryRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

func (r *HistoryRepository) migrateSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS gcurl_history_ledger (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		http_method TEXT NOT NULL,
		target_url TEXT NOT NULL,
		status_code INTEGER NOT NULL,
		bytes_transmitted INTEGER NOT NULL DEFAULT 0,
		bytes_received INTEGER NOT NULL DEFAULT 0,
		total_duration_us INTEGER NOT NULL DEFAULT 0,
		executed_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_gcurl_history_url_time 
	ON gcurl_history_ledger (target_url, executed_at DESC);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if _, err := r.db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to execute automated schema migration: %w", err)
	}

	return nil
}

func (r *HistoryRepository) WriteAuditTrail(cfg *config.RequestConfiguration) error {
	query := `
		INSERT INTO gcurl_history_ledger(
			http_method, target_url, status_code, bytes_transmitted, bytes_received, total_duration_us
		) VALUES (?, ?, ?, ?, ?, ?)
	`
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := r.db.ExecContext(
		ctx,
		query,
		cfg.Method,
		cfg.URL,
		0,
		cfg.Metrics.BytesTransmitted,
		cfg.Metrics.BytesReceived,
		cfg.Metrics.TotalDuration.Microseconds(),
	)
	if err != nil {
		return err
	}
	return r.enforceLRUBoundary()
}

func (r *HistoryRepository) enforceLRUBoundary() error {
	query := `
		DELETE FROM gcurl_history_ledger
		WHERE id NOT IN (
			SELECT id FROM gcurl_history_ledger
			ORDER BY executed_at DESC
			LIMIT 50
		);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_, err := r.db.ExecContext(ctx, query)
	return err
}
