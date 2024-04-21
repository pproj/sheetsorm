package cache

// RowUIDCache caches the row numbers for certain UIDs
// Implementations must be thread safe
// If the cache runs into any error, sheetsorm does not really care about that, so the cache can not report an error, it has to figure it out for itself
// UIDs are always strings, they are what typemagic spits out for that row
// It's generally okay to return stale data, because sheetsorm will check if the returned data has the correct UID, if not then it will invalidate the data.
// The same RowUIDCache backend could be shared among more instances of sheetsorm to improve performance.
type RowUIDCache interface {
	// CacheUID should store the uid-rowNum pair in the cache
	CacheUID(string, int)

	// GetRowNumByUID should return a row number for a specific UID if it is in the cache. If not then the second return value must be false
	// If anything goes wrong with the cache, this method should just return as if the data was not in the cache, and sheetsorm will go and fetch it
	GetRowNumByUID(string) (int, bool)

	// InvalidateUID should drop a cache entry for the specified UID
	// Note: since caches may implement any sort of forget mechanism, dropping the entire cache for this call is
	// a valid operation, but not really efficient
	InvalidateUID(string)
}

// RowCache caches entire rows of data identified by their RowNum
// While it's generally okay to return stale data in RowUIDCache, returning stale data here have more serious implications
// As if sheetsorm have a cache hit in both the RowCache and the RowUIDCache, it will not fetch any data from the sheet,
// and it is possible to drift away a lot from the actual sheet this way. Therefore, a short forget policy is a MUST for RowCache
// Generally for small applications, RowCache isn't really recommended, as it's not really useful when there are a low amount of requests.
// Errors from RowCache isn't interesting by sheetsorm either, the interface does not require error reporting capability on purpose.
// The row cache WILL NEVER be used to lookup UIDs even if they are technically could be
type RowCache interface {
	// CacheRow stores an entire row for a rowNum
	CacheRow(int, map[string]string)

	// GetRow should return a row for the rowNum if it's in the cache, if the rowNum is not in the cache, then the second return value must be false
	GetRow(int) (map[string]string, bool)

	// InvalidateRow should drop the cache entry for the rowNum. Or drop the entire cache, it's fine either way.
	InvalidateRow(int)
}
