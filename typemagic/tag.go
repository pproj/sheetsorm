package typemagic

import (
	"github.com/pproj/sheetsorm/column"
	"strings"
)

/*
example struct tag:

`sheet:"A,uid,True=1,False=0"`


*/

type Tag struct {
	Column string
	IsUID  bool

	// IsReadOnly is in the context of the sheet, in other words, if this is set to true, the value will be read from the sheet, but never written to the sheet
	IsReadOnly         bool
	BoolRepresentation BoolRepresentation
}

func (t Tag) HasColumn() bool {
	return t.Column != "-"
}

type BoolRepresentation struct {
	True  string
	False string

	// Unknown represents the default value to be used, when it can not be decided
	Unknown bool
}

func (br BoolRepresentation) UnRepresent(val string) bool {
	switch val {
	case br.True:
		return true
	case br.False:
		return false
	default:
		return br.Unknown
	}
}
func (br BoolRepresentation) Represent(val bool) string {
	if val {
		return br.True
	} else {
		return br.False
	}
}

func NewDefaultTag() Tag {
	return Tag{
		IsUID:      false,
		IsReadOnly: false,
		BoolRepresentation: BoolRepresentation{
			True:    "1",
			False:   "0",
			Unknown: false,
		},
	}
}

func ParseTagValString(tagVal string) Tag {
	if tagVal == "" {
		panic("tag string could not be empty")
	}
	elems := strings.Split(tagVal, ",")
	if len(elems) == 0 {
		panic("wtf")
	}
	t := NewDefaultTag()
	t.Column = elems[0]

	if t.HasColumn() {
		if !column.IsValidCol(t.Column) {
			panic("invalid column name defined")
		}
	}

	if len(elems) == 1 {
		return t
	}

	for _, elem := range elems[1:] { // we will no longer need the first element
		// If this gets out of hand, we should just split on the first = and use the first part in a split case
		if elem == SheetTagOptionUID {
			t.IsUID = true
			continue
		}
		if elem == SheetTagOptionReadOnly {
			t.IsReadOnly = true
			continue
		}
		if elem == SheetTagOptionUnknownIsTrue {
			t.BoolRepresentation.Unknown = true
			continue
		}
		if strings.HasPrefix(elem, SheetTagOptionTrueRepr) {
			t.BoolRepresentation.True = strings.TrimPrefix(elem, SheetTagOptionTrueRepr)
			continue
		}
		if strings.HasPrefix(elem, SheetTagOptionFalseRepr) {
			t.BoolRepresentation.False = strings.TrimPrefix(elem, SheetTagOptionFalseRepr)
			continue
		}

	}

	return t
}
