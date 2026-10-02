package records

import (
	"context"
	"muttley/sqlite/entities"
)

type RecordsRepository interface {
	ListLocationRecords(ctx context.Context, userID int64) ([]entities.ListLocationRecordsRow, error)
	GetLocationRecordSnapshot(ctx context.Context, arg entities.GetLocationRecordSnapshotParams) (entities.GetLocationRecordSnapshotRow, error)
}
