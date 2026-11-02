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

func runPollLoopTest(
	t *testing.T,
	maxPolls, bufferedTicks int,
	cancelCtx bool,
	pollFn func(callCount *int) func(context.Context) (string, bool, error),
	wantCalls int,
	wantToken, wantErr string,
	wantErrIs error,
) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if cancelCtx {
		cancel()
	}

	tick := make(chan time.Time, bufferedTicks)
	for range bufferedTicks {
		tick <- time.Now()
	}

	var calls int
	poll := pollFn(&calls)

	token, err := pollLoop(ctx, tick, maxPolls, poll)
	if calls != wantCalls {
		t.Fatalf("call count = %d, want %d", calls, wantCalls)
	}
	if token != wantToken {
		t.Fatalf("token = %q, want %q", token, wantToken)
	}
	assertTestErr(t, err, wantErr, wantErrIs)
}

func TestPollLoop(t *testing.T) {
	t.Parallel()

	t.Run("pending_forever", func(t *testing.T) {
		t.Parallel()
		runPollLoopTest(t, 3, 5, false, func(callCount *int) func(context.Context) (string, bool, error) {
			return func(_ context.Context) (string, bool, error) {
				*callCount++
				return "", false, nil
			}
		}, 3, "", "simkl: device code expired", nil)
	})

	t.Run("success_on_second_tick", func(t *testing.T) {
		t.Parallel()
		runPollLoopTest(t, 5, 5, false, func(callCount *int) func(context.Context) (string, bool, error) {
			return func(_ context.Context) (string, bool, error) {
				*callCount++
				if *callCount == 2 {
					return "secret-token", true, nil
				}
				return "", false, nil
			}
		}, 2, "secret-token", "", nil)
	})

	t.Run("ctx_canceled", func(t *testing.T) {
		t.Parallel()
		runPollLoopTest(t, 3, 1, true, func(callCount *int) func(context.Context) (string, bool, error) {
			return func(_ context.Context) (string, bool, error) {
				*callCount++
				return "", false, nil
			}
		}, 0, "", "simkl: device poll: context canceled", context.Canceled)
	})
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
			ctx, cancel := context.WithDeadlineCause(context.Background(), time.Now().Add(-time.Hour), errCodeExpired)
			return ctx, cancel, func(context.Context) (string, bool, error) { return "", false, nil }
		}, "simkl: device code expired", errCodeExpired, false},
		{"poll_fn_error_after_expiry_cause", func() (context.Context, func(), func(context.Context) (string, bool, error)) {
			ctx, cancel := context.WithCancelCause(context.Background())
			return ctx, func() { cancel(nil) }, func(context.Context) (string, bool, error) {
				cancel(errCodeExpired)
				return "", false, errors.New("dropped")
			}
		}, "simkl: device code expired", errCodeExpired, false},
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
			if tc.wantNotExp && errors.Is(err, errCodeExpired) {
				t.Fatalf("expected error not to be errCodeExpired, got: %v", err)
			}
		})
	}
}
