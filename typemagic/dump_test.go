package typemagic

import (
	"database/sql/driver"
	"github.com/google/uuid"
	"github.com/pproj/sheets-orm/column"
	"github.com/pproj/sheets-orm/errors"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

var calledValuers int

// Valuers, pointer reciever

type TestValuerIntPtrRcv struct {
	called bool
}

func (tv *TestValuerIntPtrRcv) Value() (driver.Value, error) {
	calledValuers++
	tv.called = true
	return int64(1), nil
}

type TestValuerFloatPtrRcv struct {
	called bool
}

func (tv *TestValuerFloatPtrRcv) Value() (driver.Value, error) {
	calledValuers++
	tv.called = true
	return float64(1.2), nil
}

type TestValuerBoolPtrRcv struct {
	called bool
}

func (tv *TestValuerBoolPtrRcv) Value() (driver.Value, error) {
	calledValuers++
	tv.called = true
	return true, nil
}

type TestValuerBytesPtrRcv struct {
	called bool
}

func (tv *TestValuerBytesPtrRcv) Value() (driver.Value, error) {
	calledValuers++
	tv.called = true
	return []byte("hello"), nil
}

type TestValuerStringPtrRcv struct {
	called bool
}

func (tv *TestValuerStringPtrRcv) Value() (driver.Value, error) {
	calledValuers++
	tv.called = true
	return "hello", nil
}

type TestValuerTimePtrRcv struct {
	called   bool
	testTime time.Time
}

func (tv *TestValuerTimePtrRcv) Value() (driver.Value, error) {
	calledValuers++
	tv.called = true
	return tv.testTime, nil
}

type TestValuerInvalidPtrRcv struct {
	called bool
}

func (tv *TestValuerInvalidPtrRcv) Value() (driver.Value, error) {
	calledValuers++
	tv.called = true
	return struct{}{}, nil
}

// Valuers

type TestValuerInt struct{}

func (tv TestValuerInt) Value() (driver.Value, error) {
	calledValuers++
	return int64(1), nil
}

type TestValuerFloat struct{}

func (tv TestValuerFloat) Value() (driver.Value, error) {
	calledValuers++
	return float64(1.2), nil
}

type TestValuerBool struct{}

func (tv TestValuerBool) Value() (driver.Value, error) {
	calledValuers++
	return true, nil
}

type TestValuerBytes struct{}

func (tv TestValuerBytes) Value() (driver.Value, error) {
	calledValuers++
	return []byte("hello"), nil
}

type TestValuerString struct{}

func (tv TestValuerString) Value() (driver.Value, error) {
	calledValuers++
	return "hello", nil
}

type TestValuerTime struct {
	testTime time.Time
}

func (tv TestValuerTime) Value() (driver.Value, error) {
	calledValuers++
	return tv.testTime, nil
}

type TestValuerInvalid struct{}

func (tv TestValuerInvalid) Value() (driver.Value, error) {
	calledValuers++
	return struct{}{}, nil
}

// More valuers

type TestValuerThatIsNotStruct int64

func (tv TestValuerThatIsNotStruct) Value() (driver.Value, error) {
	calledValuers++
	return int64(tv + 1), nil
}

type TestValuerThatIsNotStructPtrRcv int64

func (tv *TestValuerThatIsNotStructPtrRcv) Value() (driver.Value, error) {
	calledValuers++
	return int64((*tv) + 1), nil
}

type TestValuerErrors struct{}

func (tv TestValuerErrors) Value() (driver.Value, error) {
	calledValuers++
	return int64(0), errors.ErrColsNotInOrder
}

type TestValuerErrorsPtrRcv struct {
	called bool
}

func (tv *TestValuerErrorsPtrRcv) Value() (driver.Value, error) {
	calledValuers++
	tv.called = true
	return int64(0), errors.ErrColsNotInOrder
}

func TestWorkOutValue(t *testing.T) {

	type testStruct struct {
		Str     string
		Int     int
		Int8    int8
		Int16   int16
		Int32   int32
		Int64   int64
		UInt    uint
		UInt8   uint8
		UInt16  uint16
		UInt32  uint32
		UInt64  uint64
		Float64 float64
		Float32 float32
		UUID    uuid.UUID
		Bool    bool
		Bool2   bool
		Time    time.Time
		Wild    []int //this should trigger the %v path

		IntValuer     TestValuerInt
		FloatValuer   TestValuerFloat
		BoolValuer    TestValuerBool
		BytesValuer   TestValuerBytes
		StringValuer  TestValuerString
		TimeValuer    TestValuerTime
		InvalidValuer TestValuerInvalid

		IntValuerPtr     *TestValuerIntPtrRcv
		FloatValuerPtr   *TestValuerFloatPtrRcv
		BoolValuerPtr    *TestValuerBoolPtrRcv
		BytesValuerPtr   *TestValuerBytesPtrRcv
		StringValuerPtr  *TestValuerStringPtrRcv
		TimeValuerPtr    *TestValuerTimePtrRcv
		InvalidValuerPtr *TestValuerInvalidPtrRcv

		NotStructValuer    TestValuerThatIsNotStruct
		NotStructValuerPtr *TestValuerThatIsNotStructPtrRcv
		ErrorValuer        TestValuerErrors
		ErrorValuerPtr     *TestValuerErrorsPtrRcv
	}

	testTimestamp := time.Now()
	testValNotStructValuerPtrRcv := TestValuerThatIsNotStructPtrRcv(785)
	testVal := testStruct{
		Str:                "hello",
		Int:                -2,
		Int8:               -64,
		Int16:              -1024,
		Int32:              -1048576,
		Int64:              -281474976710656,
		UInt:               3,
		UInt8:              64,
		UInt16:             1024,
		UInt32:             1048576,
		UInt64:             281474976710656,
		Float64:            4.2,
		Float32:            2.4,
		UUID:               uuid.New(),
		Bool:               true,
		Bool2:              false,
		Time:               time.Now(),
		Wild:               []int{1, 2, 3},
		IntValuer:          TestValuerInt{},
		FloatValuer:        TestValuerFloat{},
		BoolValuer:         TestValuerBool{},
		BytesValuer:        TestValuerBytes{},
		StringValuer:       TestValuerString{},
		TimeValuer:         TestValuerTime{testTime: testTimestamp},
		InvalidValuer:      TestValuerInvalid{},
		IntValuerPtr:       &TestValuerIntPtrRcv{},
		FloatValuerPtr:     &TestValuerFloatPtrRcv{},
		BoolValuerPtr:      &TestValuerBoolPtrRcv{},
		BytesValuerPtr:     &TestValuerBytesPtrRcv{},
		StringValuerPtr:    &TestValuerStringPtrRcv{},
		TimeValuerPtr:      &TestValuerTimePtrRcv{testTime: testTimestamp},
		InvalidValuerPtr:   &TestValuerInvalidPtrRcv{},
		NotStructValuer:    TestValuerThatIsNotStruct(77),
		NotStructValuerPtr: &testValNotStructValuerPtrRcv,
		ErrorValuer:        TestValuerErrors{},
		ErrorValuerPtr:     &TestValuerErrorsPtrRcv{},
	}

	expectedStrings := []string{
		"hello",
		"-2",
		"-64",
		"-1024",
		"-1048576",
		"-281474976710656",
		"3",
		"64",
		"1024",
		"1048576",
		"281474976710656",
		"4.2",
		"2.4",
		testVal.UUID.String(),
		"yes",
		"no",
		testVal.Time.String(),
		"[1 2 3]", // default
		"1",
		"1.2",
		"yes",
		"hello",
		"hello",
		testTimestamp.String(),
		"!!!PANIC!!!",
		"1",
		"1.2",
		"yes",
		"hello",
		"hello",
		testTimestamp.String(),
		"!!!PANIC!!!",
		"78",
		"786",
		"!!!PANIC!!!",
		"!!!PANIC!!!",
	}

	val := reflect.ValueOf(testVal)
	for i := 0; i < val.NumField(); i++ {
		t.Run(val.Type().Field(i).Name, func(t *testing.T) {

			expectedPanic := expectedStrings[i] == "!!!PANIC!!!"
			f := val.Field(i)

			if expectedPanic {
				assert.Panics(t, func() {
					workOutValue(f, BoolRepresentation{})
				})
			} else {
				assert.NotPanics(t, func() {
					valStr := workOutValue(f, BoolRepresentation{
						True:    "yes",
						False:   "no",
						Unknown: false,
					})
					assert.Equal(t, expectedStrings[i], valStr)
				})
			}

		})

	}

	t.Run("check_valuers_called_properly", func(t *testing.T) {
		assert.True(t, testVal.IntValuerPtr.called)
		assert.True(t, testVal.FloatValuerPtr.called)
		assert.True(t, testVal.BoolValuerPtr.called)
		assert.True(t, testVal.BytesValuerPtr.called)
		assert.True(t, testVal.StringValuerPtr.called)
		assert.True(t, testVal.TimeValuerPtr.called)
		assert.True(t, testVal.IntValuerPtr.called)
		assert.True(t, testVal.ErrorValuerPtr.called)
		assert.Equal(t, 18, calledValuers)
	})
}

func TestDumpCols(t *testing.T) {
	type test1 struct {
		A string `sheet:"A,uid"`
		B string `sheet:"B"`
		C string `sheet:"C"`
	}
	type test2 struct {
		A string `sheet:"A"`
		B string `sheet:"A"`
		C string `sheet:"C"`
	}
	type test3 struct {
		A string `sheet:"A"`
		B string `sheet:"-"`
		C string `sheet:"C"`
	}
	type test4 struct {
		A string `sheet:"A"`
		B string
		C string `sheet:"C"`
	}
	type test5 struct {
		A string `sheet:"A"`
		B string `sheet:"aa12"`
		C string `sheet:"C"`
	}
	type test6 struct {
		A *string  `sheet:"A"`
		B *int     `sheet:"B"`
		C *bool    `sheet:"C"`
		D bool     `sheet:"D"`
		E struct{} `sheet:"E"`
	}
	type test7 struct {
		A *string  `sheet:"A"`
		B *int     `sheet:"A"`
		C *bool    `sheet:"A"`
		D bool     `sheet:"D"`
		E struct{} `sheet:"E"`
	}
	type test8 struct {
		A string   `sheet:"A"`
		b int      `sheet:"A"`
		c bool     `sheet:"A"`
		D bool     `sheet:"D"`
		E struct{} `sheet:"E"`
	}
	type test9Sub struct {
		C string `sheet:"C"`
		D string `sheet:"D"`
	}
	type test9 struct {
		A    string `sheet:"A"`
		B    string `sheet:"B"`
		Data test9Sub
	}

	testCases := []struct {
		name          string
		item          interface{}
		expectedCols  column.Cols
		expectedPanic bool
	}{
		{
			name:         "simple",
			item:         test1{},
			expectedCols: []string{"A", "B", "C"},
		},
		{
			name:          "duplicate",
			item:          test2{},
			expectedPanic: true,
		},
		{
			name:         "explicitly_skipped_one",
			item:         test3{},
			expectedCols: []string{"A", "C"},
		},
		{
			name:         "not_tagged_one",
			item:         test4{},
			expectedCols: []string{"A", "C"},
		},
		{
			name:          "invalid",
			item:          test5{},
			expectedPanic: true,
		},
		{
			name:          "empty",
			item:          struct{}{},
			expectedPanic: true,
		},
		{
			name:         "works_with_nil_and_empty",
			item:         test6{},
			expectedCols: []string{"A", "B", "C", "D", "E"},
		},
		{
			name:          "panics_nil_multi_col",
			item:          test7{},
			expectedPanic: true,
		},
		{
			name:         "unexported_ignored",
			item:         test8{},
			expectedCols: []string{"A", "D", "E"},
		},
		{
			name:         "nested",
			item:         test9{},
			expectedCols: []string{"A", "B", "C", "D"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectedPanic {
				assert.Panics(t, func() {
					DumpCols(tc.item)
				})
			} else {
				assert.NotPanics(t, func() {
					result := DumpCols(tc.item)
					assert.Equal(t, tc.expectedCols, result)
				})
			}
		})
	}
}

func TestDumpUID(t *testing.T) {

	type test1 struct {
		A string `sheet:"A,uid"`
		B string `sheet:"B"`
		C string `sheet:"C"`
	}
	type test2 struct {
		A string `sheet:"A"`
		B string `sheet:"B"`
		C string `sheet:"C"`
	}
	type test3 struct {
		C string `sheet:"C"`
		B string `sheet:"B"`
		A string `sheet:"A"`
	}
	type test4 struct {
		A int    `sheet:"A,uid"`
		B string `sheet:"B"`
		C string `sheet:"C"`
	}
	type test5 struct {
		A uuid.UUID `sheet:"A,uid"`
		B string    `sheet:"B"`
		C string    `sheet:"C"`
	}
	type test6 struct {
		A uuid.UUID `sheet:"A"`
		B string    `sheet:"B,uid"`
		C string    `sheet:"C"`
	}
	type test7 struct {
		A *string `sheet:"A,uid"`
		B string  `sheet:"B"`
		C string  `sheet:"C"`
	}
	type test8 struct {
		A *string `sheet:"A"`
		B string  `sheet:"B"`
		C string  `sheet:"C"`
	}
	type test9 struct {
		A string  `sheet:"B"`
		B *string `sheet:"D,uid"`
		C string  `sheet:"C"`
	}
	type test10 struct {
		A string  `sheet:"B"`
		B *string `sheet:"A"`
		C string  `sheet:"C"`
	}
	testUUID := uuid.New()

	testCases := []struct {
		name           string
		item           interface{}
		expectedUID    string
		expectedUIDCol string
		expectedPanic  bool
	}{
		{
			name: "explicit",
			item: test1{
				A: "alma",
				B: "barack",
				C: "citrom",
			},
			expectedUID:    "alma",
			expectedUIDCol: "A",
		}, {
			name: "leftmost",
			item: test2{
				A: "alma",
				B: "barack",
				C: "citrom",
			},
			expectedUID:    "alma",
			expectedUIDCol: "A",
		}, {
			name: "leftmost__unordered",
			item: test3{
				A: "alma",
				B: "barack",
				C: "citrom",
			},
			expectedUID:    "alma",
			expectedUIDCol: "A",
		}, {
			name: "explicit__int",
			item: test4{
				A: 12,
				B: "barack",
				C: "citrom",
			},
			expectedUID:    "12",
			expectedUIDCol: "A",
		}, {
			name: "explicit__uuid",
			item: test5{
				A: testUUID,
				B: "barack",
				C: "citrom",
			},
			expectedUID:    testUUID.String(),
			expectedUIDCol: "A",
		}, {
			name: "explicit__not_left_most",
			item: test6{
				A: testUUID,
				B: "barack",
				C: "citrom",
			},
			expectedUID:    "barack",
			expectedUIDCol: "B",
		}, {
			name:          "empty_struct",
			item:          struct{}{},
			expectedPanic: true,
		}, {
			name: "irrelevant_stuff",
			item: struct {
				Name  string `sheet:"-"`
				Age   int
				ohBoi bool
			}{},
			expectedPanic: true,
		}, {
			name:           "explicit_but_nil",
			item:           test7{},
			expectedUID:    "",
			expectedUIDCol: "A",
		}, {
			name:           "implicit_but_nil",
			item:           test8{},
			expectedUID:    "",
			expectedUIDCol: "A",
		}, {
			name:           "explicit_unordered_but_nil",
			item:           test9{},
			expectedUID:    "",
			expectedUIDCol: "D",
		}, {
			name:           "implicit_unordered_but_nil",
			item:           test10{},
			expectedUID:    "",
			expectedUIDCol: "A",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			if tc.expectedPanic {
				assert.Panics(t, func() {
					DumpUID(tc.item)
				})
				assert.Panics(t, func() {
					DumpUIDCol(tc.item)
				})
			} else {
				assert.NotPanics(t, func() {
					result := DumpUID(tc.item)
					assert.Equal(t, tc.expectedUID, result)
					resultCol := DumpUIDCol(tc.item)
					assert.Equal(t, tc.expectedUIDCol, resultCol)
				})
			}
		})
	}
}

