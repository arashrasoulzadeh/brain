package plugin

import (
	"context"

	"brain/internal/cortex"
)

// BindOperatorDrains initializes host-only placement metadata before execution.
func (s *SQLiteStorage) BindOperatorDrains(ctx context.Context, scheduler *cortex.Scheduler) error {
	if scheduler == nil {
		return cortex.ErrInvalidConfig
	}
	if _, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS cortex_operator_drains (
		neuron_id TEXT PRIMARY KEY NOT NULL
	)`); err != nil {
		return err
	}
	return scheduler.BindOperatorDrainStore(ctx, s)
}

func (s *SQLiteStorage) LoadOperatorDrains(ctx context.Context, limit int) ([]string, error) {
	if limit < 1 || limit > 4097 {
		return nil, cortex.ErrInvalidConfig
	}
	rows, err := s.db.QueryContext(ctx, `SELECT neuron_id FROM cortex_operator_drains ORDER BY neuron_id LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (s *SQLiteStorage) SaveOperatorDrain(ctx context.Context, id string, draining bool) error {
	if draining {
		_, err := s.db.ExecContext(ctx, `INSERT INTO cortex_operator_drains(neuron_id) VALUES (?) ON CONFLICT(neuron_id) DO NOTHING`, id)
		return err
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM cortex_operator_drains WHERE neuron_id=?`, id)
	return err
}
