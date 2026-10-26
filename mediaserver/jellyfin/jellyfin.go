package jellyfin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultTimeout    = 30 * time.Second
	defaultClient     = "goenvoy"
	defaultVersion    = "0.0.1"
	defaultDeviceID   = "goenvoy-client"
	defaultDeviceName = "GoEnvoy"
	defaultUserAgent  = "goenvoy/0.0.1"
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

// WithAccessToken sets a pre-existing access token, skipping the need to call [Client.AuthenticateByName].
func WithAccessToken(token string) Option {
	return func(cl *Client) { cl.accessToken = token }
}

// WithDeviceID sets the device identifier sent in the Authorization header.
func WithDeviceID(id string) Option {
	return func(cl *Client) { cl.deviceID = id }
}

// WithUserAgent sets the User-Agent header for all requests.
func WithUserAgent(ua string) Option {
	return func(cl *Client) { cl.userAgent = ua }
}

// Client is a Jellyfin Media Server API client.
type Client struct {
	rawBaseURL  string
	accessToken string
	httpClient  *http.Client
	clientName  string
	version     string
	deviceID    string
	deviceName  string
	userAgent   string
}

// New creates a Jellyfin [Client] for the server at baseURL.
// It returns an error if baseURL is not a valid HTTP/HTTPS URL.
func New(baseURL string, opts ...Option) (*Client, error) {
	baseURL = strings.TrimRight(baseURL, "/")

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("jellyfin: invalid base URL %q: %w", baseURL, err)
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, fmt.Errorf("jellyfin: invalid base URL %q: must be http(s) with a host", baseURL)
	}

	c := &Client{
		rawBaseURL: baseURL,
		httpClient: &http.Client{Timeout: defaultTimeout, Transport: sharedTransport},
		clientName: defaultClient,
		version:    defaultVersion,
		deviceID:   defaultDeviceID,
		deviceName: defaultDeviceName,
		userAgent:  defaultUserAgent,
	}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

var (
	// ErrEmptyName is returned when a collection name is empty.
	ErrEmptyName = errors.New("name cannot be empty")
	// ErrEmptyCollectionID is returned when a collection ID is empty.
	ErrEmptyCollectionID = errors.New("collection ID cannot be empty")
	// ErrEmptyItemIDs is returned when an item ID is blank, or no item ID is given where one is required.
	ErrEmptyItemIDs = errors.New("item IDs cannot be empty")
	// ErrEmptyUserID is returned when a user ID is empty.
	ErrEmptyUserID = errors.New("user ID cannot be empty")
	// ErrEmptySecret is returned when a quick connect secret is empty.
	ErrEmptySecret = errors.New("secret cannot be empty")
	// ErrEmptyCode is returned when a quick connect code is empty.
	ErrEmptyCode = errors.New("code cannot be empty")
)

