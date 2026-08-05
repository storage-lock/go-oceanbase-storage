package oceanbase_storage

import (
	"context"
	"os"
	"testing"

	storage_test_helper "github.com/storage-lock/go-storage-test-helper"
	"github.com/stretchr/testify/assert"
)

func TestOceanBaseStorage(t *testing.T) {
	dsn := os.Getenv("STORAGE_LOCK_OCEANBASE_DSN")
	if dsn == "" {
		t.Skip("STORAGE_LOCK_OCEANBASE_DSN not set, skipping OceanBase test")
	}

	ctx := context.Background()
	cm := NewOceanBaseConnectionManagerFromDSN(dsn)
	s, err := NewOceanBaseStorage(ctx, &OceanBaseStorageOptions{
		ConnectionManager: cm,
		TableName:         storage_test_helper.TestTableName,
	})
	assert.Nil(t, err)
	storage_test_helper.TestStorage(t, s)
}