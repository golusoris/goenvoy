package lidarr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/lusoris/goenvoy/arr"
)

// Client is a Lidarr API client.
type Client struct {
	base *arr.BaseClient
}

// New creates a Lidarr [Client] for the instance at baseURL.
func New(baseURL, apiKey string, opts ...arr.Option) (*Client, error) {
	base, err := arr.NewBaseClient(baseURL, apiKey, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{base: base}, nil
}

// GetAllArtists returns every artist configured in Lidarr.
func (c *Client) GetAllArtists(ctx context.Context) ([]Artist, error) {
	var out []Artist
	if err := c.base.Get(ctx, "/api/v1/artist", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetArtist returns a single artist by its database ID.
func (c *Client) GetArtist(ctx context.Context, id int) (*Artist, error) {
	var out Artist
	path := fmt.Sprintf("/api/v1/artist/%d", id)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AddArtist adds a new artist to Lidarr.
func (c *Client) AddArtist(ctx context.Context, artist *Artist) (*Artist, error) {
	var out Artist
	if err := c.base.Post(ctx, "/api/v1/artist", artist, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateArtist updates an existing artist. Set moveFiles to true to relocate
// files when the artist path changes.
func (c *Client) UpdateArtist(ctx context.Context, artist *Artist, moveFiles bool) (*Artist, error) {
	var out Artist
	path := fmt.Sprintf("/api/v1/artist/%d?moveFiles=%t", artist.ID, moveFiles)
	if err := c.base.Put(ctx, path, artist, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteArtist removes an artist. Set deleteFiles to true to also delete
// downloaded files from disk.
func (c *Client) DeleteArtist(ctx context.Context, id int, deleteFiles, addImportListExclusion bool) error {
	path := fmt.Sprintf("/api/v1/artist/%d?deleteFiles=%t&addImportListExclusion=%t", id, deleteFiles, addImportListExclusion)
	return c.base.Delete(ctx, path, nil, nil)
}

// LookupArtist searches for an artist by name.
func (c *Client) LookupArtist(ctx context.Context, term string) ([]Artist, error) {
	var out []Artist
	path := "/api/v1/artist/lookup?term=" + url.QueryEscape(term)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetAlbums returns albums for the given artist.
func (c *Client) GetAlbums(ctx context.Context, artistID int) ([]Album, error) {
	var out []Album
	path := fmt.Sprintf("/api/v1/album?artistId=%d", artistID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetAlbum returns a single album by its database ID.
func (c *Client) GetAlbum(ctx context.Context, id int) (*Album, error) {
	var out Album
	path := fmt.Sprintf("/api/v1/album/%d", id)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AddAlbum adds a new album to Lidarr.
func (c *Client) AddAlbum(ctx context.Context, album *Album) (*Album, error) {
	var out Album
	if err := c.base.Post(ctx, "/api/v1/album", album, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateAlbum updates an existing album.
func (c *Client) UpdateAlbum(ctx context.Context, album *Album) (*Album, error) {
	var out Album
	path := fmt.Sprintf("/api/v1/album/%d", album.ID)
	if err := c.base.Put(ctx, path, album, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteAlbum removes an album.
func (c *Client) DeleteAlbum(ctx context.Context, id int, deleteFiles, addImportListExclusion bool) error {
	path := fmt.Sprintf("/api/v1/album/%d?deleteFiles=%t&addImportListExclusion=%t", id, deleteFiles, addImportListExclusion)
	return c.base.Delete(ctx, path, nil, nil)
}

// LookupAlbum searches for an album by name.
func (c *Client) LookupAlbum(ctx context.Context, term string) ([]Album, error) {
	var out []Album
	path := "/api/v1/album/lookup?term=" + url.QueryEscape(term)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// MonitorAlbums sets the monitored status for a list of albums.
func (c *Client) MonitorAlbums(ctx context.Context, req *AlbumsMonitoredResource) error {
	return c.base.Put(ctx, "/api/v1/album/monitor", req, nil)
}

// GetTracks returns tracks for the given artist and album.
func (c *Client) GetTracks(ctx context.Context, artistID, albumID int) ([]Track, error) {
	var out []Track
	path := fmt.Sprintf("/api/v1/track?artistId=%d&albumId=%d", artistID, albumID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetTrack returns a single track by its database ID.
func (c *Client) GetTrack(ctx context.Context, id int) (*Track, error) {
	var out Track
	path := fmt.Sprintf("/api/v1/track/%d", id)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTrackFiles returns all track files for the given artist.
func (c *Client) GetTrackFiles(ctx context.Context, artistID int) ([]TrackFile, error) {
	var out []TrackFile
	path := fmt.Sprintf("/api/v1/trackfile?artistId=%d", artistID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetTrackFile returns a single track file by its database ID.
func (c *Client) GetTrackFile(ctx context.Context, id int) (*TrackFile, error) {
	var out TrackFile
	path := fmt.Sprintf("/api/v1/trackfile/%d", id)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteTrackFile removes a single track file by its database ID.
func (c *Client) DeleteTrackFile(ctx context.Context, id int) error {
	path := fmt.Sprintf("/api/v1/trackfile/%d", id)
	return c.base.Delete(ctx, path, nil, nil)
}

// DeleteTrackFiles removes multiple track files by their IDs.
func (c *Client) DeleteTrackFiles(ctx context.Context, ids []int) error {
	body := TrackFileListResource{TrackFileIDs: ids}
	return c.base.Delete(ctx, "/api/v1/trackfile/bulk", &body, nil)
}

// GetCalendar returns albums with releases between start and end (RFC 3339 timestamps).
func (c *Client) GetCalendar(ctx context.Context, start, end string, unmonitored bool) ([]Album, error) {
	var out []Album
	path := fmt.Sprintf("/api/v1/calendar?start=%s&end=%s&unmonitored=%t",
		url.QueryEscape(start), url.QueryEscape(end), unmonitored)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SendCommand triggers a named command (e.g. "RefreshArtist", "AlbumSearch").
func (c *Client) SendCommand(ctx context.Context, cmd arr.CommandRequest) (*arr.CommandResponse, error) {
	var out arr.CommandResponse
	if err := c.base.Post(ctx, "/api/v1/command", cmd, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetCommands returns all currently queued or running commands.
func (c *Client) GetCommands(ctx context.Context) ([]arr.CommandResponse, error) {
	var out []arr.CommandResponse
	if err := c.base.Get(ctx, "/api/v1/command", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetCommand returns the status of a single command by its ID.
func (c *Client) GetCommand(ctx context.Context, id int) (*arr.CommandResponse, error) {
	var out arr.CommandResponse
	path := fmt.Sprintf("/api/v1/command/%d", id)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Parse parses a release title and returns the extracted information.
func (c *Client) Parse(ctx context.Context, title string) (*ParseResult, error) {
	var out ParseResult
	path := "/api/v1/parse?title=" + url.QueryEscape(title)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Search performs a general search and returns matching artists and albums.
func (c *Client) Search(ctx context.Context, term string) ([]SearchResult, error) {
	var out []SearchResult
	path := "/api/v1/search?term=" + url.QueryEscape(term)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetSystemStatus returns Lidarr system information.
func (c *Client) GetSystemStatus(ctx context.Context) (*arr.StatusResponse, error) {
	var out arr.StatusResponse
	if err := c.base.Get(ctx, "/api/v1/system/status", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetHealth returns current health check results.
func (c *Client) GetHealth(ctx context.Context) ([]arr.HealthCheck, error) {
	var out []arr.HealthCheck
	if err := c.base.Get(ctx, "/api/v1/health", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetDiskSpace returns disk usage information for configured paths.
func (c *Client) GetDiskSpace(ctx context.Context) ([]arr.DiskSpace, error) {
	var out []arr.DiskSpace
	if err := c.base.Get(ctx, "/api/v1/diskspace", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetQueue returns the current download queue with pagination.
func (c *Client) GetQueue(ctx context.Context, page, pageSize int) (*arr.PagingResource[arr.QueueRecord], error) {
	var out arr.PagingResource[arr.QueueRecord]
	path := fmt.Sprintf("/api/v1/queue?page=%d&pageSize=%d", page, pageSize)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteQueueItem removes an item from the download queue.
func (c *Client) DeleteQueueItem(ctx context.Context, id int, removeFromClient, blocklist bool) error {
	path := fmt.Sprintf("/api/v1/queue/%d?removeFromClient=%t&blocklist=%t", id, removeFromClient, blocklist)
	return c.base.Delete(ctx, path, nil, nil)
}

// GetQualityProfiles returns all configured quality profiles.
func (c *Client) GetQualityProfiles(ctx context.Context) ([]arr.QualityProfile, error) {
	var out []arr.QualityProfile
	if err := c.base.Get(ctx, "/api/v1/qualityprofile", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetMetadataProfiles returns all configured metadata profiles.
func (c *Client) GetMetadataProfiles(ctx context.Context) ([]MetadataProfile, error) {
	var out []MetadataProfile
	if err := c.base.Get(ctx, "/api/v1/metadataprofile", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetTags returns all tags.
func (c *Client) GetTags(ctx context.Context) ([]arr.Tag, error) {
	var out []arr.Tag
	if err := c.base.Get(ctx, "/api/v1/tag", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateTag creates a new tag and returns it with its assigned ID.
func (c *Client) CreateTag(ctx context.Context, label string) (*arr.Tag, error) {
	var out arr.Tag
	if err := c.base.Post(ctx, "/api/v1/tag", arr.Tag{Label: label}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetRootFolders returns all configured root folders.
func (c *Client) GetRootFolders(ctx context.Context) ([]arr.RootFolder, error) {
	var out []arr.RootFolder
	if err := c.base.Get(ctx, "/api/v1/rootfolder", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetHistory returns the download history with pagination.
func (c *Client) GetHistory(ctx context.Context, page, pageSize int) (*arr.PagingResource[HistoryRecord], error) {
	var out arr.PagingResource[HistoryRecord]
	path := fmt.Sprintf("/api/v1/history?page=%d&pageSize=%d", page, pageSize)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetWantedMissing returns albums with missing tracks (paginated).
func (c *Client) GetWantedMissing(ctx context.Context, page, pageSize int) (*arr.PagingResource[Album], error) {
	var out arr.PagingResource[Album]
	path := fmt.Sprintf("/api/v1/wanted/missing?page=%d&pageSize=%d", page, pageSize)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetWantedCutoff returns albums not meeting quality cutoff (paginated).
func (c *Client) GetWantedCutoff(ctx context.Context, page, pageSize int) (*arr.PagingResource[Album], error) {
	var out arr.PagingResource[Album]
	path := fmt.Sprintf("/api/v1/wanted/cutoff?page=%d&pageSize=%d", page, pageSize)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetImportListExclusions returns all import list exclusions.
func (c *Client) GetImportListExclusions(ctx context.Context) ([]ImportListExclusion, error) {
	var out []ImportListExclusion
	if err := c.base.Get(ctx, "/api/v1/importlistexclusion", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// EditArtists performs a batch update on multiple artists.
func (c *Client) EditArtists(ctx context.Context, editor *ArtistEditorResource) error {
	return c.base.Put(ctx, "/api/v1/artist/editor", editor, nil)
}

// DeleteArtists performs a batch delete of multiple artists.
func (c *Client) DeleteArtists(ctx context.Context, editor *ArtistEditorResource) error {
	return c.base.Delete(ctx, "/api/v1/artist/editor", editor, nil)
}
