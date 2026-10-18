package metadata_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/golusoris/goenvoy/metadata"
)

type tokenCB func(access string)

func TestOAuthState_ZeroValue(t *testing.T) {
	t.Parallel()
	var s metadata.OAuthState[tokenCB]
	if s.ClientSecret() != "" || s.AccessToken() != "" || s.RefreshToken() != "" {
		t.Fatal("zero OAuthState holds values")
	}
	if cb := s.StoreAccessToken("a"); cb != nil {
		t.Fatal("zero OAuthState returned a callback")
	}
}

func TestOAuthState_SetAndRead(t *testing.T) {
	t.Parallel()
	var s metadata.OAuthState[tokenCB]
	s.SetClientSecret("sec")
	s.SetAccessToken("acc")
	s.SetRefreshToken("ref")
	if s.ClientSecret() != "sec" || s.AccessToken() != "acc" || s.RefreshToken() != "ref" {
		t.Fatalf("got %q %q %q", s.ClientSecret(), s.AccessToken(), s.RefreshToken())
	}
	if rt, cs := s.RefreshCredentials(); rt != "ref" || cs != "sec" {
		t.Fatalf("RefreshCredentials = %q, %q", rt, cs)
	}
}

func TestOAuthState_StoreReturnsCallback(t *testing.T) {
	t.Parallel()
	var s metadata.OAuthState[tokenCB]
	var got string
	s.SetTokenCallback(func(access string) { got = access })
	s.StoreAccessToken("a1")("a1")
	if got != "a1" || s.AccessToken() != "a1" {
		t.Fatalf("StoreAccessToken: callback %q, token %q", got, s.AccessToken())
	}
	cb := s.StoreTokens("a2", "r2")
	cb("a2")
	if got != "a2" || s.AccessToken() != "a2" || s.RefreshToken() != "r2" {
		t.Fatalf("StoreTokens: callback %q, tokens %q %q", got, s.AccessToken(), s.RefreshToken())
	}
}

func TestOAuthState_Concurrent(t *testing.T) {
	t.Parallel()
	var s metadata.OAuthState[tokenCB]
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			s.StoreTokens("a", "r")
			_ = s.AccessToken()
			_, _ = s.RefreshCredentials()
		})
	}
	wg.Wait()
}

type answer struct {
	N int `json:"n"`
}

func TestPostAs(t *testing.T) {
	t.Parallel()
	var gotPath string
	var gotBody any
	post := func(_ context.Context, path string, body, dst any) error {
		gotPath, gotBody = path, body
		dst.(*answer).N = 7
		return nil
	}
	out, err := metadata.PostAs[answer](context.Background(), post, "/sync/history", "items")
	if err != nil || out.N != 7 || gotPath != "/sync/history" || gotBody != "items" {
		t.Fatalf("PostAs = %+v, %v; path %q body %v", out, err, gotPath, gotBody)
	}
}

func TestPostAs_Error(t *testing.T) {
	t.Parallel()
	want := errors.New("boom")
	post := func(context.Context, string, any, any) error { return want }
	out, err := metadata.PostAs[answer](context.Background(), post, "/x", nil)
	if !errors.Is(err, want) || out != nil {
		t.Fatalf("PostAs = %+v, %v; want nil and the post error", out, err)
	}
}

func TestPageParams(t *testing.T) {
	t.Parallel()
	cases := []struct {
		page, limit int
		want        string
	}{
		{0, 0, ""},
		{-1, -5, ""},
		{1, 0, "page=1"},
		{0, 10, "limit=10"},
		{2, 50, "limit=50&page=2"},
	}
	for _, tc := range cases {
		if got := metadata.PageParams(tc.page, tc.limit).Encode(); got != tc.want {
			t.Errorf("PageParams(%d, %d) = %q; want %q", tc.page, tc.limit, got, tc.want)
		}
	}
}
