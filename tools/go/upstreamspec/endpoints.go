package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.yaml.in/yaml/v3"
)

const (
	maxOperands   = 256
	maxParenDepth = 64
)

var (
	reFmtParam  = regexp.MustCompile(`%[dsv]`)
	rePathParam = regexp.MustCompile(`\{[^}]*\}`)
)

type operation struct {
	verb string
	path string
}

type codeOp struct {
	funcName string
	file     string
	line     int
	verb     string
	path     string
}

type allowEntry struct {
	Module  string `yaml:"module"`
	Func    string `yaml:"func"`
	Reason  string `yaml:"reason"`
	Expires string `yaml:"expires"`
}

type allowDoc struct {
	Allow []allowEntry `yaml:"allow"`
}

type endpointsChecker struct {
	checker checker
	root    string
}

func (c endpointsChecker) all(ctx context.Context, pins []pin, allowList []allowEntry) error {
	var failed []error
	matched := make(map[int]bool)
	for _, p := range pins {
		errs, err := c.processPin(ctx, p, allowList, matched)
		if err != nil {
			return err
		}
		failed = append(failed, errs...)
	}
	for i, entry := range allowList {
		if !matched[i] {
			failed = append(failed, fmt.Errorf("allow entry %s %s matches no off-spec call; remove it", entry.Module, entry.Func))
		}
	}
	return errors.Join(failed...)
}

func (c endpointsChecker) processPin(ctx context.Context, p pin, allowList []allowEntry, matched map[int]bool) ([]error, error) {
	if isYAML(p.Path) {
		if p.Module != "" {
			return nil, fmt.Errorf("%s: yaml not supported", p.Name)
		}
		if _, err := fmt.Fprintf(c.checker.out, "upstreamspec: %s: skipped: yaml spec\n", p.Name); err != nil {
			return nil, fmt.Errorf("write summary: %w", err)
		}
		return nil, nil
	}
	if p.Module == "" {
		return nil, nil
	}
	findings, err := c.checkPin(ctx, p, allowList, matched)
	if err != nil {
		return nil, err
	}
	errs := make([]error, len(findings))
	for i, f := range findings {
		errs[i] = errors.New(f)
	}
	return errs, nil
}

func (c endpointsChecker) checkPin(ctx context.Context, p pin, allowList []allowEntry, matched map[int]bool) ([]string, error) {
	data, err := c.checker.fetch(ctx, p)
	if err != nil {
		return nil, err
	}
	specOps, err := parseSpec(data, p.Prefix)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", p.Name, err)
	}

	moduleDir := filepath.Join(c.root, p.Module)
	codeOps, err := parseModule(moduleDir)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", p.Name, err)
	}

	findings, inSpecCount, calledCount := evalOps(p, codeOps, specOps, allowList, matched)

	uncalled := len(specOps) - calledCount
	if _, err := fmt.Fprintf(c.checker.out, "upstreamspec: %s: code ops %d, in spec %d, spec ops not called %d\n",
		p.Name, len(codeOps), inSpecCount, uncalled); err != nil {
		return nil, fmt.Errorf("write summary: %w", err)
	}
	return findings, nil
}

func evalOps(p pin, ops []codeOp, specOps map[operation]bool, allowList []allowEntry, matched map[int]bool) ([]string, int, int) {
	inSpecCount := 0
	calledSpecOps := make(map[operation]bool)
	var findings []string
	now := time.Now()
	expiredReported := make(map[int]bool)

	tag := p.Tag
	if tag == "" {
		tag = p.Branch
	}

	for _, op := range ops {
		if op.verb == "" {
			findings = append(findings, fmt.Sprintf("%s/%s:%d %s: verb unknown", p.Module, op.file, op.line, op.funcName))
			continue
		}
		key := operation{verb: op.verb, path: op.path}
		if specOps[key] {
			inSpecCount++
			calledSpecOps[key] = true
			continue
		}
		findings = append(findings, evalOffSpecOp(p.Module, tag, p.Name, op, allowList, matched, expiredReported, now)...)
	}
	return findings, inSpecCount, len(calledSpecOps)
}

