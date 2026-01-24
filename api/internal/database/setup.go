package database

import (
	"context"
	"fmt"

	"mars/internal/database/sql"
)

func ApplySchema(ctx context.Context, db DBTX) error {
	if _, err := db.Exec(ctx, sql.Schema()); err != nil {
		return fmt.Errorf("applying schema: %w", err)
	}
	return nil
}
