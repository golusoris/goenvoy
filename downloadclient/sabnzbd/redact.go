package sabnzbd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

// scrub replaces the raw, query-escaped, path-escaped, Go-quoted, and JSON-escaped API key in s with REDACTED, ignoring hex-digit case in percent escapes of the two URL-escaped forms.
func (c *Client) scrub(s string) string {
	for _, form := range c.keyForms() {
		s = strings.ReplaceAll(s, form, "REDACTED")
	}
	var b strings.Builder
	for range len(s) {
		start, end, ok := c.percentSpan(s)
		if !ok {
			break
		}
		b.WriteString(s[:start])
		b.WriteString("REDACTED")
		s = s[end:]
	}
	if b.Len() == 0 {
		return s
	}
	b.WriteString(s)
	return b.String()
}

// keyForms returns the distinct literal raw, query-escaped, path-escaped, Go-quoted, and JSON-escaped forms of the API key, or nil when the key is empty; hex-digit case of percent escapes is handled by percentSpan.
func (c *Client) keyForms() []string {
	// Why: ReplaceAll with an empty old inserts at every rune and Contains with an empty key matches every string.
	if c.apiKey == "" {
		return nil
	}
	quoted := strconv.Quote(c.apiKey)
	query, path := url.QueryEscape(c.apiKey), url.PathEscape(c.apiKey)
	// Why: a server that echoes the request URL may write query or path escaping.
	candidates := []string{c.apiKey, query, path, quoted[1 : len(quoted)-1]}
	// Why: encoding/json writes <, >, &, U+2028 and U+2029 as \uXXXX escapes that no other form matches.
	if b, err := json.Marshal(c.apiKey); err == nil {
		candidates = append(candidates, string(b[1:len(b)-1]))
	}
	forms := make([]string, 0, len(candidates))
	for _, form := range candidates {
		if !slices.Contains(forms, form) {
			forms = append(forms, form)
		}
	}
	return forms
}

// percentSpan returns the byte span of the first query-escaped or path-escaped API key match in s, ignoring hex-digit case in percent escapes.
func (c *Client) percentSpan(s string) (int, int, bool) {
	if c.apiKey == "" {
		return 0, 0, false
	}
	query, path := url.QueryEscape(c.apiKey), url.PathEscape(c.apiKey)
	for i := range len(s) {
		if len(query) <= len(s)-i && matchPercentForm(s[i:], query) {
			return i, i + len(query), true
		}
		if path != query && len(path) <= len(s)-i && matchPercentForm(s[i:], path) {
			return i, i + len(path), true
		}
	}
	return 0, 0, false
}

// matchPercentForm reports whether s starts with form, ignoring hex-digit case in the two bytes after each '%' in form.
func matchPercentForm(s, form string) bool {
	hex := 0
	for j := range len(form) {
		ok := s[j] == form[j]
		if hex > 0 {
			ok = sameHex(s[j], form[j])
			hex--
		} else if form[j] == '%' {
			hex = 2
		}
		if !ok {
			return false
		}
	}
	return true
}

// sameHex reports whether a and b are equal after folding lower-case ASCII hex digits to upper case.
func sameHex(a, b byte) bool {
	if a >= 'a' && a <= 'f' {
		a -= 'a' - 'A'
	}
	return a == b
}

type redactedError struct {
	msg string
	// Why: As sanitizes a copied URL error with the same key scrub as the message.
	c *Client
	// Why: no Unwrap method, so callers cannot reach text that holds the key.
	err error
}

// Error returns the sanitized message.
func (e *redactedError) Error() string { return e.msg }

// Is matches the hidden error without exposing it.
func (e *redactedError) Is(target error) bool {
	// Why: like errors.Is, compare only a comparable target, so an uncomparable one never panics.
	canCompare := target != nil && reflect.TypeOf(target).Comparable()
	found := false
	e.walk(func(node error) bool {
		if canCompare && node == target {
			found = true
		} else if m, ok := node.(interface{ Is(error) bool }); ok && m.Is(target) {
			found = true
		}
		return found
	})
	return found
}

// Timeout reports whether the hidden error timed out.
func (e *redactedError) Timeout() bool {
	// Why: url.Error.Timeout type-asserts its Err directly, so the wrapper must forward the first node that has it.
	reported := false
	e.walk(func(node error) bool {
		t, ok := node.(interface{ Timeout() bool })
		if ok {
			reported = t.Timeout()
		}
		return ok
	})
	return reported
}

