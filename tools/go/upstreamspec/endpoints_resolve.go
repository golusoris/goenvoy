package main

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

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
