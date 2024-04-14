package sheets

import (
	"github.com/pproj/sheets-orm/errors"
)

type StructureConfig struct {
	DocID string
	Sheet string

	SkipRows int
}

func (st StructureConfig) Validate() error {
	if st.SkipRows < 0 {
		return errors.ErrConfigInvalid
	}
	if st.DocID == "" {
		return errors.ErrConfigInvalid
	}
	return nil
}
