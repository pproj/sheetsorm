package cache

// NullCache is a simple "implementation" of cache that is used when no cache is set by the user of the library
// it implements both the Row and UID caches.
type NullCache struct {
}

func (nc *NullCache) CacheUID(uid string, rowNum int) {
}

func (nc *NullCache) GetRowNumByUID(uid string) (int, bool) {
	return 0, false
}

func (nc *NullCache) InvalidateUID(uid string) {

}

func (nc *NullCache) CacheRow(rowNum int, rowData map[string]string) {
}

func (nc *NullCache) GetRow(rowNum int) (map[string]string, bool) {
	return nil, false
}

func (nc *NullCache) InvalidateRow(rowNum int) {

}
