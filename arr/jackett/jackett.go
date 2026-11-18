package jackett

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	defaultTimeout = 30 * time.Second
	availableYes   = "yes"
)

// Option configures a [Client].
type Option func(*Client)

// WithHTTPClient sets a custom [http.Client].
func WithHTTPClient(c *http.Client) Option {
	return func(cl *Client) { cl.httpClient = c }
}

// WithTimeout overrides the default HTTP request timeout.
func WithTimeout(d time.Duration) Option {
	return func(cl *Client) { cl.httpClient.Timeout = d }
}

// Client is a Jackett API client.
type Client struct {
	rawBaseURL string
	apiKey     string
	httpClient *http.Client
}

// New creates a Jackett [Client] for the instance at baseURL with the given API key.
// It returns an error if baseURL is not a valid HTTP/HTTPS URL, or if it contains a query or fragment.
func New(baseURL, apiKey string, opts ...Option) (*Client, error) {
	baseURL = strings.TrimRight(baseURL, "/")
	if strings.ContainsAny(baseURL, "?#") {
		// Why: message omits baseURL because a base URL query may itself carry the key.
		return nil, errors.New("jackett: invalid base URL: query or fragment not allowed")
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, errors.New("jackett: invalid base URL: must be http(s) with a host")
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, errors.New("jackett: invalid base URL: must be http(s) with a host")
	}

	c := &Client{
		rawBaseURL: baseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: defaultTimeout, Transport: sharedTransport},
	}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

// APIError is returned when the API responds with a non-2xx status.
type APIError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e *APIError) Error() string {
	if e.Body != "" {
		return fmt.Sprintf("jackett: HTTP %d: %s", e.StatusCode, e.Body)
	}
	return fmt.Sprintf("jackett: HTTP %d", e.StatusCode)
}

// doRequest executes an HTTP request and returns the response body.
func (c *Client) doRequest(ctx context.Context, u string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("jackett: create request: %w", c.redactError(err))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("jackett: request: %w", c.redactError(err))
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("jackett: read response: %w", c.redactError(err))
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &APIError{StatusCode: resp.StatusCode, Status: c.scrub(resp.Status), Body: c.scrubBody(string(data))}
	}

	return data, nil
}

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

// redactedError presents sanitized message text and hides the error it replaces, whose text holds the API key.
type redactedError struct {
	msg string
	// Why: no Unwrap method, so callers cannot reach text that holds the key.
	err error
	// Why: As sanitizes a copied URL error with the same key scrub as the message.
	c *Client
}

// Error returns the sanitized error message.
func (e *redactedError) Error() string {
	return e.msg
}

// Is reports whether the hidden error matches target, so errors.Is keeps working without exposing it.
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

// Format prints only the sanitized message for every fmt verb and flag, so no verb reaches the hidden error.
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

// appendWrapped appends the non-nil errors that err wraps directly, single or joined, to queue, never growing it past maxErrNodes+1 errors.
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

// redactError redacts root URL errors and scrubs raw, query-escaped, quoted, and JSON-escaped API key text.
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

// redactURL replaces the apikey query parameter in raw with REDACTED; it drops the whole query when the URL or its query does not parse.
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

func (c *Client) getJSON(ctx context.Context, path string, params url.Values, v any) error {
	if params == nil {
		params = url.Values{}
	}
	params.Set("apikey", c.apiKey)
	u := c.rawBaseURL + path + "?" + params.Encode()

	data, err := c.doRequest(ctx, u)
	if err != nil {
		return err
	}

	if v != nil {
		if err := json.Unmarshal(data, v); err != nil {
			// Why: decode errors quote response text, which may echo the key.
			return fmt.Errorf("jackett: decode json: %w", c.scrubError(err))
		}
	}
	return nil
}

func (c *Client) getXML(ctx context.Context, path string, params url.Values, v any) error {
	if params == nil {
		params = url.Values{}
	}
	params.Set("apikey", c.apiKey)
	u := c.rawBaseURL + path + "?" + params.Encode()

	data, err := c.doRequest(ctx, u)
	if err != nil {
		return err
	}

	if v != nil {
		if err := xml.Unmarshal(data, v); err != nil {
			// Why: decode errors quote response text, which may echo the key.
			return fmt.Errorf("jackett: decode xml: %w", c.scrubError(err))
		}
	}
	return nil
}

func (c *Client) torznabSearch(ctx context.Context, indexer string, params url.Values) ([]SearchResult, error) {
	path := "/api/v2.0/indexers/" + url.PathEscape(indexer) + "/results/torznab"
	var rss rssResponse
	if err := c.getXML(ctx, path, params, &rss); err != nil {
		return nil, err
	}
	return convertItems(rss.Channel.Items), nil
}

// applyNumericAttr parses and assigns numeric RSS attributes.
func applyNumericAttr(r *SearchResult, name, val string) bool {
	switch name {
	case "seeders":
		n, err := strconv.Atoi(val)
		if err != nil {
			// Why: malformed indexer attr defaults to zero.
			n = 0
		}
		r.Seeders = n
		return true
	case "peers":
		n, err := strconv.Atoi(val)
		if err != nil {
			// Why: malformed indexer attr defaults to zero.
			n = 0
		}
		r.Peers = n
		return true
	default:
		return false
	}
}

// applyMetaAttr assigns metadata-related RSS attributes.
func applyMetaAttr(r *SearchResult, name, val string) bool {
	switch name {
	case "category":
		r.Category = val
	case "description":
		r.CategoryDesc = val
	case "infohash":
		r.InfoHash = val
	case "magneturl":
		r.MagnetURL = val
	case "indexer":
		r.Indexer = val
	default:
		return false
	}
	return true
}