// Temporary reports whether the hidden error is temporary.
func (e *redactedError) Temporary() bool {
	// Why: url.Error.Temporary type-asserts its Err directly, so the wrapper must forward the first node that has it.
	reported := false
	e.walk(func(node error) bool {
		t, ok := node.(interface{ Temporary() bool })
		if ok {
			reported = t.Temporary()
		}
		return ok
	})
	return reported
}

// walkFrame tracks the current child index inside a slice of wrapped errors during depth-first traversal.
type walkFrame struct {
	children []error
	index    int
}

// walk visits the hidden tree depth-first, in errors.Is order, and stops when visit reports true, the walk exceeds maxWalkSteps nodes, or the tree holds more than maxErrNodes joined errors.
func (e *redactedError) walk(visit func(error) bool) {
	// Why: errors.Is and errors.As visit every path, exponential on a fan-out tree; only joined errors spend the maxErrNodes budget, so a long quiet chain stays reachable.
	stack := []walkFrame{{children: []error{e.err}}}
	for steps, branches := 0, 0; len(stack) > 0; {
		steps++
		// Why: cyclic single-unwrap chain must terminate; bound far above any real chain.
		if steps > maxWalkSteps {
			return
		}
		top := stack[len(stack)-1]
		stack = pushNextChild(stack[:len(stack)-1], top.children, top.index+1)
		node := top.children[top.index]
		if visit(node) {
			return
		}
		children, joined := unwrapChildren(node)
		if joined {
			branches++
		}
		if branches > maxErrNodes {
			return
		}
		stack = pushNextChild(stack, children, 0)
	}
}

// unwrapChildren returns the direct child slice of err and whether err joins several errors.
func unwrapChildren(err error) ([]error, bool) {
	if w := errors.Unwrap(err); w != nil {
		return []error{w}, false
	}
	joined, ok := err.(interface{ Unwrap() []error })
	if !ok {
		return nil, false
	}
	return joined.Unwrap(), true
}

// pushNextChild appends a frame positioned at the first non-nil error in children[from:], if any.
func pushNextChild(stack []walkFrame, children []error, from int) []walkFrame {
	for i := from; i < len(children); i++ {
		if children[i] != nil {
			return append(stack, walkFrame{children: children, index: i})
		}
	}
	return stack
}

// As sets target to a sanitized copy of the first *url.Error in the hidden tree and reports false for every other target type, so no key-bearing value escapes.
func (e *redactedError) As(target any) bool {
	dst, ok := target.(**url.Error)
	if !ok {
		return false
	}
	// Why: bounded breadth-first walk over the hidden tree; errors.As on it would search unbounded.
	queue := []error{e.err}
	for i := 0; i < maxErrNodes && i < len(queue); i++ {
		if uerr, ok := queue[i].(*url.Error); ok && uerr != nil { //nolint:errorlint // first URL error of a bounded walk; errors.As is the unbounded search
			// Why: hand out a scrubbed copy, never the original that holds the key.
			*dst = &url.Error{Op: e.c.scrub(uerr.Op), URL: e.c.scrub(redactURL(uerr.URL)), Err: e.c.scrubError(uerr.Err)}
			return true
		}
		queue = appendWrapped(queue, queue[i])
	}
	return false
}

// Format prints only the sanitized message for every verb and flag.
func (e *redactedError) Format(f fmt.State, verb rune) {
	if _, err := fmt.Fprintf(f, fmt.FormatString(f, verb), e.msg); err != nil {
		// Why: Format has no error return; a failed write leaves nothing more to print.
		return
	}
}

const maxErrNodes = 64

// maxWalkSteps bounds the total nodes visited by redactedError.walk across single-unwrap chains.
const maxWalkSteps = 4096

// errTreeTooLargeMsg replaces the formatted message of an error tree that exceeds maxErrNodes.
const errTreeTooLargeMsg = "error tree too large: details redacted"

// scrubError hides error trees that carry the API key or exceed the traversal limit.
func (c *Client) scrubError(err error) error {
	if err == nil {
		return nil
	}
	if exceedsErrNodes(err) {
		return &redactedError{msg: errTreeTooLargeMsg, err: err, c: c}
	}
	if !c.carriesKey(err) {
		return err
	}
	return &redactedError{msg: c.scrub(err.Error()), err: err, c: c}
}

