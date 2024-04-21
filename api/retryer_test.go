package api

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
	"testing"
	"time"
)

func TestRetryerStopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(time.Millisecond * 100)
		cancel()
	}()

	l := zaptest.NewLogger(t)
	testErr := errors.New("hello")
	var calls int

	err := DoRetry(ctx, l, func() error {
		calls++
		time.Sleep(time.Millisecond * 150)
		return testErr
	}, func(err error) bool {
		return true
	})

	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, 1, calls)
}

func TestRetryerDoesAllIfCan(t *testing.T) {
	ctx := context.Background()

	l := zaptest.NewLogger(t)
	testErr := errors.New("hello")
	var calls int

	err := DoRetry(ctx, l, func() error {
		calls++
		return testErr
	}, func(err error) bool {
		return true
	})

	assert.ErrorIs(t, err, testErr)
	assert.Equal(t, maxRetryCount, calls)
}

func TestRetryerNoErrorNoRetry(t *testing.T) {
	ctx := context.Background()

	l := zaptest.NewLogger(t)
	var calls int
	var retryCheckerCalled bool

	err := DoRetry(ctx, l, func() error {
		calls++
		return nil
	}, func(err error) bool {
		retryCheckerCalled = true
		return true
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, calls)
	assert.False(t, retryCheckerCalled)
}

func TestRetryerDontRetryIfShouldNot(t *testing.T) {
	ctx := context.Background()

	l := zaptest.NewLogger(t)
	testErr := errors.New("hello")
	var calls int

	err := DoRetry(ctx, l, func() error {
		calls++
		return testErr
	}, func(err error) bool {
		return false
	})

	assert.ErrorIs(t, err, testErr)
	assert.Equal(t, 1, calls)
}
