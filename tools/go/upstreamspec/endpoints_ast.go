package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	maxOperands     = 256
	maxParenDepth   = 64
	maxResolveDepth = 4
)

var reFmtParam = regexp.MustCompile(`%[dsv]`)

type moduleWalker struct {
	dir    string
	fset   *token.FileSet
	allOps []codeOp
}

func (w *moduleWalker) walk(path string, d os.DirEntry, err error) error {
	if err != nil {
		return fmt.Errorf("walk %s: %w", path, err)
	}
	if d.IsDir() {
		return w.walkDir(path, d.Name())
	}
	if !isSourceFile(d.Name()) {
		return nil
	}
	ops, err := parseFile(w.fset, w.dir, path)
	if err != nil {
		return err
	}
	w.allOps = append(w.allOps, ops...)
	return nil
}

func (w *moduleWalker) walkDir(path, name string) error {
	if path != w.dir && shouldSkipDir(w.dir, path, name) {
		return filepath.SkipDir
	}
	return nil
}

// parseModule recursively scans a module directory and parses non-test Go source files.
func parseModule(dir string) ([]codeOp, error) {
	w := moduleWalker{
		dir:  dir,
		fset: token.NewFileSet(),
	}
	if err := filepath.WalkDir(dir, w.walk); err != nil {
		return nil, fmt.Errorf("read module %s: %w", dir, err)
	}
	return w.allOps, nil
}

// isSourceFile reports whether a directory entry represents a non-test Go source file.
func isSourceFile(name string) bool {
	return strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
}

// shouldSkipDir reports whether a directory should be skipped during module scanning.
func shouldSkipDir(rootDir, path, name string) bool {
	if name == "testdata" || name == "vendor" {
		return true
	}
	if path != rootDir {
		if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
			return true
		}
	}
	return false
}

// parseFile parses a single Go file and extracts all endpoint operations.
func parseFile(fset *token.FileSet, moduleDir, filePath string) ([]codeOp, error) {
	// #nosec G304 -- filePath comes from the walk of the validated module directory
	f, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", filePath, err)
	}
	relFile, err := filepath.Rel(moduleDir, filePath)
	if err != nil {
		return nil, fmt.Errorf("rel path %s: %w", filePath, err)
	}
	relFile = filepath.ToSlash(relFile)
	fileConsts := collectFileConsts(f.Decls)
	var fileOps []codeOp
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		fileOps = append(fileOps, parseFunc(fset, relFile, fn, fileConsts)...)
	}
	return fileOps, nil
}

// collectFileConsts collects top-level string constants in a file.
func collectFileConsts(decls []ast.Decl) map[string]string {
	consts := make(map[string]string)
	for _, decl := range decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			extractConstValues(spec, consts)
		}
	}
	return consts
}

// extractConstValues extracts constant identifiers with string literal values from a spec.
func extractConstValues(spec ast.Spec, consts map[string]string) {
	valSpec, ok := spec.(*ast.ValueSpec)
	if !ok {
		return
	}
	for i, name := range valSpec.Names {
		if i >= len(valSpec.Values) {
			continue
		}
		lit, ok := valSpec.Values[i].(*ast.BasicLit)
		if !ok {
			continue
		}
		if s, ok := unquote(lit); ok {
			consts[name.Name] = s
		}
	}
}

// funcName computes the formatted name of a function or method.
func funcName(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return fn.Name.Name
	}
	recvType := fn.Recv.List[0].Type
	if star, ok := recvType.(*ast.StarExpr); ok {
		recvType = star.X
	}
	if id, ok := recvType.(*ast.Ident); ok {
		return id.Name + "." + fn.Name.Name
	}
	if idx, ok := recvType.(*ast.IndexExpr); ok {
		if id, ok := idx.X.(*ast.Ident); ok {
			return id.Name + "." + fn.Name.Name
		}
	}
	return fn.Name.Name
}

// collectSingleAssigns records identifiers assigned exactly once within a function body.
func collectSingleAssigns(body *ast.BlockStmt) map[string]ast.Expr {
	assignCounts := make(map[string]int)
	assignExprs := make(map[string]ast.Expr)
	ast.Inspect(body, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.AssignStmt:
			recordAssignStmt(stmt, assignCounts, assignExprs)
		case *ast.DeclStmt:
			recordDeclStmt(stmt, assignCounts, assignExprs)
		}
		return true
	})
	single := make(map[string]ast.Expr)
	for name, count := range assignCounts {
		if count == 1 {
			single[name] = assignExprs[name]
		}
	}
	return single
}

// recordAssignStmt records assignments from short and standard assign statements.
func recordAssignStmt(stmt *ast.AssignStmt, counts map[string]int, exprs map[string]ast.Expr) {
	if isQuerySuffix(stmt) {
		return
	}
	for i, lhs := range stmt.Lhs {
		id, ok := lhs.(*ast.Ident)
		if !ok {
			continue
		}
		counts[id.Name]++
		if i < len(stmt.Rhs) && counts[id.Name] == 1 {
			exprs[id.Name] = stmt.Rhs[i]
		}
	}
}

