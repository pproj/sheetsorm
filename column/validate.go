package column

import "regexp"

var colNameValidator = regexp.MustCompile(`^[A-Z]+$`)

func IsValidCol(colname string) bool {
	return colNameValidator.MatchString(colname)
}
