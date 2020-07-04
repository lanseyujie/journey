package orm

import "database/sql"

type Engine struct {
    dialect Dialect
    db      *sql.DB
}

var engineMap = make(map[string]*Engine)

// NewEngine
func NewEngine(name string, d Dialect) *Engine {
    if e, exist := engineMap[name]; exist {
        return e
    }

    e:=  &Engine{
        dialect: d,
    }

    engineMap[name] = e

    return e
}

// GetEngine
func GetEngine(name string) *Engine {
    if e, exist := engineMap[name]; exist {
        return e
    }

    return nil
}

// GetDsn
func (e *Engine) Open() (err error) {
    e.db, err = sql.Open(e.dialect.GetDriverName(), e.dialect.GetDsn())
    if err != nil {
        return
    }

    e.dialect.Config(e.db)

    return e.db.Ping()
}

// Close
func (e *Engine) Close() {
    if e != nil && e.db != nil {
        _ = e.db.Close()
    }
}

// Db
func (e *Engine) Db() *sql.DB {
    if e != nil && e.db != nil {
        return e.db
    }

    return nil
}
