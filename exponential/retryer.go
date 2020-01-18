package exponential

import (
	"context"
	"math"
	"math/rand"
	"sync"
	"time"
)

type lockedSource struct {
	lk  sync.Mutex
	src rand.Source
}

func (r *lockedSource) Int63() (n int64) {
	r.lk.Lock()
	n = r.src.Int63()
	r.lk.Unlock()
	return
}

func (r *lockedSource) Seed(seed int64) {
	r.lk.Lock()
	r.src.Seed(seed)
	r.lk.Unlock()
}

var seededRand = rand.New(&lockedSource{src: rand.NewSource(time.Now().UnixNano())})

const (
	DefaultRetryerMaxNumRetries = 3
	DefaultRetryerMinRetryDelay = 30 * time.Millisecond
	DefaultRetryerMaxRetryDelay = 300 * time.Second

	RetryOver = time.Duration(-1)
)

type Retryer struct {
	NumMaxRetries int
	MinRetryDelay time.Duration
	MaxRetryDelay time.Duration
}

func (d *Retryer) setRetryerDefaults() {
	if d.NumMaxRetries == 0 {
		d.NumMaxRetries = DefaultRetryerMaxNumRetries
	}
	if d.MinRetryDelay == 0 {
		d.MinRetryDelay = DefaultRetryerMinRetryDelay
	}
	if d.MaxRetryDelay == 0 {
		d.MaxRetryDelay = DefaultRetryerMaxRetryDelay
	}
}

func (d *Retryer) Do(ctx context.Context, operation func(ctx context.Context) error, isErrorRetryable func(err error) bool) (over bool, err error) {
	var retryCount int
	for {
		err := operation(ctx)
		if err == nil {
			return false, nil
		}

		if !isErrorRetryable(err) {
			return false, err
		}

		delay := d.NextBackOff(retryCount)
		if delay == RetryOver {
			return true, err
		}

		err = sleepWithContext(ctx, delay)
		if err != nil {
			return false, err
		}

		retryCount++
	}
}

func (d Retryer) MaxRetries() int {
	return d.NumMaxRetries
}

func (d Retryer) NextBackOff(retryCount int) time.Duration {

	if d.MaxRetries() <= retryCount {
		return RetryOver
	}

	d.setRetryerDefaults()

	minDelay := d.MinRetryDelay
	maxDelay := d.MaxRetryDelay

	var delay time.Duration

	actualRetryCount := int(math.Log2(float64(minDelay))) + 1
	if actualRetryCount < 63-retryCount {
		delay = time.Duration(1<<uint64(retryCount)) * getJitterDelay(minDelay)
		if delay > maxDelay {
			delay = getJitterDelay(maxDelay / 2)
		}
	} else {
		delay = getJitterDelay(maxDelay / 2)
	}
	return delay
}

func getJitterDelay(duration time.Duration) time.Duration {
	return time.Duration(seededRand.Int63n(int64(duration)) + int64(duration))
}

func sleepWithContext(ctx context.Context, dur time.Duration) error {
	t := time.NewTimer(dur)
	defer t.Stop()

	select {
	case <-t.C:
		break
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}
