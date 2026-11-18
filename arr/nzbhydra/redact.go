package nzbhydra

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// scrub replaces the raw, query-escaped, path-escaped, Go-quoted, and JSON-escaped API key in s with REDACTED, with hex-digit case of percent escapes ignored.
func (c *Client) scrub(s string) string {
	for _, form := range c.keyForms() {
		s = strings.ReplaceAll(s, form, "REDACTED")
	}
	if re := c.keyPattern(); re != nil {
		s = re.ReplaceAllLiteralString(s, "REDACTED")
	}
	return s
}

// keyForms returns the distinct raw, query-escaped, path-escaped, Go-quoted, and JSON-escaped forms of the API key, or nil when the key is empty; hex-digit case of percent escapes is ignored by keyPattern.
func (c *Client) keyForms() []string {
	// Why: ReplaceAll with an empty old inserts at every rune and Contains with an empty key matches every string.
	if c.apiKey == "" {
		return nil
	}
	quoted := strconv.Quote(c.apiKey)
	query, path := url.QueryEscape(c.apiKey), url.PathEscape(c.apiKey)
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

// keyPattern returns a regular expression matching the query-escaped and path-escaped API key with case-insensitive percent-escape hex digits, or nil when the key is empty.
func (c *Client) keyPattern() *regexp.Regexp {
	if c.apiKey == "" {
		return nil
	}
	query, path := percentPattern(url.QueryEscape(c.apiKey)), percentPattern(url.PathEscape(c.apiKey))
	if query == path {
		return regexp.MustCompile(query)
	}
	return regexp.MustCompile(query + "|" + path)
}

// percentPattern returns a regular expression matching escaped with case-insensitive percent-escape hex digits.
func percentPattern(escaped string) string {
	out := make([]byte, 0, len(escaped)*2)
	for i := 0; i < len(escaped); i++ {
		if escaped[i] == '%' && i+2 < len(escaped) {
			out = append(out, '%')
			out = appendHexClass(out, escaped[i+1])
			out = appendHexClass(out, escaped[i+2])
			i += 2
			continue
		}
		out = append(out, regexp.QuoteMeta(escaped[i:i+1])...)
	}
	return string(out)
}

// appendHexClass appends a two-letter case class for an upper-case hex letter or d itself.
func appendHexClass(out []byte, d byte) []byte {
	if d >= 'A' && d <= 'F' {
		return append(out, '[', d+'a'-'A', d, ']')
	}
	return append(out, d)
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

// walk visits the hidden tree depth-first, in errors.Is order, and stops when visit reports true, visited nodes exceed maxWalkSteps, or joined errors exceed maxErrNodes.
func (e *redactedError) walk(visit func(error) bool) {
	// Why: errors.Is and errors.As visit every path, exponential on a fan-out tree; only joined errors spend the maxErrNodes budget, so a long quiet chain stays reachable.
	// Why: cyclic single-unwrap chain must terminate; bound far above any real chain.
	stack := []error{e.err}
	for steps, branches := 0, 0; len(stack) > 0; {
		steps++
		if steps > maxWalkSteps {
			return
		}
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
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
		for i := len(children) - 1; i >= 0; i-- {
			stack = append(stack, children[i])
		}
	}
}

// wrappedChildren returns the non-nil errors err wraps directly and whether err joins several.
func wrappedChildren(err error) ([]error, bool) {
	if w := errors.Unwrap(err); w != nil {
		return []error{w}, false
	}
	joined, ok := err.(interface{ Unwrap() []error })
	if !ok {
		return nil, false
	}
	var children []error
	for _, child := range joined.Unwrap() {
		if child != nil {
			children = append(children, child)
		}
	}
	return children, true
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
		if uerr, ok := queue[i].(*url.Error); ok { //nolint:errorlint // first URL error of a bounded walk; errors.As is the unbounded search
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

// maxErrNodes bounds the error tree walk in carriesKey and the joined-error budget in walk.
const maxErrNodes = 64

// maxWalkSteps bounds the total nodes visited by walk so cyclic unwrap chains terminate.
const maxWalkSteps = 4096

// scrubError hides error trees that carry the API key or exceed the traversal limit.
func (c *Client) scrubError(err error) error {
	if err == nil || !c.carriesKey(err) {
		return err
	}
	return &redactedError{msg: c.scrub(err.Error()), err: err, c: c}
}

// carriesKey reports whether an error tree contains the API key or exceeds the node limit.
func (c *Client) carriesKey(err error) bool {
	if err == nil {
		return false
	}
	queue := []error{err}
	for i := 0; i < maxErrNodes && i < len(queue); i++ {
		if c.containsKey(queue[i].Error()) {
			return true
		}
		queue = appendWrapped(queue, queue[i])
	}
	// Why: a tree past the bound is hidden, never trusted unread.
	return len(queue) > maxErrNodes
}

// containsKey reports whether s holds the nonempty raw, query-escaped, path-escaped, Go-quoted, or JSON-escaped API key, with hex-digit case of percent escapes ignored.
func (c *Client) containsKey(s string) bool {
	if slices.ContainsFunc(c.keyForms(), func(form string) bool { return strings.Contains(s, form) }) {
		return true
	}
	re := c.keyPattern()
	return re != nil && re.MatchString(s)
}

// appendWrapped appends non-nil wrapped errors while bounding queue growth.
func appendWrapped(queue []error, err error) []error {
	if len(queue) > maxErrNodes {
		// Why: carriesKey already treats this tree as too large; more entries only cost memory.
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
	if uerr, ok := err.(*url.Error); ok { //nolint:errorlint // root identity wanted; tree search is the defect
		// Why: copy only when sanitizing changed the URL text or hid the cause; an untouched root keeps its identity for errors.Is.
		// Why: a copy never mutates an error value net/http or a custom transport may share.
		cleanOp, cleanURL := c.scrub(uerr.Op), c.scrub(redactURL(uerr.URL))
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
