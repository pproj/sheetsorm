package sheetsorm

import (
	"context"
	"errors"
	"github.com/pproj/sheetsorm/api"
	"github.com/pproj/sheetsorm/cache"
	"github.com/pproj/sheetsorm/column"
	e "github.com/pproj/sheetsorm/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
	"google.golang.org/api/sheets/v4"
	"testing"
)

func TestToolkit_uidsToRowNums(t *testing.T) {
	testError := errors.New("hello")

	testCases := []struct {
		name string

		toolkitSkipRows int
		toolkitUidCol   string

		apiResult *sheets.ValueRange
		apiError  error

		apiExpectedRange string
		apiExpectCalled  bool

		uids []string

		expectedRowNums    []int
		expectedCacheCount int
		expectedErr        error
	}{
		{
			name:            "happy__simple",
			toolkitSkipRows: 0,
			toolkitUidCol:   "A",
			apiResult: &sheets.ValueRange{
				MajorDimension: "ROWS",
				Range:          "A1:A",
				Values:         [][]interface{}{{"1"}, {"2"}, {"3"}},
			},
			apiError:           nil,
			apiExpectedRange:   "A1:A",
			apiExpectCalled:    true,
			uids:               []string{"1", "2", "3"},
			expectedRowNums:    []int{1, 2, 3},
			expectedCacheCount: 3,
			expectedErr:        nil,
		},
		{
			name:            "happy__simple_other_col",
			toolkitSkipRows: 0,
			toolkitUidCol:   "D",
			apiResult: &sheets.ValueRange{
				MajorDimension: "ROWS",
				Range:          "D1:D",
				Values:         [][]interface{}{{"1"}, {"2"}, {"3"}},
			},
			apiError:           nil,
			apiExpectedRange:   "D1:D",
			apiExpectCalled:    true,
			uids:               []string{"1", "2", "3"},
			expectedRowNums:    []int{1, 2, 3},
			expectedCacheCount: 3,
			expectedErr:        nil,
		},
		{
			name:            "happy__get_gap",
			toolkitSkipRows: 0,
			toolkitUidCol:   "A",
			apiResult: &sheets.ValueRange{
				MajorDimension: "ROWS",
				Range:          "A1:A",
				Values:         [][]interface{}{{"1"}, {"2"}, {"3"}},
			},
			apiError:           nil,
			apiExpectedRange:   "A1:A",
			apiExpectCalled:    true,
			uids:               []string{"1", "3"},
			expectedRowNums:    []int{1, 3},
			expectedCacheCount: 3,
			expectedErr:        nil,
		},
		{
			name:            "happy__skip_rows",
			toolkitSkipRows: 3,
			toolkitUidCol:   "A",
			apiResult: &sheets.ValueRange{
				MajorDimension: "ROWS",
				Range:          "A4:A",
				Values:         [][]interface{}{{"1"}, {"2"}, {"3"}},
			},
			apiError:           nil,
			apiExpectedRange:   "A4:A",
			apiExpectCalled:    true,
			uids:               []string{"1", "2", "3"},
			expectedRowNums:    []int{4, 5, 6},
			expectedCacheCount: 3,
			expectedErr:        nil,
		},
		{
			name:            "happy__get_rows_with_some_empty",
			toolkitSkipRows: 0,
			toolkitUidCol:   "A",
			apiResult: &sheets.ValueRange{
				MajorDimension: "ROWS",
				Range:          "A1:A",
				Values:         [][]interface{}{{"1"}, {"2"}, {}, {"3"}, {}, {""}, {"4"}},
			},
			apiError:           nil,
			apiExpectedRange:   "A1:A",
			apiExpectCalled:    true,
			uids:               []string{"1", "4"},
			expectedRowNums:    []int{1, 7},
			expectedCacheCount: 4,
			expectedErr:        nil,
		},
		{
			name:            "happy__multi_fill",
			toolkitSkipRows: 0,
			toolkitUidCol:   "A",
			apiResult: &sheets.ValueRange{
				MajorDimension: "ROWS",
				Range:          "A1:A",
				Values:         [][]interface{}{{"1"}, {"2"}, {"3"}},
			},
			apiError:           nil,
			apiExpectedRange:   "A1:A",
			apiExpectCalled:    true,
			uids:               []string{"1", "1", "2", "2"},
			expectedRowNums:    []int{1, 1, 2, 2},
			expectedCacheCount: 3,
			expectedErr:        nil,
		},
		{
			name:            "happy__no_request_quick",
			toolkitSkipRows: 0,
			toolkitUidCol:   "A",
			apiResult: &sheets.ValueRange{
				MajorDimension: "ROWS",
				Range:          "A1:A",
				Values:         [][]interface{}{{"1"}, {"2"}, {"3"}},
			},
			apiError:           nil,
			apiExpectedRange:   "A1:A",
			apiExpectCalled:    false,
			uids:               []string{},
			expectedRowNums:    nil,
			expectedCacheCount: 0,
			expectedErr:        nil,
		},
		{
			name:            "error__uid_not_found",
			toolkitSkipRows: 0,
			toolkitUidCol:   "A",
			apiResult: &sheets.ValueRange{
				MajorDimension: "ROWS",
				Range:          "A1:A",
				Values:         [][]interface{}{{"1"}, {"2"}, {"3"}},
			},
			apiError:           nil,
			apiExpectedRange:   "A1:A",
			apiExpectCalled:    true,
			uids:               []string{"4"},
			expectedCacheCount: 3,
			expectedErr:        e.ErrRecordNotFound,
		},
		{
			name:            "error__uid_multi_not_found_one",
			toolkitSkipRows: 0,
			toolkitUidCol:   "A",
			apiResult: &sheets.ValueRange{
				MajorDimension: "ROWS",
				Range:          "A1:A",
				Values:         [][]interface{}{{"1"}, {"2"}, {"3"}},
			},
			apiError:           nil,
			apiExpectedRange:   "A1:A",
			apiExpectCalled:    true,
			uids:               []string{"1", "2", "4"},
			expectedCacheCount: 3,
			expectedErr:        e.ErrRecordNotFound,
		},
		{
			name:            "error__uid_not_provided",
			toolkitSkipRows: 0,
			toolkitUidCol:   "A",
			apiResult: &sheets.ValueRange{
				MajorDimension: "ROWS",
				Range:          "A1:A",
				Values:         [][]interface{}{{"1"}, {"2"}, {"3"}},
			},
			apiError:           nil,
			apiExpectedRange:   "A1:A",
			apiExpectCalled:    false,
			uids:               []string{""},
			expectedCacheCount: 0,
			expectedErr:        e.ErrEmptyUID,
		},
		{
			name:            "error__api_error_propagated",
			toolkitSkipRows: 0,
			toolkitUidCol:   "A",
			apiResult: &sheets.ValueRange{
				MajorDimension: "ROWS",
				Range:          "A1:A",
				Values:         [][]interface{}{{"1"}, {"2"}, {"3"}},
			},
			apiError:           testError,
			apiExpectedRange:   "A1:A",
			apiExpectCalled:    true,
			uids:               []string{"1", "2"},
			expectedCacheCount: 0,
			expectedErr:        testError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			maw := &api.MockApiWrapper{}
			ctx := context.Background()

			var apiCalled bool

			maw.On("GetRange", ctx, tc.apiExpectedRange).Run(func(_ mock.Arguments) {
				apiCalled = true
			}).Return(tc.apiResult, tc.apiError)

			var cacheCalled int

			muic := &cache.MockRowUIDCache{}
			muic.On("CacheUID", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
				cacheCalled++
				uid := args.String(0)
				rowNum := args.Int(1)
				assert.Equal(t, tc.apiResult.Values[rowNum-tc.toolkitSkipRows-1][0], uid)
			})

			testLogger := zaptest.NewLogger(t)
			toolkit := &sheetsToolkit{
				aw:       maw,
				skipRows: tc.toolkitSkipRows,
				uidCol:   tc.toolkitUidCol,
				logger:   testLogger,
				uidCache: muic,
			}

			result, err := toolkit.uidsToRowNums(ctx, tc.uids)

			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedRowNums, result)
			}

			assert.Equal(t, tc.apiExpectCalled, apiCalled)
			assert.Equal(t, tc.expectedCacheCount, cacheCalled)
		})
	}
}

