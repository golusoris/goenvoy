package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"go.yaml.in/yaml/v3"
)

var rePathParam = regexp.MustCompile(`\{[^}]*\}`)

type specDoc struct {
	Paths map[string]map[string]any `json:"paths" yaml:"paths"`
}

// isYAML reports whether the spec file path has a YAML extension.
func isYAML(filePath string) bool {
	return strings.HasSuffix(filePath, ".yml") || strings.HasSuffix(filePath, ".yaml")
}

// parseSpec decodes an OpenAPI document (JSON or YAML) and normalizes endpoint operations.
func parseSpec(data []byte, prefix string, yamlDoc bool) (map[operation]bool, error) {
	var doc specDoc
	if yamlDoc {
		if err := yaml.Unmarshal(data, &doc); err != nil {
			return nil, fmt.Errorf("decode openapi yaml: %w", err)
		}
	} else {
		if err := json.Unmarshal(data, &doc); err != nil {
			return nil, fmt.Errorf("decode openapi json: %w", err)
		}
	}
	return extractSpecOps(doc.Paths, prefix), nil
}

// extractSpecOps converts raw OpenAPI paths and methods into normalized operations.
func extractSpecOps(paths map[string]map[string]any, prefix string) map[operation]bool {
	ops := make(map[operation]bool)
	for rawPath, methods := range paths {
		specPath := rawPath
		if prefix != "" {
			specPath = strings.TrimSuffix(prefix, "/") + "/" + strings.TrimPrefix(rawPath, "/")
		}
		// Why: dynamic operands in spec normalize to {} so mid-segment and simple params match.
		normPath := rePathParam.ReplaceAllString(specPath, "{}")
		for method := range methods {
			m := strings.ToUpper(method)
			switch m {
			case "GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "TRACE":
				ops[operation{verb: m, path: normPath}] = true
			}
		}
	}
	return ops
}
