package orm

// fork from gorm

import (
	"context"
	"database/sql"
	"sync"
)

// StmtQuery implemented by type *sql.DB / *sql.Tx
type StmtQuery interface {
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// TxBeginner
type TxBeginner interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

// StmtQueryBeginner
type StmtQueryBeginner interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (StmtQuery, error)
}

// Stmt
type Stmt struct {
	*sql.Stmt
	Transaction bool
}

// StmtDB
type StmtDB struct {
	PreparedSQL []string
	Stmts       map[string]Stmt
	Mux         *sync.RWMutex
	StmtQuery
}

// NewStmtDB
func NewStmtDB(db StmtQuery) *StmtDB {
	return &StmtDB{
		PreparedSQL: make([]string, 0, 100),
		Stmts:       map[string]Stmt{},
		Mux:         &sync.RWMutex{},
		StmtQuery:   db,
	}
}

// Close
func (db *StmtDB) Close() {
	db.Mux.Lock()
	for _, query := range db.PreparedSQL {
		if stmt, ok := db.Stmts[query]; ok {
			delete(db.Stmts, query)
			_ = stmt.Close()
		}
	}
	db.Mux.Unlock()
}

func (db *StmtDB) prepare(ctx context.Context, conn StmtQuery, isTransaction bool, query string) (Stmt, error) {
	db.Mux.RLock()
	if stmt, ok := db.Stmts[query]; ok && (!stmt.Transaction || isTransaction) {
		db.Mux.RUnlock()

		return stmt, nil
	}
	db.Mux.RUnlock()

	db.Mux.Lock()
	if stmt, ok := db.Stmts[query]; ok && (!stmt.Transaction || isTransaction) {
		db.Mux.Unlock()

		return stmt, nil
	} else if ok {
		_ = stmt.Close()
	}

	stmt, err := conn.PrepareContext(ctx, query)
	if err == nil {
		db.PreparedSQL = append(db.PreparedSQL, query)
		db.Stmts[query] = Stmt{Stmt: stmt, Transaction: isTransaction}
	}
	db.Mux.Unlock()

	return db.Stmts[query], err
}

// BeginTx
func (db *StmtDB) BeginTx(ctx context.Context, opt *sql.TxOptions) (StmtQuery, error) {
	if beginner, ok := db.StmtQuery.(TxBeginner); ok {
		tx, err := beginner.BeginTx(ctx, opt)

		return &StmtTX{Tx: tx, StmtDB: db}, err
	}

	return nil, ErrInvalidTransaction
}

// ExecContext
func (db *StmtDB) ExecContext(ctx context.Context, query string, args ...interface{}) (result sql.Result, err error) {
	var stmt Stmt
	stmt, err = db.prepare(ctx, db.StmtQuery, false, query)
	if err != nil {
		return
	}

	result, err = stmt.ExecContext(ctx, args...)
	if err != nil {
		db.Mux.Lock()
		_ = stmt.Close()
		delete(db.Stmts, query)
		db.Mux.Unlock()
	}

	return
}

// QueryContext
func (db *StmtDB) QueryContext(ctx context.Context, query string, args ...interface{}) (rows *sql.Rows, err error) {
	var stmt Stmt
	stmt, err = db.prepare(ctx, db.StmtQuery, false, query)
	if err != nil {
		return
	}

	rows, err = stmt.QueryContext(ctx, args...)
	if err != nil {
		db.Mux.Lock()
		_ = stmt.Close()
		delete(db.Stmts, query)
		db.Mux.Unlock()
	}

	return
}

// QueryRowContext
func (db *StmtDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) (row *sql.Row) {
	stmt, err := db.prepare(ctx, db.StmtQuery, false, query)
	if err != nil {
		row = &sql.Row{}
	} else {
		row = stmt.QueryRowContext(ctx, args...)
	}

	return
}

// #

// StmtTX
type StmtTX struct {
	*sql.Tx
	StmtDB *StmtDB
}

// Commit
func (tx *StmtTX) Commit() error {
	if tx.Tx != nil {
		return tx.Tx.Commit()
	}

	return ErrInvalidTransaction
}

// Rollback
func (tx *StmtTX) Rollback() error {
	if tx.Tx != nil {
		return tx.Tx.Rollback()
	}

	return ErrInvalidTransaction
}

// ExecContext
func (tx *StmtTX) ExecContext(ctx context.Context, query string, args ...interface{}) (result sql.Result, err error) {
	var stmt Stmt
	stmt, err = tx.StmtDB.prepare(ctx, tx.Tx, true, query)
	if err != nil {
		return
	}

	result, err = tx.Tx.StmtContext(ctx, stmt.Stmt).ExecContext(ctx, args...)
	if err != nil {
		tx.StmtDB.Mux.Lock()
		_ = stmt.Close()
		delete(tx.StmtDB.Stmts, query)
		tx.StmtDB.Mux.Unlock()
	}

	return
}

// QueryContext
func (tx *StmtTX) QueryContext(ctx context.Context, query string, args ...interface{}) (rows *sql.Rows, err error) {
	var stmt Stmt
	stmt, err = tx.StmtDB.prepare(ctx, tx.Tx, true, query)
	if err != nil {
		return
	}

	rows, err = tx.Tx.Stmt(stmt.Stmt).QueryContext(ctx, args...)
	if err != nil {
		tx.StmtDB.Mux.Lock()
		_ = stmt.Close()
		delete(tx.StmtDB.Stmts, query)
		tx.StmtDB.Mux.Unlock()
	}

	return
}

// QueryRowContext
func (tx *StmtTX) QueryRowContext(ctx context.Context, query string, args ...interface{}) (row *sql.Row) {
	stmt, err := tx.StmtDB.prepare(ctx, tx.Tx, true, query)
	if err != nil {
		row = &sql.Row{}
	} else {
		row = tx.Tx.StmtContext(ctx, stmt.Stmt).QueryRowContext(ctx, args...)
	}

	return
}
