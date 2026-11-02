package simkl

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func assertTestErr(t *testing.T, err error, wantErr string, wantErrIs error) {
	t.Helper()
	if wantErr != "" {
		if err == nil || !strings.Contains(err.Error(), wantErr) {
			t.Fatalf("err = %v, want error containing %q", err, wantErr)
		}
	} else if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wantErrIs != nil && !errors.Is(err, wantErrIs) {
		t.Fatalf("errors.Is(err, %v) = false; err: %v", wantErrIs, err)
	}
}

type pollLoopCase struct {
	name          string
	maxPolls      int
	bufferedTicks int
	cancelCtx     bool
	pollFn        func(callCount *int) func(context.Context) (string, bool, error)
	wantCalls     int
	wantToken     string
	wantErr       string
	wantErrIs     error
}

func runPollLoopTest(t *testing.T, tc pollLoopCase) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if tc.cancelCtx {
		cancel()
	}

	tick := make(chan time.Time, tc.bufferedTicks)
	for range tc.bufferedTicks {
		tick <- time.Now()
	}

	var calls int
	poll := tc.pollFn(&calls)

	token, err := pollLoop(ctx, tick, tc.maxPolls, poll)
	if calls != tc.wantCalls {
		t.Fatalf("call count = %d, want %d", calls, tc.wantCalls)
	}
	if token != tc.wantToken {
		t.Fatalf("token = %q, want %q", token, tc.wantToken)
	}
	assertTestErr(t, err, tc.wantErr, tc.wantErrIs)
}

func pollLoopCases() []pollLoopCase {
	return []pollLoopCase{
		{
			name:          "pending_forever",
			maxPolls:      3,
			bufferedTicks: 5,
			pollFn: func(callCount *int) func(context.Context) (string, bool, error) {
				return func(_ context.Context) (string, bool, error) {
					*callCount++
					return "", false, nil
				}
			},
			wantCalls: 3,
			wantErr:   "simkl: device code expired",
		},
		{
			name:          "success_on_second_tick",
			maxPolls:      5,
			bufferedTicks: 5,
			pollFn: func(callCount *int) func(context.Context) (string, bool, error) {
				return func(_ context.Context) (string, bool, error) {
					*callCount++
					if *callCount == 2 {
						return "secret-token", true, nil
					}
					return "", false, nil
				}
			},
			wantCalls: 2,
			wantToken: "secret-token",
		},
		{
			name:          "ctx_canceled",
			maxPolls:      3,
			bufferedTicks: 1,
			cancelCtx:     true,
			pollFn: func(callCount *int) func(context.Context) (string, bool, error) {
				return func(_ context.Context) (string, bool, error) {
					*callCount++
					return "", false, nil
				}
			},
			wantErr:   "simkl: device poll: context canceled",
			wantErrIs: context.Canceled,
		},
	}
}

func TestPollLoop(t *testing.T) {
	t.Parallel()

	for _, tc := range pollLoopCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runPollLoopTest(t, tc)
		})
	}
}

func TestPollLoop_ContextErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setup      func() (context.Context, func(), func(context.Context) (string, bool, error))
		wantErr    string
		wantErrIs  error
		wantNotExp bool
	}{
		{"parent_deadline_expired", func() (context.Context, func(), func(context.Context) (string, bool, error)) {
			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Hour))
			return ctx, cancel, func(context.Context) (string, bool, error) { return "", false, nil }
		}, "simkl: device poll", context.DeadlineExceeded, true},
		{"poll_ctx_expiry_cause", func() (context.Context, func(), func(context.Context) (string, bool, error)) {
			ctx, cancel := context.WithDeadlineCause(context.Background(), time.Now().Add(-time.Hour), ErrCodeExpired)
			return ctx, cancel, func(context.Context) (string, bool, error) { return "", false, nil }
		}, "simkl: device code expired", ErrCodeExpired, false},
		{"poll_fn_error_after_expiry_cause", func() (context.Context, func(), func(context.Context) (string, bool, error)) {
			ctx, cancel := context.WithCancelCause(context.Background())
			return ctx, func() { cancel(nil) }, func(context.Context) (string, bool, error) {
				cancel(ErrCodeExpired)
				return "", false, errors.New("dropped")
			}
		}, "simkl: device code expired", ErrCodeExpired, false},
		{"poll_fn_error_after_caller_cancel", func() (context.Context, func(), func(context.Context) (string, bool, error)) {
			ctx, cancel := context.WithCancel(context.Background())
			return ctx, cancel, func(context.Context) (string, bool, error) {
				cancel()
				return "", false, errors.New("dropped")
			}
		}, "simkl: device poll: context canceled", context.Canceled, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel, poll := tc.setup()
			defer cancel()

			tick := make(chan time.Time, 1)
			tick <- time.Now()

			token, err := pollLoop(ctx, tick, 1, poll)
			if token != "" {
				t.Fatalf("token = %q, want empty", token)
			}
			assertTestErr(t, err, tc.wantErr, tc.wantErrIs)
			if tc.wantNotExp && errors.Is(err, ErrCodeExpired) {
				t.Fatalf("expected error not to be ErrCodeExpired, got: %v", err)
			}
		})
	}
}
