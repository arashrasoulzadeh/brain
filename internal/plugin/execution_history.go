package plugin

import (
	"context"
	"encoding/json"

	"brain/internal/cortex"
)

// BindExecutionHistory keeps the bounded Cortex journal in the existing host
// database, separate from project-visible tables and committed effects.
func (s *SQLiteStorage) BindExecutionHistory(ctx context.Context, scheduler *cortex.Scheduler) error {
	if scheduler == nil {
		return cortex.ErrInvalidConfig
	}
	if _, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS cortex_execution_history (
		sequence INTEGER PRIMARY KEY AUTOINCREMENT,
		finished_at INTEGER NOT NULL,
		trace_json TEXT NOT NULL
	)`); err != nil {
		return err
	}
	return scheduler.BindExecutionHistory(ctx, s)
}

func (s *SQLiteStorage) LoadRecentExecutions(ctx context.Context, limit int) ([]cortex.ExecutionTrace, error) {
	if limit <= 0 || limit > 128 {
		return nil, cortex.ErrInvalidConfig
	}
	rows, err := s.db.QueryContext(ctx, `SELECT trace_json FROM cortex_execution_history
		ORDER BY finished_at DESC, sequence DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]cortex.ExecutionTrace, 0, limit)
	for rows.Next() {
		var encoded string
		var trace cortex.ExecutionTrace
		if err := rows.Scan(&encoded); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(encoded), &trace); err != nil {
			return nil, err
		}
		result = append(result, trace)
	}
	return result, rows.Err()
}

func (s *SQLiteStorage) AppendExecutionTrace(ctx context.Context, trace cortex.ExecutionTrace, limit int) error {
	if limit <= 0 || limit > 128 {
		return cortex.ErrInvalidConfig
	}
	encoded, err := json.Marshal(trace)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `INSERT INTO cortex_execution_history(finished_at, trace_json)
		VALUES (?, ?)`, trace.FinishedAt.UnixNano(), string(encoded)); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM cortex_execution_history WHERE sequence NOT IN
		(SELECT sequence FROM cortex_execution_history ORDER BY finished_at DESC, sequence DESC LIMIT ?)`, limit); err != nil {
		return err
	}
	return tx.Commit()
}
