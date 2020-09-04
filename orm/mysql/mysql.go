package mysql

import (
    "database/sql"
    "fmt"
    _ "github.com/go-sql-driver/mysql"
    "net/url"
    "reflect"
    "strconv"
    "time"
)

type MySql struct {
    host     string
    port     int
    username string
    password string
    database string
    prefix   string
    options  map[string]string
}

// NewMySql
func NewMySql(host string, port int, username, password, database, prefix string, options map[string]string) *MySql {
    return &MySql{
        host:     host,
        port:     port,
        username: username,
        password: password,
        database: database,
        prefix:   prefix,
        options:  options,
    }
}

// GetDriverName
func (m *MySql) GetDriverName() string {
    return "mysql"
}

// GetDsn
func (m *MySql) GetDsn() string {
    values := make(url.Values)
    for key, val := range m.options {
        values.Set(key, val)
    }

    u := url.URL{
        User:     url.UserPassword(m.username, m.password),
        Host:     "tcp(" + m.host + ":" + strconv.FormatInt(int64(m.port), 10) + ")",
        Path:     m.database,
        RawQuery: values.Encode(),
    }

    return u.String()[len("//"):]
}

// Config
func (m *MySql) Config(db *sql.DB) {
    db.SetMaxOpenConns(1e3)
    db.SetMaxIdleConns(4)
    db.SetConnMaxLifetime(time.Hour)
}

// GetTablePrefix
func (m *MySql) GetTablePrefix() string {
    return m.prefix
}

// GetColumnSqlType
func (m *MySql) GetColumnSqlType(typ reflect.Value) string {
    switch typ.Kind() {
    case reflect.Bool:
        return "BOOL"
    case reflect.Int8:
        return "TINYINT"
    case reflect.Int16:
        return "SMALLINT"
    case reflect.Int, reflect.Int32:
        return "INTEGER"
    case reflect.Int64:
        return "BIGINT"
    case reflect.Uint8:
        return "TINYINT UNSIGNED"
    case reflect.Uint16:
        return "SMALLINT UNSIGNED"
    case reflect.Uint, reflect.Uint32:
        return "INTEGER UNSIGNED"
    case reflect.Uint64:
        return "BIGINT UNSIGNED"
    case reflect.Float32, reflect.Float64:
        return "REAL"
    case reflect.String:
        return "TEXT"
    case reflect.Array, reflect.Slice:
        return "BLOB"
    case reflect.Struct:
        switch typ.Interface().(type) {
        case sql.NullBool:
            return "BOOL"
        case sql.NullInt32:
            return "INTEGER"
        case sql.NullInt64:
            return "BIGINT"
        case sql.NullFloat64:
            return "REAL"
        case sql.NullString:
            return "TEXT"
        case time.Time:
            return "DATETIME"
        }
    }

    panic(fmt.Sprintf("orm: unsupported type %s in model %s", typ.Kind(), typ.Type().Name()))
}

// TableExistSql
func (m *MySql) GetExistTableSql(table string) (preSql string, params []interface{}) {
    preSql = "SELECT COUNT(*) AS `exist` FROM `information_schema`.`TABLES` WHERE `TABLE_SCHEMA` = ? AND `TABLE_NAME` = ?;"
    params = []interface{}{m.database, table}

    return
}
