package sheetsorm

import (
	"context"
	"errors"
	"fmt"
	"github.com/pproj/sheetsorm/api"
	"github.com/pproj/sheetsorm/cache"
	e "github.com/pproj/sheetsorm/errors"
	"github.com/pproj/sheetsorm/typemagic"
	"go.uber.org/zap"
	"google.golang.org/api/sheets/v4"
	"reflect"
	"slices"
	"sync"
)

type Sheet interface {
	// GetRecord fetches a single record from the sheet, the passed struct must have its UID field filled, or it returns an error
	GetRecord(ctx context.Context, out interface{}) error

	// GetAllRecords returns all valid records from the the sheet, the argument must be a list of structs
	GetAllRecords(ctx context.Context, out interface{}) error

	// UpdateRecords take individual records, or list of records, or both as vararg. The UID field of each record must be filled, otherwise it returns an error
	UpdateRecords(ctx context.Context, records ...interface{}) error
}

type SheetImpl struct {
	mu *sync.RWMutex
	aw api.ApiWrapper

	logger   *zap.Logger
	skipRows int

	uidCache cache.RowUIDCache
	rowCache cache.RowCache
}

type SheetInitializationOption func(*SheetImpl)

func NewSheet(srv *sheets.Service, st StructureConfig, opts ...SheetInitializationOption) (*SheetImpl, error) {

	err := st.Validate()
	if err != nil {
		return nil, err
	}

	nc := &cache.NullCache{}
	nl := zap.NewNop()

	si := &SheetImpl{
		mu:       &sync.RWMutex{},
		aw:       nil, // will be initialized after applying options, because they configure the logger as well
		logger:   nl,
		skipRows: st.SkipRows,
		uidCache: nc,
		rowCache: nc,
	}

	for _, o := range opts {
		o(si)
	}

	si.aw = api.NewApiWrapper(srv, st.DocID, st.Sheet, si.logger)

	return si, nil
}

func WithRowUIDCache(c cache.RowUIDCache) SheetInitializationOption {
	return func(si *SheetImpl) {
		si.uidCache = c
	}
}

func WithRowCache(c cache.RowCache) SheetInitializationOption {
	return func(si *SheetImpl) {
		si.rowCache = c
	}
}

func WithLogger(l *zap.Logger) SheetInitializationOption {
	return func(si *SheetImpl) {
		si.logger = l
	}
}

func typeAssert(val interface{}, expectedKinds ...reflect.Kind) bool {
	typ := reflect.TypeOf(val)
	for i, expectedKind := range expectedKinds {
		if typ.Kind() != expectedKind {
			return false
		}
		if i < len(expectedKinds)-1 {
			typ = typ.Elem()
		}
	}
	return true
}

func (si *SheetImpl) getToolkit(sample interface{}) (*sheetsToolkit, error) {
	cols := typemagic.DumpCols(sample)
	uidCol := typemagic.DumpUIDCol(sample)

	return newToolkit(si.aw, cols, uidCol, si.skipRows, si.logger, si.uidCache, si.rowCache)
}

func (si *SheetImpl) GetRecord(ctx context.Context, out interface{}) error {
	si.mu.RLock()
	defer si.mu.RUnlock()

	if !typeAssert(out, reflect.Ptr, reflect.Struct) {
		return errors.Join(e.ErrInvalidType, fmt.Errorf("expected pointer to a struct"))
	}

	toolkit, err := si.getToolkit(out)
	if err != nil {
		si.logger.Error("Failed to initialize toolkit", zap.Error(err))
		return err
	}

	uid := typemagic.DumpUID(out)
	if uid == "" {
		return e.ErrEmptyUID
	}

	var data map[string]string
	data, err = toolkit.getRecordData(ctx, uid)
	if err != nil {
		si.logger.Error("error while getting record data")
		return err
	}

	return typemagic.LoadIntoStruct(data, out)
}

func (si *SheetImpl) GetAllRecords(ctx context.Context, out interface{}) error {
	si.mu.RLock()
	defer si.mu.RUnlock()

	if !typeAssert(out, reflect.Ptr, reflect.Slice, reflect.Struct) {
		return errors.Join(e.ErrInvalidType, fmt.Errorf("expected pointer to a struct"))
	}

	// create a sample instance first
	inst := reflect.New(reflect.TypeOf(out).Elem().Elem())

	toolkit, err := si.getToolkit(inst.Elem().Interface()) // dereference it
	if err != nil {
		si.logger.Error("Failed to initialize toolkit", zap.Error(err))
		return err
	}

	var ch <-chan map[string]string
	ch, err = toolkit.getAllRecordsData(ctx)
	if err != nil {
		si.logger.Error("Failure while getting records", zap.Error(err))
		return err
	}

	outSlicePtr := reflect.New(reflect.TypeOf(out).Elem())
	outSlice := outSlicePtr.Elem()

loop:
	for {
		select {
		case data, ok := <-ch:
			if !ok {
				break loop
			}

			inst = reflect.New(reflect.TypeOf(out).Elem().Elem())

			err = typemagic.LoadIntoStruct(data, inst.Interface())
			if err != nil {
				return err
			}

			outSlice.Set(reflect.Append(outSlice, inst.Elem()))
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	reflect.ValueOf(out).Elem().Set(outSlicePtr.Elem())

	return nil
}

// UpdateRecords the corresponding uid field must be filled in the records in receives, if the uid can not be found in the table, it throws an error
func (si *SheetImpl) UpdateRecords(ctx context.Context, records ...interface{}) error {
	si.mu.Lock()
	defer si.mu.Unlock()

	if len(records) == 0 {
		return nil
	}

	unwrappedRecords := make([]interface{}, 0) // <- will store just pointers to structs

	for _, r := range records {
		if typeAssert(r, reflect.Ptr, reflect.Struct) {
			unwrappedRecords = append(unwrappedRecords, r)
			continue
		}
		if typeAssert(r, reflect.Slice, reflect.Ptr, reflect.Struct) {
			val := reflect.ValueOf(r)
			for i := 0; i < val.Len(); i++ {
				unwrappedRecords = append(unwrappedRecords, val.Index(i).Interface())
			}
			continue
		}

		return errors.Join(e.ErrInvalidType, fmt.Errorf("expected pointer to struct or a slice of pointers to structs"))
	}

	// create a sample first, for the toolkit
	inst := reflect.New(reflect.TypeOf(unwrappedRecords[0]).Elem())

	toolkit, err := si.getToolkit(inst.Elem().Interface())
	if err != nil {
		si.logger.Error("Failed to initialize toolkit", zap.Error(err))
		return err
	}

	allData := make([]map[string]string, len(unwrappedRecords))
	uids := make([]string, len(unwrappedRecords))
	for i, r := range unwrappedRecords {
		uid := typemagic.DumpUID(r)

		if slices.Contains(uids, uid) {
			return e.ErrMultiUpdate
		}

		uids[i] = uid
		if uids[i] == "" {
			return e.ErrEmptyUID
		}

		allData[i] = typemagic.DumpStruct(r, true)

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	var updatedData []map[string]string
	updatedData, err = toolkit.updateRecords(ctx, uids, allData)
	if err != nil {
		si.logger.Error("error while updating records", zap.Error(err))
		return err
	}

	for i, r := range unwrappedRecords {
		err = typemagic.LoadIntoStruct(updatedData[i], r)
		if err != nil {
			return err
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	return nil
}
