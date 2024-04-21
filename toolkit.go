package sheetsorm

import (
	"context"
	"fmt"
	"github.com/pproj/sheetsorm/api"
	"github.com/pproj/sheetsorm/cache"
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

	logger   *zap.Logger
	uidCache cache.RowUIDCache
	rowCache cache.RowCache
}

func newToolkit(
	aw api.ApiWrapper,
	cols column.Cols,
	uidCol string,
	skipRows int,
	logger *zap.Logger,
	uidCache cache.RowUIDCache,
	rowCache cache.RowCache,
) (*sheetsToolkit, error) {

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

		logger:   logger,
		uidCache: uidCache,
		rowCache: rowCache,
	}, nil
}

// uidsToRowNums does only a single API call, and can resolve multiple UIDs to row numbers.
// it does store received data in cache, but does not do lookups to it
// (the reason for that is that we want to explicit control over when we want data from cache)
func (st *sheetsToolkit) uidsToRowNums(ctx context.Context, uids []string) ([]int, error) {
	if len(uids) == 0 {
		return nil, nil // nothing to do
	}
	for _, uid := range uids {
		if uid == "" {
			return nil, errors.ErrEmptyUID
		}
	}

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

		rowUid := row[0].(string)
		if rowUid == "" {
			// We do not consider empty uids as valid
			// it would be hard to distinguish in sheets as well
			continue
		}

		// we have rowNum - rowUid pairs here, let's greedy cache them...
		st.uidCache.CacheUID(rowUid, rowNum)

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
// it does store received data in cache, but does not do lookups to it
// (the reason for that is that we want to explicit control over when we want data from cache)
func (st *sheetsToolkit) uidToRowNum(ctx context.Context, uid string) (int, error) {
	nums, err := st.uidsToRowNums(ctx, []string{uid}) // < cache saving implemented here
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

// getDataMapFromRowNum stores received data in cache, but does not do lookups to it
// (the reason for that is that we want to explicit control over when we want data from cache)
func (st *sheetsToolkit) getDataMapFromRowNum(ctx context.Context, rowNum int) (map[string]string, error) {
	rangeStr := fmt.Sprintf("%[1]s%[3]d:%[2]s%[3]d", st.firstCol, st.lastCol, rowNum)

	vals, err := st.aw.GetRange(ctx, rangeStr)
	if err != nil {
		st.logger.Error("Failed to get row from sheet", zap.Error(err), zap.String("range", rangeStr))
		return nil, err
	}

	row := vals.Values[0]
	rowData := st.translateFullRowToMap(row)

	st.rowCache.CacheRow(rowNum, rowData)
	return rowData, nil
}

// getDataMapsFromRowNums does store recieved data in cache, but does not do lookups against it.
// (the reason for that is that we want to explicit control over when we want data from cache)
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
		st.rowCache.CacheRow(rowNums[i], out[i])
	}

	return out, nil
}

