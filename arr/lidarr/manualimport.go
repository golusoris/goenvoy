package lidarr

import (
	"context"
	"errors"
	"net/url"
	"strconv"

	"github.com/golusoris/goenvoy/arr/v2"
)

// ImportMode selects how a manual import treats the source files.
type ImportMode string

// Import modes accepted by the ManualImport command.
const (
	ImportModeAuto ImportMode = "auto"
	ImportModeMove ImportMode = "move"
	ImportModeCopy ImportMode = "copy"
)

// ErrNoImportFiles is returned by SendManualImport when no file is given.
var ErrNoImportFiles = errors.New("lidarr: manual import needs at least one file")

// ErrNoImportUpdates is returned by RescoreManualImport when no item is given.
var ErrNoImportUpdates = errors.New("lidarr: rescore manual import needs at least one item")

// ErrInvalidImportMode is returned by SendManualImport for a mode outside auto, move and copy.
var ErrInvalidImportMode = errors.New("lidarr: invalid manual import mode")

// ManualImportOptions selects the candidates GetManualImportCandidates lists.
//
// Field mappings to Lidarr query parameters:
//   - Folder: folder (not sent when empty)
//   - DownloadID: downloadId (not sent when empty)
//   - ArtistID: artistId (not sent when zero)
//   - FilterExistingFiles: filterExistingFiles (always sent; Lidarr defaults it to true)
//   - ReplaceExistingFiles: replaceExistingFiles (always sent; Lidarr defaults it to true)
type ManualImportOptions struct {
	Folder               string
	DownloadID           string
	ArtistID             int
	FilterExistingFiles  bool
	ReplaceExistingFiles bool
}

// ImportRejection is one reason Lidarr refuses to import a candidate.
type ImportRejection struct {
	Reason string `json:"reason"`
	// Type is "permanent" or "temporary".
	Type string `json:"type"`
}

// ManualImportCandidate is one file GET /api/v1/manualimport offers for import,
// with the artist and album Lidarr mapped it to (from grab history or parsing).
type ManualImportCandidate struct {
	ID                      int               `json:"id"`
	Path                    string            `json:"path"`
	Name                    string            `json:"name"`
	Size                    int64             `json:"size"`
	Artist                  *Artist           `json:"artist,omitempty"`
	Album                   *Album            `json:"album,omitempty"`
	AlbumReleaseID          int               `json:"albumReleaseId"`
	Tracks                  []Track           `json:"tracks"`
	Quality                 QualityModel      `json:"quality"`
	ReleaseGroup            string            `json:"releaseGroup,omitempty"`
	QualityWeight           int               `json:"qualityWeight"`
	DownloadID              string            `json:"downloadId,omitempty"`
	IndexerFlags            int               `json:"indexerFlags"`
	Rejections              []ImportRejection `json:"rejections"`
	AdditionalFile          bool              `json:"additionalFile"`
	ReplaceExistingFiles    bool              `json:"replaceExistingFiles"`
	DisableReleaseSwitching bool              `json:"disableReleaseSwitching"`
}

// ManualImportFile is one file of a ManualImport command.
type ManualImportFile struct {
	Path                    string       `json:"path"`
	ArtistID                int          `json:"artistId"`
	AlbumID                 int          `json:"albumId"`
	AlbumReleaseID          int          `json:"albumReleaseId"`
	TrackIDs                []int        `json:"trackIds"`
	Quality                 QualityModel `json:"quality"`
	ReleaseGroup            string       `json:"releaseGroup,omitempty"`
	IndexerFlags            int          `json:"indexerFlags"`
	DownloadID              string       `json:"downloadId,omitempty"`
	DisableReleaseSwitching bool         `json:"disableReleaseSwitching"`
}

