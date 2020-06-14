package dao

import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "strconv"
    "time"
)

type MySql struct {
    host      string
    port      uint
    username  string
    password  string
    database  string
    prefix    string
    charset   string
    collation string
}

var (
    db    *sql.DB
    mysql *MySql
)

// NewMySql
// host, port, username, password, database, prefix, charset, collation
func NewMySql(host string, port uint, username, password, database string, params ...string) *MySql {
    if mysql != nil {
        mysql.Close()
    }

    prefix := ""
    charset := "utf8mb4"
    collation := "utf8mb4_general_ci"
    if len(params) > 0 {
        for i, val := range params {
            if val != "" {
                if i == 0 {
                    prefix = val
                } else if i == 1 {
                    charset = val
                } else if i == 2 {
                    collation = val
                }
            }
        }
    }

    return &MySql{
        host:      host,
        port:      port,
        username:  username,
        password:  password,
        database:  database,
        prefix:    prefix,
        charset:   charset,
        collation: collation,
    }
}

// GetDsn
func (ms *MySql) GetDsn() string {
    return ms.username + ":" + ms.password + "@tcp(" + ms.host + ":" + strconv.FormatInt(int64(ms.port), 10) + ")/" + ms.database + "?charset=" + ms.charset + "&collation=" + ms.collation
}

// GetDsn
func (ms *MySql) Open() (err error) {
    db, err = sql.Open("mysql", ms.GetDsn())
    if err != nil {
        return
    }

    db.SetMaxOpenConns(1e3)
    db.SetMaxIdleConns(4)
    db.SetConnMaxLifetime(time.Hour)

    err = db.Ping()
    if err != nil {
        return
    }

    mysql = ms

    return
}

// Close
func (ms *MySql) Close() {
    if db != nil {
        _ = db.Close()
    }
}
