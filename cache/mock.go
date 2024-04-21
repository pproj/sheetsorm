package cache

import "github.com/stretchr/testify/mock"

// MockRowUIDCache is a mock implementation of the RowUIDCache interface.
type MockRowUIDCache struct {
	mock.Mock
}

// CacheUID mocks the CacheUID method of the RowUIDCache interface.
func (m *MockRowUIDCache) CacheUID(uid string, rowNum int) {
	m.Called(uid, rowNum)
}

// GetRowNumByUID mocks the GetRowNumByUID method of the RowUIDCache interface.
func (m *MockRowUIDCache) GetRowNumByUID(uid string) (int, bool) {
	args := m.Called(uid)
	return args.Int(0), args.Bool(1)
}

// InvalidateUID mocks the InvalidateUID method of the RowUIDCache interface.
func (m *MockRowUIDCache) InvalidateUID(uid string) {
	m.Called(uid)
}

// MockRowCache is a mock implementation of the RowCache interface.
type MockRowCache struct {
	mock.Mock
}

// CacheRow mocks the CacheRow method of the RowCache interface.
func (m *MockRowCache) CacheRow(rowNum int, rowData map[string]string) {
	m.Called(rowNum, rowData)
}

// GetRow mocks the GetRow method of the RowCache interface.
func (m *MockRowCache) GetRow(rowNum int) (map[string]string, bool) {
	args := m.Called(rowNum)
	return args.Get(0).(map[string]string), args.Bool(1)
}

// InvalidateRow mocks the InvalidateRow method of the RowCache interface.
func (m *MockRowCache) InvalidateRow(rowNum int) {
	m.Called(rowNum)
}
