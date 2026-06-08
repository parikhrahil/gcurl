package audit

import (
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
	query := fmt.Sprintf("SELECT id, http_method, target_url, status_code, bytes_transmitted, bytes_received, total_duration_us, executed_at FROM gcurl_history_ledger ORDER BY executed_at DESC LIMIT %d;", limit)
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to scan audit history table: %w", err)
	}
	defer rows.Close()

	var logs []HistoricalLog
	for rows.Next() {
		var log HistoricalLog
		var durationUs int64
		var executedAtStr string

		err := rows.Scan(
			&log.ID,
			&log.Method,
			&log.URL,
			&log.StatusCode,
			&log.BytesTransmitted,
			&log.BytesReceived,
			&durationUs,
			&executedAtStr,
		)
		if err != nil {
			return nil, fmt.Errorf("row mapping corruption detected during historical parse: %w", err)
		}

		log.Duration = time.Duration(durationUs) * time.Microsecond
		parsedTime, parseErr := time.Parse(time.RFC3339, executedAtStr)
		if parseErr != nil {
			parsedTime, parseErr = time.Parse("2006-01-02 15:04:05", executedAtStr)
		}

		if parseErr != nil {
			log.ExecutedAt = time.Now().UTC()
		} else {
			log.ExecutedAt = parsedTime
		}
		logs = append(logs, log)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("historical log streaming encountered an unexpected driver error: %w", err)
	}
	return logs, nil
}
