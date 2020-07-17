package orm

import (
    "journey/cardinal/utils"
    "strings"
)

type Table struct {
    Prefix  string
    Name    string // for model
    Alias   string // for sql
    Model   interface{}
    Columns []*Column
}

// GetAlias
func (t *Table) GetAlias() string {
    if t.Alias != "" {
        return t.Alias
    }

    return utils.UnderScoreCase(t.Prefix + t.Name)
}

// Create
func (t *Table) Create() (sql string) {
    var columns []string
    for _, col := range t.Columns {
        columns = append(columns, "`"+col.Alias+"` "+col.Type+" "+col.Options)
    }

    return "CREATE TABLE IF NOT EXISTS `" + t.Alias + "` (" + strings.Join(columns, ",") + ");"
}

// Drop
func (t *Table) Drop() string {
    return "DROP TABLE IF EXISTS `" + t.Alias + "`;"
}

// Exist
func (t *Table) Exist(dialect Dialect) (string, []interface{}) {
    return dialect.GetExistTableSql(t.Alias)
}

// Alter
func Alter() string {
    // TODO://

    return ""
}

// Truncate
func (t *Table) Truncate() string {
    return "TRUNCATE TABLE `" + t.GetAlias() + "`;"
}