// APIError is returned when the API responds with a non-2xx status.
type APIError struct {
	StatusCode int    `json:"-"`
	Message    string `json:"message"`
	RawBody    string `json:"-"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("jellyfin: HTTP %d: %s", e.StatusCode, e.Message)
	}
	if e.RawBody != "" {
		return fmt.Sprintf("jellyfin: HTTP %d: %s", e.StatusCode, e.RawBody)
	}
	return fmt.Sprintf("jellyfin: HTTP %d", e.StatusCode)
}

func (c *Client) authHeader() string {
	h := fmt.Sprintf("MediaBrowser Client=%q, Device=%q, DeviceId=%q, Version=%q",
		c.clientName, c.deviceName, c.deviceID, c.version)
	if c.accessToken != "" {
		h += fmt.Sprintf(", Token=%q", c.accessToken)
	}
	return h
}

// buildRequest creates an HTTP request with JSON body and Jellyfin auth headers.
func (c *Client) buildRequest(ctx context.Context, method, rawURL string, body any) (*http.Request, error) {
	var reqBody io.Reader = http.NoBody
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("jellyfin: marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, rawURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("jellyfin: create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("User-Agent", c.userAgent)
	return req, nil
}

// parseResponse checks for API errors and decodes the response body.
func parseResponse(statusCode int, respBody []byte, dst any) error {
	if statusCode < 200 || statusCode >= 300 {
		apiErr := &APIError{StatusCode: statusCode}
		if jsonErr := json.Unmarshal(respBody, apiErr); jsonErr != nil {
			apiErr.RawBody = string(respBody)
		}
		return apiErr
	}

	if dst != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, dst); err != nil {
			return fmt.Errorf("jellyfin: decode response: %w", err)
		}
	}
	return nil
}

func (c *Client) doRequest(ctx context.Context, method, path string, body, dst any) error {
	u, err := url.Parse(c.rawBaseURL + path)
	if err != nil {
		return fmt.Errorf("jellyfin: parse URL: %w", err)
	}

	req, err := c.buildRequest(ctx, method, u.String(), body)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("jellyfin: %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("jellyfin: read response: %w", err)
	}

	return parseResponse(resp.StatusCode, respBody, dst)
}

func (c *Client) get(ctx context.Context, path string, params url.Values) ([]byte, error) {
	if params != nil {
		path += "?" + params.Encode()
	}
	u, err := url.Parse(c.rawBaseURL + path)
	if err != nil {
		return nil, fmt.Errorf("jellyfin: parse URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("jellyfin: create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("jellyfin: GET %s: %w", path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("jellyfin: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &APIError{StatusCode: resp.StatusCode}
		if jsonErr := json.Unmarshal(body, apiErr); jsonErr != nil {
			apiErr.RawBody = string(body)
		}
		return nil, apiErr
	}
	return body, nil
}

// Authentication.

// AuthenticateUserByName authenticates with the server, stores the access token, and returns the result.
func (c *Client) AuthenticateUserByName(ctx context.Context, username, password string) (*AuthenticationResult, error) {
	var result AuthenticationResult
	err := c.doRequest(ctx, http.MethodPost, "/Users/AuthenticateByName", &authRequest{
		Username: username,
		Pw:       password,
	}, &result)
	if err != nil {
		return nil, err
	}
	c.accessToken = result.AccessToken
	return &result, nil
}

// AuthenticateByName authenticates with the server and stores the access token.
func (c *Client) AuthenticateByName(ctx context.Context, username, password string) error {
	_, err := c.AuthenticateUserByName(ctx, username, password)
	return err
}

// AuthenticateWithQuickConnect authenticates using an authorized Quick Connect secret.
func (c *Client) AuthenticateWithQuickConnect(ctx context.Context, secret string) (*AuthenticationResult, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, fmt.Errorf("jellyfin: authenticate with quick connect: %w", ErrEmptySecret)
	}
	var result AuthenticationResult
	err := c.doRequest(ctx, http.MethodPost, "/Users/AuthenticateWithQuickConnect", &quickConnectRequest{
		Secret: secret,
	}, &result)
	if err != nil {
		return nil, err
	}
	c.accessToken = result.AccessToken
	return &result, nil
}

// System endpoints.

// GetPublicSystemInfo returns server info without authentication.
func (c *Client) GetPublicSystemInfo(ctx context.Context) (*PublicSystemInfo, error) {
	data, err := c.get(ctx, "/System/Info/Public", nil)
	if err != nil {
		return nil, err
	}
	var info PublicSystemInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("jellyfin: decode public info: %w", err)
	}
	return &info, nil
}

// GetSystemInfo returns server info (requires authentication).
func (c *Client) GetSystemInfo(ctx context.Context) (*SystemInfo, error) {
	data, err := c.get(ctx, "/System/Info", nil)
	if err != nil {
		return nil, err
	}
	var info SystemInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("jellyfin: decode system info: %w", err)
	}
	return &info, nil
}

// Ping tests server connectivity.
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.get(ctx, "/System/Ping", nil)
	return err
}

// Users.

// GetUsers returns all users.
func (c *Client) GetUsers(ctx context.Context) ([]UserDto, error) {
	data, err := c.get(ctx, "/Users", nil)
	if err != nil {
		return nil, err
	}
	var users []UserDto
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, fmt.Errorf("jellyfin: decode users: %w", err)
	}
	return users, nil
}

// GetCurrentUser returns the current authenticated user.
func (c *Client) GetCurrentUser(ctx context.Context) (*UserDto, error) {
	data, err := c.get(ctx, "/Users/Me", nil)
	if err != nil {
		return nil, err
	}
	var user UserDto
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("jellyfin: decode user: %w", err)
	}
	return &user, nil
}

// GetUser returns a user by ID.
func (c *Client) GetUser(ctx context.Context, userID string) (*UserDto, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("jellyfin: get user: %w", ErrEmptyUserID)
	}
	data, err := c.get(ctx, "/Users/"+url.PathEscape(userID), nil)
	if err != nil {
		return nil, err
	}
	var user UserDto
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("jellyfin: decode user: %w", err)
	}
	return &user, nil
}

// Sessions.

// GetSessions returns all active sessions.
func (c *Client) GetSessions(ctx context.Context) ([]SessionInfoDto, error) {
	data, err := c.get(ctx, "/Sessions", nil)
	if err != nil {
		return nil, err
	}
	var sessions []SessionInfoDto
	if err := json.Unmarshal(data, &sessions); err != nil {
		return nil, fmt.Errorf("jellyfin: decode sessions: %w", err)
	}
	return sessions, nil
}

// Items/Libraries.

// GetItems returns items with optional query parameters.
func (c *Client) GetItems(ctx context.Context, params url.Values) (*ItemsResult, error) {
	if params == nil {
		params = url.Values{}
	}
	data, err := c.get(ctx, "/Items", params)
	if err != nil {
		return nil, err
	}
	var result ItemsResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("jellyfin: decode items: %w", err)
	}
	return &result, nil
}

// GetItemsByParent returns items in a specific folder/library.
func (c *Client) GetItemsByParent(ctx context.Context, parentID string, startIndex, limit int) (*ItemsResult, error) {
	p := url.Values{
		"ParentId": {parentID},
	}
	if startIndex >= 0 {
		p.Set("StartIndex", strconv.Itoa(startIndex))
	}
	if limit > 0 {
		p.Set("Limit", strconv.Itoa(limit))
	}
	return c.GetItems(ctx, p)
}

// GetItem returns a specific item by ID.
func (c *Client) GetItem(ctx context.Context, itemID string) (*BaseItemDto, error) {
	data, err := c.get(ctx, "/Items/"+url.PathEscape(itemID), nil)
	if err != nil {
		return nil, err
	}
	var item BaseItemDto
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, fmt.Errorf("jellyfin: decode item: %w", err)
	}
	return &item, nil
}

// GetUserViews returns the user's library views.
func (c *Client) GetUserViews(ctx context.Context, userID string) (*ItemsResult, error) {
	data, err := c.get(ctx, "/UserViews", url.Values{"userId": {userID}})
	if err != nil {
		return nil, err
	}
	var result ItemsResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("jellyfin: decode views: %w", err)
	}
	return &result, nil
}

func hasBlankItemID(ids []string) bool {
	for _, id := range ids {
		if strings.TrimSpace(id) == "" {
			return true
		}
	}
	return false
}

func hasEmptyItemIDs(ids []string) bool {
	return len(ids) == 0 || hasBlankItemID(ids)
}

// Collections.

// CreateCollection creates a new collection holding itemIDs; no itemIDs creates it empty.
func (c *Client) CreateCollection(ctx context.Context, name string, itemIDs []string, parentID string) (CollectionCreationResult, error) {
	if strings.TrimSpace(name) == "" {
		return CollectionCreationResult{}, fmt.Errorf("jellyfin: create collection: %w", ErrEmptyName)
	}
	if hasBlankItemID(itemIDs) {
		return CollectionCreationResult{}, fmt.Errorf("jellyfin: create collection: %w", ErrEmptyItemIDs)
	}

	q := url.Values{"name": {name}}
	if len(itemIDs) > 0 {
		q.Set("ids", strings.Join(itemIDs, ","))
	}
	if parentID != "" {
		q.Set("parentId", parentID)
	}

	var result CollectionCreationResult
	if err := c.doRequest(ctx, http.MethodPost, "/Collections?"+q.Encode(), nil, &result); err != nil {
		return CollectionCreationResult{}, err
	}
	return result, nil
}

// AddToCollection adds items to an existing collection.
func (c *Client) AddToCollection(ctx context.Context, collectionID string, itemIDs []string) error {
	if strings.TrimSpace(collectionID) == "" {
		return fmt.Errorf("jellyfin: add to collection: %w", ErrEmptyCollectionID)
	}
	if hasEmptyItemIDs(itemIDs) {
		return fmt.Errorf("jellyfin: add to collection: %w", ErrEmptyItemIDs)
	}

	path := "/Collections/" + url.PathEscape(collectionID) + "/Items?" + url.Values{
		"ids": {strings.Join(itemIDs, ",")},
	}.Encode()
	return c.doRequest(ctx, http.MethodPost, path, nil, nil)
}

// RemoveFromCollection removes items from a collection.
func (c *Client) RemoveFromCollection(ctx context.Context, collectionID string, itemIDs []string) error {
	if strings.TrimSpace(collectionID) == "" {
		return fmt.Errorf("jellyfin: remove from collection: %w", ErrEmptyCollectionID)
	}
	if hasEmptyItemIDs(itemIDs) {
		return fmt.Errorf("jellyfin: remove from collection: %w", ErrEmptyItemIDs)
	}

	path := "/Collections/" + url.PathEscape(collectionID) + "/Items?" + url.Values{
		"ids": {strings.Join(itemIDs, ",")},
	}.Encode()
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}

// Quick Connect.

// QuickConnectEnabled returns whether Quick Connect is enabled on the server.
func (c *Client) QuickConnectEnabled(ctx context.Context) (bool, error) {
	data, err := c.get(ctx, "/QuickConnect/Enabled", nil)
	if err != nil {
		return false, err
	}
	var enabled bool
	if err := json.Unmarshal(data, &enabled); err != nil {
		return false, fmt.Errorf("jellyfin: decode quick connect enabled: %w", err)
	}
	return enabled, nil
}

// InitiateQuickConnect initiates a new Quick Connect session.
func (c *Client) InitiateQuickConnect(ctx context.Context) (*QuickConnectResult, error) {
	var result QuickConnectResult
	if err := c.doRequest(ctx, http.MethodPost, "/QuickConnect/Initiate", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetQuickConnectState checks the state of an initiated Quick Connect session.
func (c *Client) GetQuickConnectState(ctx context.Context, secret string) (*QuickConnectResult, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, fmt.Errorf("jellyfin: get quick connect state: %w", ErrEmptySecret)
	}
	data, err := c.get(ctx, "/QuickConnect/Connect", url.Values{"secret": {secret}})
	if err != nil {
		return nil, err
	}
	var result QuickConnectResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("jellyfin: decode quick connect result: %w", err)
	}
	return &result, nil
}

// AuthorizeQuickConnect authorizes an active Quick Connect session.
func (c *Client) AuthorizeQuickConnect(ctx context.Context, code, userID string) (bool, error) {
	if strings.TrimSpace(code) == "" {
		return false, fmt.Errorf("jellyfin: authorize quick connect: %w", ErrEmptyCode)
	}
	q := url.Values{"code": {code}}
	if userID != "" {
		q.Set("userId", userID)
	}
	var authorized bool
	if err := c.doRequest(ctx, http.MethodPost, "/QuickConnect/Authorize?"+q.Encode(), nil, &authorized); err != nil {
		return false, err
	}
	return authorized, nil
}
