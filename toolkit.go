package sheetsorm

import (
	"context"
	"fmt"
	"github.com/pproj/sheetsorm/api"
	"github.com/pproj/sheetsorm/column"
	"github.com/pproj/sheetsorm/errors"
	"go.uber.org/zap"
	"google.golang.org/api/sheets/v4"
	"slices"
)

// sheetsToolkit is a toolkit used internally to work with sheets. A toolkit is bound to a specific type of record in a specific sheet
// The goal of toolkit is to separate record manipulation logic for sheet manipulation logic, therefore it uses only rowData maps for its functionality
type sheetsToolkit struct {
	aw api.ApiWrapper

	// data
	skipRows int
	firstCol string
	lastCol  string
	colShift int
	cols     []string
	uidCol   string

	logger *zap.Logger
}

func newToolkit(aw api.ApiWrapper, cols column.Cols, uidCol string, skipRows int, logger *zap.Logger) (*sheetsToolkit, error) {
	if !cols.Contains(uidCol) {
		return nil, errors.InvalidUIDCol
	}
	return &sheetsToolkit{
		aw: aw,

		skipRows: skipRows,
		firstCol: cols.First(), // saved for faster access
		lastCol:  cols.Last(),
		colShift: cols.Shift(),
		cols:     cols,
		uidCol:   uidCol,

		logger: logger,
	}, nil
}

// uidsToRowNums does only a single API call, and can resolve multiple UIDs to row numbers.
func (st *sheetsToolkit) uidsToRowNums(ctx context.Context, uids []string) ([]int, error) {
	if len(uids) == 0 {
		return nil, nil // nothing to do
	}
	for _, uid := range uids {
		if uid == "" {
			return nil, errors.ErrEmptyUID
		}
	}

	// TODO: add caching here
	uidColRange := fmt.Sprintf("%[1]s%[2]d:%[1]s", st.uidCol, st.skipRows+1)
	vals, err := st.aw.GetRange(ctx, uidColRange)
	if err != nil {
		st.logger.Error("Failed to get uid column", zap.String("range", uidColRange), zap.Error(err))
		return nil, err
	}

	rowNums := make([]int, len(uids))

	for rowI, row := range vals.Values { // the header is already skipped by the request
		rowNum := rowI + 1 + st.skipRows // zero index correction plus skipped rows

		if len(row) == 0 {
			// this happens for empty rows it seems...
			continue
		}

		rowUid := row[0]
		if rowUid == "" {
			// We do not consider empty uids as valid
			// it would be hard to distinguish in sheets as well
			continue
		}

		// we have rowNum - rowUid pairs here

		for i, uid := range uids {
			if rowUid == uid {
				st.logger.Debug("Translated uid to row num", zap.String("uid", uid), zap.Int("rowNum", rowNum))
				rowNums[i] = rowNum
			}
			if ctx.Err() != nil { // context cancelled
				return nil, ctx.Err()
			}
		}

	}

	if slices.Contains(rowNums, 0) {
		// 0 is not a valid row number in sheets, so if we found one, it means that one record could not be paired
		return nil, errors.ErrRecordNotFound
	}

	return rowNums, nil
}

// uidToRowNum resolves a single uid to a row number
func (st *sheetsToolkit) uidToRowNum(ctx context.Context, uid string) (int, error) {
	nums, err := st.uidsToRowNums(ctx, []string{uid})
	if err != nil {
		return 0, err
	}
	if len(nums) != 1 {
		return 0, fmt.Errorf("wtf")
	}
	return nums[0], nil
}

func (st *sheetsToolkit) translateFullRowToMap(row []interface{}) map[string]string {
	resultSet := make(map[string]string, len(st.cols))
	for _, col := range st.cols {
		idx := column.ColIndex(col)

		val := row[idx-st.colShift]
		resultSet[col] = val.(string)
	}
	return resultSet
}

func (st *sheetsToolkit) getDataMapFromRowNum(ctx context.Context, rowNum int) (map[string]string, error) {
	// TODO: maybe add caching here??
	rangeStr := fmt.Sprintf("%[1]s%[3]d:%[2]s%[3]d", st.firstCol, st.lastCol, rowNum)

	vals, err := st.aw.GetRange(ctx, rangeStr)
	if err != nil {
		st.logger.Error("Failed to get row from sheet", zap.Error(err), zap.String("range", rangeStr))
		return nil, err
	}

	row := vals.Values[0]
	return st.translateFullRowToMap(row), nil
}

func (st *sheetsToolkit) getDataMapsFromRowNums(ctx context.Context, rowNums []int) ([]map[string]string, error) {

	ranges := make([]string, len(rowNums))

	for i, rowNum := range rowNums {
		ranges[i] = fmt.Sprintf("%[1]s%[3]d:%[2]s%[3]d", st.firstCol, st.lastCol, rowNum)
	}

	vals, err := st.aw.BatchGetRanges(ctx, ranges)
	if err != nil {
		st.logger.Error("Failed to get rows from sheet", zap.Error(err), zap.Strings("ranges", ranges))
		return nil, err
	}

	out := make([]map[string]string, len(rowNums))
	for i, row := range vals.ValueRanges { // yay: The order of the ValueRanges is the same as the order of the requested ranges
		out[i] = st.translateFullRowToMap(row.Values[0])
	}

	return out, nil
}

