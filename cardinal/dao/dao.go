package dao

import (
    "database/sql"
)

type Dao struct {
    isTx  bool
    table string
}

func NewDao(table string) *Dao {
    return &Dao{
        table: table,
    }
}

// Exec
func (dao *Dao) Exec(preSql string, params ...interface{}) (err error) {
    var stmt *sql.Stmt
    if dao.isTx {
        stmt, err = tx.Prepare(preSql)
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

    _, err = stmt.Exec(params...)

    return
}

// Query
func (dao *Dao) Query(preSql string, params ...interface{}) (result []map[string]interface{}, err error) {
    var (
        stmt *sql.Stmt
        rows *sql.Rows
        cols []string
    )

    if dao.isTx {
        stmt, err = tx.Prepare(preSql)
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

    rows, err = stmt.Query(params...)
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

    return
}