// getRecordData first tries to look up data from caches, if it fails loads the data from the sheet
func (st *sheetsToolkit) getRecordData(ctx context.Context, uid string) (map[string]string, error) {
	var err error

	rowNum, uidCacheHit := st.uidCache.GetRowNumByUID(uid)
	st.logger.Debug("uid cache lookup complete", zap.Int("cachedRowNum", rowNum), zap.Bool("cacheHit", uidCacheHit), zap.String("uid", uid))

	if !uidCacheHit {
		rowNum, err = st.uidToRowNum(ctx, uid)
		if err != nil {
			st.logger.Error("Failed to translate UID to row number", zap.Error(err), zap.String("uid", uid))
			return nil, err
		}
	}

	recordDataMap, rowCacheHit := st.rowCache.GetRow(rowNum)
	st.logger.Debug("row cache lookup complete", zap.Int("rowNum", rowNum), zap.Bool("cacheHit", rowCacheHit))

	if !rowCacheHit {
		recordDataMap, err = st.getDataMapFromRowNum(ctx, rowNum)
		if err != nil {
			st.logger.Error("Failed to get data for row", zap.Error(err), zap.String("uid", uid), zap.Int("rowNum", rowNum))
			return nil, err
		}
	}

	uidOut := recordDataMap[st.uidCol]
	if uidOut != uid {
		// seems like the data is changed, between the getRowNum and getRow calls,
		// if caches were involved, let's retry the calls without them.
		// if we still get inconsistent data then something must be wrong, that we can not figure out...

		if !(uidCacheHit || rowCacheHit) {
			// caches were not involved, the data returned is just bad...
			st.logger.Error("The requested UID does not match the UID returned from the API", zap.String("uidRequested", uid), zap.String("uidReturned", uidOut))
			return nil, errors.ErrInconsistentData
		}

		// Seems like there could be cache inconsistency, we drop all data and retry...
		st.logger.Debug("There were some cache inconsistency, we re-try fetching stuff directly from the API",
			zap.String("uidRequested", uid), zap.String("uidReturned", uidOut),
			zap.Bool("uidCacheHit", uidCacheHit), zap.Bool("rowCacheHit", rowCacheHit),
		)

		// invalidate data
		if uidCacheHit {
			st.uidCache.InvalidateUID(uid)
			st.uidCache.InvalidateUID(uidOut)
		}
		if rowCacheHit {
			st.rowCache.InvalidateRow(rowNum)
		}

		// read data as fresh...

		// first the row num (if it was cached, if not then we shouldn't trash the api requests)
		if uidCacheHit { // the UID was cached, let's gather it freshly from the api...
			rowNum, err = st.uidToRowNum(ctx, uid)
			if err != nil {
				st.logger.Error("Failed to translate UID to row number", zap.Error(err), zap.String("uid", uid))
				return nil, err
			}
		}

		// always read new row data...
		recordDataMap, err = st.getDataMapFromRowNum(ctx, rowNum)
		if err != nil {
			st.logger.Error("Failed to get data for row", zap.Error(err), zap.String("uid", uid), zap.Int("rowNum", rowNum))
			return nil, err
		}

		// check success one last time if still wrong, give up...
		uidOut = recordDataMap[st.uidCol]
		if uidOut != uid {
			st.logger.Error("The requested UID does not match the UID returned from the API", zap.String("uidRequested", uid), zap.String("uidReturned", uidOut))
			return nil, errors.ErrInconsistentData
		}

	}

	return recordDataMap, nil
}

// getAllRecordsData gets all records via a single API call, it does not look up data from cache, but updates it
func (st *sheetsToolkit) getAllRecordsData(ctx context.Context) (<-chan map[string]string, error) {
	rangeStr := fmt.Sprintf("%s%d:%s", st.firstCol, st.skipRows+1, st.lastCol)

	vals, err := st.aw.GetRange(ctx, rangeStr)
	if err != nil {
		st.logger.Error("Failed to get rows from sheet", zap.Error(err), zap.String("rangeStr", rangeStr))
		return nil, err
	}

	outChan := make(chan map[string]string)

	go func() {
		defer close(outChan)
		var c int
		for i, val := range vals.Values {
			if ctx.Err() != nil { // context cancelled
				return
			}

			rowNum := i + st.skipRows + 1
			dataMap := st.translateFullRowToMap(val)
			uid := dataMap[st.uidCol]

			if uid != "" {
				// greedy caching of data...
				st.rowCache.CacheRow(rowNum, dataMap)
				st.uidCache.CacheUID(uid, rowNum)

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
// Also this function drops related entries from the cache, so they could be refreshed there as well. It does not look up anything from cache.
// This is a very taxing call, as it does at least 3 API calls each time it is called, and two of those calls are batch calls.
func (st *sheetsToolkit) updateRecords(ctx context.Context, uids []string, records []map[string]string) ([]map[string]string, error) {

	// Resolve all uids to row numbers using a single API call, we don't want to use the cache here, because
	// if we base our update call on stale data, that will cause headache
	// Note: we might want to use cache if we figure out a way to do proper transactions against the sheets api
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

		// check if UID is needed to be dropped from the cache
		newUID := r[st.uidCol]
		oldUID := uids[i]
		if newUID != "" && newUID != oldUID {
			// There possibly will be an update in the UID column, so we might want these cache entries to be dropped
			st.uidCache.InvalidateUID(newUID) // the new uid
			st.uidCache.InvalidateUID(oldUID) // the old uid
			st.logger.Debug("Invalidated UIDs in cache", zap.Strings("uids", []string{newUID, oldUID}))
		}

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

	// Before doing the actual update, drop all row cache data that would go stale
	for _, rowNum := range rowNums {
		st.rowCache.InvalidateRow(rowNum)
	}
	st.logger.Debug("Invalidated row data in cache", zap.Ints("rowNums", rowNums))

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

	return st.getDataMapsFromRowNums(ctx, rowNums) // see? we don't want to load stuff from cache,... even if it's invalidated, but we want to fill it up with the new values, which is done by this function automagically
}
