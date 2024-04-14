package typemagic

import (
	"database/sql/driver"
	"fmt"
	"github.com/pproj/sheetsorm/column"
	"reflect"
	"slices"
	"strconv"
	"time"
)

var stringerType = reflect.TypeOf((*fmt.Stringer)(nil)).Elem() // yes...
var valuerType = reflect.TypeOf((*driver.Valuer)(nil)).Elem()  // yes...

func workOutValue(value reflect.Value, br BoolRepresentation) string {
	// value should never be a pointer...

	v := value.Interface()

	if !value.IsValid() {
		return "" // TODO, maybe panic?
	}

	// try valuer first
	if value.Type().Implements(valuerType) {
		valuer := v.(driver.Valuer)
		valuedValue, err := valuer.Value()
		if err != nil {
			panic(err)
		}

		switch valuedValue := valuedValue.(type) { // the only possible values should be this, by documentation
		case int64:
			return fmt.Sprintf("%d", valuedValue)
		case float64:
			return strconv.FormatFloat(valuedValue, 'f', -1, 64)
		case bool:
			return br.Represent(valuedValue)
		case []byte:
			return string(valuedValue)
		case string:
			return valuedValue
		case time.Time:
			return valuedValue.String()
		default:
			panic("valuer implemented, but returned invalid value")
		}

	}

	// then stringer
	if value.Type().Implements(stringerType) {
		s := v.(fmt.Stringer)
		return s.String()
	} else { // this is getting weird...
		if value.CanAddr() {
			vPtr := value.Addr()
			if vPtr.Type().Implements(stringerType) {
				s := vPtr.Interface().(fmt.Stringer)
				return s.String()
			}
		}
	}

	// then try "manually" serialize

	switch v := v.(type) {
	case bool:
		return br.Represent(v)
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	}

	if value.CanInt() || value.CanUint() {
		return fmt.Sprintf("%d", v)
	}

	// if none working, resort to this
	return fmt.Sprintf("%v", value)
}

func magicDumpIter(item interface{}, iterator func(valueValid bool, value reflect.Value, t Tag) bool) {
	val := reflect.ValueOf(item)

	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			panic("this is nil")
		}

		val = val.Elem()
		if val.Kind() == reflect.Interface { // okay, I probably should read some documentation at this point...
			if val.IsNil() {
				panic("this is nil")
			}
			val = val.Elem()
		}
	}

	if val.Kind() != reflect.Struct {
		panic("expected struct or pointer to struct, not " + val.Kind().String())
	}

	for i := 0; i < val.NumField(); i++ {
		f := val.Field(i)

		if !val.Type().Field(i).IsExported() {
			// ignore unexported fields
			continue
		}

		valueValid := true // in other words... non-nil
		if f.Kind() == reflect.Ptr {
			if f.IsNil() {
				// skip nil field
				valueValid = false
			} else {
				// pointer but not nil
				f = f.Elem() // up one level

				if f.Kind() == reflect.Ptr { // check if it's not another pointer
					panic("multi-level pointers are not supported") // because I'm lazy
				}
			}
		}

		// Okay, this is getting tricky... so how we decide if we should recurse into a struct or not?
		// Naive idea: Let's just use the struct tag... if it has one, we won't recurse, if not then we should...
		// probably a better idea would be to take into consideration if it implements valuer or stringer...

		tagVal := val.Type().Field(i).Tag.Get(SheetTag)

		if tagVal == "" {
			if valueValid && f.Kind() == reflect.Struct {
				// if another struct, then recurse into it
				magicDumpIter(f.Interface(), iterator)
				continue
			}

			// ignore un-configured, non-struct
			continue
		}

		// parse struct tag
		tag := ParseTagValString(tagVal)

		if !tag.HasColumn() { // has a "-" as the column, should be ignored ...
			continue
		}

		shouldContinue := iterator(valueValid, f, tag)
		if !shouldContinue {
			return // stop iterator
		}
	}
}

// DumpStruct dumps the structure into a rowData map based on the sheet:"..." struct tag. It can omit fields marked as read-only
// The uid field is not read-only by default, so if you want to omit it from the dump, you must mark it as read-only in the struct tag.
func DumpStruct(item interface{}, omitReadOnly bool) map[string]string {
	// We are writing type-safe type-unsafe code here...

	data := make(map[string]string)

	magicDumpIter(item, func(valid bool, value reflect.Value, t Tag) bool {
		if !valid {
			return true
		}
		if t.IsReadOnly && omitReadOnly {
			// If a value is read-only then we might not want to get it dumped, for example for updates..
			return true
		}

		_, ok := data[t.Column]
		if ok {
			panic("multiple values assigned to the same column")
		}
		data[t.Column] = workOutValue(value, t.BoolRepresentation)
		return true
	})

	return data

}

// DumpUID extracts the UID value from the struct, if it is not configured it will use the left-most value, it dumps the value even if the uid col is marked read-only
func DumpUID(item interface{}) string {

	var result string // used in place of uid if not defined (the left most column)
	var minCol = -1   // invalid
	var found bool

	magicDumpIter(item, func(valid bool, value reflect.Value, t Tag) bool {
		if t.IsUID {
			if valid {
				result = workOutValue(value, t.BoolRepresentation)
			} else {
				result = ""
			}
			found = true
			return false // found the explicit uid definition
		}

		// if not explicitly configured, then check if it's lefter than the previous
		colIdx := column.ColIndex(t.Column)
		if minCol == -1 || colIdx < minCol {
			if valid {
				result = workOutValue(value, t.BoolRepresentation)
			} else {
				result = ""
			}
			minCol = colIdx
			found = true
		}
		return true

	})
	if !found {
		panic("no suitable field for uid found")
	}

	return result
}

// DumpUIDCol is the same as DumpUID but with the column itself
func DumpUIDCol(item interface{}) string {

	var result string // used in place of uid if not defined (the left most column)
	var minCol = -1   // invalid

	magicDumpIter(item, func(_ bool, _ reflect.Value, t Tag) bool {
		if t.IsUID {
			result = t.Column
			return false // found the explicit uid definition
		}

		// if not explicitly configured, then check if it's lefter than the previous
		colIdx := column.ColIndex(t.Column)
		if minCol == -1 || colIdx < minCol {
			result = t.Column
			minCol = colIdx
		}
		return true

	})
	if result == "" {
		panic("no suitable field for uid found")
	}

	return result
}

// DumpCols returns column.Cols that are used for this type (regardless if the column has a valid value or not)
func DumpCols(item interface{}) column.Cols {

	var resultS []string

	magicDumpIter(item, func(_ bool, _ reflect.Value, t Tag) bool {

		if slices.Contains(resultS, t.Column) {
			panic("multiple values assigned to the same column")
		}

		resultS = append(resultS, t.Column)
		return true
	})

	slices.SortFunc(resultS, func(a, b string) int {
		return column.ColIndex(a) - column.ColIndex(b)
	})

	result := column.Cols(resultS)

	err := result.Validate()
	if err != nil {
		panic(err)
	}

	return result
}
