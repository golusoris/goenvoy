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

func positiveCases(specSha string) []endpointCase {
	return append(positiveBasicCases(specSha), positivePlantedCases(specSha)...)
}

func positiveBasicCases(specSha string) []endpointCase {
	return []endpointCase{
		{
			name:        "(a) good path passes",
			code:        "func (c *Client) Fetch(ctx context.Context) error { return c.Get(ctx, \"/items\") }\n",
			module:      "testmod",
			sha:         specSha,
			wantCodeOps: 1,
			wantInSpec:  1,
		},
		{
			name:        "(d) Sprintf %d path matches {id}",
			code:        "func (c *Client) Fetch(ctx context.Context, id int) error { return c.Get(ctx, fmt.Sprintf(\"/items/%d\", id)) }\n",
			module:      "testmod",
			sha:         specSha,
			wantCodeOps: 1,
			wantInSpec:  1,
		},
		{
			name:        "(e) concat \"/x/\"+id matches",
			code:        "func (c *Client) Send(ctx context.Context, id string) error { return c.Post(ctx, \"/x/\" + id) }\n",
			module:      "testmod",
			sha:         specSha,
			wantCodeOps: 1,
			wantInSpec:  1,
		},
		{
			name:        "(f) query string stripped",
			code:        "func (c *Client) Search(ctx context.Context, q string) error { return c.Get(ctx, \"/search?q=\" + q) }\n",
			module:      "testmod",
			sha:         specSha,
			wantCodeOps: 1,
			wantInSpec:  1,
		},
		{
			name:        "(j) func with Get(\"/a\") then Post(\"/b\")",
			code:        "func (c *Client) Fetch(ctx context.Context) error {\n\tif err := c.Get(ctx, \"/a\"); err != nil {\n\t\treturn err\n\t}\n\treturn c.Post(ctx, \"/b\")\n}\n",
			module:      "testmod",
			sha:         specSha,
			wantCodeOps: 2,
			wantInSpec:  2,
		},
	}
}

