package trakt

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golusoris/goenvoy/metadata"
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
	pollFn func(callCount *int) func(context.Context) (*Token, bool, error),
	wantCalls int,
	wantToken *Token,
	wantErr string,
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
	if wantToken != nil {
		if token == nil || token.AccessToken != wantToken.AccessToken {
			t.Fatalf("token = %v, want %v", token, wantToken)
		}
	} else if token != nil {
		t.Fatalf("token = %v, want nil", token)
	}
	assertTestErr(t, err, wantErr, wantErrIs)
}

func TestPollLoop(t *testing.T) {
	t.Parallel()

	t.Run("pending_forever", func(t *testing.T) {
		t.Parallel()
		runPollLoopTest(t, 3, 5, false, func(callCount *int) func(context.Context) (*Token, bool, error) {
			return func(_ context.Context) (*Token, bool, error) {
				*callCount++
				return nil, false, nil
			}
		}, 3, nil, "trakt: device code expired", nil)
	})

	t.Run("success_on_second_tick", func(t *testing.T) {
		t.Parallel()
		runPollLoopTest(t, 5, 5, false, func(callCount *int) func(context.Context) (*Token, bool, error) {
			return func(_ context.Context) (*Token, bool, error) {
				*callCount++
				if *callCount == 2 {
					return &Token{AccessToken: "secret-token"}, true, nil
				}
				return nil, false, nil
			}
		}, 2, &Token{AccessToken: "secret-token"}, "", nil)
	})

	t.Run("ctx_canceled", func(t *testing.T) {
		t.Parallel()
		runPollLoopTest(t, 3, 1, true, func(callCount *int) func(context.Context) (*Token, bool, error) {
			return func(_ context.Context) (*Token, bool, error) {
				*callCount++
				return nil, false, nil
			}
		}, 0, nil, "trakt: device poll: context canceled", context.Canceled)
	})
}

func TestPollLoop_ContextErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setup      func() (context.Context, func(), func(context.Context) (*Token, bool, error))
		wantErr    string
		wantErrIs  error
		wantNotExp bool
	}{
		{"parent_deadline_expired", func() (context.Context, func(), func(context.Context) (*Token, bool, error)) {
			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Hour))
			return ctx, cancel, func(context.Context) (*Token, bool, error) { return nil, false, nil }
		}, "trakt: device poll", context.DeadlineExceeded, true},
		{"poll_ctx_expiry_cause", func() (context.Context, func(), func(context.Context) (*Token, bool, error)) {
			ctx, cancel := context.WithDeadlineCause(context.Background(), time.Now().Add(-time.Hour), errCodeExpired)
			return ctx, cancel, func(context.Context) (*Token, bool, error) { return nil, false, nil }
		}, "trakt: device code expired", errCodeExpired, false},
		{"poll_fn_error_after_expiry_cause", func() (context.Context, func(), func(context.Context) (*Token, bool, error)) {
			ctx, cancel := context.WithCancelCause(context.Background())
			return ctx, func() { cancel(nil) }, func(context.Context) (*Token, bool, error) {
				cancel(errCodeExpired)
				return nil, false, errors.New("dropped")
			}
		}, "trakt: device code expired", errCodeExpired, false},
		{"poll_fn_error_after_caller_cancel", func() (context.Context, func(), func(context.Context) (*Token, bool, error)) {
			ctx, cancel := context.WithCancel(context.Background())
			return ctx, cancel, func(context.Context) (*Token, bool, error) {
				cancel()
				return nil, false, errors.New("dropped")
			}
		}, "trakt: device poll: context canceled", context.Canceled, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel, poll := tc.setup()
			defer cancel()

			tick := make(chan time.Time, 1)
			tick <- time.Now()

			token, err := pollLoop(ctx, tick, 1, poll)
			if token != nil {
				t.Fatalf("token = %v, want nil", token)
			}
			assertTestErr(t, err, tc.wantErr, tc.wantErrIs)
			if tc.wantNotExp && errors.Is(err, errCodeExpired) {
				t.Fatalf("expected error not to be errCodeExpired, got: %v", err)
			}
		})
	}
}

func TestPollDeviceToken_Retry429Floor(t *testing.T) {
	t.Parallel()

	var reqCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := atomic.AddInt32(&reqCount, 1)
		if n == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"slow_down"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"valid-token","token_type":"bearer","expires_in":7200,"refresh_token":"valid-refresh","scope":"public","created_at":1000}`))
	}))
	t.Cleanup(srv.Close)

	c := New("test-client", metadata.WithBaseURL(srv.URL))
	var slept time.Duration
	c.pollSleep = func(d time.Duration) {
		slept = d
	}

	code := &DeviceCode{
		DeviceCode: "test-code",
		Interval:   0,
		ExpiresIn:  10,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := c.PollDeviceToken(ctx, code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if slept < 1*time.Second {
		t.Fatalf("expected 429 retry back-off >= 1s, got %v", slept)
	}
}
