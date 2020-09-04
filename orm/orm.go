package orm

import (
    "context"
    "database/sql"
    "errors"
    "go/ast"
    "journey/cardinal/utils"
    "reflect"
    "time"
)

type Orm struct {
    engine *Engine
}

type CtxFn func() context.Context

var (
    ErrEngineIsNil     = errors.New("orm: engine is nil")
    ErrConnIsNotOpen   = errors.New("orm: connection is not open")
    ErrTransNotStart   = errors.New("orm: transaction not started")
    ErrUnsupportedType = errors.New("orm: unsupported model type")
    ErrUnexpectedField = errors.New("orm: unexpected model field")
)

// NewOrm
func NewOrm(engine *Engine) (*Orm, error) {
    if engine == nil {
        return nil, ErrEngineIsNil
    }

    if engine.db == nil {
        return nil, ErrConnIsNotOpen
    }

    if err := engine.db.Ping(); err != nil {
        return nil, err
    }

    return &Orm{engine: engine}, nil
}

// Query
func (orm *Orm) Query(fn CtxFn) *Query {
    if fn == nil {
        fn = func() context.Context {
            ctx, _ := context.WithTimeout(context.Background(), time.Second*3)

            return ctx
        }
    }
    return &Query{
        db:    orm.engine.db,
        ctxFn: fn,
    }
}

// ParseModel
func (orm *Orm) ParseModel(model interface{}) (table *Table, err error) {
    typ := reflect.Indirect(reflect.ValueOf(model)).Type()
    if typ.Kind() != reflect.Struct {
        return nil, ErrUnsupportedType
    }

    // table = cache.Get(typ.String())
    // if table != nil {
    //     return
    // }

    table = &Table{
        Prefix: orm.engine.dialect.GetTablePrefix(),
        Name:   typ.Name(),
        Model:  model,
    }
    table.Alias = table.GetAlias()

    for i := 0; i < typ.NumField(); i++ {
        field := typ.Field(i)
        if !field.Anonymous && ast.IsExported(field.Name) {
            column := &Column{
                Name: field.Name,
            }

            if value, exist := field.Tag.Lookup("orm"); exist {
                column.Alias = value
            } else {
                column.Alias = column.GetAlias()
            }

            if value, exist := field.Tag.Lookup("type"); exist {
                column.Type = value
            } else {
                column.Type = orm.engine.dialect.GetColumnSqlType(reflect.Indirect(reflect.New(field.Type)))
            }

            if value, exist := field.Tag.Lookup("opt"); exist {
                column.Options = value
            }

            table.Columns = append(table.Columns, column)
        }
    }

    // cache.Put(typ.String(), table)

    return table, nil
}

// CreateTable
func (orm *Orm) CreateTable(model interface{}) error {
    table, err := orm.ParseModel(model)
    if err != nil {
        return err
    }

    _, err = orm.Query(nil).Exec(table.Create())

    return err
}

// CreateTable
func (orm *Orm) DropTable(model interface{}) error {
    table, err := orm.ParseModel(model)
    if err != nil {
        return err
    }

    _, err = orm.Query(nil).Exec(table.Drop())

    return err
}

// TableAlias
func (orm *Orm) TableAlias(table interface{}) (alias string, err error) {
    prefix := orm.engine.dialect.GetTablePrefix()
    switch tab := table.(type) {
    case string:
        alias = prefix + tab
    case *Table:
        alias = tab.GetAlias()
    default:
        typ := reflect.Indirect(reflect.ValueOf(tab)).Type()
        if typ.Kind() != reflect.Struct {
            return "", ErrUnsupportedType
        } else {
            alias = utils.UnderScoreCase(prefix + typ.Name())
        }
    }

    return
}

// ExistTable
func (orm *Orm) ExistTable(table interface{}) (exist bool, err error) {
    var (
        preSql string
        params []interface{}
    )

    switch tab := table.(type) {
    case string:
        preSql, params = orm.engine.dialect.GetExistTableSql(tab)
    case *Table:
        preSql, params = tab.Exist(orm.engine.dialect)
    default:
        return false, ErrUnsupportedType
    }

    err = orm.Query(nil).QueryRow(preSql, params, &exist)

    return
}

// AlterTable
func (orm *Orm) AlterTable(oldModel, newModel interface{}) error {
    // TODO://

    return nil
}

// TruncateTable
func (orm *Orm) TruncateTable(model interface{}) error {
    // TODO://

    return nil
}

// Insert
func (orm *Orm) Insert(model ...interface{}) (ret sql.Result, err error) {
    // TODO://

    return
}

// Delete
func (orm *Orm) Delete() {
    // TODO://
}

// Update
func (orm *Orm) Update() {
    // TODO://
}

// Select
func (orm *Orm) Select(models interface{}) {
    // TODO://
}

// Migrate
func (orm *Orm) Migrate() {
    // TODO://
}