// carriesKey reports whether an error tree contains the API key or exceeds the node limit.
func (c *Client) carriesKey(err error) bool {
	if err == nil {
		return false
	}
	// Why: a tree past the bound is hidden before Error() can format exponential joined fan-out.
	if exceedsErrNodes(err) {
		return true
	}
	queue := []error{err}
	for i := 0; i < maxErrNodes && i < len(queue); i++ {
		if uerr, ok := queue[i].(*url.Error); (!ok || uerr != nil) && c.containsKey(queue[i].Error()) { //nolint:errorlint // guard typed-nil *url.Error before Error()
			return true
		}
		queue = appendWrapped(queue, queue[i])
	}
	return false
}

// exceedsErrNodes reports whether err's unwrap tree holds more than maxErrNodes nodes without formatting any node.
func exceedsErrNodes(err error) bool {
	queue := []error{err}
	for i := 0; i < maxErrNodes && i < len(queue); i++ {
		queue = appendWrapped(queue, queue[i])
	}
	return len(queue) > maxErrNodes
}

// containsKey reports whether s holds the nonempty raw, query-escaped, path-escaped, Go-quoted, or JSON-escaped API key, ignoring hex-digit case in percent escapes of the two URL-escaped forms.
func (c *Client) containsKey(s string) bool {
	if slices.ContainsFunc(c.keyForms(), func(form string) bool { return strings.Contains(s, form) }) {
		return true
	}
	_, _, ok := c.percentSpan(s)
	return ok
}

// appendWrapped appends non-nil wrapped errors while bounding queue growth.
func appendWrapped(queue []error, err error) []error {
	if len(queue) > maxErrNodes {
		// Why: carriesKey already treats this tree as too large; more entries only cost memory.
		return queue
	}
	if uerr, ok := err.(*url.Error); ok && uerr == nil { //nolint:errorlint // typed-nil *url.Error panics inside errors.Unwrap
		return queue
	}
	if w := errors.Unwrap(err); w != nil {
		return append(queue, w)
	}
	joined, ok := err.(interface{ Unwrap() []error })
	if !ok {
		return queue
	}
	children := joined.Unwrap()
	for i := 0; i < len(children) && len(queue) <= maxErrNodes; i++ {
		if children[i] != nil {
			queue = append(queue, children[i])
		}
	}
	return queue
}

// redactError sanitizes URL errors and hides error trees that may expose the API key.
func (c *Client) redactError(err error) error {
	// Why: only the root can be a *url.Error to copy; errors.As would search the whole tree unbounded.
	if uerr, ok := err.(*url.Error); ok && uerr != nil { //nolint:errorlint // root identity wanted; tree search is the defect
		// Why: copy only when sanitizing changed the URL text or hid the cause; an untouched root keeps its identity for errors.Is.
		// Why: a copy never mutates an error value net/http or a custom transport may share.
		cleanOp, cleanURL := c.scrub(uerr.Op), c.scrub(redactURL(uerr.URL))
		if cleanOp != uerr.Op || cleanURL != uerr.URL || (uerr.Err != nil && c.carriesKey(uerr.Err)) {
			err = &url.Error{Op: cleanOp, URL: cleanURL, Err: c.scrubError(uerr.Err)}
		}
	}
	return c.scrubError(err)
}

// redactBaseURL drops userinfo and redacts apikey query values, or drops userinfo and the query when parsing fails.
func redactBaseURL(raw string) string {
	if u, err := url.Parse(raw); err == nil {
		u.User = nil
		return redactURL(u.String())
	}
	base := redactURL(raw)
	scheme, rest, ok := strings.Cut(base, "://")
	if !ok {
		return base
	}
	end := strings.IndexAny(rest, "/?#")
	if end < 0 {
		end = len(rest)
	}
	if at := strings.LastIndexByte(rest[:end], '@'); at >= 0 {
		rest = rest[at+1:]
	}
	return scheme + "://" + rest
}

// redactURL redacts apikey query values or drops the query when parsing fails.
func redactURL(raw string) string {
	u, err := url.Parse(raw)
	var q url.Values
	if err == nil {
		q, err = url.ParseQuery(u.RawQuery)
	}
	if err != nil {
		// Why: an unparseable URL or query may still carry the key, so drop the query whole.
		base, _, _ := strings.Cut(raw, "?")
		return base
	}
	if !q.Has("apikey") {
		return raw
	}
	q.Set("apikey", "REDACTED")
	u.RawQuery = q.Encode()
	return u.String()
}
