package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryTransaction_BasicOperations(t *testing.T) {
	backend := NewMemoryBackend()
	defer backend.Close()

	err := backend.Initialize()
	require.NoError(t, err)

	// 在后端设置一些初始数据
	err = backend.Set("existing_key", "existing_value", 0)
	require.NoError(t, err)

	// 创建事务
	tx := NewMemoryTransaction(backend)
	require.NotNil(t, tx)
	require.NotEmpty(t, tx.GetID())
	require.True(t, tx.IsActive())

	// 在事务中读取现有数据
	value, err := tx.Get("existing_key")
	require.NoError(t, err)
	assert.Equal(t, "existing_value", value)

	// 在事务中设置新值
	err = tx.Set("tx_key", "tx_value")
	require.NoError(t, err)

	// 在事务中读取新设置的值
	value, err = tx.Get("tx_key")
	require.NoError(t, err)
	assert.Equal(t, "tx_value", value)

	// 后端应该还没有这个值
	_, err = backend.Get("tx_key")
	assert.Error(t, err)

	// 提交事务
	err = tx.Commit()
	require.NoError(t, err)

	// 现在后端应该有这个值
	value, err = backend.Get("tx_key")
	require.NoError(t, err)
	assert.Equal(t, "tx_value", value)

	// 事务应该不再活跃
	assert.False(t, tx.IsActive())
	assert.Equal(t, TransactionStateCommitted, tx.GetState())
}

func TestMemoryTransaction_Rollback(t *testing.T) {
	backend := NewMemoryBackend()
	defer backend.Close()

	err := backend.Initialize()
	require.NoError(t, err)

	// 在后端设置初始数据
	err = backend.Set("key1", "original_value", 0)
	require.NoError(t, err)

	// 创建事务
	tx := NewMemoryTransaction(backend)
	require.NotNil(t, tx)

	// 在事务中修改值
	err = tx.Set("key1", "modified_value")
	require.NoError(t, err)

	// 在事务中添加新值
	err = tx.Set("key2", "new_value")
	require.NoError(t, err)

	// 验证事务中的值
	value, err := tx.Get("key1")
	require.NoError(t, err)
	assert.Equal(t, "modified_value", value)

	value, err = tx.Get("key2")
	require.NoError(t, err)
	assert.Equal(t, "new_value", value)

	// 回滚事务
	err = tx.Rollback()
	require.NoError(t, err)

	// 验证后端数据未被修改
	value, err = backend.Get("key1")
	require.NoError(t, err)
	assert.Equal(t, "original_value", value)

	_, err = backend.Get("key2")
	assert.Error(t, err)

	// 事务应该不再活跃
	assert.False(t, tx.IsActive())
	assert.Equal(t, TransactionStateRolledBack, tx.GetState())
}

func TestMemoryTransaction_Delete(t *testing.T) {
	backend := NewMemoryBackend()
	defer backend.Close()

	err := backend.Initialize()
	require.NoError(t, err)

	// 在后端设置初始数据
	err = backend.Set("key1", "value1", 0)
	require.NoError(t, err)
	err = backend.Set("key2", "value2", 0)
	require.NoError(t, err)

	// 创建事务
	tx := NewMemoryTransaction(backend)
	require.NotNil(t, tx)

	// 在事务中删除一个键
	err = tx.Delete("key1")
	require.NoError(t, err)

	// 在事务中应该找不到被删除的键
	_, err = tx.Get("key1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key not found")

	// 但是后端中应该还存在
	value, err := backend.Get("key1")
	require.NoError(t, err)
	assert.Equal(t, "value1", value)

	// 提交事务
	err = tx.Commit()
	require.NoError(t, err)

	// 现在后端中也应该被删除了
	_, err = backend.Get("key1")
	assert.Error(t, err)

	// key2应该还存在
	value, err = backend.Get("key2")
	require.NoError(t, err)
	assert.Equal(t, "value2", value)
}

func TestMemoryTransaction_OperationOrder(t *testing.T) {
	backend := NewMemoryBackend()
	defer backend.Close()

	err := backend.Initialize()
	require.NoError(t, err)

	// 创建事务
	tx := NewMemoryTransaction(backend)
	require.NotNil(t, tx)

	// 执行一系列操作
	err = tx.Set("key1", "value1")
	require.NoError(t, err)

	err = tx.Set("key1", "value2") // 覆盖
	require.NoError(t, err)

	err = tx.Delete("key1") // 删除
	require.NoError(t, err)

	err = tx.Set("key1", "value3") // 重新设置
	require.NoError(t, err)

	// 最终值应该是最后设置的值
	value, err := tx.Get("key1")
	require.NoError(t, err)
	assert.Equal(t, "value3", value)

	// 检查操作数量
	assert.Equal(t, 4, tx.GetOperationCount())

	// 提交事务
	err = tx.Commit()
	require.NoError(t, err)

	// 验证后端的最终值
	value, err = backend.Get("key1")
	require.NoError(t, err)
	assert.Equal(t, "value3", value)
}

