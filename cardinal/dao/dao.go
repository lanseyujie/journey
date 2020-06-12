package dao

import (
    "database/sql"
    "strings"
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
    preSql = strings.TrimSpace(preSql)
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

    preSql = strings.TrimSpace(preSql)
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
    // store the value of each field in a row
    row := make([]sql.RawBytes, count)
    scans := make([]interface{}, count)
    // let each row of data be filled in []sql.RawBytes
    for i := range row {
        scans[i] = &row[i]
    }

    for rows.Next() {
        // scans[i] = &values[i]
        err = rows.Scan(scans...)
        if err == nil {
            // a row of data, column => value
            values := make(map[string]interface{})
            for i, v := range row {
                if v != nil {
                    values[cols[i]] = string(v)
                } else {
                    // SQL NULL
                    values[cols[i]] = nil
                }
            }
            result = append(result, values)
        }
    }

    return
}
