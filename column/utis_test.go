package column

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestIndexing(t *testing.T) {
	wg := sync.WaitGroup{}

	wg.Add(10)
	for i := 0; i < 10; i++ {
		batch := i
		go func(batch int) {
			defer wg.Done()
			start := batch * 1000000
			end := (batch + 1) * 1000000
			for j := start; j < end; j++ {
				col := ColFromIndex(j)
				assert.True(t, IsValidCol(col))
				assert.Equal(t, j, ColIndex(col))
			}
		}(batch)
	}

	wg.Wait()
}
