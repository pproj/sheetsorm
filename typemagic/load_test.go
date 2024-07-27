package typemagic

import (
	"github.com/google/uuid"
	"github.com/pproj/sheetsorm/errors"
	"github.com/stretchr/testify/assert"
	"net"
	"reflect"
	"strconv"
	"testing"
)

type TestScanner struct {
	called bool
}

func (ts *TestScanner) Scan(src any) error {
	ts.called = true
	return nil
}

var TestScannerNotStructBasedScanCalled int

type TestScannerNotStructBased int64

func (ts *TestScannerNotStructBased) Scan(src any) error {
	TestScannerNotStructBasedScanCalled++
	return nil
}

type TestScannerErrors struct {
	called bool
}

func (ts *TestScannerErrors) Scan(src any) error {
	ts.called = true
	return errors.ErrOverflow
}

func TestLoadIntoStruct(t *testing.T) {
	type testStructSimple struct {
		Name string `sheet:"A"`
	}
	type testStructExtended struct {
		Str   string    `sheet:"A"`
		Str2  string    `sheet:"B"`
		Int   int       `sheet:"C"`
		Uint  uint      `sheet:"D"`
		Float float64   `sheet:"E"`
		Bool  bool      `sheet:"F"`
		UUID  uuid.UUID `sheet:"G"`
	}
	type testStructExtendedPtr struct {
		Str   *string    `sheet:"A"`
		Str2  *string    `sheet:"B"`
		Int   *int       `sheet:"C"`
		Uint  *uint      `sheet:"D"`
		Float *float64   `sheet:"E"`
		Bool  *bool      `sheet:"F"`
		UUID  *uuid.UUID `sheet:"G"`
	}
	type testStructExtendedBools struct {
		Bool1 bool  `sheet:"A"`
		Bool2 bool  `sheet:"B,true=alma,false=barack"`
		Bool3 bool  `sheet:"C,utrue"`
		Bool4 *bool `sheet:"D"`
		Bool5 *bool `sheet:"E,true=alma,false=barack"`
		Bool6 *bool `sheet:"F,utrue"`
	}
	type testStructInts struct {
		Int    int  `sheet:"A"`
		IntPtr *int `sheet:"B"`
	}
	type testStructInts2 struct {
		Int     int  `sheet:"-"`
		IntPtr  *int `sheet:"-"`
		Int2    int
		IntPtr2 *int
	}
	type testStructIPs struct {
		IP    net.IP  `sheet:"A"`
		IPPtr *net.IP `sheet:"B"`
	}
	type testStructScanner struct {
		Scanner    TestScanner  `sheet:"A"`
		ScannerPtr *TestScanner `sheet:"B"`
	}
	type testStructScannerNonStruct struct {
		Scanner    TestScannerNotStructBased  `sheet:"A"`
		ScannerPtr *TestScannerNotStructBased `sheet:"B"`
	}
	type testStructScannerErrors struct {
		Scanner    TestScannerErrors  `sheet:"A"`
		ScannerPtr *TestScannerErrors `sheet:"B"`
	}
	type testStructNested struct {
		Str  string `sheet:"A"`
		Data struct {
			Name string `sheet:"B"`
			Age  int    `sheet:"C"`
		}
	}
	type testStructNestedDeeper struct {
		GoDeeper testStructNested
		Hello    string `sheet:"D"`
	}
	type testStructNestedDeeperPtr struct {
		GoDeeper *testStructNested
		Hello    string `sheet:"D"`
	}
	type testStructOverflow struct {
		Int8    int8    `sheet:"A"`
		Unt8    uint8   `sheet:"B"`
		Float32 float32 `sheet:"C"`
	}
	type testStructUnexported struct {
		unexported  int `sheet:"A"`
		unexported2 int `sheet:"B"`
		Exported    int `sheet:"C"`
	}
	type testStructMultiptr struct {
		Multiptr **string `sheet:"A"`
	}
	type testStructMultiptrButIgnored struct {
		Multiptr **string // but should be ignored...
		Hello    string   `sheet:"A"`
	}
	type testStructInvalidType struct {
		InvalidType struct{} `sheet:"A"`
	}

	testUUID := uuid.New()
	testVal := "test"
	testValInt := -1
	testValInt69 := 69
	testValUint := uint(1)
	testValBoolTrue := true
	testValBoolFalse := false
	testValFloat := 12.2
	testValIPv4 := net.IPv4(127, 0, 0, 1)
	testValIPv6 := net.ParseIP("fe80::")
	testScannerNonStruct := TestScannerNotStructBased(0)

	testCases := []struct {
		name          string
		data          map[string]string
		item          interface{}
		expectedItem  interface{}
		expectedErr   error
		expectedPanic bool
	}{
		{
			name: "simple",
			data: map[string]string{
				"A": "alma",
			},
			item:         &testStructSimple{},
			expectedItem: testStructSimple{Name: "alma"},
		}, {
			name: "not_simple",
			data: map[string]string{
				"A": "alma",
				"B": "",
				"C": "-1",
				"D": "1",
				"E": "12.2",
				"F": "1",
				"G": testUUID.String(),
			},
			item: &testStructExtended{},
			expectedItem: testStructExtended{
				Str:   "alma",
				Str2:  "",
				Int:   -1,
				Uint:  1,
				Float: 12.2,
				Bool:  true,
				UUID:  testUUID,
			},
		}, {
			name: "not_simple_pointers",
			data: map[string]string{
				"A": testVal,
				"B": testVal,
				"C": "-1",
				"D": "1",
				"E": "12.2",
				"F": "1",
				"G": testUUID.String(),
			},
			item: &testStructExtendedPtr{},
			expectedItem: testStructExtendedPtr{
				Str:   &testVal,
				Str2:  &testVal,
				Int:   &testValInt,
				Uint:  &testValUint,
				Float: &testValFloat,
				Bool:  &testValBoolTrue,
				UUID:  &testUUID,
			},
		}, {
			name: "bools",
			data: map[string]string{
				"A": "0",
				"B": "alma",
				"C": "alma",
				"D": "0",
				"E": "alma",
				"F": "alma",
			},
			item: &testStructExtendedBools{},
			expectedItem: testStructExtendedBools{
				Bool1: false,
				Bool2: true,
				Bool3: true,
				Bool4: &testValBoolFalse,
				Bool5: &testValBoolTrue,
				Bool6: &testValBoolTrue,
			},
		}, {
			name: "bools_2",
			data: map[string]string{
				"A": "1",
				"B": "barack",
				"C": "barack",
				"D": "1",
				"F": "barack",
				"E": "barack",
			},
			item: &testStructExtendedBools{},
			expectedItem: testStructExtendedBools{
				Bool1: true,
				Bool2: false,
				Bool3: true,
				Bool4: &testValBoolTrue,
				Bool5: &testValBoolFalse,
				Bool6: &testValBoolTrue,
			},
		}, {
			name: "error__int_parse",
			data: map[string]string{
				"A": "alma",
				"B": "alma",
			},
			item:        &testStructInts{},
			expectedErr: strconv.ErrSyntax,
		}, {
			name: "error__uint_parse",
			data: map[string]string{
				"D": "alma",
			},
			item:        &testStructExtended{},
			expectedErr: strconv.ErrSyntax,
		}, {
			name: "error__float_parse",
			data: map[string]string{
				"E": "alma",
			},
			item:        &testStructExtended{},
			expectedErr: strconv.ErrSyntax,
		}, {
			name: "error__int_parse_ptr",
			data: map[string]string{
				"A": "12",
				"B": "alma",
			},
			item:        &testStructInts{},
			expectedErr: strconv.ErrSyntax,
		}, {
			name: "error__int_parse_nested",
			data: map[string]string{
				"A": "12",
				"B": "alma",
				"C": "alma",
			},
			item:        &testStructNested{},
			expectedErr: strconv.ErrSyntax,
		},
		{
			name: "ips_v4",
			data: map[string]string{
				"A": "127.0.0.2",
				"B": "127.0.0.1",
			},
			item: &testStructIPs{},
			expectedItem: testStructIPs{
				IP:    net.IPv4(127, 0, 0, 2),
				IPPtr: &testValIPv4,
			},
			expectedErr:   nil,
			expectedPanic: false,
		},
		{
			name: "ips_v6",
			data: map[string]string{
				"A": "fe80::1",
				"B": "fe80::",
			},
			item: &testStructIPs{},
			expectedItem: testStructIPs{
				IP:    net.ParseIP("fe80::1"),
				IPPtr: &testValIPv6,
			},
			expectedErr:   nil,
			expectedPanic: false,
		},
		{
			name: "overwrite",
			data: map[string]string{
				"A": "420",
				"B": "69",
			},
			item: &testStructInts{
				Int:    12,
				IntPtr: &testValInt,
			},
			expectedItem: testStructInts{
				Int:    420,
				IntPtr: &testValInt69,
			},
		}, {
			name: "ignored_untouched",
			data: map[string]string{
				"A": "420",
				"B": "69",
				"C": "123",
				"D": "456",
			},
			item: &testStructInts2{
				Int:     12,
				IntPtr:  &testValInt,
				Int2:    1,
				IntPtr2: nil,
			},
			expectedItem: testStructInts2{
				Int:     12,
				IntPtr:  &testValInt,
				Int2:    1,
				IntPtr2: nil,
			},
		}, {
			name: "ignore_no_val",
			data: map[string]string{
				"A": "hello",
				"B": "world",
			},
			item: &testStructExtended{},
			expectedItem: testStructExtended{
				Str:  "hello",
				Str2: "world",
			},
		}, {
			name: "scanner",
			data: map[string]string{
				"A": "420",
				"B": "barack",
			},
			item: &testStructScanner{},
			expectedItem: testStructScanner{
				Scanner:    TestScanner{called: true},
				ScannerPtr: &TestScanner{called: true},
			},
		}, {
			name: "scanner_non_struct",
			data: map[string]string{
				"A": "alma",
				"B": "barack",
			},
			item: &testStructScannerNonStruct{},
			expectedItem: testStructScannerNonStruct{
				Scanner:    TestScannerNotStructBased(0),
				ScannerPtr: &testScannerNonStruct,
			},
		}, {
			name: "scanner_errors",
			data: map[string]string{
				"A": "alma",
				"B": "barack",
			},
			item:        &testStructScannerErrors{},
			expectedErr: errors.ErrOverflow,
		}, {
			name: "nested",
			data: map[string]string{
				"A": "alma",
				"B": "barack",
				"C": "420",
			},
			item: &testStructNested{},
			expectedItem: testStructNested{
				Str: "alma",
				Data: struct {
					Name string `sheet:"B"`
					Age  int    `sheet:"C"`
				}{
					Name: "barack",
					Age:  420,
				},
			},
		}, {
			name: "nested_deep",
			data: map[string]string{
				"A": "alma",
				"B": "barack",
				"C": "420",
				"D": "Hello",
			},
			item: &testStructNestedDeeper{},
			expectedItem: testStructNestedDeeper{
				GoDeeper: testStructNested{
					Str: "alma",
					Data: struct {
						Name string `sheet:"B"`
						Age  int    `sheet:"C"`
					}{
						Name: "barack",
						Age:  420,
					},
				},
				Hello: "Hello",
			},
		}, {
			name: "nested_deep_ptr",
			data: map[string]string{
				"A": "alma",
				"B": "barack",
				"C": "420",
				"D": "Hello",
			},
			item: &testStructNestedDeeperPtr{},
			expectedItem: testStructNestedDeeperPtr{
				GoDeeper: &testStructNested{
					Str: "alma",
					Data: struct {
						Name string `sheet:"B"`
						Age  int    `sheet:"C"`
					}{
						Name: "barack",
						Age:  420,
					},
				},
				Hello: "Hello",
			},
		}, {
			name: "nested_deep_ptr_but_not_nil",
			data: map[string]string{
				"A": "alma",
				"B": "barack",
				"C": "420",
				"D": "Hello",
			},
			item: &testStructNestedDeeperPtr{GoDeeper: &testStructNested{
				Str: "hello",
			}},
			expectedItem: testStructNestedDeeperPtr{
				GoDeeper: &testStructNested{
					Str: "alma",
					Data: struct {
						Name string `sheet:"B"`
						Age  int    `sheet:"C"`
					}{
						Name: "barack",
						Age:  420,
					},
				},
				Hello: "Hello",
			},
		}, {
			name: "error__overflow_int",
			data: map[string]string{
				"A": "2048",
			},
			item:        &testStructOverflow{},
			expectedErr: errors.ErrOverflow,
		}, {
			name: "error__overflow_uint",
			data: map[string]string{
				"B": "2048",
			},
			item:        &testStructOverflow{},
			expectedErr: errors.ErrOverflow,
		}, {
			name: "error__overflow_float32",
			data: map[string]string{
				"C": "184467440737095516161844674407370955161618446744073709551616.1844674407370955161618446744073709551616",
			},
			item:        &testStructOverflow{},
			expectedErr: errors.ErrOverflow,
		}, {
			name: "ignored_unexported",
			data: map[string]string{
				"A": "1",
				"B": "2",
				"C": "3",
			},
			item: &testStructUnexported{},
			expectedItem: testStructUnexported{
				unexported:  0,
				unexported2: 0,
				Exported:    3,
			},
		}, {
			name: "panic__multiptr",
			data: map[string]string{
				"A": "1", // placeholder
				"B": "2",
				"C": "3",
			},
			item:          &testStructMultiptr{},
			expectedPanic: true,
		}, {
			name: "multiptr_ignored",
			data: map[string]string{
				"A": "world",
			},
			item: &testStructMultiptrButIgnored{},
			expectedItem: testStructMultiptrButIgnored{
				Multiptr: nil,
				Hello:    "world",
			},
		}, {
			name: "panic__nil",
			data: map[string]string{
				"A": "1", // placeholder
				"B": "2",
				"C": "3",
			},
			item:          nil,
			expectedPanic: true,
		}, {
			name: "panic__unexpected_ptr",
			data: map[string]string{
				"A": "1", // placeholder
				"B": "2",
				"C": "3",
			},
			item:          &testValBoolFalse,
			expectedPanic: true,
		}, {
			name: "panic__store_into_invalid_type",
			data: map[string]string{
				"A": "1",
			},
			item:          &testStructInvalidType{},
			expectedPanic: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			if tc.expectedPanic {
				assert.Panics(t, func() {
					_ = LoadIntoStruct(tc.data, tc.item)
				})
			} else {
				assert.NotPanics(t, func() {
					err := LoadIntoStruct(tc.data, tc.item)
					if tc.expectedErr != nil {
						assert.Error(t, err)
						assert.ErrorIs(t, err, tc.expectedErr)
					} else {
						assert.NoError(t, err)
						assert.Equal(t, tc.expectedItem, reflect.ValueOf(tc.item).Elem().Interface())
					}
				})
			}

		})
	}

	t.Run("assert_scanner_call_counters", func(t *testing.T) {
		assert.Equal(t, 2, TestScannerNotStructBasedScanCalled)
	})

}

func TestLoadIntoStructSimplified(t *testing.T) {
	type test1 struct {
		Name string `sheet:"A"`
	}
	data := map[string]string{
		"A": "alma",
	}
	tv := test1{}
	te := test1{
		Name: "alma",
	}

	_ = LoadIntoStruct(data, &tv)
	assert.Equal(t, te, tv)
	assert.Equal(t, tv.Name, "alma")
}
