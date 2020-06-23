package orm

import (
    "context"
    "database/sql"
    "errors"
    "go/ast"
    "journey/cardinal/utils"
    "reflect"
)

type Orm struct {
    engine *Engine
    ctx    context.Context
    tx     *sql.Tx
}

var (
    ErrEngineIsNil     = errors.New("orm: engine is nil")
    ErrConnIsNotOpen   = errors.New("orm: connection is not open")
    ErrTransNotStart   = errors.New("orm: transaction not started")
    ErrUnsupportedType = errors.New("orm: unsupported model type")
    ErrUnexpectedField = errors.New("orm: unexpected model field")
)

// NewOrm
func NewOrm(engine *Engine, ctx context.Context) (*Orm, error) {
    if engine == nil {
        return nil, ErrEngineIsNil
    }

    if engine.db == nil {
        return nil, ErrConnIsNotOpen
    }

    if err := engine.db.Ping(); err != nil {
        return nil, err
    }

    if ctx == nil {
        ctx = context.Background()
    }

    return &Orm{
        engine: engine,
        ctx:    ctx,
    }, nil
}

// ParseModel
func (orm *Orm) ParseModel(model interface{}) *Table {
    typ := reflect.Indirect(reflect.ValueOf(model)).Type()
    if typ.Kind() != reflect.Struct {
        return nil
    }

    table := &Table{
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

    return table
}

// CreateTable
func (orm *Orm) CreateTable(model interface{}) error {
    // TODO://

    return nil
}

// CreateTable
func (orm *Orm) DropTable(model interface{}) error {
    // TODO://

    return nil
}

// ExistTable
func (orm *Orm) ExistTable(model interface{}) error {
    // TODO://

    return nil
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

// Begin
func (orm *Orm) Begin(opts *sql.TxOptions) error {
    tx, err := orm.engine.db.BeginTx(orm.ctx, opts)
    if err != nil {
        return err
    }

    orm.tx = tx

    return nil
}

// Rollback
func (orm *Orm) Rollback() error {
    if orm.tx != nil {
        return orm.tx.Rollback()
    } else {
        return ErrTransNotStart
    }
}

// Commit
func (orm *Orm) Commit() error {
    if orm.tx != nil {
        return orm.tx.Commit()
    } else {
        return ErrTransNotStart
    }
}

// Exec
func (orm *Orm) Exec(preSql string, params ...interface{}) (ret sql.Result, err error) {
    var stmt *sql.Stmt
    if orm.tx != nil {
        stmt, err = orm.tx.Prepare(preSql)
    } else {
        stmt, err = orm.engine.db.Prepare(preSql)
    }

    if err != nil {
        return
    }
    defer stmt.Close()

    return stmt.ExecContext(orm.ctx, params...)
}

// QueryRow
func (orm *Orm) QueryRow(model interface{}, preSql string, params ...interface{}) (err error) {
    // TODO://

    return
}

// Query
func (orm *Orm) Query(models interface{}, preSql string, params ...interface{}) (err error) {
    var (
        stmt *sql.Stmt
        rows *sql.Rows
        cols []string
    )

    rValue := reflect.Indirect(reflect.ValueOf(models))
    rType := rValue.Type()

    switch rType.Kind() {
    case reflect.Slice:
        //
    case reflect.Struct:
        return orm.QueryRow(models, preSql, params...)
    default:
        return ErrUnsupportedType
    }

    if orm.tx != nil {
        stmt, err = orm.tx.Prepare(preSql)
    } else {
        stmt, err = orm.engine.db.Prepare(preSql)
    }

    if err != nil {
        return
    }
    defer stmt.Close()

    rows, err = stmt.QueryContext(orm.ctx, params...)
    if err != nil {
        return
    }
    defer rows.Close()

    // get the column names of the query
    cols, err = rows.Columns()
    if err != nil {
        return
    }

    model := reflect.New(rType.Elem()).Elem()
    // iterate over each row
    for rows.Next() {
        var results []interface{}
        // associated with the query results to model field
        for _, col := range cols {
            // TODO:// alias tag
            field := model.FieldByName(utils.PascalCase(col))
            if field.IsValid() {
                results = append(results, field.Addr().Interface())
            } else {
                return ErrUnexpectedField
            }
        }

        err = rows.Scan(results...)
        if err != nil {
            return
        }
        rValue.Set(reflect.Append(rValue, model))
    }

    return rows.Err()
}
