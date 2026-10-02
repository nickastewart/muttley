package records

import (
	"context"
	"log/slog"
	"muttley/sqlite"
	"muttley/sqlite/entities"
)

type RecordsRepositorySqlite struct {
	queries *entities.Queries
}

func NewRecordsRepository(queries *entities.Queries) *RecordsRepositorySqlite {
	return &RecordsRepositorySqlite{
		queries: queries,
	}
}

func (r *RecordsRepositorySqlite) q(ctx context.Context) *entities.Queries {
	return sqlite.Queries(ctx, r.queries)
}

func (r *RecordsRepositorySqlite) ListLocationRecords(ctx context.Context, userID int64) ([]entities.ListLocationRecordsRow, error) {
	rows, err := r.q(ctx).ListLocationRecords(ctx, userID)
	if err != nil {
		slog.Error(err.Error())
	}
	return rows, err
}

func (r *RecordsRepositorySqlite) GetLocationRecordSnapshot(ctx context.Context, arg entities.GetLocationRecordSnapshotParams) (entities.GetLocationRecordSnapshotRow, error) {
	row, err := r.q(ctx).GetLocationRecordSnapshot(ctx, arg)
	if err != nil {
		slog.Error(err.Error())
	}
	return row, err
}
