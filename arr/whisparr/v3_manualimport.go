package whisparr

import (
	"context"
	"net/url"
	"strconv"

	"github.com/golusoris/goenvoy/arr/v2"
)

// ManualImportOptionsV3 selects the candidates GetManualImportCandidates lists for Whisparr v3.
//
// Field mappings to Whisparr query parameters:
//   - Folder: folder (not sent when empty)
//   - DownloadID: downloadId (not sent when empty)
//   - MovieID: movieId (not sent when zero)
//   - FilterExistingFiles: filterExistingFiles (always sent; Whisparr defaults it to true)
type ManualImportOptionsV3 struct {
	Folder              string
	DownloadID          string
	MovieID             int
	FilterExistingFiles bool
}

// ManualImportCandidateV3 is one file GET /api/v3/manualimport offers for import in Whisparr v3,
// with the movie Whisparr mapped it to (from grab history or parsing).
type ManualImportCandidateV3 struct {
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

// ManualImportFileV3 is one file of a Whisparr v3 ManualImport command.
type ManualImportFileV3 struct {
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
func (m *ManualImportCandidateV3) ImportFile() (ManualImportFileV3, bool) {
	if m.Movie == nil || m.Movie.ID == 0 {
		return ManualImportFileV3{}, false
	}
	return ManualImportFileV3{
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

// GetManualImportCandidates lists the files Whisparr v3 offers for manual import,
// each with its movie mapping and its rejections.
func (c *ClientV3) GetManualImportCandidates(ctx context.Context, opts ManualImportOptionsV3) ([]ManualImportCandidateV3, error) {
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

	return arr.GetList[ManualImportCandidateV3](ctx, c.base, "/api/v3/manualimport?"+q.Encode())
}

// SendManualImport queues the ManualImport command, which imports the given files.
// The command runs asynchronously: the reply only proves it was queued.
func (c *ClientV3) SendManualImport(ctx context.Context, mode ImportMode, files []ManualImportFileV3) (*arr.CommandResponse, error) {
	return arr.SendManualImport(ctx, c.base, "/api/v3/command", mode, files, ErrNoImportFiles, ErrInvalidImportMode)
}
