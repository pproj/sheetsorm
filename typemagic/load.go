package typemagic

import (
	"database/sql"
	"encoding"
	"github.com/pproj/sheetsorm/errors"
	"reflect"
	"strconv"
)

func magicLoaderIter(item interface{}, iterator func(value reflect.Value, t Tag) error) (int, error) {
	val := reflect.ValueOf(item)

	if val.Kind() != reflect.Ptr || val.IsNil() {
		panic("expected pointer to a struct, not " + val.Kind().String())
	}

	val = reflect.Indirect(val)
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		panic("expected pointer to a struct, not pointer to " + val.Kind().String())
	}

	successfullyVisitedCount := 0
	for i := 0; i < val.NumField(); i++ {
		f := val.Field(i)

		if !val.Type().Field(i).IsExported() {
			// ignore unexported fields
			continue
		}

		if !f.CanSet() {
			// ignore fields we could not set anyway
			continue
		}

		// Same issue as with dumper
		tagVal := val.Type().Field(i).Tag.Get(SheetTag)

		if tagVal == "" {
			if f.Kind() == reflect.Struct || (f.Kind() == reflect.Ptr && f.Type().Elem().Kind() == reflect.Struct) { // (this type check works with nil ptrs)

				var toRecurseInto reflect.Value
				var shouldAddBack bool
				if f.Kind() == reflect.Ptr {
					if f.IsNil() { // if nil, then create it
						toRecurseInto = reflect.New(f.Type().Elem())
						shouldAddBack = true
					} else {
						toRecurseInto = f
						shouldAddBack = false
					}
				} else {
					// it's not a pointer type of field, but we need a pointer to it to be able to modify
					toRecurseInto = f.Addr()
					shouldAddBack = false
				}

				// if another struct, then recurse into it
				subVisited, err := magicLoaderIter(toRecurseInto.Interface(), iterator) // this may panic in weird cases...
				if err != nil {
					return 0, err
				}

				if shouldAddBack && subVisited > 0 { // only add the new struct if it was useful (i.e.: don't add empty struct)
					f.Set(toRecurseInto)
				}

				continue
			}

			// ignore that does not have tagval and not even a struct
			continue
		}

		// parse struct tag
		t := ParseTagValString(tagVal)

		if !t.HasColumn() { // has a "-" as the column, should be ignored ...
			continue
		}

		if f.Kind() == reflect.Ptr { // multi level ptrs should only cause panic, if we try to use them, that's why it's after all possible skips
			if f.Type().Elem().Kind() == reflect.Ptr { // check if it's not another pointer
				panic("multi-level pointers are not supported")
			}
		}

		err := iterator(f, t)
		if err != nil {
			return 0, err
		}
		successfullyVisitedCount++
	}
	return successfullyVisitedCount, nil
}

var scannerType = reflect.TypeOf((*sql.Scanner)(nil)).Elem()                  // yes...
var unmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem() // eh

// convertAndStoreProperly returns an error only if the data is invalid for the format, otherwise it panics as usual
func convertAndStoreProperly(value reflect.Value, data string, br BoolRepresentation) error {

	// First, try if it implements scanner type
	if value.Type().Implements(scannerType) { // might happen if we already got a pointer type
		s := value.Interface().(sql.Scanner)
		return s.Scan(data)
	} else { // if we didn't get a pointer type, we might still can do something about that
		if value.CanAddr() {
			vPtr := value.Addr()
			if vPtr.Type().Implements(scannerType) {
				s := vPtr.Interface().(sql.Scanner)
				return s.Scan(data)
			}
		}
	}

	// Try the same with TextUnmarshaler https://pkg.go.dev/encoding#TextMarshaler
	if value.Type().Implements(unmarshalerType) { // might happen if we already got a pointer type
		s := value.Interface().(encoding.TextUnmarshaler)
		return s.UnmarshalText([]byte(data))
	} else { // if we didn't get a pointer type, we might still can do something about that
		if value.CanAddr() {
			vPtr := value.Addr()
			if vPtr.Type().Implements(unmarshalerType) {
				s := vPtr.Interface().(encoding.TextUnmarshaler)
				return s.UnmarshalText([]byte(data))
			}
		}
	}

	// If not, do the manual labor...
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(data, 10, 64)
		if err != nil {
			return err
		}
		if value.OverflowInt(i) {
			return errors.ErrOverflow
		}
		value.SetInt(i)
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(data, 10, 64)
		if err != nil {
			return err
		}
		if value.OverflowUint(i) {
			return errors.ErrOverflow
		}
		value.SetUint(i)
		return nil

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(data, 64)
		if err != nil {
			return err
		}
		if value.OverflowFloat(f) {
			return errors.ErrOverflow
		}
		value.SetFloat(f)
		return nil

	case reflect.String:
		value.SetString(data)
		return nil

	case reflect.Bool:
		value.SetBool(br.UnRepresent(data))
		return nil

	default:
		panic("unsupported kind: " + value.Kind().String())
	}
}

// LoadIntoStruct returns an error only if the supplied data (coming from sheets) is not valid for the type in the struct. If the struct itself has issues, it will panic as ususal.
func LoadIntoStruct(data map[string]string, item interface{}) error {
	var err error
	_, err = magicLoaderIter(item, func(value reflect.Value, t Tag) error {

		dataVal, ok := data[t.Column]
		if !ok {
			return nil // nothing to set, continue iteration..
		}

		if value.Kind() == reflect.Ptr {
			immediateVal := reflect.New(value.Type().Elem())
			internalErr := convertAndStoreProperly(immediateVal.Elem(), dataVal, t.BoolRepresentation)
			if internalErr != nil {
				return internalErr
			}
			value.Set(immediateVal)
			return nil
		} else {
			return convertAndStoreProperly(value, dataVal, t.BoolRepresentation)
		}
	})
	return err
}
