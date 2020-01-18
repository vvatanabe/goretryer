package exponential

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetryNextBackOff(t *testing.T) {
	d := Retryer{
		NumMaxRetries: 100,
		MinRetryDelay: time.Second,
		MaxRetryDelay: 5 * time.Minute,
	}
	for i := 0; i < 100; i++ {
		a := d.NextBackOff(i)
		if a < time.Second {
			t.Errorf("retry delay should never be greater than one secounds received %s for retrycount %d", a, i)
		}
		if a > 5*time.Minute {
			t.Errorf("retry delay should never be greater than five minutes, received %s for retrycount %d", a, i)
		}
	}
	for i := 0; i < 100; i++ {
		a := d.NextBackOff(i)
		if a < time.Second {
			t.Errorf("retry delay should never be greater than one secounds received %s for retrycount %d", a, i)
		}
		if a > 5*time.Minute {
			t.Errorf("retry delay should never be greater than five minutes, received %s for retrycount %d", a, i)
		}
	}
	for i := 0; i < 100; i++ {
		a := d.NextBackOff(i)
		if a < time.Second {
			t.Errorf("retry delay should never be greater than one secounds received %s for retrycount %d", a, i)
		}
		if a > 5*time.Minute {
			t.Errorf("retry delay should never be greater than five minutes, received %s for retrycount %d", a, i)
		}
	}
}

func TestRetryer_Do(t *testing.T) {
	type fields struct {
		NumMaxRetries int
		MinRetryDelay time.Duration
		MaxRetryDelay time.Duration
	}
	tests := []struct {
		fields      fields
		operation   func(ctx context.Context) error
		wantDoCount int
		wantErr     bool
	}{
		{
			fields: fields{
				NumMaxRetries: 3,
				MinRetryDelay: time.Millisecond,
				MaxRetryDelay: time.Second,
			},
			operation: func(ctx context.Context) error {
				return errors.New("foo")
			},
			wantDoCount: 4,
			wantErr:     true,
		},
		{
			fields: fields{
				NumMaxRetries: 100,
				MinRetryDelay: time.Nanosecond,
				MaxRetryDelay: time.Millisecond,
			},
			operation: func(ctx context.Context) error {
				return errors.New("foo")
			},
			wantDoCount: 101,
			wantErr:     true,
		},
		{
			fields: fields{
				NumMaxRetries: 0,
				MinRetryDelay: time.Millisecond,
				MaxRetryDelay: time.Second,
			},
			operation: func(ctx context.Context) error {
				return errors.New("foo")
			},
			wantDoCount: 1,
			wantErr:     true,
		},
		{
			fields: fields{
				NumMaxRetries: 3,
				MinRetryDelay: time.Nanosecond,
				MaxRetryDelay: time.Second,
			},
			operation: func(ctx context.Context) error {
				return nil
			},
			wantDoCount: 1,
			wantErr:     false,
		},
		{
			fields: fields{
				NumMaxRetries: 3,
				MinRetryDelay: time.Nanosecond,
				MaxRetryDelay: time.Second,
			},
			operation: func(ctx context.Context) error {
				return errors.New("bar")
			},
			wantDoCount: 1,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			d := &Retryer{
				NumMaxRetries: tt.fields.NumMaxRetries,
				MinRetryDelay: tt.fields.MinRetryDelay,
				MaxRetryDelay: tt.fields.MaxRetryDelay,
			}
			var doCount int
			ctx := context.Background()
			if _, err := d.Do(ctx, func(ctx context.Context) error {
				doCount++
				return tt.operation(ctx)
			}, func(err error) bool {
				return err.Error() == "foo"
			}); (err != nil) != tt.wantErr {
				t.Errorf("Retryer.Do() error = %v, wantErr %v", err, tt.wantErr)
			}
			if doCount != tt.wantDoCount {
				t.Errorf("Retryer.Do() doCount = %v, wantDoCount %v", doCount, tt.wantDoCount)
			}
		})
	}
}

func Test_sleepWithContext(t *testing.T) {
	type args struct {
		ctx context.Context
		dur time.Duration
	}
	tests := []struct {
		args    args
		wantErr bool
	}{
		{
			args: args{
				ctx: context.Background(),
				dur: time.Millisecond * 300,
			},
			wantErr: false,
		},
		{
			args: args{
				ctx: func() context.Context {
					ctx, _ := context.WithTimeout(context.Background(), 100*time.Millisecond)
					return ctx
				}(),
				dur: time.Millisecond * 300,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if err := sleepWithContext(tt.args.ctx, tt.args.dur); (err != nil) != tt.wantErr {
				t.Errorf("sleepWithContext() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExponentialRetryer_setRetryerDefaults(t *testing.T) {
	tests := []struct {
		NumMaxRetries int
		MinRetryDelay time.Duration
		MaxRetryDelay time.Duration
	}{
		{
			NumMaxRetries: DefaultRetryerMaxNumRetries,
			MinRetryDelay: DefaultRetryerMinRetryDelay,
			MaxRetryDelay: DefaultRetryerMaxRetryDelay,
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			var r Retryer
			r.setRetryerDefaults()
			if r.NumMaxRetries != tt.NumMaxRetries {
				t.Errorf("Retryer NumMaxRetries value result = %v, want %v", r.NumMaxRetries, tt.NumMaxRetries)
			}
			if r.MinRetryDelay != tt.MinRetryDelay {
				t.Errorf("Retryer MinRetryDelay value result = %v, want %v", r.MinRetryDelay, tt.MinRetryDelay)
			}
			if r.MaxRetryDelay != tt.MaxRetryDelay {
				t.Errorf("Retryer MaxRetryDelay value result = %v, want %v", r.MaxRetryDelay, tt.MaxRetryDelay)
			}
		})
	}
}
