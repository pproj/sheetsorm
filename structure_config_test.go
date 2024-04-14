package sheets

import (
	"github.com/pproj/sheets-orm/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStructureConfig_Validate(t *testing.T) {
	testCases := []struct {
		name        string
		sc          StructureConfig
		expectedErr error
	}{
		{
			name: "happy__simple",
			sc: StructureConfig{
				DocID:    "dummy_doc_id",
				Sheet:    "dummy_sheet_name",
				SkipRows: 12,
			},
			expectedErr: nil,
		},
		{
			name: "happy__default_sheet",
			sc: StructureConfig{
				DocID:    "dummy_doc_id",
				Sheet:    "",
				SkipRows: 12,
			},
			expectedErr: nil,
		},
		{
			name: "error__empty_doc_id",
			sc: StructureConfig{
				DocID:    "",
				Sheet:    "dummy_sheet_name",
				SkipRows: 12,
			},
			expectedErr: errors.ErrConfigInvalid,
		},
		{
			name: "error__skip_rows_negative",
			sc: StructureConfig{
				DocID:    "dummy_doc_id",
				Sheet:    "dummy_sheet_name",
				SkipRows: -12,
			},
			expectedErr: errors.ErrConfigInvalid,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.ErrorIs(t, tc.sc.Validate(), tc.expectedErr)
		})
	}
}
