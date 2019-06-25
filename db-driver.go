package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"time"
)

// if dev mode, use dbQueryLog, or use dbQuerier.
// debug is true, use dbQueryLog
// debug is false, use sql.DB(*(sql.DB) implement all dbQuerier interface in sql.go file)
type dbQuerier interface {
	Prepare(query string) (*sql.Stmt, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

var _ dbQuerier = new(dbQueryLog)
type dbQueryLog struct {
	db dbQuerier
}

func (d *dbQueryLog) Prepare(query string) (*sql.Stmt, error) {
	stmt, err := d.db.Prepare(query)
	return stmt, err
}
func (d *dbQueryLog) Exec(query string, args ...interface{}) (sql.Result, error) {
	res, err := d.db.Exec(query, args...)
	return res, err
}
func (d *dbQueryLog) Query(query string, args ...interface{}) (*sql.Rows, error) {
	res, err := d.db.Query(query, args...)
	return res, err
}
func (d *dbQueryLog) QueryRow(query string, args ...interface{}) *sql.Row {
	res := d.db.QueryRow(query, args...)

	fmt.Println("orm_log-------------", query)
	return res
}

// driver.Driver
var _ driver.Driver = new(open)

type open struct{}

func (o *open) Open(dsn string) (driver.Conn, error) {
	fmt.Println("-----Open-----")
	return new(conn), nil
}

// driver.Conn
var _ driver.Conn = new(conn)
type conn struct{}
func (c *conn) Prepare(query string) (driver.Stmt, error) {
	return new(stmt), nil
}
func (c *conn) Close() error {
	return nil
}
func (c *conn) Begin() (driver.Tx, error) {
	return new(tx), nil
}

// driver.Pinger
var _ driver.Pinger = new(ping)
type ping struct{}
func (c *ping) Ping(ctx context.Context) error {
	return nil
}

// driver.Execer
var _ driver.Execer = new(exec)
type exec struct{}
func (c *exec) Exec(query string, args []driver.Value) (driver.Result, error) {
	return new(result), nil
}

// driver.Queryer
var _ driver.Queryer = new(query)
type query struct{}
func (c *query) Query(query string, args []driver.Value) (driver.Rows, error) {
	return new(rows), nil
}

// driver.Stmt
var _ driver.Stmt = new(stmt)
type stmt struct{}
func (s *stmt) Close() error {
	return nil
}
func (s *stmt) NumInput() int {
	return 0
}
func (s *stmt) Exec(args []driver.Value) (driver.Result, error) {
	return new(result), nil
}
func (s *stmt) Query(args []driver.Value) (driver.Rows, error) {
	return new(rows), nil
}

// driver.Tx
var _ driver.Tx = new(tx)
type tx struct{}
func (t *tx) Commit() error {
	return nil
}
func (t *tx) Rollback() error {
	return nil
}

// driver.Result
var _ driver.Result = new(result)
type result struct{}
func (r *result) LastInsertId() (int64, error) {
	return 0, nil
}
func (r *result) RowsAffected() (int64, error) {
	return 0, nil
}

// driver.Rows
var _ driver.Rows = new(rows)
type rows struct{}
func (r *rows) Columns() []string {
	return []string{}
}
func (r *rows) Close() error {
	return nil
}
func (r *rows) Next(dest []driver.Value) error {
	return nil
}

// sql包中已实现基础空数据类型
// NullInt64
// NullString
// NullFloat64
// NullBool
var _ driver.Valuer = new(NullTime)
var _ sql.Scanner = new(NullTime)
type NullTime struct {
	Time  time.Time
	Valid bool // Valid is true if Time is not NULL
}
// sql.Scanner
func (n *NullTime) Scan(value interface{}) error {
	n.Time, n.Valid = value.(time.Time)
	return nil
}
// driver.Valuer
func (n *NullTime) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Time, nil
}

func init() {
	sql.Register("default", &open{})
}

func main() {
	var (
		db         *sql.DB
		debug      bool
		err        error
		dbQuerier  dbQuerier
	)

	if db, err = sql.Open("default", "root@127.0.0.1:3306/mysql?demo"); err != nil {
		fmt.Println("init DB failed...")
	}

	debug = true

	if debug {
		dbQuerier = new(dbQueryLog)
	} else {
		dbQuerier = db
	}

	fmt.Println(reflect.TypeOf(dbQuerier))
}