func TestToolkit_translateRowDataToUpdateRanges(t *testing.T) {
	testCases := []struct {
		name           string
		cols           column.Cols
		rowNum         int
		rowData        map[string]string
		expectedRanges []string
		expectedVals   [][]interface{}
	}{
		{
			name:   "happy__full",
			cols:   []string{"A", "B", "C"},
			rowNum: 1,
			rowData: map[string]string{
				"A": "1",
				"B": "2",
				"C": "3",
			},
			expectedRanges: []string{"A1:C1"},
			expectedVals:   [][]interface{}{{"1", "2", "3"}},
		},
		{
			name:   "happy__gap",
			cols:   []string{"A", "B", "C"},
			rowNum: 1,
			rowData: map[string]string{
				"A": "1",
				"C": "3",
			},
			expectedRanges: []string{"A1", "C1"},
			expectedVals:   [][]interface{}{{"1"}, {"3"}},
		},
		{
			name:   "happy__gap_and_span",
			cols:   []string{"A", "B", "C", "D", "E"},
			rowNum: 1,
			rowData: map[string]string{
				"A": "1",
				"C": "3",
				"D": "4",
				"E": "5",
			},
			expectedRanges: []string{"A1", "C1:E1"},
			expectedVals:   [][]interface{}{{"1"}, {"3", "4", "5"}},
		},
		{
			name:   "happy__first_only",
			cols:   []string{"A", "B", "C", "D", "E"},
			rowNum: 1,
			rowData: map[string]string{
				"A": "1",
			},
			expectedRanges: []string{"A1"},
			expectedVals:   [][]interface{}{{"1"}},
		},
		{
			name:   "happy__last_only",
			cols:   []string{"A", "B", "C", "D", "E"},
			rowNum: 1,
			rowData: map[string]string{
				"E": "5",
			},
			expectedRanges: []string{"E1"},
			expectedVals:   [][]interface{}{{"5"}},
		},
		{
			name:   "happy__last_only_span",
			cols:   []string{"A", "B", "C", "D", "E"},
			rowNum: 1,
			rowData: map[string]string{
				"D": "4",
				"E": "5",
			},
			expectedRanges: []string{"D1:E1"},
			expectedVals:   [][]interface{}{{"4", "5"}},
		},
		{
			name:   "happy__first_only_span",
			cols:   []string{"A", "B", "C", "D", "E"},
			rowNum: 1,
			rowData: map[string]string{
				"A": "1",
				"B": "2",
			},
			expectedRanges: []string{"A1:B1"},
			expectedVals:   [][]interface{}{{"1", "2"}},
		},
		{
			name:   "happy__middle_only_span",
			cols:   []string{"A", "B", "C", "D", "E"},
			rowNum: 1,
			rowData: map[string]string{
				"B": "2",
				"C": "3",
				"D": "4",
			},
			expectedRanges: []string{"B1:D1"},
			expectedVals:   [][]interface{}{{"2", "3", "4"}},
		},
		{
			name:   "happy__three_span",
			cols:   []string{"A", "B", "C", "D", "E", "F", "G", "H"},
			rowNum: 1,
			rowData: map[string]string{
				"A": "1",
				"B": "2",
				// C
				"D": "4",
				"E": "5",
				// F
				"G": "7",
				"H": "8",
			},
			expectedRanges: []string{"A1:B1", "D1:E1", "G1:H1"},
			expectedVals:   [][]interface{}{{"1", "2"}, {"4", "5"}, {"7", "8"}},
		},
		{
			name:   "happy__the_inverse_of_three_span",
			cols:   []string{"A", "B", "C", "D", "E", "F", "G", "H"},
			rowNum: 1,
			rowData: map[string]string{
				"C": "3",
				"F": "6",
			},
			expectedRanges: []string{"C1", "F1"},
			expectedVals:   [][]interface{}{{"3"}, {"6"}},
		},
		{
			name:   "error__missing_col_from_cols",
			cols:   []string{"A", "B", "C", "E", "F", "G", "H"}, // <- no D here
			rowNum: 1,
			rowData: map[string]string{
				"C": "3",
				"D": "6",
			},
			expectedRanges: []string{"C1"},
			expectedVals:   [][]interface{}{{"3"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tk := sheetsToolkit{
				cols:   tc.cols,
				logger: zaptest.NewLogger(t),
			}
			valueRanges := tk.translateRowDataToUpdateRanges(tc.rowNum, tc.rowData)

			assert.Equal(t, len(tc.expectedRanges), len(valueRanges))
			var valsC int
			var valsEC int
			for i, vr := range valueRanges {
				assert.Equal(t, tc.expectedRanges[i], vr.Range)
				assert.Equal(t, tc.expectedVals[i], vr.Values[0])
				assert.Len(t, vr.Values, 1)
				valsEC += len(tc.expectedVals[i])
				valsC += len(vr.Values[0])
			}
			assert.Equal(t, valsEC, valsC)

		})
	}

}
