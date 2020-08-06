package orm

import (
    "database/sql"
    "journey/cardinal/utils"
    "reflect"
)

type Query struct {
    db    *sql.DB
    tx    *sql.Tx
    ctxFn CtxFn
}

// NewQuery
func NewQuery(db *sql.DB, fn CtxFn) *Query {
    return &Query{
        db:    db,
        ctxFn: fn,
    }
}

// Begin
func (q *Query) Begin(opts *sql.TxOptions) error {
    tx, err := q.db.BeginTx(q.ctxFn(), opts)
    if err != nil {
        return err
    }

    q.tx = tx

    return nil
}

// Rollback
func (q *Query) Rollback() error {
    if q.tx != nil {
        return q.tx.Rollback()
    } else {
        return ErrTransNotStart
    }
}

// Commit
func (q *Query) Commit() error {
    if q.tx != nil {
        return q.tx.Commit()
    } else {
        return ErrTransNotStart
    }
}

// GetStmt
func (q *Query) GetStmt(preSql string) (stmt *sql.Stmt, err error) {
    if q.tx != nil {
        stmt, err = q.tx.Prepare(preSql)
    } else {
        stmt, err = q.db.Prepare(preSql)
    }

    return
}

// Exec
func (q *Query) Exec(preSql string, params ...interface{}) (sql.Result, error) {
    stmt, err := q.GetStmt(preSql)
    if err != nil {
        return nil, err
    }
    defer stmt.Close()

    return stmt.ExecContext(q.ctxFn(), params...)
}

// QueryRow
func (q *Query) QueryRow(preSql string, params []interface{}, values ...interface{}) error {
    stmt, err := q.GetStmt(preSql)
    if err != nil {
        return err
    }
    defer stmt.Close()

    row := stmt.QueryRowContext(q.ctxFn(), params...)

    return row.Scan(values...)
}

// Query
func (q *Query) Query(models interface{}, preSql string, params ...interface{}) (err error) {
    var (
        stmt *sql.Stmt
        rows *sql.Rows
        cols []string
    )

    rValue := reflect.Indirect(reflect.ValueOf(models))
    rType := rValue.Type()
    if rType.Kind() != reflect.Slice {
        return ErrUnsupportedType
    }

    stmt, err = q.GetStmt(preSql)
    if err != nil {
        return err
    }
    defer stmt.Close()

    rows, err = stmt.QueryContext(q.ctxFn(), params...)
    if err != nil {
        return
    }
    defer rows.Close()

    // get the column names of the query
    cols, err = rows.Columns()
    if err != nil {
        return
    }

    var modelType reflect.Type
    isPtr := rType.Elem().Kind() == reflect.Ptr
    if isPtr {
        // e.g. []*Member{}
        modelType = rType.Elem().Elem()
    } else {
        // e.g. []Member
        modelType = rType.Elem()
    }

    alias := q.GetFields(modelType)

    // iterate over each row
    for rows.Next() {
        // new model
        model := reflect.New(modelType).Elem()

        // associated with the query results to model field
        var results []interface{}
        for _, col := range cols {
            field := model.FieldByName(alias[col])
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

        if isPtr {
            rValue.Set(reflect.Append(rValue, model.Addr()))
        } else {
            rValue.Set(reflect.Append(rValue, model))
        }
    }

    return rows.Err()
}

// GetFields
func (q *Query) GetFields(model reflect.Type) map[string]string {
    // sql column name to model field name
    fields := make(map[string]string)
    for i := 0; i < model.NumField(); i++ {
        field := model.Field(i)
        name := field.Name
        if tag, exist := field.Tag.Lookup("orm"); exist {
            fields[tag] = name
        } else {
            fields[utils.UnderScoreCase(name)] = name
        }
    }

    return fields
}
