package dao

type Dao struct {
    table string
}

func NewDao(table string) *Dao {
    return &Dao{
        table: table,
    }
}

// Exec
func (dao *Dao) Exec(preSql string, params ...interface{}) (err error) {
    // TODO://

    return
}

// Query
func (dao *Dao) Query(preSql string, params ...interface{}) (result []map[string]interface{}, err error) {
    // TODO://

    return
}
