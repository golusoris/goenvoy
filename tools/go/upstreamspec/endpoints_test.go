package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const fixtureSpec = `{
  "openapi": "3.0.0",
  "paths": {
    "/items": {
      "get": {}
    },
    "/items/{id}": {
      "get": {}
    },
    "/items-{id}": {
      "get": {}
    },
    "/x/{id}": {
      "post": {}
    },
    "/x/{id}.jpg": {
      "post": {}
    },
    "/search": {
      "get": {}
    },
    "/allowed": {
      "delete": {}
    },
    "/a": {
      "get": {}
    },
    "/b": {
      "post": {}
    }
  }
}`

const fixtureYAMLSpec = `openapi: 3.0.0
paths:
  /items:
    get: {}
  /items/{id}:
    get: {}
`

// mkdirOwner creates dir, whose parent must exist, readable only by the owner.
// Why: t.TempDir creates owner-only (0o700) directories; renaming keeps that mode.
func mkdirOwner(t *testing.T, dir string) {
	t.Helper()
	if err := os.Rename(t.TempDir(), dir); err != nil {
		t.Fatalf("mkdirOwner %s: rename: %v", dir, err)
	}
}

func serveSpec(t *testing.T, body []byte, contentType string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/Owner/Project/"+someCommit+"/src/") {
			http.NotFound(w, r)
			return
		}
		ct := contentType
		if ct == "" {
			ct = "application/json"
		}
		w.Header().Set("Content-Type", ct)
		if _, err := io.Copy(w, bytes.NewReader(body)); err != nil {
			t.Errorf("serveSpec: write fixture: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func setupTestRoot(t *testing.T, code, allowContent string) string {
	t.Helper()
	root := t.TempDir()
	modDir := filepath.Join(root, "testmod")
	mkdirOwner(t, modDir)
	if code != "" {
		src := "package testmod\n\nimport (\n\t\"context\"\n\t\"fmt\"\n)\n\ntype Client struct{}\n\n" +
			"func (c *Client) Get(ctx context.Context, p string) error { return nil }\n" +
			"func (c *Client) Post(ctx context.Context, p string) error { return nil }\n" +
			"func (c *Client) Delete(ctx context.Context, p string) error { return nil }\n\n" +
			code
		if err := os.WriteFile(filepath.Join(modDir, "client.go"), []byte(src), 0o600); err != nil {
			t.Fatalf("write client.go: %v", err)
		}
	}
	if allowContent != "" {
		upDir := filepath.Join(root, "upstream")
		mkdirOwner(t, upDir)
		if err := os.WriteFile(filepath.Join(upDir, "endpoints-allow.yaml"), []byte(allowContent), 0o600); err != nil {
			t.Fatalf("write allow.yaml: %v", err)
		}
	}
	return root
}

func writeManifest(t *testing.T, file, sha, module string) {
	t.Helper()
	entryContent := entry(map[string]string{
		"prefix": "",
		"sha256": "sha256: " + sha,
		"module": "module: " + module,
	})
	writeFile(t, file, entryContent)
}

type endpointCase struct {
	name        string
	code        string
	allow       string
	module      string
	sha         string
	wantPart    string
	wantCodeOps int
	wantInSpec  int
}

func runEndpointCase(t *testing.T, tt endpointCase, serverURL string) {
	t.Helper()
	root := setupTestRoot(t, tt.code, tt.allow)
	manifestFile := filepath.Join(root, "specs.yaml")
	writeManifest(t, manifestFile, tt.sha, tt.module)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var out bytes.Buffer
	args := []string{"endpoints", "-manifest", manifestFile, "-root", root}
	err := run(ctx, args, &out, serverURL)

	if tt.wantPart == "" {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		wantSummary := fmt.Sprintf("code ops %d, in spec %d", tt.wantCodeOps, tt.wantInSpec)
		if !strings.Contains(out.String(), wantSummary) {
			t.Errorf("summary got %q, want %q", out.String(), wantSummary)
		}
		return
	}
	if err == nil || !strings.Contains(err.Error(), tt.wantPart) {
		t.Fatalf("run = %v, want error containing %q", err, tt.wantPart)
	}
}

func TestEndpointsTableDriven(t *testing.T) {
	t.Parallel()
	specBytes := []byte(fixtureSpec)
	server := serveSpec(t, specBytes, "application/json")

	for _, tt := range endpointTestCases(digest(specBytes)) {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			runEndpointCase(t, tt, server.URL)
		})
	}
}

func TestEndpointsNestedPackages(t *testing.T) {
	t.Parallel()
	specBytes := []byte(fixtureSpec)
	server := serveSpec(t, specBytes, "application/json")

	root := setupTestRoot(t, "", "")
	subDir := filepath.Join(root, "testmod", "sub")
	mkdirOwner(t, subDir)
	src := "package sub\n\nimport \"context\"\n\ntype Client struct{}\n" +
		"func (c *Client) Sub(ctx context.Context) error { return c.Get(ctx, \"/items\") }\n" +
		"func (c *Client) Get(ctx context.Context, p string) error { return nil }\n"
	if err := os.WriteFile(filepath.Join(subDir, "sub.go"), []byte(src), 0o600); err != nil {
		t.Fatalf("write sub.go: %v", err)
	}
	manifestFile := filepath.Join(root, "specs.yaml")
	writeManifest(t, manifestFile, digest(specBytes), "testmod")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var out bytes.Buffer
	err := run(ctx, []string{"endpoints", "-manifest", manifestFile, "-root", root}, &out, server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "code ops 1, in spec 1") {
		t.Errorf("summary = %q, want 'code ops 1, in spec 1'", out.String())
	}
}

