package sheetsorm

import (
	"github.com/pproj/sheetsorm/column"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
	"testing"
)

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
