package oceanbase_storage

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/go-sql-driver/mysql"
	"github.com/storage-lock/go-storage"
)

// OceanBaseConnectionManager OceanBase 连接管理器
// OceanBase 兼容 MySQL 协议，使用 MySQL 驱动
type OceanBaseConnectionManager struct {
	DSN string

	mu sync.Mutex
	db *sql.DB

	open        func(driverName, dataSourceName string) (*sql.DB, error)
	ping        func(context.Context, *sql.DB) error
	initialized bool
}

var _ storage.ConnectionManager[*sql.DB] = &OceanBaseConnectionManager{}

// NewOceanBaseConnectionManagerFromDSN 从数据源名称创建连接管理器
// DSN 格式: username:password@tcp(host:port)/database
func NewOceanBaseConnectionManagerFromDSN(dsn string) *OceanBaseConnectionManager {
	return &OceanBaseConnectionManager{
		DSN: dsn,
	}
}

// NewOceanBaseConnectionManagerFromConfig 从配置参数创建连接管理器
func NewOceanBaseConnectionManagerFromConfig(host string, port int, database, username, password string) *OceanBaseConnectionManager {
	// OceanBase 兼容 MySQL，使用 MySQL DSN 格式
	cfg := mysql.Config{
		User:   username,
		Passwd: password,
		Net:    "tcp",
		Addr:   fmt.Sprintf("%s:%d", host, port),
		DBName: database,
	}
	return &OceanBaseConnectionManager{
		DSN: cfg.FormatDSN(),
	}
}

func (x *OceanBaseConnectionManager) Name() string {
	return "oceanbase-connection-manager"
}

func (x *OceanBaseConnectionManager) Take(ctx context.Context) (*sql.DB, error) {
	x.mu.Lock()
	defer x.mu.Unlock()

	if x.initialized {
		return x.db, nil
	}

	open := x.open
	if open == nil {
		open = sql.Open
	}
	db, err := open("mysql", x.DSN)
	if err != nil {
		return nil, err
	}

	// 配置连接池
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	// 测试连接
	ping := x.ping
	if ping == nil {
		ping = func(ctx context.Context, db *sql.DB) error { return db.PingContext(ctx) }
	}
	if err := ping(ctx, db); err != nil {
		db.Close()
		return nil, err
	}

	x.db = db
	x.initialized = true
	return x.db, nil
}

func (x *OceanBaseConnectionManager) Return(ctx context.Context, conn *sql.DB) error {
	return nil
}

func (x *OceanBaseConnectionManager) Shutdown(ctx context.Context) error {
	x.mu.Lock()
	defer x.mu.Unlock()

	if x.db != nil {
		return x.db.Close()
	}
	return nil
}
