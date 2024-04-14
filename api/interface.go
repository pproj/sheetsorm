package api

import (
	"context"
	"google.golang.org/api/sheets/v4"
)

// ApiWrapper is a simple wrapper around a single google sheet page
type ApiWrapper interface {
	GetSpreadsheet(ctx context.Context) (*sheets.Spreadsheet, error)
	GetRange(ctx context.Context, range_ string) (*sheets.ValueRange, error)
	BatchGetRanges(ctx context.Context, ranges []string) (*sheets.BatchGetValuesResponse, error)
	BatchUpdate(ctx context.Context, values []*sheets.ValueRange) (*sheets.BatchUpdateValuesResponse, error)
}
