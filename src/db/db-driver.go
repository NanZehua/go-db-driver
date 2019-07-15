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
// Driver是一个数据库驱动的接口，它定义了一个method： Open(name string)，这个方法返回一个数据库的Conn接口，返回的Conn只能用来进行一次goroutine的操作。
var _ driver.Driver = new(open)
type open struct{}
func (o *open) Open(dsn string) (driver.Conn, error) {
	fmt.Println("-----Open-----")
	return new(conn), nil
}

// driver.
// Conn是一个数据库连接的接口定义，他定义了一系列方法，这个Conn只能应用在一个goroutine里面
// Prepare函数返回与当前连接相关的执行Sql语句的准备状态，可以进行查询、删除等操作。
// Close函数关闭当前的连接，执行释放连接拥有的资源等清理工作。因为驱动实现了database/sql里面建议的conn pool，
//      所以你不用再去实现缓存conn之类的，这样会容易引起问题。
// Begin函数返回一个代表事务处理的Tx，通过它你可以进行查询,更新等操作，或者对事务进行回滚、递交。
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

// driver.Stmt
// Stmt是一种准备好的状态，和Conn相关联，而且只能应用于一个goroutine中，不能应用于多个goroutine。
// Close函数关闭当前的链接状态，但是如果当前正在执行query，query还是有效返回rows数据。
// NumInput函数返回当前预留参数的个数，当返回>=0时数据库驱动就会智能检查调用者的参数。当数据库驱动包不知道预留参数的时候，返回-1。
// Exec函数执行Prepare准备好的sql，传入参数执行update/insert等操作，返回Result数据
// Query函数执行Prepare准备好的sql，传入需要的参数执行select操作，返回Rows结果集
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
// 事务处理一般就两个过程，递交或者回滚。
var _ driver.Tx = new(tx)
type tx struct{}
func (t *tx) Commit() error {
	return nil
}
func (t *tx) Rollback() error {
	return nil
}

// driver.Pinger
var _ driver.Pinger = new(ping)
type ping struct{}
func (c *ping) Ping(ctx context.Context) error {
	return nil
}

// driver.Execer
// 这是一个Conn可选择实现的接口, 如果这个接口没有定义，那么在调用DB.Exec,就会首先调用Prepare返回Stmt，然后执行Stmt的Exec，然后关闭Stmt
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

// driver.Result
// 这个是执行Update/Insert等操作返回的结果接口定义
// LastInsertId函数返回由数据库执行插入操作得到的自增ID号。
// RowsAffected函数返回query操作影响的数据条目数。
var _ driver.Result = new(result)
type result struct{}
func (r *result) LastInsertId() (int64, error) {
	return 0, nil
}
func (r *result) RowsAffected() (int64, error) {
	return 0, nil
}

// driver.Rows
// Rows是执行查询返回的结果集接口定义
// Columns函数返回查询数据库表的字段信息，这个返回的slice和sql查询的字段一一对应，而不是返回整个表的所有字段。
// Close函数用来关闭Rows迭代器。
// Next函数用来返回下一条数据，把数据赋值给dest。
// dest里面的元素必须是driver.Value的值除了string，返回的数据里面所有的string都必须要转换成[]byte。如果最后没数据了，Next函数最后返回io.EOF。
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
	// 这个存在于database/sql的函数是用来注册数据库驱动的，当第三方开发者开发数据库驱动时，都会实现init函数，
	// 在init里面会调用这个Register(name string, driver driver.Driver)完成本驱动的注册。
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
