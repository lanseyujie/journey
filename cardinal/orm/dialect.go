package orm

import (
    "database/sql"
    "reflect"
)

// Dialect
type Dialect interface {
    GetDriverName() string
    GetDsn() string
    Config(db *sql.DB)
    GetTablePrefix() string
    GetColumnSqlType(typ reflect.Value) string
    GetExistTableSql(table string) string
}

var dialectMap = make(map[string]Dialect)

// RegisterDialect
func RegisterDialect(name string, dialect Dialect) {
    dialectMap[name] = dialect
}

// GetDialect
func GetDialect(name string) Dialect {
    if dialect, exist := dialectMap[name]; exist {
        return dialect
    }

    return nil
}
