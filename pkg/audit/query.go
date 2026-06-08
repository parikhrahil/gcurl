package audit

import (
	"context"
	"fmt"
	"time"
)

type HistoricalLog struct {
	ID               int
	Method           string
	URL              string
	StatusCode       int
	BytesTransmitted int64
	BytesReceived    int64
	Duration         time.Duration
	ExecutedAt       time.Time
}

func (r *HistoryRepository) FetchHistory(limit int) ([]HistoricalLog, error) {
	query := `
		SELECT id, http_method, target_url, status_code, bytes_transmitted, bytes_received, total_duration_us, executed_at
		FROM gcurl_history_ledger
		ORDER BY executed_at DESC
		LIMIT ?;
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to scan audit history table: %w", err)
	}
	defer rows.Close()

	var logs []HistoricalLog
	for rows.Next() {
		var log HistoricalLog
		var durationUs int64
		var timeStr string

		err := rows.Scan(
			&log.ID,
			&log.Method,
			&log.URL,
			&log.StatusCode,
			&log.BytesTransmitted,
			&log.BytesReceived,
			&durationUs,
			&timeStr,
		)
		if err != nil {
			return nil, fmt.Errorf("row mapping corruption detected during historical parse: %w", err)
		}

		log.Duration = time.Duration(durationUs) * time.Microsecond
		log.ExecutedAt, _ = time.Parse("2006-01-02 15:04:05", timeStr)
		logs = append(logs, log)
	}
	return logs, nil
}
