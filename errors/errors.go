package errors

import "errors"

var ErrRecordNotFound = errors.New("record with this uid not found in the sheet")

var ErrEmptyUID = errors.New("empty uid provided where uid expected")
var ErrMultiUpdate = errors.New("updating the same record multiple times in the same request")

var ErrInconsistentData = errors.New("inconsistent data, maybe changed between api calls") // we can not place a global "lock" on the sheet, sadly...

var ErrInvalidType = errors.New("invalid type passed to function")

var InvalidUIDCol = errors.New("the uid colum is invalid") // either not present in the list of columns, or something else

var ErrColsNotInOrder = errors.New("columns are not in order")
var ErrColsInvalid = errors.New("columns are invalid")

var ErrConfigInvalid = errors.New("structure config of sheet is invalid")

var ErrOverflow = errors.New("integer/float overflow error")
