package orm

import "journey/cardinal/utils"

type Column struct {
    Name    string // for model
    Alias   string // for sql
    Type    string
    Options string
}

// GetAlias
func (col *Column) GetAlias() string {
    if col.Alias != "" {
        return col.Alias
    }

    return utils.UnderScoreCase(col.Name)
}
