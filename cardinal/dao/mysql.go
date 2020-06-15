package dao

import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "net/url"
    "strconv"
    "time"
)

type MySql struct {
    host     string
    port     uint
    username string
    password string
    database string
    prefix   string
    params   map[string]string
}

var (
    db    *sql.DB
    mysql *MySql
)

// NewMySql
// host, port, username, password, database, prefix, charset, collation, parseTime
func NewMySql(host string, port uint, username, password, database, prefix string, params map[string]string) *MySql {
    if mysql != nil {
        mysql.Close()
    }

    return &MySql{
        host:     host,
        port:     port,
        username: username,
        password: password,
        database: database,
        prefix:   prefix,
        params:   params,
    }
}

// GetDsn
func (ms *MySql) GetDsn() string {
    values := make(url.Values)
    for key, val := range ms.params {
        values.Set(key, val)
    }

    u := url.URL{
        User:     url.UserPassword(ms.username, ms.password),
        Host:     "tcp(" + ms.host + ":" + strconv.FormatInt(int64(ms.port), 10) + ")",
        Path:     ms.database,
        RawQuery: values.Encode(),
    }

    return u.String()[len("//"):]
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

func Db() *sql.DB {
    return db
}

func Mysql() *MySql {
    return mysql
}
