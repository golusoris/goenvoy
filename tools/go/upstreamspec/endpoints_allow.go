package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"go.yaml.in/yaml/v3"
)

// allowEntry defines an approved off-spec endpoint exception with expiration and audit metadata.
type allowEntry struct {
	Module  string `yaml:"module"`
	Func    string `yaml:"func"`
	Method  string `yaml:"method"`
	Path    string `yaml:"path"`
	Reason  string `yaml:"reason"`
	Expires string `yaml:"expires"`
}

type allowDoc struct {
	Allow []allowEntry `yaml:"allow"`
}

// readAllowList loads and validates the allow list file.
func readAllowList(path string) ([]allowEntry, error) {
	// #nosec G304 -- allow-list path is built below the repository root
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var doc allowDoc
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&doc); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if err := validateAllowList(path, doc.Allow); err != nil {
		return nil, err
	}
	return doc.Allow, nil
}

// validateAllowList verifies that all required fields are populated and unexpired.
func validateAllowList(path string, list []allowEntry) error {
	seen := make(map[string]bool)
	for i, entry := range list {
		if err := validateAllowEntry(path, i, entry); err != nil {
			return err
		}
		key := entry.Module + "\x00" + entry.Func + "\x00" + entry.Method + "\x00" + entry.Path
		if seen[key] {
			return fmt.Errorf("%s: allow[%d]: duplicate entry for (%s, %s, %s, %s)", path, i, entry.Module, entry.Func, entry.Method, entry.Path)
		}
		seen[key] = true
	}
	return nil
}

// validateAllowEntry validates all required fields of a single allow entry.
func validateAllowEntry(path string, i int, entry allowEntry) error {
	switch {
	case entry.Module == "":
		return fmt.Errorf("%s: allow[%d]: module is required", path, i)
	case entry.Func == "":
		return fmt.Errorf("%s: allow[%d]: func is required", path, i)
	case entry.Method == "":
		return fmt.Errorf("%s: allow[%d]: method is required", path, i)
	case entry.Path == "":
		return fmt.Errorf("%s: allow[%d]: path is required", path, i)
	case entry.Reason == "":
		return fmt.Errorf("%s: allow[%d]: reason is required", path, i)
	}
	if _, err := time.Parse(time.DateOnly, entry.Expires); err != nil {
		return fmt.Errorf("%s: allow[%d]: invalid expires %q (must be YYYY-MM-DD): %w", path, i, entry.Expires, err)
	}
	return nil
}

// findAllowMatch finds an unexpired matching allow entry for an off-spec operation.
func findAllowMatch(module string, op codeOp, allowList []allowEntry) int {
	for i, entry := range allowList {
		if entry.Module == module && entry.Func == op.funcName && entry.Method == op.verb && entry.Path == op.path {
			return i
		}
	}
	return -1
}