func TestDumpStruct(t *testing.T) {

	var testValString = "test"
	var testValStringPtr = &testValString
	var testValInt = 12
	var testValBool = true

	testCases := []struct {
		name          string
		item          interface{}
		omitReadOnly  bool
		expectedDump  map[string]string
		expectedPanic bool
	}{
		{
			name: "test__simple",
			item: struct {
				Name string `sheet:"A"`
			}{
				Name: "alma",
			},
			expectedDump: map[string]string{
				"A": "alma",
			},
		}, {
			name: "test__more",
			item: struct {
				Name    string  `sheet:"A"`
				Age     int     `sheet:"B"`
				Valid   bool    `sheet:"C"`
				PtrVal1 *string `sheet:"D"`
				PtrVal2 *int    `sheet:"E"`
				PtrVal3 *bool   `sheet:"F"`
			}{
				Name:    "alma",
				Age:     12,
				Valid:   true,
				PtrVal1: &testValString,
				PtrVal2: &testValInt,
				PtrVal3: &testValBool,
			},
			expectedDump: map[string]string{
				"A": "alma",
				"B": "12",
				"C": "1",
				"D": "test",
				"E": "12",
				"F": "1",
			},
		}, {
			name: "test__unexported_ignored",
			item: struct {
				name  string `sheet:"A"`
				Name2 string `sheet:"B"`
			}{
				name:  "alma",
				Name2: "alma",
			},
			expectedDump: map[string]string{
				"B": "alma",
			},
		}, {
			name: "test__nil_ignored",
			item: struct {
				Name    string  `sheet:"A"`
				NamePtr *string `sheet:"B"`
			}{
				Name:    "alma",
				NamePtr: nil,
			},
			expectedDump: map[string]string{
				"A": "alma",
			},
		}, {
			name: "test__un-configured_ignored",
			item: struct {
				Name  string `sheet:"A"`
				Name2 string
			}{
				Name:  "alma",
				Name2: "barack",
			},
			expectedDump: map[string]string{
				"A": "alma",
			},
		}, {
			name: "test__ignored_ignored",
			item: struct {
				Name  string `sheet:"A"`
				Name2 string `sheet:"-"`
			}{
				Name:  "alma",
				Name2: "barack",
			},
			expectedDump: map[string]string{
				"A": "alma",
			},
		}, {
			name: "test__ignore_readonly",
			item: struct {
				Name  string `sheet:"A"`
				Name2 string `sheet:"B,readonly"`
			}{
				Name:  "alma",
				Name2: "barack",
			},
			omitReadOnly: true,
			expectedDump: map[string]string{
				"A": "alma",
			},
		}, {
			name: "test__not_ignore_readonly",
			item: struct {
				Name  string `sheet:"A"`
				Name2 string `sheet:"B,readonly"`
			}{
				Name:  "alma",
				Name2: "barack",
			},
			omitReadOnly: false,
			expectedDump: map[string]string{
				"A": "alma",
				"B": "barack",
			},
		}, {
			name: "test__ignore_readonly_recurse",
			item: struct {
				Name string `sheet:"A"`
				Data struct {
					Item1 string `sheet:"B"`
					Item2 string `sheet:"C,readonly"`
				}
			}{
				Name: "alma",
				Data: struct {
					Item1 string `sheet:"B"`
					Item2 string `sheet:"C,readonly"`
				}{
					Item1: "barack",
					Item2: "korte",
				},
			},
			omitReadOnly: true,
			expectedDump: map[string]string{
				"A": "alma",
				"B": "barack",
			},
		}, {
			name: "test__not_ignore_readonly_recurse",
			item: struct {
				Name string `sheet:"A"`
				Data struct {
					Item1 string `sheet:"B"`
					Item2 string `sheet:"C,readonly"`
				}
			}{
				Name: "alma",
				Data: struct {
					Item1 string `sheet:"B"`
					Item2 string `sheet:"C,readonly"`
				}{
					Item1: "barack",
					Item2: "korte",
				},
			},
			omitReadOnly: false,
			expectedDump: map[string]string{
				"A": "alma",
				"B": "barack",
				"C": "korte",
			},
		}, {
			name: "test__recurse",
			item: struct {
				Name string `sheet:"A"`
				Data struct {
					Age  int `sheet:"B"`
					Year int `sheet:"C"`
				}
			}{
				Name: "alma",
				Data: struct {
					Age  int `sheet:"B"`
					Year int `sheet:"C"`
				}{
					Age:  12,
					Year: 2002,
				},
			},
			expectedDump: map[string]string{
				"A": "alma",
				"B": "12",
				"C": "2002",
			},
		}, {
			name: "test__recurse_deeper",
			item: struct {
				Name string `sheet:"A"`
				Data struct {
					Age       int `sheet:"B"`
					Year      int `sheet:"C"`
					OtherData struct {
						NoIdea string `sheet:"D"`
					}
				}
			}{
				Name: "alma",
				Data: struct {
					Age       int `sheet:"B"`
					Year      int `sheet:"C"`
					OtherData struct {
						NoIdea string `sheet:"D"`
					}
				}{
					Age:  12,
					Year: 2002,
					OtherData: struct {
						NoIdea string `sheet:"D"`
					}{
						NoIdea: "ohboi",
					},
				},
			},
			expectedDump: map[string]string{
				"A": "alma",
				"B": "12",
				"C": "2002",
				"D": "ohboi",
			},
		}, {
			name: "test__recurse_deeper_ptr",
			item: struct {
				Name string `sheet:"A"`
				Data *struct {
					Age       int `sheet:"B"`
					Year      int `sheet:"C"`
					OtherData struct {
						NoIdea string `sheet:"D"`
					}
				}
			}{
				Name: "alma",
				Data: &struct {
					Age       int `sheet:"B"`
					Year      int `sheet:"C"`
					OtherData struct {
						NoIdea string `sheet:"D"`
					}
				}{
					Age:  12,
					Year: 2002,
					OtherData: struct {
						NoIdea string `sheet:"D"`
					}{
						NoIdea: "ohboi",
					},
				},
			},
			expectedDump: map[string]string{
				"A": "alma",
				"B": "12",
				"C": "2002",
				"D": "ohboi",
			},
		}, {
			name: "test__multiptr_panic",
			item: struct {
				Name       string   `sheet:"A"`
				NamePtrPtr **string `sheet:"B"`
			}{
				Name:       "alma",
				NamePtrPtr: &testValStringPtr,
			},
			expectedPanic: true,
		}, {
			name: "test__duplicate_panic",
			item: struct {
				Name       string `sheet:"A"`
				NamePtrPtr string `sheet:"A"`
			}{
				Name:       "alma",
				NamePtrPtr: "barack",
			},
			expectedPanic: true,
		}, {
			name:          "test__nil_panic",
			item:          nil,
			expectedPanic: true,
		}, {
			name:          "test__int_panic",
			item:          12,
			expectedPanic: true,
		}, {
			name:         "test__empty",
			item:         struct{}{},
			expectedDump: map[string]string{},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectedPanic {
				assert.Panics(t, func() {
					DumpStruct(tc.item, tc.omitReadOnly)
				})
			} else {
				assert.NotPanics(t, func() {
					result := DumpStruct(tc.item, tc.omitReadOnly)
					assert.Equal(t, tc.expectedDump, result)

					// this should be valid too
					result2 := DumpStruct(&tc.item, tc.omitReadOnly)
					assert.Equal(t, tc.expectedDump, result2)

				})
			}
		})
	}
}
