package dao

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "time"
)

type Dao struct {
    ctx context.Context
    tx  *sql.Tx
}

type Result []map[string]interface{}

var ErrTransNotStart = errors.New("dao: transaction not started")

// NewDao
func NewDao(ctx ...context.Context) *Dao {
    var c context.Context
    if len(ctx) == 1 {
        c = ctx[0]
    } else {
        c = context.Background()
    }

    return &Dao{
        ctx: c,
    }
}

// Begin
func (dao *Dao) Begin(opts *sql.TxOptions) error {
    tx, err := db.BeginTx(dao.ctx, opts)
    if err != nil {
        return err
    }

    dao.tx = tx

    return nil
}

// Rollback
func (dao *Dao) Rollback() error {
    if dao.tx != nil {
        return dao.tx.Rollback()
    } else {
        return ErrTransNotStart
    }
}

// Commit
func (dao *Dao) Commit() error {
    if dao.tx != nil {
        return dao.tx.Commit()
    } else {
        return ErrTransNotStart
    }
}

// Exec
func (dao *Dao) Exec(preSql string, params ...interface{}) (ret sql.Result, err error) {
    var stmt *sql.Stmt
    if dao.tx != nil {
        stmt, err = dao.tx.Prepare(preSql)
        if err != nil {
            return
        }
    } else {
        stmt, err = db.Prepare(preSql)
        if err != nil {
            return
        }
    }
    defer stmt.Close()

    return stmt.ExecContext(dao.ctx, params...)
}

// ExecMulti used to insert multiple data at once
func (dao *Dao) ExecMulti(preSql string, params ...[]interface{}) (err error) {
    var stmt *sql.Stmt
    if dao.tx != nil {
        stmt, err = dao.tx.Prepare(preSql)
        if err != nil {
            return
        }
    } else {
        stmt, err = db.Prepare(preSql)
        if err != nil {
            return
        }
    }
    defer stmt.Close()

    for _, p := range params {
        _, err = stmt.ExecContext(dao.ctx, p...)
        if err != nil {
            return
        }
    }

    return
}

// Query
func (dao *Dao) Query(preSql string, params ...interface{}) (result Result, err error) {
    var (
        stmt *sql.Stmt
        rows *sql.Rows
        cols []string
    )

    if dao.tx != nil {
        stmt, err = dao.tx.Prepare(preSql)
        if err != nil {
            return
        }
    } else {
        stmt, err = db.Prepare(preSql)
        if err != nil {
            return
        }
    }
    defer stmt.Close()

    rows, err = stmt.QueryContext(dao.ctx, params...)
    if err != nil {
        return
    }
    defer rows.Close()

    // get the column names of the query
    cols, err = rows.Columns()
    if err != nil {
        return
    }

    count := len(cols)
    // iterate over each row
    for rows.Next() {
        // store the value of each field in a row
        row := make([]interface{}, count)
        for i := range row {
            row[i] = new(interface{})
        }

        err = rows.Scan(row...)
        if err != nil {
            return
        }

        // a row of data, column => value
        values := make(map[string]interface{}, count)
        for index, col := range cols {
            values[col] = row[index]
        }
        result = append(result, values)
    }

    err = rows.Err()

    return
}

// QueryRow
func (dao *Dao) QueryRow(preSql string, params []interface{}, values ...interface{}) (err error) {
    var stmt *sql.Stmt
    if dao.tx != nil {
        stmt, err = dao.tx.Prepare(preSql)
        if err != nil {
            return
        }
    } else {
        stmt, err = db.Prepare(preSql)
        if err != nil {
            return
        }
    }
    defer stmt.Close()

    row := stmt.QueryRowContext(dao.ctx, params...)

    return row.Scan(values...)
}

// String
func (data Result) String() string {
    if len(data) <= 0 {
        return ""
    }

    str := ""
    cols := make([]string, 0, len(data[0]))
    for col := range data[0] {
        cols = append(cols, col)
    }

    for _, row := range data {
        for _, col := range cols {
            if val, exist := row[col]; exist {
                str += col + ":"
                switch v := (*(val.(*interface{}))).(type) {
                case nil:
                    str += "NULL"
                case []byte:
                    str += string(v)
                case time.Time:
                    str += v.Format("2006-01-02 15:04:05.000")
                default:
                    str += fmt.Sprint(v)
                }
                str += "\t"
            }
        }
        str += "\n"
    }

    return str
}