// maxConcatDepth bounds the walk down a left-nested concatenation.
const maxConcatDepth = 64

// isQuerySuffix reports whether stmt is `x += "?..." + ...`, which only appends a query string.
func isQuerySuffix(stmt *ast.AssignStmt) bool {
	if stmt.Tok != token.ADD_ASSIGN || len(stmt.Rhs) != 1 {
		return false
	}
	cur := stmt.Rhs[0]
	for range maxConcatDepth {
		bin, ok := cur.(*ast.BinaryExpr)
		if !ok || bin.Op != token.ADD {
			break
		}
		cur = bin.X
	}
	lit, ok := cur.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return false
	}
	value, err := strconv.Unquote(lit.Value)
	return err == nil && strings.HasPrefix(value, "?")
}

// recordDeclStmt records variable declarations within a function body.
func recordDeclStmt(stmt *ast.DeclStmt, counts map[string]int, exprs map[string]ast.Expr) {
	gen, ok := stmt.Decl.(*ast.GenDecl)
	if !ok || gen.Tok != token.VAR {
		return
	}
	for _, spec := range gen.Specs {
		valSpec, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for i, name := range valSpec.Names {
			counts[name.Name]++
			if i < len(valSpec.Values) && counts[name.Name] == 1 {
				exprs[name.Name] = valSpec.Values[i]
			}
		}
	}
}

// parseFunc extracts endpoint operations from a single function or method.
func parseFunc(fset *token.FileSet, relFile string, fn *ast.FuncDecl, fileConsts map[string]string) []codeOp {
	name := funcName(fn)
	singleAssigns := collectSingleAssigns(fn.Body)
	var ops []codeOp
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		op, isReq := inspectCall(fset, relFile, name, call, singleAssigns, fileConsts)
		if isReq {
			ops = append(ops, op)
		}
		return true
	})
	return ops
}

// inspectCall checks if a call expression is a request call and attributes verb and path.
func inspectCall(fset *token.FileSet, relFile, fnName string, call *ast.CallExpr, singleAssigns map[string]ast.Expr, fileConsts map[string]string) (codeOp, bool) {
	if isIgnoredCall(call) {
		return codeOp{}, false
	}
	cName := callName(call)
	verb := verbForSelector(cName)
	pathArg := extractPathArg(call)
	if pathArg == nil {
		if verb != "" {
			// Why: request call with recognized verb but no path argument fails closed.
			line := fset.Position(call.Pos()).Line
			return codeOp{funcName: fnName, file: relFile, line: line, verb: verb, unresolved: true}, true
		}
		return codeOp{}, false
	}

	path, line, ok := resolvePath(fset, pathArg, singleAssigns, fileConsts)
	if verb != "" {
		if !ok {
			// Why: recognized request verb with unresolvable path argument fails closed.
			return codeOp{funcName: fnName, file: relFile, line: line, verb: verb, unresolved: true}, true
		}
		return codeOp{funcName: fnName, file: relFile, line: line, verb: verb, path: path}, true
	}

	// Why: calls without recognized HTTP verbs that take resolved paths fail closed as unknown verbs.
	if ok && strings.HasPrefix(path, "/") {
		return codeOp{funcName: fnName, file: relFile, line: line, verb: "", path: path}, true
	}
	return codeOp{}, false
}

// resolvePath resolves an expression to a clean path string or reports unresolvable.
func resolvePath(fset *token.FileSet, expr ast.Expr, singleAssigns map[string]ast.Expr, fileConsts map[string]string) (string, int, bool) {
	cur := expr
	line := fset.Position(expr.Pos()).Line
	for range maxResolveDepth {
		id, ok := cur.(*ast.Ident)
		if !ok {
			break
		}
		if assign, ok := singleAssigns[id.Name]; ok {
			cur = assign
			continue
		}
		if val, ok := fileConsts[id.Name]; ok {
			if strings.HasPrefix(val, "/") {
				return cleanPath(val), line, true
			}
			return "", line, false
		}
		// Why: unresolved identifiers in request call path arguments fail closed.
		return "", line, false
	}

	switch node := cur.(type) {
	case *ast.BasicLit:
		return resolveBasicLit(node, line)
	case *ast.CallExpr:
		return resolveCallExpr(node, line)
	case *ast.BinaryExpr:
		return resolveBinaryExpr(node, line)
	default:
		// Why: selectors or complex expressions that cannot be statically resolved fail closed.
		return "", line, false
	}
}

