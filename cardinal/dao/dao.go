package dao

import (
    "database/sql"
    "fmt"
    "time"
)

type Dao struct {
    isTx bool
}

type Result []map[string]interface{}

func NewDao() *Dao {
    return &Dao{}
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
func (dao *Dao) Query(preSql string, params ...interface{}) (result Result, err error) {
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

// QueryRow
func (dao *Dao) QueryRow(preSql string, params []interface{}, values ...interface{}) (err error) {
    var stmt *sql.Stmt
    if tx != nil {
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

    row := stmt.QueryRow(params...)
    err = row.Scan(values...)

    return
}

// String
func (data Result) String() string {
    if len(data) <= 0 {
        return ""
    }

    str := ""
    cols := make([]string, len(data[0]))
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
