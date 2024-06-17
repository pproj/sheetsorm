package api

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/sheets/v4"
	"net/http"
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

func IsTooManyRequests(err error) bool {
	var apiErr *googleapi.Error
	return errors.As(err, &apiErr) && apiErr.Code == http.StatusTooManyRequests
}

func IsServiceUnavailable(err error) bool { // sometimes google api does this lol...
	var apiErr *googleapi.Error
	return errors.As(err, &apiErr) && apiErr.Code == http.StatusServiceUnavailable
}

func ShouldRetryAPICall(err error) bool {
	return IsTooManyRequests(err) || IsServiceUnavailable(err)
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
	var result *sheets.Spreadsheet
	return result, DoRetry(ctx, aw.logger, func() error {
		var err error
		result, err = aw.srv.Spreadsheets.Get(aw.docID).Context(ctx).Do()
		return err
	}, ShouldRetryAPICall)
}

func (aw *ApiWrapperImpl) GetRange(ctx context.Context, range_ string) (*sheets.ValueRange, error) {
	boundRange := aw.bindRange(range_)
	aw.logger.Debug("Attempting to get data from sheet", zap.String("range", boundRange))

	var result *sheets.ValueRange
	return result, DoRetry(ctx, aw.logger, func() error {
		var err error
		result, err = aw.srv.Spreadsheets.Values.Get(aw.docID, boundRange).Context(ctx).Do()
		return err
	}, ShouldRetryAPICall)
}

func (aw *ApiWrapperImpl) BatchGetRanges(ctx context.Context, ranges []string) (*sheets.BatchGetValuesResponse, error) {
	boundRanges := make([]string, len(ranges))
	for i, r := range ranges {
		boundRanges[i] = aw.bindRange(r)
	}
	aw.logger.Debug("Attempting to batch get data from sheet", zap.Strings("ranges", boundRanges))

	var result *sheets.BatchGetValuesResponse
	return result, DoRetry(ctx, aw.logger, func() error {
		var err error
		result, err = aw.srv.Spreadsheets.Values.BatchGet(aw.docID).Context(ctx).Ranges(boundRanges...).Do()
		return err
	}, ShouldRetryAPICall)
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

	var result *sheets.BatchUpdateValuesResponse
	return result, DoRetry(ctx, aw.logger, func() error {
		var err error
		result, err = aw.srv.Spreadsheets.Values.BatchUpdate(aw.docID, &req).Context(ctx).Do()
		return err
	}, ShouldRetryAPICall)
}
