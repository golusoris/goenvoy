package main

import (
	"bytes"
	"context"
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
    "/x/{id}": {
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

func serveSpec(t *testing.T, body []byte) *httptest.Server {
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

func setupTestRoot(t *testing.T, code, allowContent string) string {
	t.Helper()
	root := t.TempDir()
	modDir := filepath.Join(root, "testmod")
	if err := os.MkdirAll(modDir, 0o750); err != nil {
		t.Fatalf("mkdir mod: %v", err)
	}
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
		if err := os.MkdirAll(upDir, 0o750); err != nil {
			t.Fatalf("mkdir upstream: %v", err)
		}
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
	name     string
	code     string
	allow    string
	module   string
	sha      string
	wantPart string
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
		if !strings.Contains(out.String(), "in spec ") {
			t.Errorf("summary missing from out: %q", out.String())
		}
		return
	}
	if err == nil || !strings.Contains(err.Error(), tt.wantPart) {
		t.Fatalf("run = %v, want error containing %q", err, tt.wantPart)
	}
}

func positiveCases(specSha string) []endpointCase {
	return []endpointCase{
		{
			name:     "(a) good path passes",
			code:     "func (c *Client) Fetch(ctx context.Context) error { return c.Get(ctx, \"/items\") }\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "",
		},
		{
			name:     "(d) Sprintf %d path matches {id}",
			code:     "func (c *Client) Fetch(ctx context.Context, id int) error { return c.Get(ctx, fmt.Sprintf(\"/items/%d\", id)) }\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "",
		},
		{
			name:     "(e) concat \"/x/\"+id matches",
			code:     "func (c *Client) Send(ctx context.Context, id string) error { return c.Post(ctx, \"/x/\" + id) }\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "",
		},
		{
			name:     "(f) query string stripped",
			code:     "func (c *Client) Search(ctx context.Context, q string) error { return c.Get(ctx, \"/search?q=\" + q) }\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "",
		},
		{
			name:     "(j) func with Get(\"/a\") then Post(\"/b\")",
			code:     "func (c *Client) Fetch(ctx context.Context) error {\n\tif err := c.Get(ctx, \"/a\"); err != nil {\n\t\treturn err\n\t}\n\treturn c.Post(ctx, \"/b\")\n}\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "",
		},
	}
}

func negativeBasicCases(specSha string) []endpointCase {
	return []endpointCase{
		{
			name:     "(b) wrong path refused",
			code:     "func (c *Client) Fetch(ctx context.Context) error { return c.Get(ctx, \"/wrong\") }\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "/wrong not in sample v1.2.3 spec",
		},
		{
			name:     "(c) wrong verb refused",
			code:     "func (c *Client) Fetch(ctx context.Context) error { return c.Post(ctx, \"/items\") }\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "POST /items not in sample v1.2.3 spec",
		},
		{
			name:     "(h) sha mismatch refused",
			code:     "func (c *Client) Fetch(ctx context.Context) error { return c.Get(ctx, \"/items\") }\n",
			module:   "testmod",
			sha:      strings.Repeat("0", 64),
			wantPart: "has sha256",
		},
		{
			name:     "(i) module with .. refused",
			code:     "func (c *Client) Fetch(ctx context.Context) error { return c.Get(ctx, \"/items\") }\n",
			module:   "../testmod",
			sha:      specSha,
			wantPart: "one clean path",
		},
	}
}

func negativeAllowCases(specSha string) []endpointCase {
	return []endpointCase{
		{
			name:     "(g) expired allow entry fails",
			code:     "func (c *Client) Old(ctx context.Context) error { return c.Delete(ctx, \"/not-in-spec\") }\n",
			allow:    "allow:\n  - module: testmod\n    func: Old\n    reason: test\n    expires: 2020-01-01\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "DELETE /not-in-spec not in sample v1.2.3 spec",
		},
		{
			name:     "(k) unknown key refused",
			code:     "func (c *Client) Fetch(ctx context.Context) error { return c.Get(ctx, \"/items\") }\n",
			allow:    "allow:\n  - module: testmod\n    func: Fetch\n    reason: test\n    expires: 2027-01-01\n    unknown_key: bad\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "unknown_key",
		},
		{
			name:     "(l) duplicate entry refused",
			code:     "func (c *Client) Fetch(ctx context.Context) error { return c.Get(ctx, \"/items\") }\n",
			allow:    "allow:\n  - module: testmod\n    func: Fetch\n    reason: test 1\n    expires: 2027-01-01\n  - module: testmod\n    func: Fetch\n    reason: duplicate\n    expires: 2027-01-01\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "duplicate",
		},
		{
			name:     "(m) missing reason refused",
			code:     "func (c *Client) Fetch(ctx context.Context) error { return c.Get(ctx, \"/items\") }\n",
			allow:    "allow:\n  - module: testmod\n    func: Fetch\n    expires: 2027-01-01\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "reason",
		},
		{
			name:     "(n) malformed expiry refused",
			code:     "func (c *Client) Fetch(ctx context.Context) error { return c.Get(ctx, \"/items\") }\n",
			allow:    "allow:\n  - module: testmod\n    func: Fetch\n    reason: test\n    expires: 2027-13-45\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "expires",
		},
		{
			name:     "(o) stale entry refused",
			code:     "func (c *Client) Fetch(ctx context.Context) error { return c.Get(ctx, \"/items\") }\n",
			allow:    "allow:\n  - module: testmod\n    func: Fetch\n    reason: test\n    expires: 2027-01-01\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "matches no off-spec call; remove it",
		},
	}
}

func negativeCases(specSha string) []endpointCase {
	return append(negativeBasicCases(specSha), negativeAllowCases(specSha)...)
}

func endpointTestCases(specSha string) []endpointCase {
	return append(positiveCases(specSha), negativeCases(specSha)...)
}

func TestEndpointsTableDriven(t *testing.T) {
	t.Parallel()
	specBytes := []byte(fixtureSpec)
	server := serveSpec(t, specBytes)

	for _, tt := range endpointTestCases(digest(specBytes)) {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			runEndpointCase(t, tt, server.URL)
		})
	}
}

