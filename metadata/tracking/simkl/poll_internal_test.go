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

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
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
