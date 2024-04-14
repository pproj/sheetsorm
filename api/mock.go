package api

import (
	"context"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/sheets/v4"
)

type MockApiWrapper struct {
	mock.Mock
}

func (m *MockApiWrapper) GetSpreadsheet(ctx context.Context) (*sheets.Spreadsheet, error) {
	args := m.Called(ctx)
	return args.Get(0).(*sheets.Spreadsheet), args.Error(1)
}

func (m *MockApiWrapper) GetRange(ctx context.Context, range_ string) (*sheets.ValueRange, error) {
	args := m.Called(ctx, range_)
	return args.Get(0).(*sheets.ValueRange), args.Error(1)
}

func (m *MockApiWrapper) BatchGetRanges(ctx context.Context, ranges []string) (*sheets.BatchGetValuesResponse, error) {
	args := m.Called(ctx, ranges)
	return args.Get(0).(*sheets.BatchGetValuesResponse), args.Error(1)
}

func (m *MockApiWrapper) BatchUpdate(ctx context.Context, values []*sheets.ValueRange) (*sheets.BatchUpdateValuesResponse, error) {
	args := m.Called(ctx, values)
	return args.Get(0).(*sheets.BatchUpdateValuesResponse), args.Error(1)
}
