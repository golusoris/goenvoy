package sonarr

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
var ErrNoImportFiles = errors.New("sonarr: manual import needs at least one file")

// ErrInvalidImportMode is returned by SendManualImport for a mode outside auto, move and copy.
var ErrInvalidImportMode = errors.New("sonarr: invalid manual import mode")

// ManualImportOptions selects the candidates GetManualImportCandidates lists.
//
// Field mappings to Sonarr query parameters:
//   - Folder: folder (not sent when empty)
//   - DownloadID: downloadId (not sent when empty)
//   - SeriesID: seriesId (not sent when zero)
//   - SeasonNumber: seasonNumber (not sent when nil)
//   - FilterExistingFiles: filterExistingFiles (always sent; Sonarr defaults it to true)
type ManualImportOptions struct {
	Folder              string
	DownloadID          string
	SeriesID            int
	SeasonNumber        *int
	FilterExistingFiles bool
}

// ImportRejection is one reason Sonarr refuses to import a candidate.
type ImportRejection struct {
	Reason string `json:"reason"`
	// Type is "permanent" or "temporary".
	Type string `json:"type"`
}

// ManualImportCandidate is one file GET /api/v3/manualimport offers for import,
// with the series and episodes Sonarr mapped it to (from grab history or parsing).
type ManualImportCandidate struct {
	ID                int                        `json:"id"`
	Path              string                     `json:"path"`
	RelativePath      string                     `json:"relativePath"`
	FolderName        string                     `json:"folderName"`
	Name              string                     `json:"name"`
	Size              int64                      `json:"size"`
	Series            *Series                    `json:"series,omitempty"`
	SeasonNumber      *int                       `json:"seasonNumber,omitempty"`
	Episodes          []Episode                  `json:"episodes"`
	EpisodeFileID     *int                       `json:"episodeFileId,omitempty"`
	ReleaseGroup      string                     `json:"releaseGroup"`
	Quality           QualityModel               `json:"quality"`
	Languages         []Language                 `json:"languages"`
	QualityWeight     int                        `json:"qualityWeight"`
	DownloadID        string                     `json:"downloadId"`
	CustomFormats     []arr.CustomFormatResource `json:"customFormats"`
	CustomFormatScore int                        `json:"customFormatScore"`
	IndexerFlags      int                        `json:"indexerFlags"`
	// ReleaseType is "unknown", "singleEpisode", "multiEpisode" or "seasonPack".
	ReleaseType string            `json:"releaseType"`
	Rejections  []ImportRejection `json:"rejections"`
}

// ManualImportFile is one file of a ManualImport command.
type ManualImportFile struct {
	Path          string       `json:"path"`
	FolderName    string       `json:"folderName,omitempty"`
	SeriesID      int          `json:"seriesId"`
	EpisodeIDs    []int        `json:"episodeIds"`
	EpisodeFileID *int         `json:"episodeFileId,omitempty"`
	Quality       QualityModel `json:"quality"`
	Languages     []Language   `json:"languages"`
	ReleaseGroup  string       `json:"releaseGroup,omitempty"`
	IndexerFlags  int          `json:"indexerFlags"`
	ReleaseType   string       `json:"releaseType,omitempty"`
	DownloadID    string       `json:"downloadId,omitempty"`
}

// ImportFile converts the candidate into a command file. It reports false when
// the candidate is not mapped to a series and at least one episode; rejections
// are left for the caller to judge.
func (m *ManualImportCandidate) ImportFile() (ManualImportFile, bool) {
	if m.Series == nil || m.Series.ID == 0 || len(m.Episodes) == 0 {
		return ManualImportFile{}, false
	}
	ids := make([]int, 0, len(m.Episodes))
	for i := range m.Episodes {
		if m.Episodes[i].ID == 0 {
			return ManualImportFile{}, false
		}
		ids = append(ids, m.Episodes[i].ID)
	}
	return ManualImportFile{
		Path:          m.Path,
		FolderName:    m.FolderName,
		SeriesID:      m.Series.ID,
		EpisodeIDs:    ids,
		EpisodeFileID: m.EpisodeFileID,
		Quality:       m.Quality,
		Languages:     m.Languages,
		ReleaseGroup:  m.ReleaseGroup,
		IndexerFlags:  m.IndexerFlags,
		ReleaseType:   m.ReleaseType,
		DownloadID:    m.DownloadID,
	}, true
}

// GetManualImportCandidates lists the files Sonarr offers for manual import,
// each with its series and episode mapping and its rejections.
func (c *Client) GetManualImportCandidates(ctx context.Context, opts ManualImportOptions) ([]ManualImportCandidate, error) {
	q := make(url.Values)
	if opts.Folder != "" {
		q.Set("folder", opts.Folder)
	}
	if opts.DownloadID != "" {
		q.Set("downloadId", opts.DownloadID)
	}
	if opts.SeriesID != 0 {
		q.Set("seriesId", strconv.Itoa(opts.SeriesID))
	}
	if opts.SeasonNumber != nil {
		q.Set("seasonNumber", strconv.Itoa(*opts.SeasonNumber))
	}
	q.Set("filterExistingFiles", strconv.FormatBool(opts.FilterExistingFiles))

	return arr.GetList[ManualImportCandidate](ctx, c.base, "/api/v3/manualimport?"+q.Encode())
}

// SendManualImport queues the ManualImport command, which imports the given files.
// The command runs asynchronously: the reply only proves it was queued.
func (c *Client) SendManualImport(ctx context.Context, mode ImportMode, files []ManualImportFile) (*arr.CommandResponse, error) {
	return arr.SendManualImport(ctx, c.base, "/api/v3/command", mode, files, ErrNoImportFiles, ErrInvalidImportMode)
}
