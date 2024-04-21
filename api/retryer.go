package api

import (
	"context"
	"go.uber.org/zap"
	"math"
	"math/rand"
	"time"
)

const (
	maxRetryCount  = 7
	maxBackoffTime = time.Second * 32
)

// DoRetry is a lazy implementation based on https://developers.google.com/sheets/api/limits
func DoRetry(ctx context.Context, logger *zap.Logger, retryableFunc func() error, shouldRetry func(error) bool) error {

	var nextTryTimeout time.Duration = 0
	var tryCount int
	for {
		tryCount++
		select {
		case <-time.After(nextTryTimeout):
			logger.Debug("Attempting to do retryable request...", zap.Int("tryCount", tryCount), zap.Int("maxRetryCount", maxRetryCount))
			err := retryableFunc()

			if err == nil {
				// no error, we are good
				logger.Debug("Retryable request succeeded", zap.Int("tryCount", tryCount), zap.Int("maxRetryCount", maxRetryCount))
				return nil
			}

			// there was an error
			if tryCount >= maxRetryCount { // if tries exhausted, return
				logger.Warn("Maximum retry limit reached... giving up...", zap.Int("tryCount", tryCount), zap.Int("maxRetryCount", maxRetryCount))
				return err
			}
			if !shouldRetry(err) { // otherwise, check if we should retry
				logger.Warn("The error is not retryable... giving up...", zap.Int("tryCount", tryCount), zap.Int("maxRetryCount", maxRetryCount), zap.Error(err))
				return err
			}

			// if we should retry, set the timeout
			randomNumberMilliseconds := time.Duration(rand.Intn(1000)) * time.Millisecond // #nosec G404 this is just a retryer
			exp := time.Duration(math.Pow(2, float64(tryCount))) * time.Second
			nextTryTimeout = min(exp+randomNumberMilliseconds, maxBackoffTime)
			logger.Debug("The retryable request failed. Scheduled a retry", zap.Int("tryCount", tryCount), zap.Int("maxRetryCount", maxRetryCount), zap.Error(err), zap.Duration("nextTryTimeout", nextTryTimeout))
		case <-ctx.Done():
			logger.Debug("context cancelled", zap.Error(ctx.Err()), zap.Int("tryCount", tryCount), zap.Int("maxRetryCount", maxRetryCount))
			return ctx.Err()
		}
	}

}