// applyRatioAttr assigns ratio and traffic-related RSS attributes.
func applyRatioAttr(r *SearchResult, name, val string) {
	switch name {
	case "minimumratio":
		r.MinimumRatio = val
	case "minimumseedtime":
		r.MinimumSeedTime = val
	case "downloadvolumefactor":
		r.DownloadVolumeFactor = val
	case "uploadvolumefactor":
		r.UploadVolumeFactor = val
	}
}

// applyAttr maps an RSS attribute onto a SearchResult.
func applyAttr(r *SearchResult, a rssAttr) {
	if a.XMLName.Local != "attr" {
		return
	}
	if applyNumericAttr(r, a.Name, a.Value) || applyMetaAttr(r, a.Name, a.Value) {
		return
	}
	applyRatioAttr(r, a.Name, a.Value)
}

// convertItem translates an individual RSS item into a SearchResult.
func convertItem(item rssItem) SearchResult {
	r := SearchResult{
		Title:    item.Title,
		GUID:     item.GUID,
		Link:     item.Link,
		Comments: item.Comments,
		PubDate:  item.PubDate,
		Size:     item.Size,
	}
	for _, a := range item.Attrs {
		applyAttr(&r, a)
	}
	return r
}

func convertItems(items []rssItem) []SearchResult {
	results := make([]SearchResult, 0, len(items))
	for _, item := range items {
		results = append(results, convertItem(item))
	}
	return results
}

func catParams(categories []int) string {
	if len(categories) == 0 {
		return ""
	}
	s := make([]string, len(categories))
	for i, c := range categories {
		s[i] = strconv.Itoa(c)
	}
	return strings.Join(s, ",")
}

// Search queries all indexers for the given term and optional category filter.
func (c *Client) Search(ctx context.Context, query string, categories []int) ([]SearchResult, error) {
	params := url.Values{"t": {"search"}, "q": {query}}
	if cats := catParams(categories); cats != "" {
		params.Set("cat", cats)
	}
	return c.torznabSearch(ctx, "all", params)
}

// SearchIndexer queries a specific indexer by ID.
func (c *Client) SearchIndexer(ctx context.Context, indexerID, query string, categories []int) ([]SearchResult, error) {
	params := url.Values{"t": {"search"}, "q": {query}}
	if cats := catParams(categories); cats != "" {
		params.Set("cat", cats)
	}
	return c.torznabSearch(ctx, indexerID, params)
}

// TVSearch performs a TV-specific search across all indexers.
func (c *Client) TVSearch(ctx context.Context, query string, season, episode int, imdbID string) ([]SearchResult, error) {
	params := url.Values{"t": {"tvsearch"}, "q": {query}}
	if season > 0 {
		params.Set("season", strconv.Itoa(season))
	}
	if episode > 0 {
		params.Set("ep", strconv.Itoa(episode))
	}
	if imdbID != "" {
		params.Set("imdbid", imdbID)
	}
	return c.torznabSearch(ctx, "all", params)
}

// MovieSearch performs a movie-specific search across all indexers.
func (c *Client) MovieSearch(ctx context.Context, query, imdbID string) ([]SearchResult, error) {
	params := url.Values{"t": {"movie"}, "q": {query}}
	if imdbID != "" {
		params.Set("imdbid", imdbID)
	}
	return c.torznabSearch(ctx, "all", params)
}

// MusicSearch performs a music-specific search across all indexers.
func (c *Client) MusicSearch(ctx context.Context, query string) ([]SearchResult, error) {
	params := url.Values{"t": {"music"}, "q": {query}}
	return c.torznabSearch(ctx, "all", params)
}

// BookSearch performs a book-specific search across all indexers.
func (c *Client) BookSearch(ctx context.Context, query string) ([]SearchResult, error) {
	params := url.Values{"t": {"book"}, "q": {query}}
	return c.torznabSearch(ctx, "all", params)
}

// GetCapabilities returns the server capabilities.
func (c *Client) GetCapabilities(ctx context.Context) (*Capabilities, error) {
	params := url.Values{"t": {"caps"}}
	var raw capsResponse
	if err := c.getXML(ctx, "/api/v2.0/indexers/all/results/torznab", params, &raw); err != nil {
		return nil, err
	}
	return convertCaps(&raw), nil
}

func convertCaps(raw *capsResponse) *Capabilities {
	caps := &Capabilities{
		Server: ServerInfo{Title: raw.Server.Title, Image: raw.Server.Image},
		Limits: Limits{Max: raw.Limits.Max, Default: raw.Limits.Default},
		Searching: Searching{
			SearchAvailable:      raw.Searching.Search.Available == availableYes,
			TVSearchAvailable:    raw.Searching.TVSearch.Available == availableYes,
			MovieSearchAvailable: raw.Searching.MovieSearch.Available == availableYes,
			MusicSearchAvailable: raw.Searching.MusicSearch.Available == availableYes,
			BookSearchAvailable:  raw.Searching.BookSearch.Available == availableYes,
		},
	}
	for _, c := range raw.Categories.Categories {
		cat := Category{ID: c.ID, Name: c.Name}
		for _, sc := range c.SubCategories {
			cat.SubCategories = append(cat.SubCategories, SubCategory(sc))
		}
		caps.Categories = append(caps.Categories, cat)
	}
	return caps
}

// GetIndexers returns the list of configured indexers.
func (c *Client) GetIndexers(ctx context.Context) ([]Indexer, error) {
	var out []Indexer
	params := url.Values{"configured": {"true"}}
	if err := c.getJSON(ctx, "/api/v2.0/indexers", params, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetServerConfig returns the Jackett server configuration.
func (c *Client) GetServerConfig(ctx context.Context) (*ServerConfig, error) {
	var out ServerConfig
	if err := c.getJSON(ctx, "/api/v2.0/server/config", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
