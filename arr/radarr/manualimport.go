package radarr

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
var ErrNoImportFiles = errors.New("radarr: manual import needs at least one file")

// ErrInvalidImportMode is returned by SendManualImport for a mode outside auto, move and copy.
var ErrInvalidImportMode = errors.New("radarr: invalid manual import mode")

// ManualImportOptions selects the candidates GetManualImportCandidates lists.
//
// Field mappings to Radarr query parameters:
//   - Folder: folder (not sent when empty)
//   - DownloadID: downloadId (not sent when empty)
//   - MovieID: movieId (not sent when zero)
//   - FilterExistingFiles: filterExistingFiles (always sent; Radarr defaults it to true)
type ManualImportOptions struct {
	Folder              string
	DownloadID          string
	MovieID             int
	FilterExistingFiles bool
}

// ImportRejection is one reason Radarr refuses to import a candidate.
type ImportRejection struct {
	Reason string `json:"reason"`
	// Type is "permanent" or "temporary".
	Type string `json:"type"`
}

// ManualImportCandidate is one file GET /api/v3/manualimport offers for import,
// with the movie Radarr mapped it to (from grab history or parsing).
type ManualImportCandidate struct {
	ID                int                        `json:"id"`
	Path              string                     `json:"path"`
	RelativePath      string                     `json:"relativePath"`
	FolderName        string                     `json:"folderName"`
	Name              string                     `json:"name"`
	Size              int64                      `json:"size"`
	Movie             *Movie                     `json:"movie,omitempty"`
	MovieFileID       *int                       `json:"movieFileId,omitempty"`
	ReleaseGroup      string                     `json:"releaseGroup"`
	Quality           QualityModel               `json:"quality"`
	Languages         []Language                 `json:"languages"`
	QualityWeight     int                        `json:"qualityWeight"`
	DownloadID        string                     `json:"downloadId"`
	CustomFormats     []arr.CustomFormatResource `json:"customFormats"`
	CustomFormatScore int                        `json:"customFormatScore"`
	IndexerFlags      int                        `json:"indexerFlags"`
	Rejections        []ImportRejection          `json:"rejections"`
}

// ManualImportFile is one file of a ManualImport command.
type ManualImportFile struct {
	Path         string       `json:"path"`
	FolderName   string       `json:"folderName,omitempty"`
	MovieID      int          `json:"movieId"`
	Quality      QualityModel `json:"quality"`
	Languages    []Language   `json:"languages"`
	ReleaseGroup string       `json:"releaseGroup,omitempty"`
	IndexerFlags int          `json:"indexerFlags"`
	DownloadID   string       `json:"downloadId,omitempty"`
}

// ImportFile converts the candidate into a command file. It reports false when
// the candidate is not mapped to a movie; rejections are left for the caller to judge.
func (m *ManualImportCandidate) ImportFile() (ManualImportFile, bool) {
	if m.Movie == nil || m.Movie.ID == 0 {
		return ManualImportFile{}, false
	}
	return ManualImportFile{
		Path:         m.Path,
		FolderName:   m.FolderName,
		MovieID:      m.Movie.ID,
		Quality:      m.Quality,
		Languages:    m.Languages,
		ReleaseGroup: m.ReleaseGroup,
		IndexerFlags: m.IndexerFlags,
		DownloadID:   m.DownloadID,
	}, true
}

// GetManualImportCandidates lists the files Radarr offers for manual import,
// each with its movie mapping and its rejections.
func (c *Client) GetManualImportCandidates(ctx context.Context, opts ManualImportOptions) ([]ManualImportCandidate, error) {
	q := make(url.Values)
	if opts.Folder != "" {
		q.Set("folder", opts.Folder)
	}
	if opts.DownloadID != "" {
		q.Set("downloadId", opts.DownloadID)
	}
	if opts.MovieID != 0 {
		q.Set("movieId", strconv.Itoa(opts.MovieID))
	}
	q.Set("filterExistingFiles", strconv.FormatBool(opts.FilterExistingFiles))

	return arr.GetList[ManualImportCandidate](ctx, c.base, "/api/v3/manualimport?"+q.Encode())
}

// SendManualImport queues the ManualImport command, which imports the given files.
// The command runs asynchronously: the reply only proves it was queued.
func (c *Client) SendManualImport(ctx context.Context, mode ImportMode, files []ManualImportFile) (*arr.CommandResponse, error) {
	return arr.SendManualImport(ctx, c.base, "/api/v3/command", mode, files, ErrNoImportFiles, ErrInvalidImportMode)
}
