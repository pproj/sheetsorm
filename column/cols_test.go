package column

import (
	"github.com/pproj/sheetsorm/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCols_Validate(t *testing.T) {
	testCases := []struct {
		name        string
		testVal     Cols
		expectedErr error
	}{
		{
			name:        "happy__simple",
			testVal:     []string{"A", "B", "C"},
			expectedErr: nil,
		},
		{
			name:        "happy__simple_with_hole",
			testVal:     []string{"A", "C"},
			expectedErr: nil,
		},
		{
			name:        "error__unordered",
			testVal:     []string{"A", "C", "B"},
			expectedErr: errors.ErrColsNotInOrder,
		},
		{
			name:        "error__double",
			testVal:     []string{"A", "A", "B"},
			expectedErr: errors.ErrColsNotInOrder,
		},
		{
			name:        "error__double_not_next_to",
			testVal:     []string{"A", "B", "A"},
			expectedErr: errors.ErrColsNotInOrder,
		},
		{
			name:        "error__empty",
			testVal:     []string{},
			expectedErr: errors.ErrColsInvalid,
		},
		{
			name:        "error__invalid_data_1",
			testVal:     []string{"a", "A"},
			expectedErr: errors.ErrColsInvalid,
		},
		{
			name:        "error__invalid_data_2",
			testVal:     []string{"", "A"},
			expectedErr: errors.ErrColsInvalid,
		},
		{
			name:        "error__invalid_data_3",
			testVal:     []string{"12", "A"},
			expectedErr: errors.ErrColsInvalid,
		},
		{
			name:        "error__invalid_data_4",
			testVal:     []string{"❤️", "A"},
			expectedErr: errors.ErrColsInvalid,
		},
		{
			name:        "error__invalid_data_5",
			testVal:     []string{"A1", "A"},
			expectedErr: errors.ErrColsInvalid,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.testVal.Validate()
			if tc.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			}
		})
	}
}

func TestCols_Shift(t *testing.T) {
	for i := 0; i < 100; i++ {
		cols := Cols{ColFromIndex(i), ColFromIndex(i + 1), ColFromIndex(i + 2), ColFromIndex(i + 3)}
		assert.NoError(t, cols.Validate())
		assert.Equal(t, i, cols.Shift())
	}
}

func TestCols_First(t *testing.T) {
	for i := 0; i < 100; i++ {
		firstCol := ColFromIndex(i)
		cols := Cols{firstCol, ColFromIndex(i + 1), ColFromIndex(i + 2), ColFromIndex(i + 3)}
		assert.NoError(t, cols.Validate())
		assert.Equal(t, firstCol, cols.First())
	}
}

func TestCols_Last(t *testing.T) {
	for i := 0; i < 100; i++ {
		lastCol := ColFromIndex(i + 3)
		cols := Cols{ColFromIndex(i), ColFromIndex(i + 1), ColFromIndex(i + 2), lastCol}
		assert.NoError(t, cols.Validate())
		assert.Equal(t, lastCol, cols.Last())
	}
}

func TestCols_Contains(t *testing.T) {
	for i := 0; i < 100; i++ {
		cols := Cols{ColFromIndex(i), ColFromIndex(i + 1), ColFromIndex(i + 2), ColFromIndex(i + 3)}
		assert.True(t, cols.Contains(ColFromIndex(i)))
		assert.True(t, cols.Contains(ColFromIndex(i+1)))
		assert.True(t, cols.Contains(ColFromIndex(i+2)))
		assert.True(t, cols.Contains(ColFromIndex(i+3)))
		assert.False(t, cols.Contains(ColFromIndex(i+4)))
		assert.False(t, cols.Contains("asd")) // not valid
	}
}