func evalOffSpecOp(module, tag, specName string, op codeOp, allowList []allowEntry, matched, expiredReported map[int]bool, now time.Time) []string {
	matchIdx := findAllowMatch(module, op.funcName, allowList)
	if matchIdx != -1 {
		matched[matchIdx] = true
		entry := allowList[matchIdx]
		if expired, err := isExpired(entry.Expires, now); err == nil && !expired {
			return nil
		}
		if !expiredReported[matchIdx] {
			expiredReported[matchIdx] = true
			return []string{
				fmt.Sprintf("allow entry %s %s expired %s", entry.Module, entry.Func, entry.Expires),
				fmt.Sprintf("%s/%s:%d %s: %s %s not in %s %s spec", module, op.file, op.line, op.funcName, op.verb, op.path, specName, tag),
			}
		}
	}
	return []string{
		fmt.Sprintf("%s/%s:%d %s: %s %s not in %s %s spec", module, op.file, op.line, op.funcName, op.verb, op.path, specName, tag),
	}
}

func findAllowMatch(module, funcName string, allowList []allowEntry) int {
	for i, entry := range allowList {
		if entry.Module == module && entry.Func == funcName {
			return i
		}
	}
	return -1
}

func parseSpec(data []byte, prefix string) (map[operation]bool, error) {
	var doc struct {
		Paths map[string]map[string]any `json:"paths"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("decode openapi json: %w", err)
	}
	ops := make(map[operation]bool)
	for rawPath, methods := range doc.Paths {
		specPath := rawPath
		if prefix != "" {
			specPath = strings.TrimSuffix(prefix, "/") + "/" + strings.TrimPrefix(rawPath, "/")
		}
		normPath := rePathParam.ReplaceAllString(specPath, "{}")
		for method := range methods {
			m := strings.ToUpper(method)
			switch m {
			case "GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "TRACE":
				ops[operation{verb: m, path: normPath}] = true
			}
		}
	}
	return ops, nil
}

func parseModule(dir string) ([]codeOp, error) {
	// #nosec G304 -- module directory path is validated
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", dir, err)
	}
	fset := token.NewFileSet()
	var allOps []codeOp
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		filePath := filepath.Join(dir, entry.Name())
		// #nosec G304 -- safe file inside validated module dir
		f, err := parser.ParseFile(fset, filePath, nil, 0)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", filePath, err)
		}
		ops := parseFileDecls(fset, entry.Name(), f.Decls)
		allOps = append(allOps, ops...)
	}
	return allOps, nil
}

func parseFileDecls(fset *token.FileSet, fileName string, decls []ast.Decl) []codeOp {
	var ops []codeOp
	for _, decl := range decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		fnOps := parseFunc(fset, fileName, fn)
		ops = append(ops, fnOps...)
	}
	return ops
}

func parseFunc(fset *token.FileSet, fileName string, fn *ast.FuncDecl) []codeOp {
	fallbackVerb := detectVerb(fn.Body)
	paths := extractPaths(fset, fn.Body)
	ops := make([]codeOp, 0, len(paths))
	for _, pe := range paths {
		verb := pe.verb
		if verb == "" {
			verb = fallbackVerb
		}
		ops = append(ops, codeOp{
			funcName: fn.Name.Name,
			file:     fileName,
			line:     pe.line,
			verb:     verb,
			path:     pe.path,
		})
	}
	return ops
}

func detectVerb(body *ast.BlockStmt) string {
	var verbs []string
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		name := callName(call)
		if v := verbForSelector(name); v != "" {
			verbs = append(verbs, v)
		}
		return true
	})
	if len(verbs) > 0 {
		return verbs[0]
	}
	return ""
}

type pathEntry struct {
	path string
	line int
	verb string
}

type pathExtractor struct {
	fset      *token.FileSet
	verbStack []string
	paths     *[]pathEntry
	handled   map[ast.Node]bool
}

func (pe pathExtractor) currentVerb() string {
	if len(pe.verbStack) > 0 {
		return pe.verbStack[len(pe.verbStack)-1]
	}
	return ""
}

func (pe pathExtractor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}
	if pe.handled[node] {
		return nil
	}
	if call, ok := node.(*ast.CallExpr); ok {
		if v := verbForSelector(callName(call)); v != "" {
			newStack := make([]string, len(pe.verbStack)+1)
			copy(newStack, pe.verbStack)
			newStack[len(pe.verbStack)] = v
			pe.verbStack = newStack
		}
	}
	if entry, ok := tryPath(pe.fset, node, pe.handled); ok {
		entry.verb = pe.currentVerb()
		*pe.paths = append(*pe.paths, entry)
		return nil
	}
	return pe
}

func extractPaths(fset *token.FileSet, body *ast.BlockStmt) []pathEntry {
	var paths []pathEntry
	handled := make(map[ast.Node]bool)
	extractor := pathExtractor{
		fset:    fset,
		paths:   &paths,
		handled: handled,
	}
	ast.Walk(extractor, body)
	return paths
}

func tryPath(fset *token.FileSet, n ast.Node, handled map[ast.Node]bool) (pathEntry, bool) {
	switch node := n.(type) {
	case *ast.CallExpr:
		if callName(node) == "Sprintf" && len(node.Args) > 0 {
			return handleSprintf(fset, node.Args[0], handled)
		}
	case *ast.BinaryExpr:
		if node.Op == token.ADD {
			return handleBinaryAdd(fset, node, handled)
		}
	case *ast.BasicLit:
		return handleBasicLit(fset, node)
	}
	return pathEntry{}, false
}

func handleSprintf(fset *token.FileSet, arg ast.Expr, handled map[ast.Node]bool) (pathEntry, bool) {
	lit, ok := arg.(*ast.BasicLit)
	if !ok {
		return pathEntry{}, false
	}
	s, ok := unquote(lit)
	if !ok || !strings.HasPrefix(s, "/") {
		return pathEntry{}, false
	}
	handled[arg] = true
	return pathEntry{
		path: cleanPath(s),
		line: fset.Position(lit.Pos()).Line,
	}, true
}

func appendOperand(current string, op ast.Expr) string {
	if nextLit, ok := op.(*ast.BasicLit); ok {
		if nextS, ok := unquote(nextLit); ok {
			return current + nextS
		}
	}
	if idx := strings.IndexByte(current, '?'); idx != -1 {
		current = current[:idx]
	}
	if strings.HasSuffix(current, "/") {
		return current + "{}"
	}
	return current
}

func joinOperands(operands []ast.Expr, init string) string {
	cur := init
	for _, op := range operands[1:] {
		cur = appendOperand(cur, op)
	}
	return cur
}

func handleBinaryAdd(fset *token.FileSet, node *ast.BinaryExpr, handled map[ast.Node]bool) (pathEntry, bool) {
	operands := flattenAdd(node)
	for _, op := range operands {
		handled[op] = true
	}
	markSubAdd(node, handled)
	if len(operands) == 0 {
		return pathEntry{}, false
	}
	lit, ok := operands[0].(*ast.BasicLit)
	if !ok {
		return pathEntry{}, false
	}
	s, ok := unquote(lit)
	if !ok || !strings.HasPrefix(s, "/") {
		return pathEntry{}, false
	}
	return pathEntry{
		path: cleanPath(joinOperands(operands, s)),
		line: fset.Position(lit.Pos()).Line,
	}, true
}

func markSubAdd(expr ast.Expr, handled map[ast.Node]bool) {
	stack := []ast.Expr{expr}
	for i := 0; i < maxOperands && len(stack) > 0; i++ {
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if b, ok := top.(*ast.BinaryExpr); ok && b.Op == token.ADD {
			handled[b] = true
			stack = append(stack, b.X, b.Y)
		}
	}
}

func handleBasicLit(fset *token.FileSet, node *ast.BasicLit) (pathEntry, bool) {
	s, ok := unquote(node)
	if !ok || !strings.HasPrefix(s, "/") {
		return pathEntry{}, false
	}
	return pathEntry{
		path: cleanPath(s),
		line: fset.Position(node.Pos()).Line,
	}, true
}

func cleanPath(s string) string {
	if idx := strings.IndexByte(s, '?'); idx != -1 {
		s = s[:idx]
	}
	return reFmtParam.ReplaceAllString(s, "{}")
}

func verbForSelector(name string) string {
	switch name {
	case "Upload", "SendManualImport":
		return "POST"
	}
	switch {
	case strings.HasPrefix(name, "Get"):
		return "GET"
	case strings.HasPrefix(name, "Post"):
		return "POST"
	case strings.HasPrefix(name, "Put"):
		return "PUT"
	case strings.HasPrefix(name, "Delete"):
		return "DELETE"
	case strings.HasPrefix(name, "Patch"):
		return "PATCH"
	case strings.HasPrefix(name, "Head"):
		return "HEAD"
	default:
		return ""
	}
}

func callName(call *ast.CallExpr) string {
	switch fun := call.Fun.(type) {
	case *ast.SelectorExpr:
		return fun.Sel.Name
	case *ast.IndexExpr:
		if sel, ok := fun.X.(*ast.SelectorExpr); ok {
			return sel.Sel.Name
		}
		if id, ok := fun.X.(*ast.Ident); ok {
			return id.Name
		}
	case *ast.IndexListExpr:
		if sel, ok := fun.X.(*ast.SelectorExpr); ok {
			return sel.Sel.Name
		}
		if id, ok := fun.X.(*ast.Ident); ok {
			return id.Name
		}
	case *ast.Ident:
		return fun.Name
	}
	return ""
}

func unparen(e ast.Expr) ast.Expr {
	for range maxParenDepth {
		p, ok := e.(*ast.ParenExpr)
		if !ok {
			return e
		}
		e = p.X
	}
	return e
}

func flattenAdd(expr ast.Expr) []ast.Expr {
	var operands []ast.Expr
	stack := []ast.Expr{expr}
	for i := 0; i < maxOperands && len(stack) > 0; i++ {
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		u := unparen(top)
		if b, ok := u.(*ast.BinaryExpr); ok && b.Op == token.ADD {
			stack = append(stack, b.Y, b.X)
		} else {
			operands = append(operands, u)
		}
	}
	return operands
}

func unquote(lit *ast.BasicLit) (string, bool) {
	if lit.Kind != token.STRING {
		return "", false
	}
	s, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "", false
	}
	return s, true
}

func isYAML(filePath string) bool {
	return strings.HasSuffix(filePath, ".yml") || strings.HasSuffix(filePath, ".yaml")
}

func isExpired(expiry string, today time.Time) (bool, error) {
	date, err := time.Parse(time.DateOnly, expiry)
	if err != nil {
		return false, fmt.Errorf("expiry %q: %w", expiry, err)
	}
	day := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	return day.After(date), nil
}

func readAllowList(path string) ([]allowEntry, error) {
	// #nosec G304 -- allow-list path below root
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

func validateAllowList(path string, list []allowEntry) error {
	seen := make(map[string]bool)
	for i, entry := range list {
		if entry.Module == "" {
			return fmt.Errorf("%s: allow[%d]: module is required", path, i)
		}
		if entry.Func == "" {
			return fmt.Errorf("%s: allow[%d]: func is required", path, i)
		}
		if entry.Reason == "" {
			return fmt.Errorf("%s: allow[%d]: reason is required", path, i)
		}
		if _, err := time.Parse(time.DateOnly, entry.Expires); err != nil {
			return fmt.Errorf("%s: allow[%d]: invalid expires %q (must be YYYY-MM-DD): %w", path, i, entry.Expires, err)
		}
		key := entry.Module + "\x00" + entry.Func
		if seen[key] {
			return fmt.Errorf("%s: allow[%d]: duplicate entry for (%s, %s)", path, i, entry.Module, entry.Func)
		}
		seen[key] = true
	}
	return nil
}
