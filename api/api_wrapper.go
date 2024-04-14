package api

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/api/sheets/v4"
	"strings"
)

type ApiWrapperImpl struct {
	srv *sheets.Service

	docID string
	sheet string

	logger *zap.Logger
}

func NewApiWrapper(srv *sheets.Service, docID string, sheet string, logger *zap.Logger) *ApiWrapperImpl {
	return &ApiWrapperImpl{
		srv:    srv,
		docID:  docID,
		sheet:  sheet,
		logger: logger,
	}
}

func (aw *ApiWrapperImpl) bindRange(range_ string) string {

	if strings.Contains(range_, "!") {
		panic("the range must be unbound from a sheet")
	}

	if aw.sheet == "" {
		// default
		return range_
	}
	return fmt.Sprintf("%s!%s", aw.sheet, range_)
}

func (aw *ApiWrapperImpl) GetSpreadsheet(ctx context.Context) (*sheets.Spreadsheet, error) {
	return aw.srv.Spreadsheets.Get(aw.docID).Context(ctx).Do()
}

func (aw *ApiWrapperImpl) GetRange(ctx context.Context, range_ string) (*sheets.ValueRange, error) {
	boundRange := aw.bindRange(range_)
	aw.logger.Debug("Attempting to get data from sheet", zap.String("range", boundRange))
	return aw.srv.Spreadsheets.Values.Get(aw.docID, boundRange).Context(ctx).Do()
}

func (aw *ApiWrapperImpl) BatchGetRanges(ctx context.Context, ranges []string) (*sheets.BatchGetValuesResponse, error) {
	boundRanges := make([]string, len(ranges))
	for i, r := range ranges {
		boundRanges[i] = aw.bindRange(r)
	}
	aw.logger.Debug("Attempting to batch get data from sheet", zap.Strings("ranges", boundRanges))
	return aw.srv.Spreadsheets.Values.BatchGet(aw.docID).Context(ctx).Ranges(boundRanges...).Do()
}

func (aw *ApiWrapperImpl) BatchUpdate(ctx context.Context, values []*sheets.ValueRange) (*sheets.BatchUpdateValuesResponse, error) {
	boundRanges := make([]string, len(values)) // only for debug purposes...
	boundVals := make([]*sheets.ValueRange, len(values))
	for i, val := range values {
		boundRange := aw.bindRange(val.Range)
		boundVals[i] = &sheets.ValueRange{
			MajorDimension:  val.MajorDimension,
			Range:           boundRange,
			Values:          val.Values,
			ServerResponse:  val.ServerResponse,
			ForceSendFields: val.ForceSendFields,
			NullFields:      val.NullFields,
		}
		boundRanges[i] = boundRange
	}

	aw.logger.Debug("Attempting to batch update sheet", zap.Strings("ranges", boundRanges))

	req := sheets.BatchUpdateValuesRequest{
		Data:                      boundVals,
		IncludeValuesInResponse:   false, // we will get them like animals
		ResponseValueRenderOption: "UNFORMATTED_VALUE",
		ValueInputOption:          "USER_ENTERED", // this is needed because otherwise the cell will be overwritten entirely
	}

	return aw.srv.Spreadsheets.Values.BatchUpdate(aw.docID, &req).Context(ctx).Do()
}