func TestMemoryTransaction_InactiveOperations(t *testing.T) {
	backend := NewMemoryBackend()
	defer backend.Close()

	err := backend.Initialize()
	require.NoError(t, err)

	// 创建事务
	tx := NewMemoryTransaction(backend)
	require.NotNil(t, tx)

	// 提交事务
	err = tx.Commit()
	require.NoError(t, err)

	// 在非活跃事务上执行操作应该失败
	err = tx.Set("key1", "value1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transaction is not active")

	_, err = tx.Get("key1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transaction is not active")

	err = tx.Delete("key1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transaction is not active")

	// 重复提交应该失败
	err = tx.Commit()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transaction is not active")
}

func TestMemoryTransaction_Timeout(t *testing.T) {
	backend := NewMemoryBackend()
	defer backend.Close()

	err := backend.Initialize()
	require.NoError(t, err)

	// 创建事务
	tx := NewMemoryTransaction(backend)
	require.NotNil(t, tx)

	// 设置很短的超时时间
	tx.SetTimeout(50 * time.Millisecond)
	assert.Equal(t, 50*time.Millisecond, tx.GetTimeout())

	// 等待超时
	time.Sleep(100 * time.Millisecond)

	// 事务应该不再活跃
	assert.False(t, tx.IsActive())
	assert.Equal(t, TransactionStateAborted, tx.GetState())

	// 在超时的事务上执行操作应该失败
	err = tx.Set("key1", "value1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transaction is not active")
}

func TestMemoryTransaction_GetOperations(t *testing.T) {
	backend := NewMemoryBackend()
	defer backend.Close()

	err := backend.Initialize()
	require.NoError(t, err)

	// 创建事务
	tx := NewMemoryTransaction(backend)
	require.NotNil(t, tx)

	// 执行一些操作
	err = tx.Set("key1", "value1")
	require.NoError(t, err)

	err = tx.Set("key2", "value2")
	require.NoError(t, err)

	err = tx.Delete("key1")
	require.NoError(t, err)

	// 获取操作列表
	operations := tx.GetOperations()
	assert.Len(t, operations, 3)

	// 验证操作顺序和内容
	assert.Equal(t, TransactionOpSet, operations[0].Operation)
	assert.Equal(t, "key1", operations[0].Key)
	assert.Equal(t, "value1", operations[0].Value)

	assert.Equal(t, TransactionOpSet, operations[1].Operation)
	assert.Equal(t, "key2", operations[1].Key)
	assert.Equal(t, "value2", operations[1].Value)

	assert.Equal(t, TransactionOpDelete, operations[2].Operation)
	assert.Equal(t, "key1", operations[2].Key)
	assert.Nil(t, operations[2].Value)

	// 验证时间戳
	for _, op := range operations {
		assert.False(t, op.Timestamp.IsZero())
	}
}

func TestMemoryTransaction_CommitFailure(t *testing.T) {
	// 创建一个会失败的后端（通过关闭它）
	backend := NewMemoryBackend()
	err := backend.Initialize()
	require.NoError(t, err)

	// 创建事务
	tx := NewMemoryTransaction(backend)
	require.NotNil(t, tx)

	// 在事务中设置值
	err = tx.Set("key1", "value1")
	require.NoError(t, err)

	// 关闭后端，使提交失败
	err = backend.Close()
	require.NoError(t, err)

	// 提交应该失败
	err = tx.Commit()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to commit")

	// 事务状态应该是中止
	assert.Equal(t, TransactionStateAborted, tx.GetState())
}

func TestTransactionState_String(t *testing.T) {
	tests := []struct {
		state    TransactionState
		expected string
	}{
		{TransactionStateActive, "active"},
		{TransactionStateCommitted, "committed"},
		{TransactionStateRolledBack, "rolled_back"},
		{TransactionStateAborted, "aborted"},
		{TransactionState(999), "unknown"},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.state.String())
	}
}

func TestMemoryTransaction_CreatedAt(t *testing.T) {
	backend := NewMemoryBackend()
	defer backend.Close()

	err := backend.Initialize()
	require.NoError(t, err)

	before := time.Now()
	tx := NewMemoryTransaction(backend)
	after := time.Now()

	createdAt := tx.GetCreatedAt()
	assert.True(t, createdAt.After(before) || createdAt.Equal(before))
	assert.True(t, createdAt.Before(after) || createdAt.Equal(after))
}