func (st *sheetsToolkit) getRecordData(ctx context.Context, uid string) (map[string]string, error) {
	rowNum, err := st.uidToRowNum(ctx, uid)
	if err != nil {
		st.logger.Error("Failed to translate UID to row number", zap.Error(err), zap.String("uid", uid))
		return nil, err
	}
	var recordDataMap map[string]string
	recordDataMap, err = st.getDataMapFromRowNum(ctx, rowNum)
	if err != nil {
		st.logger.Error("Failed to get data for row", zap.Error(err), zap.String("uid", uid), zap.Int("rowNum", rowNum))
		return nil, err
	}

	uidOut := recordDataMap[st.uidCol]
	if uidOut != uid {
		// TODO: Cache is inconsistent, must be dropped, and retried
		return nil, fmt.Errorf("wtf")
	}

	return recordDataMap, nil
}

func (st *sheetsToolkit) getAllRecordsData(ctx context.Context) (<-chan map[string]string, error) {
	rangeStr := fmt.Sprintf("%s%d:%s", st.firstCol, st.skipRows, st.lastCol)

	vals, err := st.aw.GetRange(ctx, rangeStr)
	if err != nil {
		st.logger.Error("Failed to get rows from sheet", zap.Error(err), zap.String("rangeStr", rangeStr))
		return nil, err
	}

	outChan := make(chan map[string]string)

	go func() {
		defer close(outChan)
		var c int
		for _, val := range vals.Values {
			if ctx.Err() != nil { // context cancelled
				return
			}

			dataMap := st.translateFullRowToMap(val)
			uid := dataMap[st.uidCol]

			if uid != "" {
				st.logger.Debug("Passing a new row", zap.String("uid", uid))
				c++
				outChan <- dataMap
			}
		}
		st.logger.Debug("Done, passed all valid-looking rows.", zap.Int("count", c))
	}()

	return outChan, nil
}

// translateRowDataToUpdateRanges iterates over st.cols, and it tries to group together updates in batches
func (st *sheetsToolkit) translateRowDataToUpdateRanges(rowNum int, row map[string]string) []*sheets.ValueRange {

	valRanges := make([]*sheets.ValueRange, 0)

	var startCol string
	var prevCol string
	var curVals []interface{} // stupid google sheets api won't accept string here

	for _, col := range st.cols {
		val, ok := row[col]
		if !ok { // end of sequence?
			if startCol == "" {
				continue // a range was not started, so don't care
			}
			// we have a range started here
			var newRange string
			if startCol == prevCol {
				newRange = fmt.Sprintf("%s%d", prevCol, rowNum)
			} else {
				newRange = fmt.Sprintf("%[1]s%[3]d:%[2]s%[3]d", startCol, prevCol, rowNum)
			}

			valRanges = append(valRanges, &sheets.ValueRange{
				MajorDimension: "ROWS",
				Range:          newRange,
				Values:         [][]interface{}{curVals},
			})
			curVals = nil
			startCol = ""
		} else { // found a val
			if startCol == "" { // but there was no active sequence, so let's start one
				startCol = col
				curVals = make([]interface{}, 0)
			}
			curVals = append(curVals, val)
		}
		prevCol = col
	}
	if startCol != "" { // one range left
		var newRange string
		if startCol == prevCol {
			newRange = fmt.Sprintf("%s%d", prevCol, rowNum)
		} else {
			newRange = fmt.Sprintf("%[1]s%[3]d:%[2]s%[3]d", startCol, prevCol, rowNum)
		}

		valRanges = append(valRanges, &sheets.ValueRange{
			MajorDimension: "ROWS",
			Range:          newRange,
			Values:         [][]interface{}{curVals},
		})
	}

	st.logger.Debug("Grouped updates to range spans", zap.Int("len(valRanges)", len(valRanges)), zap.Int("len(row)", len(row)))

	return valRanges
}

// updateRecords the uids should be a list of uids and the records should be the corresponding records, but omitting read-only fields, and fields don't wanted to be updated
// this function does not modify data in-place, instead it returns the new data, in the same order it got it... TODO: consider using channels (some other funcs may need to be changed as well)
// make sure there are no duplicates in the uids,...
func (st *sheetsToolkit) updateRecords(ctx context.Context, uids []string, records []map[string]string) ([]map[string]string, error) {

	// Resolve all uids to row numbers using a single API call
	rowNums, err := st.uidsToRowNums(ctx, uids)
	if err != nil {
		st.logger.Error("Failure while resolving uids to row nums", zap.Error(err))
		return nil, err
	}

	// Then group all updates as necessary
	valRanges := make([]*sheets.ValueRange, 0)
	for i, r := range records {
		if len(r) == 0 {
			// skip empty maps
			continue
		}

		valRangesForRecord := st.translateRowDataToUpdateRanges(rowNums[i], r)
		valRanges = append(valRanges, valRangesForRecord...)

		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
	}
	st.logger.Debug("Translated batch update to range updates", zap.Int("len(valRanges)", len(valRanges)))

	if len(valRanges) == 0 {
		// nothing to do
		st.logger.Debug("nothing to update...")
		return nil, nil
	}

	var resp *sheets.BatchUpdateValuesResponse
	resp, err = st.aw.BatchUpdate(ctx, valRanges)
	if err != nil {
		st.logger.Error("Batch update failed", zap.Error(err))
		return nil, err
	}

	st.logger.Debug("Batch update completed, reading back data...",
		zap.Int64("TotalUpdatedCells", resp.TotalUpdatedCells),
		zap.Int64("TotalUpdatedRows", resp.TotalUpdatedRows),
		zap.Int64("TotalUpdatedColumns", resp.TotalUpdatedColumns),
	)

	return st.getDataMapsFromRowNums(ctx, rowNums)
}
