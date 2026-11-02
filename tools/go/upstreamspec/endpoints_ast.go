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

// collectSingleAssigns records identifiers assigned exactly once within a function declaration.
func collectSingleAssigns(fn *ast.FuncDecl) map[string]ast.Expr {
	if fn == nil || fn.Body == nil {
		return nil
	}
	assignCounts := make(map[string]int)
	assignExprs := make(map[string]ast.Expr)
	seedParamCounts(fn.Recv, assignCounts)
	if fn.Type != nil {
		seedParamCounts(fn.Type.Params, assignCounts)
	}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
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
		if count == 1 && assignExprs[name] != nil {
			single[name] = assignExprs[name]
		}
	}
	return single
}

// seedParamCounts records parameter and receiver names with an initial count of 1.
func seedParamCounts(fields *ast.FieldList, counts map[string]int) {
	if fields == nil {
		return
	}
	for _, field := range fields.List {
		for _, name := range field.Names {
			if name != nil && name.Name != "" && name.Name != "_" {
				counts[name.Name] = 1
			}
		}
	}
}

// recordAssignStmt records assignments from short and standard assign statements.
func recordAssignStmt(stmt *ast.AssignStmt, counts map[string]int, exprs map[string]ast.Expr) {
	if isQuerySuffix(stmt) {
		return
	}
	isCompound := stmt.Tok != token.ASSIGN && stmt.Tok != token.DEFINE
	for i, lhs := range stmt.Lhs {
		id, ok := lhs.(*ast.Ident)
		if !ok {
			continue
		}
		if isCompound {
			// Why: compound assign depends on prior value, so ident cannot be single-assigned.
			counts[id.Name] += 2
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
	singleAssigns := collectSingleAssigns(fn)
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

// verbForSelector maps an exact helper method or function name to an HTTP verb.
func verbForSelector(name string) string {
	switch name {
	case "Get", "GetRaw", "GetText", "GetOne", "GetList", "GetValue", "GetQueuePage":
		return "GET"
	case "Post", "Upload", "SendManualImport", "PostOne", "PostList":
		return "POST"
	case "Put", "PutOne", "PutList":
		return "PUT"
	case "Delete":
		return "DELETE"
	case "Patch":
		return "PATCH"
	case "Head":
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
