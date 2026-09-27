package headtohead

import (
	"context"
	"muttley/sqlite/entities"
)

type HeadToHeadRepository interface {
	GetHeadToHead(ctx context.Context, userId int64) ([]entities.GetHeadToHeadRow, error)
}
