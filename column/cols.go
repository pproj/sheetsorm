package column

import (
	"github.com/pproj/sheets-orm/errors"
	"slices"
)

// Cols are columns, must be in order
type Cols []string

// Shift is the amount of columns the first column is shifted from left
// Can be useful when translating column number to column name in a result set that does not start from the left most column
func (c Cols) Shift() int {
	return ColIndex(c[0]) // the columns are validated to be in order, and non-empty
}

func (c Cols) Last() string {
	return c[len(c)-1]
}

func (c Cols) First() string {
	return c[0]
}

func (c Cols) Validate() error {
	if len(c) == 0 {
		return errors.ErrColsInvalid
	}

	lastIdx := -1
	for _, col := range c {
		if !IsValidCol(col) {
			return errors.ErrColsInvalid
		}

		idx := ColIndex(col)
		if idx <= lastIdx { // This also ensures no duplicate
			return errors.ErrColsNotInOrder
		}
		lastIdx = idx
	}
	return nil
}

func (c Cols) Contains(col string) bool {
	if !IsValidCol(col) {
		return false
	}
	return slices.Contains(c, col)
}
