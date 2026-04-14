package stash

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultTimeout   = 30 * time.Second
	defaultUserAgent = "goenvoy/0.0.1"
)

// GraphQL field fragments.
const sceneFields = `
	id title code details director date rating100
	o_counter organized play_count play_duration resume_time
	urls
	studio { id name }
	performers { id name }
	tags { id name }
	galleries { id title }
	groups { group { id name } scene_index }
	scene_markers { id title seconds primary_tag { id name } tags { id name } }
	files { path size duration video_codec audio_codec width height frame_rate bit_rate }
	stash_ids { endpoint stash_id }
	created_at updated_at
`

const performerFields = `
	id name disambiguation gender
	birthdate death_date country ethnicity
	hair_color eye_color height_cm weight
	measurements fake_tits tattoos piercings
	career_length urls details aliases
	image_path
	tags { id name }
	scene_count image_count gallery_count
	rating100 favorite
	stash_ids { endpoint stash_id }
	created_at updated_at
`

const studioFields = `
	id name url details image_path
	parent_studio { id name }
	child_studios { id name }
	aliases scene_count
	stash_ids { endpoint stash_id }
	rating100 created_at updated_at
`

const tagFields = `
	id name description image_path aliases
	scene_count performer_count gallery_count image_count
	created_at updated_at
`

const galleryFields = `
	id title details date rating100 organized urls
	studio { id name }
	performers { id name }
	tags { id name }
	image_count created_at updated_at
`

const imageFields = `
	id title date rating100 o_counter organized
	studio { id name }
	performers { id name }
	tags { id name }
	galleries { id title }
	visual_files { ... on ImageFile { path size width height } }
	created_at updated_at
`

const groupFields = `
	id name aliases duration date rating100
	director synopsis
	studio { id name }
	urls created_at updated_at
`

const markerFields = `
	id title seconds
	primary_tag { id name }
	tags { id name }
	created_at updated_at
`

