package mylar

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// scrub replaces the raw, query-escaped, and Go-quoted API key with REDACTED.
func (c *Client) scrub(s string) string {
	if c.apiKey == "" {
		// Why: ReplaceAll with empty old inserts at every rune.
		return s
	}
	s = strings.ReplaceAll(s, c.apiKey, "REDACTED")
	s = strings.ReplaceAll(s, url.QueryEscape(c.apiKey), "REDACTED")
	quoted := strconv.Quote(c.apiKey)
	return strings.ReplaceAll(s, quoted[1:len(quoted)-1], "REDACTED")
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
func (e *redactedError) Is(target error) bool { return errors.Is(e.err, target) }

// Timeout reports whether the hidden error timed out.
func (e *redactedError) Timeout() bool {
	// Why: url.Error.Timeout type-asserts its Err directly, so the wrapper must forward it.
	var t interface{ Timeout() bool }
	return errors.As(e.err, &t) && t.Timeout()
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
			*dst = &url.Error{Op: uerr.Op, URL: e.c.scrub(redactURL(uerr.URL)), Err: e.c.scrubError(uerr.Err)}
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

// scrubError hides error trees that carry the API key or exceed the traversal limit.
func (c *Client) scrubError(err error) error {
	if err == nil || !c.carriesKey(err) {
		return err
	}
	return &redactedError{msg: c.scrub(err.Error()), err: err, c: c}
}

// carriesKey reports whether an error tree contains the API key or exceeds the node limit.
func (c *Client) carriesKey(err error) bool {
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

// containsKey reports whether s contains the nonempty raw, query-escaped, or Go-quoted API key.
func (c *Client) containsKey(s string) bool {
	// Why: strings.Contains with an empty key matches every string.
	if c.apiKey == "" {
		return false
	}
	quoted := strconv.Quote(c.apiKey)
	return strings.Contains(s, c.apiKey) || strings.Contains(s, url.QueryEscape(c.apiKey)) ||
		strings.Contains(s, quoted[1:len(quoted)-1])
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
		// Why: root identity preserves outer wrappers and joined sentinels for nested URL errors.
		// Why: copy only the root URL error to sanitize its URL without mutating shared state.
		err = &url.Error{Op: uerr.Op, URL: c.scrub(redactURL(uerr.URL)), Err: c.scrubError(uerr.Err)}
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
