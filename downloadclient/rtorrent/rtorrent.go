package rtorrent

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	defaultTimeout   = 30 * time.Second
	defaultUserAgent = "goenvoy/0.0.1"
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

// WithUserAgent sets the User-Agent header sent with every request.
func WithUserAgent(ua string) Option {
	return func(cl *Client) { cl.userAgent = ua }
}

// WithAuth sets HTTP basic authentication credentials.
func WithAuth(username, password string) Option {
	return func(cl *Client) {
		cl.username = username
		cl.password = password
	}
}

// Client is an rTorrent XML-RPC client.
type Client struct {
	endpoint   string
	httpClient *http.Client
	userAgent  string
	username   string
	password   string
}

// New creates an rTorrent [Client] for the given XML-RPC endpoint URL.
func New(endpoint string, opts ...Option) *Client {
	c := &Client{
		endpoint:   endpoint,
		httpClient: &http.Client{Timeout: defaultTimeout},
		userAgent:  defaultUserAgent,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// APIError is returned when rTorrent returns an XML-RPC fault.
type APIError struct {
	FaultCode   int
	FaultString string
}

func (e *APIError) Error() string {
	if e.FaultString != "" {
		return "rtorrent: " + e.FaultString
	}
	return "rtorrent: fault code " + strconv.Itoa(e.FaultCode)
}

// XML-RPC structures.

type xmlMethodCall struct {
	XMLName    xml.Name   `xml:"methodCall"`
	MethodName string     `xml:"methodName"`
	Params     *xmlParams `xml:"params,omitempty"`
}

type xmlParams struct {
	Param []xmlParam `xml:"param"`
}

type xmlParam struct {
	Value xmlValue `xml:"value"`
}

type xmlValue struct {
	String  *string   `xml:"string,omitempty"`
	Int     *int      `xml:"i4,omitempty"`
	Int8    *int      `xml:"i8,omitempty"`
	Boolean *int      `xml:"boolean,omitempty"`
	Array   *xmlArray `xml:"array,omitempty"`
}

type xmlArray struct {
	Data xmlData `xml:"data"`
}

type xmlData struct {
	Values []xmlValue `xml:"value"`
}

// xmlMethodResponse represents the XML-RPC response.
type xmlMethodResponse struct {
	XMLName xml.Name   `xml:"methodResponse"`
	Params  *xmlParams `xml:"params,omitempty"`
	Fault   *xmlFault  `xml:"fault,omitempty"`
}

type xmlFault struct {
	Value xmlValue `xml:"value"`
}

func stringVal(s string) xmlValue {
	return xmlValue{String: &s}
}

func buildMulticall(view string, commands []string) xmlMethodCall {
	params := &xmlParams{}
	// First param: empty string (prefix), second param: view name
	params.Param = append(params.Param, xmlParam{Value: stringVal("")}, xmlParam{Value: stringVal(view)})
	for _, cmd := range commands {
		params.Param = append(params.Param, xmlParam{Value: stringVal(cmd)})
	}
	return xmlMethodCall{
		MethodName: "d.multicall2",
		Params:     params,
	}
}

func buildCall(method string, args ...string) xmlMethodCall {
	call := xmlMethodCall{MethodName: method}
	if len(args) > 0 {
		call.Params = &xmlParams{}
		for _, a := range args {
			call.Params.Param = append(call.Params.Param, xmlParam{Value: stringVal(a)})
		}
	}
	return call
}

func (c *Client) do(ctx context.Context, call xmlMethodCall) (*xmlMethodResponse, error) {
	buf, err := xml.Marshal(call)
	if err != nil {
		return nil, fmt.Errorf("rtorrent: marshal request: %w", err)
	}
	// Prepend XML declaration
	reqBody := append([]byte(xml.Header), buf...)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("rtorrent: create request: %w", err)
	}
	req.Header.Set("Content-Type", "text/xml")
	req.Header.Set("User-Agent", c.userAgent)
	if c.username != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("rtorrent: POST: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("rtorrent: HTTP %d: %s", resp.StatusCode, string(body))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("rtorrent: read response: %w", err)
	}

	var xmlResp xmlMethodResponse
	if err := xml.Unmarshal(respBody, &xmlResp); err != nil {
		return nil, fmt.Errorf("rtorrent: decode xml: %w", err)
	}

	if xmlResp.Fault != nil {
		code, msg := parseFault(xmlResp.Fault)
		return nil, &APIError{FaultCode: code, FaultString: msg}
	}

	return &xmlResp, nil
}

func parseFault(fault *xmlFault) (code int, msg string) {
	// Fault value is a struct with faultCode and faultString members.
	// For simplicity, extract from the raw value.
	if fault.Value.String != nil {
		return 0, *fault.Value.String
	}
	if fault.Value.Int != nil {
		return *fault.Value.Int, ""
	}
	return -1, "unknown fault"
}

func (c *Client) callString(ctx context.Context, method string, args ...string) (string, error) {
	resp, err := c.do(ctx, buildCall(method, args...))
	if err != nil {
		return "", err
	}
	if resp.Params != nil && len(resp.Params.Param) > 0 {
		v := resp.Params.Param[0].Value
		if v.String != nil {
			return *v.String, nil
		}
	}
	return "", nil
}

func (c *Client) callInt(ctx context.Context, method string, args ...string) (int, error) {
	resp, err := c.do(ctx, buildCall(method, args...))
	if err != nil {
		return 0, err
	}
	if resp.Params != nil && len(resp.Params.Param) > 0 {
		v := resp.Params.Param[0].Value
		if v.Int != nil {
			return *v.Int, nil
		}
		if v.Int8 != nil {
			return *v.Int8, nil
		}
	}
	return 0, nil
}

func (c *Client) callVoid(ctx context.Context, method string, args ...string) error {
	_, err := c.do(ctx, buildCall(method, args...))
	return err
}

// Torrent methods.

var torrentFields = []string{
	"d.hash=",
	"d.name=",
	"d.size_bytes=",
	"d.completed_bytes=",
	"d.down.rate=",
	"d.up.rate=",
	"d.up.total=",
	"d.ratio=",
	"d.is_open=",
	"d.is_active=",
	"d.complete=",
	"d.is_hash_checking=",
	"d.base_path=",
	"d.directory=",
	"d.custom1=",
	"d.timestamp.started=",
	"d.message=",
}

// GetTorrents returns all torrents in the given view (default: "main").
func (c *Client) GetTorrents(ctx context.Context, view string) ([]Torrent, error) {
	if view == "" {
		view = "main"
	}
	resp, err := c.do(ctx, buildMulticall(view, torrentFields))
	if err != nil {
		return nil, err
	}

	var torrents []Torrent
	if resp.Params == nil || len(resp.Params.Param) == 0 {
		return torrents, nil
	}

	// Result is an array of arrays
	outerArray := resp.Params.Param[0].Value.Array
	if outerArray == nil {
		return torrents, nil
	}

	for _, row := range outerArray.Data.Values {
		if row.Array == nil || len(row.Array.Data.Values) < len(torrentFields) {
			continue
		}
		vals := row.Array.Data.Values
		t := Torrent{
			Hash:           getStr(vals[0]),
			Name:           getStr(vals[1]),
			SizeBytes:      getInt64(vals[2]),
			CompletedBytes: getInt64(vals[3]),
			DownRate:       getInt64(vals[4]),
			UpRate:         getInt64(vals[5]),
			UpTotal:        getInt64(vals[6]),
			Ratio:          getInt64(vals[7]),
			IsOpen:         getBool(vals[8]),
			IsActive:       getBool(vals[9]),
			IsComplete:     getBool(vals[10]),
			IsHashChecking: getBool(vals[11]),
			BasePath:       getStr(vals[12]),
			Directory:      getStr(vals[13]),
			Label:          getStr(vals[14]),
			AddedTime:      getInt64(vals[15]),
			Message:        getStr(vals[16]),
		}
		torrents = append(torrents, t)
	}
	return torrents, nil
}

func getStr(v xmlValue) string {
	if v.String != nil {
		return *v.String
	}
	return ""
}

func getInt64(v xmlValue) int64 {
	if v.Int != nil {
		return int64(*v.Int)
	}
	if v.Int8 != nil {
		return int64(*v.Int8)
	}
	// Try parsing string value
	if v.String != nil {
		if n, err := strconv.ParseInt(*v.String, 10, 64); err == nil {
			return n
		}
	}
	return 0
}

func getBool(v xmlValue) bool {
	if v.Boolean != nil {
		return *v.Boolean != 0
	}
	if v.Int != nil {
		return *v.Int != 0
	}
	return false
}

// AddTorrentURL adds a torrent by URL or magnet link.
func (c *Client) AddTorrentURL(ctx context.Context, url string) error {
	return c.callVoid(ctx, "load.start", "", url)
}

// RemoveTorrent removes a torrent by hash.
func (c *Client) RemoveTorrent(ctx context.Context, hash string) error {
	return c.callVoid(ctx, "d.erase", hash)
}

// StartTorrent starts a torrent.
func (c *Client) StartTorrent(ctx context.Context, hash string) error {
	return c.callVoid(ctx, "d.start", hash)
}

// StopTorrent stops a torrent.
func (c *Client) StopTorrent(ctx context.Context, hash string) error {
	return c.callVoid(ctx, "d.stop", hash)
}

// PauseTorrent pauses a torrent.
func (c *Client) PauseTorrent(ctx context.Context, hash string) error {
	return c.callVoid(ctx, "d.pause", hash)
}

// ResumeTorrent resumes a paused torrent.
func (c *Client) ResumeTorrent(ctx context.Context, hash string) error {
	return c.callVoid(ctx, "d.resume", hash)
}

// RecheckTorrent rechecks torrent data.
func (c *Client) RecheckTorrent(ctx context.Context, hash string) error {
	return c.callVoid(ctx, "d.check_hash", hash)
}

// SetLabel sets the label (custom1) for a torrent.
func (c *Client) SetLabel(ctx context.Context, hash, label string) error {
	return c.callVoid(ctx, "d.custom1.set", hash, label)
}

// SetDirectory sets the download directory for a torrent.
func (c *Client) SetDirectory(ctx context.Context, hash, directory string) error {
	return c.callVoid(ctx, "d.directory.set", hash, directory)
}

// System methods.

// GetSystemInfo returns rTorrent system information.
func (c *Client) GetSystemInfo(ctx context.Context) (*SystemInfo, error) {
	version, err := c.callString(ctx, "system.client_version")
	if err != nil {
		return nil, err
	}
	libVersion, err := c.callString(ctx, "system.library_version")
	if err != nil {
		return nil, err
	}
	hostname, err := c.callString(ctx, "system.hostname")
	if err != nil {
		return nil, err
	}
	pid, err := c.callInt(ctx, "system.pid")
	if err != nil {
		return nil, err
	}

	return &SystemInfo{
		ClientVersion:  version,
		LibraryVersion: libVersion,
		Hostname:       hostname,
		PID:            pid,
	}, nil
}

// GetDownloadRate returns the current global download rate in bytes/s.
func (c *Client) GetDownloadRate(ctx context.Context) (int, error) {
	return c.callInt(ctx, "throttle.global_down.rate")
}

// GetUploadRate returns the current global upload rate in bytes/s.
func (c *Client) GetUploadRate(ctx context.Context) (int, error) {
	return c.callInt(ctx, "throttle.global_up.rate")
}

// SetDownloadLimit sets the global download limit in bytes/s. Pass 0 for unlimited.
func (c *Client) SetDownloadLimit(ctx context.Context, limit int) error {
	return c.callVoid(ctx, "throttle.global_down.max_rate.set", "", strconv.Itoa(limit))
}

// SetUploadLimit sets the global upload limit in bytes/s. Pass 0 for unlimited.
func (c *Client) SetUploadLimit(ctx context.Context, limit int) error {
	return c.callVoid(ctx, "throttle.global_up.max_rate.set", "", strconv.Itoa(limit))
}