// Pre-built queries.
const (
	queryFindScene  = `query ($id: ID!) { findScene(id: $id) {` + sceneFields + `} }`
	queryFindScenes = `query ($filter: FindFilterType, $scene_filter: SceneFilterType) {
		findScenes(filter: $filter, scene_filter: $scene_filter) { count scenes {` + sceneFields + `} }
	}`
	queryFindPerformer  = `query ($id: ID!) { findPerformer(id: $id) {` + performerFields + `} }`
	queryFindPerformers = `query ($filter: FindFilterType, $performer_filter: PerformerFilterType) {
		findPerformers(filter: $filter, performer_filter: $performer_filter) { count performers {` + performerFields + `} }
	}`
	queryFindStudio  = `query ($id: ID!) { findStudio(id: $id) {` + studioFields + `} }`
	queryFindStudios = `query ($filter: FindFilterType, $studio_filter: StudioFilterType) {
		findStudios(filter: $filter, studio_filter: $studio_filter) { count studios {` + studioFields + `} }
	}`
	queryFindTag  = `query ($id: ID!) { findTag(id: $id) {` + tagFields + `} }`
	queryFindTags = `query ($filter: FindFilterType, $tag_filter: TagFilterType) {
		findTags(filter: $filter, tag_filter: $tag_filter) { count tags {` + tagFields + `} }
	}`
	queryFindGallery   = `query ($id: ID!) { findGallery(id: $id) {` + galleryFields + `} }`
	queryFindGalleries = `query ($filter: FindFilterType, $gallery_filter: GalleryFilterType) {
		findGalleries(filter: $filter, gallery_filter: $gallery_filter) { count galleries {` + galleryFields + `} }
	}`
	queryFindImage  = `query ($id: ID!) { findImage(id: $id) {` + imageFields + `} }`
	queryFindImages = `query ($filter: FindFilterType, $image_filter: ImageFilterType) {
		findImages(filter: $filter, image_filter: $image_filter) { count images {` + imageFields + `} }
	}`
	queryFindGroup  = `query ($id: ID!) { findGroup(id: $id) {` + groupFields + `} }`
	queryFindGroups = `query ($filter: FindFilterType, $group_filter: GroupFilterType) {
		findGroups(filter: $filter, group_filter: $group_filter) { count groups {` + groupFields + `} }
	}`
	queryFindMarkers = `query ($filter: FindFilterType, $scene_marker_filter: SceneMarkerFilterType) {
		findSceneMarkers(filter: $filter, scene_marker_filter: $scene_marker_filter) { count scene_markers {` + markerFields + `} }
	}`
	queryStats        = `query { stats { scene_count scenes_size scenes_duration image_count images_size gallery_count performer_count studio_count tag_count total_o_count total_play_count } }`
	queryVersion      = `query { version { version hash build_type } }`
	querySystemStatus = `query { systemStatus { databaseSchema databasePath appSchema status os } }`
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

// WithUserAgent sets the User-Agent header for all requests.
func WithUserAgent(ua string) Option {
	return func(cl *Client) { cl.userAgent = ua }
}

// Client is a Stash GraphQL API client.
type Client struct {
	endpoint   string
	apiKey     string
	httpClient *http.Client
	userAgent  string
}

// New creates a Stash [Client].
// endpoint is the full GraphQL URL (e.g., "http://localhost:9999/graphql").
func New(endpoint, apiKey string, opts ...Option) *Client {
	c := &Client{
		endpoint:   endpoint,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: defaultTimeout},
		userAgent:  defaultUserAgent,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// GraphQLError represents a GraphQL error from the API.
type GraphQLError struct {
	Message string `json:"message"`
}

func (e *GraphQLError) Error() string {
	return "stash: " + e.Message
}

type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []GraphQLError  `json:"errors"`
}

// APIError is returned when the server responds with a non-2xx status.
type APIError struct {
	StatusCode int
	RawBody    string
}

func (e *APIError) Error() string {
	if e.RawBody != "" {
		return fmt.Sprintf("stash: HTTP %d: %s", e.StatusCode, e.RawBody)
	}
	return fmt.Sprintf("stash: HTTP %d", e.StatusCode)
}

// Query sends a raw GraphQL query and unmarshals the data field into dst.
func (c *Client) Query(ctx context.Context, query string, variables map[string]any, dst any) error {
	payload, err := json.Marshal(graphQLRequest{Query: query, Variables: variables})
	if err != nil {
		return fmt.Errorf("stash: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("stash: create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Apikey", c.apiKey)
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("stash: execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("stash: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{StatusCode: resp.StatusCode, RawBody: string(body)}
	}

	var gqlResp graphQLResponse
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		return fmt.Errorf("stash: decode response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return &gqlResp.Errors[0]
	}

	if dst != nil {
		if err := json.Unmarshal(gqlResp.Data, dst); err != nil {
			return fmt.Errorf("stash: decode data: %w", err)
		}
	}
	return nil
}

// Scenes.

// FindScene returns a single scene by ID.
func (c *Client) FindScene(ctx context.Context, id string) (*Scene, error) {
	var out struct {
		FindScene *Scene `json:"findScene"`
	}
	if err := c.Query(ctx, queryFindScene, map[string]any{"id": id}, &out); err != nil {
		return nil, err
	}
	return out.FindScene, nil
}

// FindScenes returns a paginated list of scenes.
func (c *Client) FindScenes(ctx context.Context, filter *FindFilter) ([]Scene, int, error) {
	var out struct {
		FindScenes struct {
			Count  int     `json:"count"`
			Scenes []Scene `json:"scenes"`
		} `json:"findScenes"`
	}
	vars := map[string]any{"filter": filter}
	if err := c.Query(ctx, queryFindScenes, vars, &out); err != nil {
		return nil, 0, err
	}
	return out.FindScenes.Scenes, out.FindScenes.Count, nil
}

// Performers.

// FindPerformer returns a single performer by ID.
func (c *Client) FindPerformer(ctx context.Context, id string) (*Performer, error) {
	var out struct {
		FindPerformer *Performer `json:"findPerformer"`
	}
	if err := c.Query(ctx, queryFindPerformer, map[string]any{"id": id}, &out); err != nil {
		return nil, err
	}
	return out.FindPerformer, nil
}

// FindPerformers returns a paginated list of performers.
func (c *Client) FindPerformers(ctx context.Context, filter *FindFilter) ([]Performer, int, error) {
	var out struct {
		FindPerformers struct {
			Count      int         `json:"count"`
			Performers []Performer `json:"performers"`
		} `json:"findPerformers"`
	}
	vars := map[string]any{"filter": filter}
	if err := c.Query(ctx, queryFindPerformers, vars, &out); err != nil {
		return nil, 0, err
	}
	return out.FindPerformers.Performers, out.FindPerformers.Count, nil
}

// Studios.

// FindStudio returns a single studio by ID.
func (c *Client) FindStudio(ctx context.Context, id string) (*Studio, error) {
	var out struct {
		FindStudio *Studio `json:"findStudio"`
	}
	if err := c.Query(ctx, queryFindStudio, map[string]any{"id": id}, &out); err != nil {
		return nil, err
	}
	return out.FindStudio, nil
}

// FindStudios returns a paginated list of studios.
func (c *Client) FindStudios(ctx context.Context, filter *FindFilter) ([]Studio, int, error) {
	var out struct {
		FindStudios struct {
			Count   int      `json:"count"`
			Studios []Studio `json:"studios"`
		} `json:"findStudios"`
	}
	vars := map[string]any{"filter": filter}
	if err := c.Query(ctx, queryFindStudios, vars, &out); err != nil {
		return nil, 0, err
	}
	return out.FindStudios.Studios, out.FindStudios.Count, nil
}

// Tags.

// FindTag returns a single tag by ID.
func (c *Client) FindTag(ctx context.Context, id string) (*Tag, error) {
	var out struct {
		FindTag *Tag `json:"findTag"`
	}
	if err := c.Query(ctx, queryFindTag, map[string]any{"id": id}, &out); err != nil {
		return nil, err
	}
	return out.FindTag, nil
}

// FindTags returns a paginated list of tags.
func (c *Client) FindTags(ctx context.Context, filter *FindFilter) ([]Tag, int, error) {
	var out struct {
		FindTags struct {
			Count int   `json:"count"`
			Tags  []Tag `json:"tags"`
		} `json:"findTags"`
	}
	vars := map[string]any{"filter": filter}
	if err := c.Query(ctx, queryFindTags, vars, &out); err != nil {
		return nil, 0, err
	}
	return out.FindTags.Tags, out.FindTags.Count, nil
}

// Galleries.

// FindGallery returns a single gallery by ID.
func (c *Client) FindGallery(ctx context.Context, id string) (*Gallery, error) {
	var out struct {
		FindGallery *Gallery `json:"findGallery"`
	}
	if err := c.Query(ctx, queryFindGallery, map[string]any{"id": id}, &out); err != nil {
		return nil, err
	}
	return out.FindGallery, nil
}

// FindGalleries returns a paginated list of galleries.
func (c *Client) FindGalleries(ctx context.Context, filter *FindFilter) ([]Gallery, int, error) {
	var out struct {
		FindGalleries struct {
			Count     int       `json:"count"`
			Galleries []Gallery `json:"galleries"`
		} `json:"findGalleries"`
	}
	vars := map[string]any{"filter": filter}
	if err := c.Query(ctx, queryFindGalleries, vars, &out); err != nil {
		return nil, 0, err
	}
	return out.FindGalleries.Galleries, out.FindGalleries.Count, nil
}

// Images.

// FindImage returns a single image by ID.
func (c *Client) FindImage(ctx context.Context, id string) (*Image, error) {
	var out struct {
		FindImage *Image `json:"findImage"`
	}
	if err := c.Query(ctx, queryFindImage, map[string]any{"id": id}, &out); err != nil {
		return nil, err
	}
	return out.FindImage, nil
}

// FindImages returns a paginated list of images.
func (c *Client) FindImages(ctx context.Context, filter *FindFilter) ([]Image, int, error) {
	var out struct {
		FindImages struct {
			Count  int     `json:"count"`
			Images []Image `json:"images"`
		} `json:"findImages"`
	}
	vars := map[string]any{"filter": filter}
	if err := c.Query(ctx, queryFindImages, vars, &out); err != nil {
		return nil, 0, err
	}
	return out.FindImages.Images, out.FindImages.Count, nil
}

// Groups.

// FindGroup returns a single group (movie) by ID.
func (c *Client) FindGroup(ctx context.Context, id string) (*Group, error) {
	var out struct {
		FindGroup *Group `json:"findGroup"`
	}
	if err := c.Query(ctx, queryFindGroup, map[string]any{"id": id}, &out); err != nil {
		return nil, err
	}
	return out.FindGroup, nil
}

// FindGroups returns a paginated list of groups.
func (c *Client) FindGroups(ctx context.Context, filter *FindFilter) ([]Group, int, error) {
	var out struct {
		FindGroups struct {
			Count  int     `json:"count"`
			Groups []Group `json:"groups"`
		} `json:"findGroups"`
	}
	vars := map[string]any{"filter": filter}
	if err := c.Query(ctx, queryFindGroups, vars, &out); err != nil {
		return nil, 0, err
	}
	return out.FindGroups.Groups, out.FindGroups.Count, nil
}

// Scene Markers.

// FindSceneMarkers returns a paginated list of scene markers.
func (c *Client) FindSceneMarkers(ctx context.Context, filter *FindFilter) ([]SceneMarker, int, error) {
	var out struct {
		FindSceneMarkers struct {
			Count        int           `json:"count"`
			SceneMarkers []SceneMarker `json:"scene_markers"`
		} `json:"findSceneMarkers"`
	}
	vars := map[string]any{"filter": filter}
	if err := c.Query(ctx, queryFindMarkers, vars, &out); err != nil {
		return nil, 0, err
	}
	return out.FindSceneMarkers.SceneMarkers, out.FindSceneMarkers.Count, nil
}

// System.

// GetStats returns system statistics.
func (c *Client) GetStats(ctx context.Context) (*Stats, error) {
	var out struct {
		Stats *Stats `json:"stats"`
	}
	if err := c.Query(ctx, queryStats, nil, &out); err != nil {
		return nil, err
	}
	return out.Stats, nil
}

// GetVersion returns version information.
func (c *Client) GetVersion(ctx context.Context) (*Version, error) {
	var out struct {
		Version *Version `json:"version"`
	}
	if err := c.Query(ctx, queryVersion, nil, &out); err != nil {
		return nil, err
	}
	return out.Version, nil
}

// GetSystemStatus returns system status.
func (c *Client) GetSystemStatus(ctx context.Context) (*SystemStatus, error) {
	var out struct {
		SystemStatus *SystemStatus `json:"systemStatus"`
	}
	if err := c.Query(ctx, querySystemStatus, nil, &out); err != nil {
		return nil, err
	}
	return out.SystemStatus, nil
}
