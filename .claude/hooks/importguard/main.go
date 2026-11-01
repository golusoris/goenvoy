// Command importguard reads a Claude Code Edit or Write hook payload on stdin and exits 2 when
// the Go source after the change imports a module outside the standard library, golang.org/x
// and goenvoy's own modules (ADR-0001).
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"strconv"
	"strings"
)

// payload is the subset of the hook input that importguard reads.
type payload struct {
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		FilePath   string `json:"file_path"`
		Content    string `json:"content"`
		OldString  string `json:"old_string"`
		NewString  string `json:"new_string"`
		ReplaceAll bool   `json:"replace_all"`
	} `json:"tool_input"`
}

// maxPayload bounds the hook input read from stdin.
const maxPayload = 64 << 20

func main() {
	code, err := run(os.Stdin, os.Stderr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "importguard: %v\n", err)
	}
	os.Exit(code)
}

// run returns 2 for a foreign import, 0 otherwise; err carries a note for stderr.
func run(in io.Reader, out io.Writer) (int, error) {
	raw, err := io.ReadAll(io.LimitReader(in, maxPayload))
	if err != nil {
		return 0, fmt.Errorf("read payload: %w", err)
	}
	var p payload
	if err := json.Unmarshal(raw, &p); err != nil {
		return 0, fmt.Errorf("decode payload: %w", err)
	}
	src, err := source(p)
	if err != nil {
		return 0, err
	}
	foreign, perr := foreignImports(src)
	if len(foreign) > 0 {
		fmt.Fprintf(out, "non-stdlib import blocked (%s): goenvoy is pure stdlib (ADR-0001). If you truly need a dep, write an ADR first.\n", strings.Join(foreign, " "))
		return 2, nil
	}
	return 0, perr
}

// source returns the file content after the Edit or Write is applied.
func source(p payload) (string, error) {
	in := p.ToolInput
	if p.ToolName == "Write" {
		return in.Content, nil
	}
	old, err := os.ReadFile(in.FilePath)
	if errors.Is(err, fs.ErrNotExist) {
		return in.NewString, nil
	}
	if err != nil {
		return "", fmt.Errorf("read %s: %w", in.FilePath, err)
	}
	n := 1
	if in.ReplaceAll {
		n = -1
	}
	// Why: an Edit fragment alone is no parseable file; the edited file is.
	return strings.Replace(string(old), in.OldString, in.NewString, n), nil
}

// foreignImports returns the foreign import paths of src. A fragment without a package clause is
// also parsed inside a synthetic package and import group; findings of every form count, since an
// imports-only parse stops at the first non-import token without an error.
func foreignImports(src string) ([]string, error) {
	forms := [...]string{src, "package p\n" + src, "package p\nimport (\n" + src + "\n)"}
	var foreign []string
	var errs []error
	for _, form := range forms {
		found, err := parseImports(form)
		foreign = append(foreign, found...)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == len(forms) {
		return foreign, errors.Join(errs...)
	}
	return foreign, nil
}

// parseImports parses the package clause and imports only, so syntax errors further down the
// file (common mid-edit) do not matter; specs parsed before an import-section error still count.
func parseImports(src string) ([]string, error) {
	f, err := parser.ParseFile(token.NewFileSet(), "", src, parser.ImportsOnly)
	if f == nil {
		return nil, fmt.Errorf("parse imports: %w", err)
	}
	var foreign []string
	for _, spec := range f.Imports {
		path, uerr := strconv.Unquote(spec.Path.Value)
		if uerr != nil {
			return nil, fmt.Errorf("unquote %s: %w", spec.Path.Value, uerr)
		}
		if !allowed(path) {
			foreign = append(foreign, path)
		}
	}
	if err != nil {
		return foreign, fmt.Errorf("parse imports: %w", err)
	}
	return foreign, nil
}

// allowed reports whether path is stdlib, golang.org/x or a goenvoy module.
func allowed(path string) bool {
	if strings.HasPrefix(path, "golang.org/x/") || strings.HasPrefix(path, "github.com/golusoris/goenvoy/") {
		return true
	}
	first, _, _ := strings.Cut(path, "/")
	// Why: stdlib paths never carry a dot in their first element; module paths always do.
	return !strings.Contains(first, ".")
}
