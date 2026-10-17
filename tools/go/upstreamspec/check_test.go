package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const sampleSpec = `{"openapi": "3.0.1", "paths": {}}`

// serve answers the address of the sample pin with body and every other address with 404.
func serve(t *testing.T, body []byte) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Owner/Project/"+someCommit+"/src/openapi.json" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(body)
	}))
	t.Cleanup(server.Close)
	return server
}

func samplePin(sha string) pin {
	return pin{
		Name: "sample", Repository: "Owner/Project", Tag: "v1", Commit: someCommit,
		Path: "src/openapi.json", SHA256: sha, License: "MIT",
	}
}

func check(t *testing.T, base string, pins ...pin) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var out bytes.Buffer
	err := checker{client: &http.Client{Timeout: 10 * time.Second}, base: base, out: &out}.all(ctx, pins)
	return out.String(), err
}

func TestCheckAcceptsThePinnedFile(t *testing.T) {
	t.Parallel()
	body := []byte(sampleSpec)
	server := serve(t, body)
	out, err := check(t, server.URL, samplePin(digest(body)))
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if want := "upstreamspec: sample v1 (0123456789ab): sha256 matches\n"; out != want {
		t.Errorf("summary = %q, want %q", out, want)
	}
}

// TestCheckRefuses plants each way a pin can go stale: the file changed at the commit (another
// sha256), the commit or path no longer serves it, and an oversized reply.
func TestCheckRefuses(t *testing.T) {
	t.Parallel()
	body := []byte(sampleSpec)
	server := serve(t, body)
	moved := samplePin(digest(body))
	moved.Path = "src/moved.json"
	otherCommit := samplePin(digest(body))
	otherCommit.Commit = strings.Repeat("f", 40)
	huge := []byte(strings.Repeat(" ", maxSpecBytes+1))
	hugeServer := serve(t, huge)
	tests := []struct {
		name string
		base string
		pin  pin
		want string
	}{
		{"another sha256", server.URL, samplePin(someSHA256), "has sha256 " + digest(body) + ", the manifest pins " + someSHA256},
		{"a path the commit does not have", server.URL, moved, "HTTP 404"},
		{"a commit that does not serve the file", server.URL, otherCommit, "HTTP 404"},
		{"a file above the size bound", hugeServer.URL, samplePin(digest(huge)), "larger than"},
		{"an unreachable host", "http://127.0.0.1:1", samplePin(digest(body)), "get http://127.0.0.1:1/"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			out, err := check(t, test.base, test.pin)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("check = %v, want an error with %q", err, test.want)
			}
			if out != "" {
				t.Errorf("a refused pin was reported as matching: %q", out)
			}
		})
	}
}

// TestCheckNamesEveryStalePin checks that one run reports each failing pin, not only the first.
func TestCheckNamesEveryStalePin(t *testing.T) {
	t.Parallel()
	body := []byte(sampleSpec)
	server := serve(t, body)
	good := samplePin(digest(body))
	first, second := samplePin(someSHA256), samplePin(someSHA256)
	first.Name, second.Name = "first", "second"
	out, err := check(t, server.URL, first, good, second)
	if err == nil || !strings.Contains(err.Error(), "first: ") || !strings.Contains(err.Error(), "second: ") {
		t.Fatalf("check = %v, want errors for first and second", err)
	}
	if !strings.Contains(out, "sample v1") {
		t.Errorf("the matching pin was not reported: %q", out)
	}
}

func TestRun(t *testing.T) {
	t.Parallel()
	body := []byte(sampleSpec)
	server := serve(t, body)
	manifest := filepath.Join(t.TempDir(), "specs.yaml")
	writeFile(t, manifest, strings.Replace(entry(nil), someSHA256, digest(body), 1))
	broken := filepath.Join(t.TempDir(), "specs.yaml")
	writeFile(t, broken, entry(map[string]string{"commit": "commit: main"}))
	tests := []struct {
		name string
		args []string
		want string // part of the error; empty for none
	}{
		{"check", []string{"check", "-manifest", manifest}, ""},
		{"no command", nil, "usage:"},
		{"another command", []string{"fetch", "-manifest", manifest}, "usage:"},
		{"an unknown flag", []string{"check", "-cache", "x"}, "usage:"},
		{"an extra argument", []string{"check", "-manifest", manifest, "extra"}, "usage:"},
		{"a malformed manifest", []string{"check", "-manifest", broken}, "full 40-digit commit id"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			var out bytes.Buffer
			err := run(ctx, test.args, &out, server.URL)
			switch {
			case test.want == "" && err != nil:
				t.Fatalf("run: %v", err)
			case test.want != "" && (err == nil || !strings.Contains(err.Error(), test.want)):
				t.Fatalf("run = %v, want an error with %q", err, test.want)
			}
		})
	}
}