func TestEndpointsYAMLSpec(t *testing.T) {
	t.Parallel()
	yamlBytes := []byte(fixtureYAMLSpec)
	server := serveSpec(t, yamlBytes, "application/yaml")

	root := setupTestRoot(t, "func (c *Client) Fetch(ctx context.Context) error { return c.Get(ctx, \"/items\") }\n", "")
	manifestFile := filepath.Join(root, "specs.yaml")
	entryContent := entry(map[string]string{
		"prefix": "",
		"path":   "path: src/openapi.yaml",
		"sha256": "sha256: " + digest(yamlBytes),
		"module": "module: testmod",
	})
	writeFile(t, manifestFile, entryContent)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var out bytes.Buffer
	err := run(ctx, []string{"endpoints", "-manifest", manifestFile, "-root", root}, &out, server.URL)
	if err != nil {
		t.Fatalf("unexpected error with yaml spec: %v", err)
	}
	if !strings.Contains(out.String(), "code ops 1, in spec 1") {
		t.Errorf("out = %q, want 'code ops 1, in spec 1'", out.String())
	}
}

func TestEndpointsUnknownVerb(t *testing.T) {
	t.Parallel()
	specBytes := []byte(fixtureSpec)
	specSha := digest(specBytes)
	server := serveSpec(t, specBytes, "application/json")

	code := "func (c *Client) Custom(ctx context.Context, p string) error { return nil }\n" +
		"func (c *Client) Call(ctx context.Context) error { return c.Custom(ctx, \"/items\") }\n"
	root := setupTestRoot(t, code, "")
	manifestFile := filepath.Join(root, "specs.yaml")
	writeManifest(t, manifestFile, specSha, "testmod")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var out bytes.Buffer
	err := run(ctx, []string{"endpoints", "-manifest", manifestFile, "-root", root}, &out, server.URL)
	if err == nil || !strings.Contains(err.Error(), "verb unknown") {
		t.Fatalf("err = %v, want error containing 'verb unknown'", err)
	}
}

func TestEndpointsUnexpiredAllowPasses(t *testing.T) {
	t.Parallel()
	specBytes := []byte(fixtureSpec)
	specSha := digest(specBytes)
	server := serveSpec(t, specBytes, "application/json")

	code := "func (c *Client) Old(ctx context.Context) error { return c.Delete(ctx, \"/not-in-spec\") }\n"
	allow := "allow:\n  - module: testmod\n    func: Client.Old\n    method: DELETE\n    path: /not-in-spec\n    reason: test\n    expires: 2099-01-01\n"
	root := setupTestRoot(t, code, allow)
	manifestFile := filepath.Join(root, "specs.yaml")
	writeManifest(t, manifestFile, specSha, "testmod")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var out bytes.Buffer
	err := run(ctx, []string{"endpoints", "-manifest", manifestFile, "-root", root}, &out, server.URL)
	if err != nil {
		t.Fatalf("unexpected error with valid allow: %v", err)
	}
}

// TestEndpointsExpiredAllowReportedOnce verifies that multiple off-spec operations matching one expired allow entry emit the expiry finding only once.
func TestEndpointsExpiredAllowReportedOnce(t *testing.T) {
	t.Parallel()
	specBytes := []byte(fixtureSpec)
	specSha := digest(specBytes)
	server := serveSpec(t, specBytes, "application/json")

	code := "func (c *Client) Old(ctx context.Context) error {\n" +
		"\t_ = c.Delete(ctx, \"/not-in-spec\")\n" +
		"\treturn c.Delete(ctx, \"/not-in-spec\")\n" +
		"}\n"
	allow := "allow:\n  - module: testmod\n    func: Client.Old\n    method: DELETE\n    path: /not-in-spec\n    reason: test\n    expires: 2020-01-01\n"
	root := setupTestRoot(t, code, allow)
	manifestFile := filepath.Join(root, "specs.yaml")
	writeManifest(t, manifestFile, specSha, "testmod")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var out bytes.Buffer
	err := run(ctx, []string{"endpoints", "-manifest", manifestFile, "-root", root}, &out, server.URL)
	if err == nil {
		t.Fatal("expected error with expired allow entry, got nil")
	}
	if got := strings.Count(err.Error(), "allow entry testmod Client.Old expired 2020-01-01"); got != 1 {
		t.Errorf("expired entry count = %d, want 1; err: %v", got, err)
	}
	if got := strings.Count(err.Error(), "DELETE /not-in-spec not in sample v1.2.3 spec"); got != 2 {
		t.Errorf("off-spec delete count = %d, want 2; err: %v", got, err)
	}
}

// TestReadAllowListEmpty verifies that an empty allow file returns no error and zero entries.
func TestReadAllowListEmpty(t *testing.T) {
	t.Parallel()
	allowFile := filepath.Join(t.TempDir(), "endpoints-allow.yaml")
	if err := os.WriteFile(allowFile, []byte{}, 0o600); err != nil {
		t.Fatalf("write allow file: %v", err)
	}
	entries, err := readAllowList(allowFile)
	if err != nil {
		t.Fatalf("readAllowList unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("readAllowList returned %d entries, want 0", len(entries))
	}
}