func positivePlantedCases(specSha string) []endpointCase {
	return []endpointCase{
		{
			name:        "(1) unused literal not reaching request call is ignored",
			code:        "func (c *Client) Fetch(ctx context.Context) error {\n\t_ = \"/ignored\"\n\treturn c.Get(ctx, \"/items\")\n}\n",
			module:      "testmod",
			sha:         specSha,
			wantCodeOps: 1,
			wantInSpec:  1,
		},
		{
			name:        "(4) embedded mid-segment param matches",
			code:        "func (c *Client) Fetch(ctx context.Context, id string) error { return c.Get(ctx, \"/items-\" + id) }\n",
			module:      "testmod",
			sha:         specSha,
			wantCodeOps: 1,
			wantInSpec:  1,
		},
		{
			name:        "(4b) embedded param in extension matches",
			code:        "func (c *Client) Send(ctx context.Context, id string) error { return c.Post(ctx, \"/x/\" + id + \".jpg\") }\n",
			module:      "testmod",
			sha:         specSha,
			wantCodeOps: 1,
			wantInSpec:  1,
		},
		{
			name:        "(5) query concat cuts at question mark",
			code:        "func (c *Client) Search(ctx context.Context, q string) error { return c.Get(ctx, \"/search?q=\" + q + \"&limit=10\") }\n",
			module:      "testmod",
			sha:         specSha,
			wantCodeOps: 1,
			wantInSpec:  1,
		},
		{
			name:        "(5b) query suffix appended with += keeps the path resolved",
			code:        "func (c *Client) Search(ctx context.Context, q string) error { path := \"/search\"; if q != \"\" { path += \"?q=\" + q }; return c.Get(ctx, path) }\n",
			module:      "testmod",
			sha:         specSha,
			wantCodeOps: 1,
			wantInSpec:  1,
		},
		{
			name:        "(8) allow entry matches module func method path",
			code:        "func (c *Client) Old(ctx context.Context) error { return c.Delete(ctx, \"/not-in-spec\") }\n",
			allow:       "allow:\n  - module: testmod\n    func: Client.Old\n    method: DELETE\n    path: /not-in-spec\n    reason: test\n    expires: 2099-01-01\n",
			module:      "testmod",
			sha:         specSha,
			wantCodeOps: 1,
			wantInSpec:  0,
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
		{
			name:     "(2) unresolvable ident path fails",
			code:     "func (c *Client) Fetch(ctx context.Context, p string) error { return c.Get(ctx, p) }\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "path unresolved",
		},
		{
			name:     "(2b) path segment appended with += fails",
			code:     "func (c *Client) Fetch(ctx context.Context) error { path := \"/search\"; path += \"/more\"; return c.Get(ctx, path) }\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "path unresolved",
		},
		{
			name:     "(3) dynamic prefix concat fails",
			code:     "func (c *Client) Fetch(ctx context.Context, p string) error { return c.Get(ctx, p + \"/items\") }\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "path unresolved",
		},
	}
}

func negativeAllowCases(specSha string) []endpointCase {
	return append(negativeAllowValidationCases(specSha), negativeAllowMatchingCases(specSha)...)
}

func negativeAllowValidationCases(specSha string) []endpointCase {
	return []endpointCase{
		{
			name:     "(k) unknown key refused",
			code:     "func (c *Client) Fetch(ctx context.Context) error { return c.Get(ctx, \"/items\") }\n",
			allow:    "allow:\n  - module: testmod\n    func: Client.Fetch\n    method: GET\n    path: /items\n    reason: test\n    expires: 2027-01-01\n    unknown_key: bad\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "unknown_key",
		},
		{
			name:     "(l) duplicate entry refused",
			code:     "func (c *Client) Fetch(ctx context.Context) error { return c.Get(ctx, \"/items\") }\n",
			allow:    "allow:\n  - module: testmod\n    func: Client.Fetch\n    method: GET\n    path: /items\n    reason: test 1\n    expires: 2027-01-01\n  - module: testmod\n    func: Client.Fetch\n    method: GET\n    path: /items\n    reason: duplicate\n    expires: 2027-01-01\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "duplicate",
		},
		{
			name:     "(m) missing reason refused",
			code:     "func (c *Client) Fetch(ctx context.Context) error { return c.Get(ctx, \"/items\") }\n",
			allow:    "allow:\n  - module: testmod\n    func: Client.Fetch\n    method: GET\n    path: /items\n    expires: 2027-01-01\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "reason",
		},
		{
			name:     "(m2) missing method refused",
			code:     "func (c *Client) Fetch(ctx context.Context) error { return c.Get(ctx, \"/items\") }\n",
			allow:    "allow:\n  - module: testmod\n    func: Client.Fetch\n    path: /items\n    reason: test\n    expires: 2027-01-01\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "method is required",
		},
		{
			name:     "(m3) missing path refused",
			code:     "func (c *Client) Fetch(ctx context.Context) error { return c.Get(ctx, \"/items\") }\n",
			allow:    "allow:\n  - module: testmod\n    func: Client.Fetch\n    method: GET\n    reason: test\n    expires: 2027-01-01\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "path is required",
		},
	}
}

func negativeAllowMatchingCases(specSha string) []endpointCase {
	return []endpointCase{
		{
			name:     "(g) expired allow entry fails",
			code:     "func (c *Client) Old(ctx context.Context) error { return c.Delete(ctx, \"/not-in-spec\") }\n",
			allow:    "allow:\n  - module: testmod\n    func: Client.Old\n    method: DELETE\n    path: /not-in-spec\n    reason: test\n    expires: 2020-01-01\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "DELETE /not-in-spec not in sample v1.2.3 spec",
		},
		{
			name:     "(n) malformed expiry refused",
			code:     "func (c *Client) Fetch(ctx context.Context) error { return c.Get(ctx, \"/items\") }\n",
			allow:    "allow:\n  - module: testmod\n    func: Client.Fetch\n    method: GET\n    path: /items\n    reason: test\n    expires: 2027-13-45\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "expires",
		},
		{
			name:     "(o) stale entry refused",
			code:     "func (c *Client) Fetch(ctx context.Context) error { return c.Get(ctx, \"/items\") }\n",
			allow:    "allow:\n  - module: testmod\n    func: Client.Fetch\n    method: GET\n    path: /items\n    reason: test\n    expires: 2027-01-01\n",
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
