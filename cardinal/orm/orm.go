package orm

import (
    "context"
    "database/sql"
)

type Orm struct {
    ctx context.Context
    tx  *sql.Tx
}
