package oceanbase_storage

import (
	"context"
	"database/sql"
	"strings"

	sql_based_storage "github.com/storage-lock/go-sql-based-storage"
	"github.com/storage-lock/go-storage"
	storage_lock "github.com/storage-lock/go-storage-lock"
)

// OceanBaseStorage 基于 OceanBase 数据库的存储引擎实现
// OceanBase 兼容 MySQL 协议，通过主键唯一约束和 WHERE 条件更新支持 CAS，
// 通过 UNIX_TIMESTAMP(NOW()) 提供数据库服务器时间
type OceanBaseStorage struct {
	*sql_based_storage.SqlBasedStorage
	options *OceanBaseStorageOptions
}

var _ storage.Storage = &OceanBaseStorage{}

func NewOceanBaseStorage(ctx context.Context, options *OceanBaseStorageOptions) (*OceanBaseStorage, error) {
	if err := options.Check(); err != nil {
		return nil, err
	}

	// OceanBase 兼容 MySQL，复用 SQL92 Provider
	baseStorageOption := sql_based_storage.NewSqlBasedStorageOptions().
		SetConnectionManager(options.ConnectionManager).
		SetSqlProvider(sql_based_storage.NewSql92Provider()).
		SetTableFullName(options.TableName)
	baseStorage, err := sql_based_storage.NewSqlBasedStorage(baseStorageOption)
	if err != nil {
		return nil, err
	}

	s := &OceanBaseStorage{
		SqlBasedStorage: baseStorage,
		options:         options,
	}

	if err := s.Init(ctx); err != nil {
		return nil, err
	}

	return s, nil
}

const StorageName = "oceanbase-storage"

func (x *OceanBaseStorage) GetName() string {
	return StorageName
}

// Capabilities 声明 OceanBase 存储支持的能力
// OceanBase 兼容 MySQL，通过主键唯一约束和 WHERE 条件更新天然支持 CAS，
// 通过 UNIX_TIMESTAMP(NOW()) 提供数据库服务器时间
func (x *OceanBaseStorage) Capabilities() []storage.StorageCapability {
	return []storage.StorageCapability{
		storage.CapabilityCAS,
		storage.CapabilityReliableTime,
	}
}

// CreateWithVersion OceanBase 兼容 MySQL，唯一约束冲突时返回 ErrLockAlreadyExists
func (x *OceanBaseStorage) CreateWithVersion(ctx context.Context, lockId string, version storage.Version, lockInformation *storage.LockInformation) error {
	err := x.SqlBasedStorage.CreateWithVersion(ctx, lockId, version, lockInformation)
	if err != nil {
		// OceanBase 兼容 MySQL 错误码: Error 1062 (23000)
		msg := err.Error()
		if strings.Contains(msg, "Error 1062 (23000)") {
			return storage_lock.ErrLockAlreadyExists
		}
	}
	return err
}

// OceanBaseStorageOptions OceanBase 存储选项
type OceanBaseStorageOptions struct {
	ConnectionManager storage.ConnectionManager[*sql.DB]
	TableName         string
}

func NewOceanBaseStorageOptions() *OceanBaseStorageOptions {
	return &OceanBaseStorageOptions{
		TableName: storage.DefaultStorageTableName,
	}
}

func (x *OceanBaseStorageOptions) SetConnectionManager(cm storage.ConnectionManager[*sql.DB]) *OceanBaseStorageOptions {
	x.ConnectionManager = cm
	return x
}

func (x *OceanBaseStorageOptions) SetTableName(name string) *OceanBaseStorageOptions {
	x.TableName = name
	return x
}

func (x *OceanBaseStorageOptions) Check() error {
	if x.ConnectionManager == nil {
		return ErrConnectionManagerNil
	}
	if x.TableName == "" {
		return ErrTableNameEmpty
	}
	return nil
}