package typemagic

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseTagValString(t *testing.T) {
	testCases := []struct {
		name            string
		tagValString    string
		expectedTag     Tag
		expectHasColumn bool
		expectPanic     bool
	}{
		{
			name:         "simple",
			tagValString: "A",
			expectedTag: Tag{
				Column:     "A",
				IsUID:      false,
				IsReadOnly: false,
				BoolRepresentation: BoolRepresentation{
					True:    "1",
					False:   "0",
					Unknown: false,
				},
			},
			expectHasColumn: true,
		},
		{
			name:         "simple2",
			tagValString: "AB",
			expectedTag: Tag{
				Column:     "AB",
				IsUID:      false,
				IsReadOnly: false,
				BoolRepresentation: BoolRepresentation{
					True:    "1",
					False:   "0",
					Unknown: false,
				},
			},
			expectHasColumn: true,
		},
		{
			name:         "nocol",
			tagValString: "-",
			expectedTag: Tag{
				Column:     "-",
				IsUID:      false,
				IsReadOnly: false,
				BoolRepresentation: BoolRepresentation{
					True:    "1",
					False:   "0",
					Unknown: false,
				},
			},
			expectHasColumn: false,
		},

		{
			name:         "invalidcol",
			tagValString: "abc123",
			expectPanic:  true,
		},
		{
			name:         "with_uid",
			tagValString: "AB,uid",
			expectedTag: Tag{
				Column:     "AB",
				IsUID:      true,
				IsReadOnly: false,
				BoolRepresentation: BoolRepresentation{
					True:    "1",
					False:   "0",
					Unknown: false,
				},
			},
			expectHasColumn: true,
		},
		{
			name:         "with_readonly",
			tagValString: "AB,readonly",
			expectedTag: Tag{
				Column:     "AB",
				IsUID:      false,
				IsReadOnly: true,
				BoolRepresentation: BoolRepresentation{
					True:    "1",
					False:   "0",
					Unknown: false,
				},
			},
			expectHasColumn: true,
		},
		{
			name:         "with_utrue",
			tagValString: "AB,utrue",
			expectedTag: Tag{
				Column:     "AB",
				IsUID:      false,
				IsReadOnly: false,
				BoolRepresentation: BoolRepresentation{
					True:    "1",
					False:   "0",
					Unknown: true,
				},
			},
			expectHasColumn: true,
		},
		{
			name:         "with_custom_true",
			tagValString: "AB,true=alma",
			expectedTag: Tag{
				Column:     "AB",
				IsUID:      false,
				IsReadOnly: false,
				BoolRepresentation: BoolRepresentation{
					True:    "alma",
					False:   "0",
					Unknown: false,
				},
			},
			expectHasColumn: true,
		},
		{
			name:         "with_custom_false",
			tagValString: "AB,false=alma",
			expectedTag: Tag{
				Column:     "AB",
				IsUID:      false,
				IsReadOnly: false,
				BoolRepresentation: BoolRepresentation{
					True:    "1",
					False:   "alma",
					Unknown: false,
				},
			},
			expectHasColumn: true,
		},
		{
			name:         "all",
			tagValString: "AB,uid,readonly,utrue,true=barack,false=alma",
			expectedTag: Tag{
				Column:     "AB",
				IsUID:      true,
				IsReadOnly: true,
				BoolRepresentation: BoolRepresentation{
					True:    "barack",
					False:   "alma",
					Unknown: true,
				},
			},
			expectHasColumn: true,
		},
		{
			name:         "panic_empty",
			tagValString: "",
			expectPanic:  true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			if tc.expectPanic {
				assert.Panics(t, func() {
					ParseTagValString(tc.tagValString)
				})
			} else {
				assert.NotPanics(t, func() {
					result := ParseTagValString(tc.tagValString)
					assert.Equal(t, tc.expectedTag, result)
					assert.Equal(t, tc.expectHasColumn, result.HasColumn())
				})
			}
		})
	}
}