func TestEndpointsYAML(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	manifestFile := filepath.Join(root, "specs.yaml")

	yamlEntry := entry(map[string]string{
		"path":   "path: spec.yml",
		"module": "module: testmod",
	})
	writeFile(t, manifestFile, yamlEntry)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var out bytes.Buffer
	err := run(ctx, []string{"endpoints", "-manifest", manifestFile, "-root", root}, &out, "http://127.0.0.1")
	if err == nil || !strings.Contains(err.Error(), "yaml not supported") {
		t.Fatalf("err = %v, want yaml not supported", err)
	}

	// Without module: skipped with printed line
	yamlEntryNoMod := entry(map[string]string{
		"path":   "path: spec.yml",
		"module": "",
	})
	writeFile(t, manifestFile, yamlEntryNoMod)
	out.Reset()
	err = run(ctx, []string{"endpoints", "-manifest", manifestFile, "-root", root}, &out, "http://127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "skipped: yaml spec") {
		t.Errorf("out = %q, want 'skipped: yaml spec'", out.String())
	}
}

func TestEndpointsUnknownVerb(t *testing.T) {
	t.Parallel()
	specBytes := []byte(fixtureSpec)
	specSha := digest(specBytes)
	server := serveSpec(t, specBytes)

	root := setupTestRoot(t, "func (c *Client) NoVerb(ctx context.Context) {\n\t_ = \"/items\"\n}\n", "")
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
	server := serveSpec(t, specBytes)

	code := "func (c *Client) Old(ctx context.Context) error { return c.Delete(ctx, \"/not-in-spec\") }\n"
	allow := "allow:\n  - module: testmod\n    func: Old\n    reason: test\n    expires: 2099-01-01\n"
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