// ManualImportUpdate is one item sent to POST /api/v1/manualimport to re-evaluate
// candidates against a candidate artist, album, or release mapping without importing.
type ManualImportUpdate struct {
	ID                      int          `json:"id,omitempty"`
	Path                    string       `json:"path,omitempty"`
	Name                    string       `json:"name,omitempty"`
	ArtistID                int          `json:"artistId,omitempty"`
	AlbumID                 int          `json:"albumId,omitempty"`
	AlbumReleaseID          int          `json:"albumReleaseId,omitempty"`
	TrackIDs                []int        `json:"trackIds,omitempty"`
	Quality                 QualityModel `json:"quality"`
	ReleaseGroup            string       `json:"releaseGroup,omitempty"`
	IndexerFlags            int          `json:"indexerFlags"`
	DownloadID              string       `json:"downloadId,omitempty"`
	AdditionalFile          bool         `json:"additionalFile"`
	ReplaceExistingFiles    bool         `json:"replaceExistingFiles"`
	DisableReleaseSwitching bool         `json:"disableReleaseSwitching"`
}

// ImportFile converts the candidate into a command file. It reports false when
// the candidate is not mapped to an artist and album or has no valid tracks; rejections
// are left for the caller to judge.
func (m *ManualImportCandidate) ImportFile() (ManualImportFile, bool) {
	if m.Artist == nil || m.Artist.ID == 0 || m.Album == nil || m.Album.ID == 0 || len(m.Tracks) == 0 {
		return ManualImportFile{}, false
	}
	ids := make([]int, 0, len(m.Tracks))
	for i := range m.Tracks {
		if m.Tracks[i].ID == 0 {
			return ManualImportFile{}, false
		}
		ids = append(ids, m.Tracks[i].ID)
	}
	return ManualImportFile{
		Path:                    m.Path,
		ArtistID:                m.Artist.ID,
		AlbumID:                 m.Album.ID,
		AlbumReleaseID:          m.AlbumReleaseID,
		TrackIDs:                ids,
		Quality:                 m.Quality,
		ReleaseGroup:            m.ReleaseGroup,
		IndexerFlags:            m.IndexerFlags,
		DownloadID:              m.DownloadID,
		DisableReleaseSwitching: m.DisableReleaseSwitching,
	}, true
}

// GetManualImportCandidates lists the files Lidarr offers for manual import,
// each with its artist and album mapping and its rejections.
func (c *Client) GetManualImportCandidates(ctx context.Context, opts ManualImportOptions) ([]ManualImportCandidate, error) {
	q := make(url.Values)
	if opts.Folder != "" {
		q.Set("folder", opts.Folder)
	}
	if opts.DownloadID != "" {
		q.Set("downloadId", opts.DownloadID)
	}
	if opts.ArtistID != 0 {
		q.Set("artistId", strconv.Itoa(opts.ArtistID))
	}
	q.Set("filterExistingFiles", strconv.FormatBool(opts.FilterExistingFiles))
	q.Set("replaceExistingFiles", strconv.FormatBool(opts.ReplaceExistingFiles))

	return arr.GetList[ManualImportCandidate](ctx, c.base, "/api/v1/manualimport?"+q.Encode())
}

// RescoreManualImport re-evaluates manual import candidates (POST /api/v1/manualimport)
// against given artist, album, or release mappings and returns updated candidates
// with re-evaluated rejections and track mapping.
// It imports nothing; importing is SendManualImport.
func (c *Client) RescoreManualImport(ctx context.Context, items []ManualImportUpdate) ([]ManualImportCandidate, error) {
	if len(items) == 0 {
		return nil, ErrNoImportUpdates
	}
	return arr.PostList[ManualImportCandidate](ctx, c.base, "/api/v1/manualimport", items)
}

// SendManualImport queues the ManualImport command, which imports the given files.
// The command runs asynchronously: the reply only proves it was queued.
func (c *Client) SendManualImport(ctx context.Context, mode ImportMode, files []ManualImportFile) (*arr.CommandResponse, error) {
	return arr.SendManualImport(ctx, c.base, "/api/v1/command", mode, files, ErrNoImportFiles, ErrInvalidImportMode)
}