// resolveBasicLit resolves a basic literal path expression.
func resolveBasicLit(node *ast.BasicLit, line int) (string, int, bool) {
	s, ok := unquote(node)
	if !ok || !strings.HasPrefix(s, "/") {
		return "", line, false
	}
	return cleanPath(s), line, true
}

// resolveCallExpr resolves a Sprintf call path expression.
func resolveCallExpr(node *ast.CallExpr, line int) (string, int, bool) {
	if callName(node) == "Sprintf" && len(node.Args) > 0 {
		lit, ok := node.Args[0].(*ast.BasicLit)
		if !ok {
			return "", line, false
		}
		s, ok := unquote(lit)
		if !ok || !strings.HasPrefix(s, "/") {
			return "", line, false
		}
		return cleanPath(s), line, true
	}
	return "", line, false
}

// resolveBinaryExpr resolves an addition concatenation path expression.
func resolveBinaryExpr(node *ast.BinaryExpr, line int) (string, int, bool) {
	if node.Op != token.ADD {
		return "", line, false
	}
	operands := flattenAdd(node)
	if len(operands) == 0 {
		return "", line, false
	}
	lit, ok := operands[0].(*ast.BasicLit)
	if !ok {
		// Why: dynamic prefix where first operand is not a string literal fails closed.
		return "", line, false
	}
	s, ok := unquote(lit)
	if !ok || !strings.HasPrefix(s, "/") {
		return "", line, false
	}
	res := joinOperands(operands, s)
	return cleanPath(res), line, true
}

// joinOperands concatenates operands, cuts at query string, and inserts param placeholders.
func joinOperands(operands []ast.Expr, init string) string {
	cur := init
	if idx := strings.IndexByte(cur, '?'); idx != -1 {
		return cur[:idx]
	}
	for _, op := range operands[1:] {
		next, done := appendOperand(cur, op)
		cur = next
		if done {
			break
		}
	}
	if idx := strings.IndexByte(cur, '?'); idx != -1 {
		cur = cur[:idx]
	}
	return cur
}

// appendOperand appends a single operand to the accumulating path string.
func appendOperand(cur string, op ast.Expr) (string, bool) {
	if strings.Contains(cur, "?") {
		return cur, true
	}
	lit, ok := op.(*ast.BasicLit)
	if !ok {
		// Why: dynamic operands always append parameter placeholder regardless of position.
		return cur + "{}", false
	}
	s, ok := unquote(lit)
	if !ok {
		return cur + "{}", false
	}
	if idx := strings.IndexByte(s, '?'); idx != -1 {
		return cur + s[:idx], true
	}
	return cur + s, false
}

// cleanPath normalizes query strings and format specifiers into standard placeholders.
func cleanPath(s string) string {
	if idx := strings.IndexByte(s, '?'); idx != -1 {
		s = s[:idx]
	}
	return reFmtParam.ReplaceAllString(s, "{}")
}

// flattenAdd recursively unpacks nested addition binary expressions into a flat slice.
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

// unparen unwraps nested parentheses expressions up to max depth.
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

// unquote unquotes a Go string literal.
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

// extractPathArg selects the candidate path argument from call arguments.
func extractPathArg(call *ast.CallExpr) ast.Expr {
	switch len(call.Args) {
	case 0:
		return nil
	case 1:
		return call.Args[0]
	default:
		if len(call.Args) >= 3 && isClientBase(call.Args[1]) {
			return call.Args[2]
		}
		if isContext(call.Args[0]) {
			return call.Args[1]
		}
		return call.Args[0]
	}
}

// isClientBase reports whether an expression refers to a base client receiver.
func isClientBase(e ast.Expr) bool {
	if sel, ok := e.(*ast.SelectorExpr); ok && sel.Sel.Name == "base" {
		return true
	}
	if id, ok := e.(*ast.Ident); ok && (id.Name == "base" || id.Name == "client") {
		return true
	}
	return false
}

// isContext reports whether an expression represents a context parameter.
func isContext(e ast.Expr) bool {
	if id, ok := e.(*ast.Ident); ok {
		lower := strings.ToLower(id.Name)
		return strings.Contains(lower, "ctx") || lower == "context"
	}
	return false
}

// isIgnoredCall reports whether a call expression is standard utility rather than an HTTP call.
func isIgnoredCall(call *ast.CallExpr) bool {
	switch fun := call.Fun.(type) {
	case *ast.SelectorExpr:
		if id, ok := fun.X.(*ast.Ident); ok {
			switch id.Name {
			case "fmt", "strings", "strconv", "url", "path", "filepath", "time", "errors", "os", "io", "bytes", "regexp":
				return true
			}
		}
	case *ast.Ident:
		switch fun.Name {
		case "append", "make", "len", "cap", "new", "panic", "recover", "close":
			return true
		}
	}
	return false
}

// verbForSelector maps a selector method or function name to an HTTP verb.
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

// callName retrieves the base name of a function or method invocation.
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
