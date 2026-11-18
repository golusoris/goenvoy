package tautulli

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

// scrub replaces the raw, query-escaped, path-escaped, Go-quoted, and JSON-escaped API key in s with REDACTED, with hex-digit case of percent escapes ignored.
func (c *Client) scrub(s string) string {
	for _, form := range c.keyForms() {
		s = strings.ReplaceAll(s, form, "REDACTED")
	}
	start, end := c.nextPercentKey(s)
	if start < 0 {
		return s
	}
	var b strings.Builder
	for steps, maxSpans := 0, len(s); steps < maxSpans && start >= 0; steps++ {
		b.WriteString(s[:start])
		b.WriteString("REDACTED")
		s = s[end:]
		start, end = c.nextPercentKey(s)
	}
	b.WriteString(s)
	return b.String()
}

// scrubBody redacts key forms and hides JSON containing an escaped key in any string token.
func (c *Client) scrubBody(text string) string {
	text = c.scrub(text)
	if c.apiKey == "" || !json.Valid([]byte(text)) {
		return text
	}
	decoder := json.NewDecoder(strings.NewReader(text))
	for range len(text) {
		token, err := decoder.Token()
		if err != nil {
			break
		}
		if value, ok := token.(string); ok && c.containsKey(value) {
			return "REDACTED"
		}
	}
	return text
}

// keyForms returns the distinct raw, query-escaped, path-escaped, Go-quoted, and JSON-escaped forms of the API key, or nil when the key is empty; hex-digit case of percent escapes is matched by nextPercentKey.
func (c *Client) keyForms() []string {
	// Why: ReplaceAll with an empty old inserts at every rune and Contains with an empty key matches every string.
	if c.apiKey == "" {
		return nil
	}
	quoted := strconv.Quote(c.apiKey)
	query, path := url.QueryEscape(c.apiKey), url.PathEscape(c.apiKey)
	candidates := []string{c.apiKey, query, path, (&url.URL{Path: c.apiKey}).EscapedPath(), quoted[1 : len(quoted)-1]}
	// Why: encoding/json writes <, >, &, U+2028 and U+2029 as \uXXXX escapes that no other form matches.
	if b, err := json.Marshal(c.apiKey); err == nil {
		encoded := string(b[1 : len(b)-1])
		candidates = append(candidates, encoded)
		if strings.Contains(c.apiKey, "/") {
			candidates = append(candidates, strings.ReplaceAll(encoded, "/", `\/`))
		}
	}
	forms := make([]string, 0, len(candidates))
	for _, form := range candidates {
		if !slices.Contains(forms, form) {
			forms = append(forms, form)
		}
	}
	return forms
}

// nextPercentKey returns the byte span of the next query-escaped, path-escaped, or URL-path-escaped API key match in s with percent-escape hex case ignored, or -1, -1 when absent.
func (c *Client) nextPercentKey(s string) (int, int) {
	if c.apiKey == "" {
		return -1, -1
	}
	candidates := []string{url.QueryEscape(c.apiKey), url.PathEscape(c.apiKey), (&url.URL{Path: c.apiKey}).EscapedPath()}
	forms := make([]string, 0, len(candidates))
	for _, form := range candidates {
		if !slices.Contains(forms, form) {
			forms = append(forms, form)
		}
	}
	for i := range len(s) {
		for _, form := range forms {
			if i+len(form) <= len(s) && matchPercentForm(s[i:i+len(form)], form) {
				return i, i + len(form)
			}
		}
	}
	return -1, -1
}

// matchPercentForm reports whether s (of length len(form)) equals form, ignoring hex-digit case for the two bytes after each '%' in form.
func matchPercentForm(s, form string) bool {
	for i, hex := 0, 0; i < len(form); i++ {
		got, want := s[i], form[i]
		if hex > 0 {
			hex--
			if got >= 'a' && got <= 'f' {
				got -= 'a' - 'A'
			}
		} else if want == '%' {
			hex = 2
		}
		if got != want {
			return false
		}
	}
	return true
}

type redactedError struct {
	msg string
	// Why: no Unwrap method, so callers cannot reach text that holds the key.
	err error
	// Why: As sanitizes a copied URL error with the same key scrub as the message.
	c *Client
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

// Timeout reports whether the underlying error timed out.
func (e *redactedError) Timeout() bool {
	reported := false
	e.walk(func(node error) bool {
		if _, ok := node.(*url.Error); ok { //nolint:errorlint // skip the URL wrapper and visit its cause
			return false
		}
		t, ok := node.(interface{ Timeout() bool })
		if !ok {
			if a, hasAs := node.(interface{ As(any) bool }); hasAs {
				ok = a.As(&t) && !isNilValue(t)
			}
		}
		if ok {
			reported = t.Timeout()
		}
		return ok
	})
	return reported
}

// Temporary reports whether the hidden error is temporary.
func (e *redactedError) Temporary() bool {
	reported := false
	e.walk(func(node error) bool {
		if _, ok := node.(*url.Error); ok { //nolint:errorlint // skip the URL wrapper and visit its cause
			return false
		}
		t, ok := node.(interface{ Temporary() bool })
		if !ok {
			if a, hasAs := node.(interface{ As(any) bool }); hasAs {
				ok = a.As(&t) && !isNilValue(t)
			}
		}
		if ok {
			reported = t.Temporary()
		}
		return ok
	})
	return reported
}

// walk visits the hidden tree depth-first, in errors.Is order, and stops when visit reports true, visited nodes exceed maxWalkSteps, or joined errors exceed maxErrNodes.
func (e *redactedError) walk(visit func(error) bool) {
	// Why: errors.Is and errors.As visit every path, exponential on a fan-out tree; only joined errors spend the maxErrNodes budget, so a long quiet chain stays reachable.
	// Why: cyclic single-unwrap chain must terminate; bound far above any real chain.
	if isNilError(e.err) {
		return
	}
	stack := [][]error{{e.err}}
	for steps, branches := 0, 0; steps < maxWalkSteps && len(stack) > 0; steps++ {
		var node error
		stack, node = popWalkNode(stack)
		if visit(node) {
			return
		}
		children, joined := wrappedChildren(node)
		if joined {
			branches++
		}
		if branches > maxErrNodes {
			return
		}
		if children != nil {
			stack = append(stack, children)
		}
	}
}

// popWalkNode pops the next non-nil node from the top slice of stack.
func popWalkNode(stack [][]error) ([][]error, error) {
	top := stack[len(stack)-1]
	node := top[0]
	if rest := trimNilErrors(top[1:]); rest != nil {
		stack[len(stack)-1] = rest
		return stack, node
	}
	return stack[:len(stack)-1], node
}

// wrappedChildren returns the non-nil child slice err wraps directly and whether err joins several.
func wrappedChildren(err error) ([]error, bool) {
	if isNilError(err) {
		return nil, false
	}
	if w := errors.Unwrap(err); !isNilError(w) {
		return []error{w}, false
	}
	joined, ok := err.(interface{ Unwrap() []error })
	if !ok {
		return nil, false
	}
	return trimNilErrors(joined.Unwrap()), true
}

// trimNilErrors returns errs starting at its first non-nil element, or nil when none exists.
func trimNilErrors(errs []error) []error {
	for i, err := range errs {
		if !isNilError(err) {
			return errs[i:]
		}
	}
	return nil
}

// isNilError reports whether err is nil or holds a nil pointer, map, slice, function, channel, or interface.
func isNilError(err error) bool {
	return isNilValue(err)
}

// isNilValue reports whether value is nil or holds a nil pointer, map, slice, function, channel, or interface.
func isNilValue(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	if slices.Contains([]reflect.Kind{reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice}, v.Kind()) {
		return v.IsNil()
	}
	return false
}

// As sets target to a sanitized copy of the first *url.Error among the first 64 breadth-first queued nodes; deeper URL errors are not found. The structural key scan examines 64 breadth-first nodes, while Is, Timeout, and Temporary visit at most 4096 nodes and 64 joined errors.
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
			*dst = &url.Error{Op: e.c.scrub(uerr.Op), URL: e.c.scrub(redactURL(e.c.scrub(uerr.URL))), Err: e.c.scrubError(uerr.Err)}
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

// maxErrNodes bounds the error tree walk in carriesKey and the joined-error budget in walk.
const maxErrNodes = 64

// maxWalkSteps bounds the total nodes visited by walk so cyclic unwrap chains terminate.
const maxWalkSteps = 4096

// redactedTreeMsg is the replacement message when an error tree exceeds maxErrNodes.
const redactedTreeMsg = "error tree too large: details redacted"

// scrubError hides error trees that carry the API key or exceed the traversal limit.
func (c *Client) scrubError(err error) error {
	carry, tooLarge := c.checkErrorTree(err)
	if !carry {
		return err
	}
	if tooLarge {
		return &redactedError{msg: redactedTreeMsg, err: err, c: c}
	}
	return &redactedError{msg: c.scrub(err.Error()), err: err, c: c}
}

// carriesKey reports whether an error tree contains the API key or exceeds the node limit.
func (c *Client) carriesKey(err error) bool {
	carry, _ := c.checkErrorTree(err)
	return carry
}

// checkErrorTree reports whether err carries the API key, a typed-nil node, or exceeds maxErrNodes, and whether its Error method is unsafe to call.
func (c *Client) checkErrorTree(err error) (bool, bool) {
	if err == nil {
		return false, false
	}
	queue := []error{err}
	for i := 0; i < maxErrNodes && i < len(queue); i++ {
		if hasTypedNil(queue[i]) {
			// Why: formatting a tree that holds a typed-nil error can dereference nil in Error().
			return true, true
		}
		queue = appendWrapped(queue, queue[i])
	}
	if len(queue) > maxErrNodes {
		// Why: a tree past the bound is hidden unread; formatting a fan-out root can be exponential.
		return true, true
	}
	for _, node := range queue {
		if c.containsKey(node.Error()) {
			return true, false
		}
	}
	return false, false
}

// hasTypedNil reports whether err is a typed-nil error or directly wraps one.
func hasTypedNil(err error) bool {
	if isNilError(err) {
		return true
	}
	if w := errors.Unwrap(err); w != nil {
		return isNilError(w)
	}
	joined, ok := err.(interface{ Unwrap() []error })
	if !ok {
		return false
	}
	for _, child := range joined.Unwrap() {
		if child != nil && isNilError(child) {
			return true
		}
	}
	return false
}

// containsKey reports whether s holds the nonempty raw, query-escaped, path-escaped, Go-quoted, or JSON-escaped API key, with hex-digit case of percent escapes ignored.
func (c *Client) containsKey(s string) bool {
	if slices.ContainsFunc(c.keyForms(), func(form string) bool { return strings.Contains(s, form) }) {
		return true
	}
	start, _ := c.nextPercentKey(s)
	return start >= 0
}

// appendWrapped appends non-nil wrapped errors while bounding queue growth.
func appendWrapped(queue []error, err error) []error {
	if len(queue) > maxErrNodes || isNilError(err) {
		// Why: carriesKey already treats this tree as too large; a typed-nil node has no children.
		return queue
	}
	if w := errors.Unwrap(err); !isNilError(w) {
		return append(queue, w)
	}
	joined, ok := err.(interface{ Unwrap() []error })
	if !ok {
		return queue
	}
	children := joined.Unwrap()
	for i := 0; i < len(children) && len(queue) <= maxErrNodes; i++ {
		if !isNilError(children[i]) {
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
		cleanOp, cleanURL := c.scrub(uerr.Op), c.scrub(redactURL(c.scrub(uerr.URL)))
		if cleanOp != uerr.Op || cleanURL != uerr.URL || (uerr.Err != nil && c.carriesKey(uerr.Err)) {
			err = &url.Error{Op: cleanOp, URL: cleanURL, Err: c.scrubError(uerr.Err)}
		}
	}
	return c.scrubError(err)
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
