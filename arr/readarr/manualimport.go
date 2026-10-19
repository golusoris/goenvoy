package readarr

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
var ErrNoImportFiles = errors.New("readarr: manual import needs at least one file")

// ErrInvalidImportMode is returned by SendManualImport for a mode outside auto, move and copy.
var ErrInvalidImportMode = errors.New("readarr: invalid manual import mode")

// ManualImportOptions selects the candidates GetManualImportCandidates lists.
//
// Field mappings to Readarr query parameters:
//   - Folder: folder (not sent when empty)
//   - DownloadID: downloadId (not sent when empty)
//   - AuthorID: authorId (not sent when zero)
//   - FilterExistingFiles: filterExistingFiles (always sent; Readarr defaults it to true)
//   - ReplaceExistingFiles: replaceExistingFiles (always sent; Readarr defaults it to true)
type ManualImportOptions struct {
	Folder               string
	DownloadID           string
	AuthorID             int
	FilterExistingFiles  bool
	ReplaceExistingFiles bool
}

// ImportRejection is one reason Readarr refuses to import a candidate.
type ImportRejection struct {
	Reason string `json:"reason"`
	// Type is "permanent" or "temporary".
	Type string `json:"type"`
}

// ManualImportCandidate is one file GET /api/v1/manualimport offers for import,
// with the author and book Readarr mapped it to (from grab history or parsing).
type ManualImportCandidate struct {
	ID                      int               `json:"id"`
	Path                    string            `json:"path"`
	Name                    string            `json:"name"`
	Size                    int64             `json:"size"`
	Author                  *Author           `json:"author,omitempty"`
	Book                    *Book             `json:"book,omitempty"`
	ForeignEditionID        string            `json:"foreignEditionId,omitempty"`
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
	AuthorID                int          `json:"authorId"`
	BookID                  int          `json:"bookId"`
	ForeignEditionID        string       `json:"foreignEditionId,omitempty"`
	Quality                 QualityModel `json:"quality"`
	IndexerFlags            int          `json:"indexerFlags"`
	DownloadID              string       `json:"downloadId,omitempty"`
	DisableReleaseSwitching bool         `json:"disableReleaseSwitching"`
}

// ImportFile converts the candidate into a command file. It reports false when
// the candidate is not mapped to an author and book; rejections are left for the caller to judge.
func (m *ManualImportCandidate) ImportFile() (ManualImportFile, bool) {
	if m.Author == nil || m.Author.ID == 0 || m.Book == nil || m.Book.ID == 0 {
		return ManualImportFile{}, false
	}
	return ManualImportFile{
		Path:                    m.Path,
		AuthorID:                m.Author.ID,
		BookID:                  m.Book.ID,
		ForeignEditionID:        m.ForeignEditionID,
		Quality:                 m.Quality,
		IndexerFlags:            m.IndexerFlags,
		DownloadID:              m.DownloadID,
		DisableReleaseSwitching: m.DisableReleaseSwitching,
	}, true
}

// GetManualImportCandidates lists the files Readarr offers for manual import,
// each with its author and book mapping and its rejections.
func (c *Client) GetManualImportCandidates(ctx context.Context, opts ManualImportOptions) ([]ManualImportCandidate, error) {
	q := make(url.Values)
	if opts.Folder != "" {
		q.Set("folder", opts.Folder)
	}
	if opts.DownloadID != "" {
		q.Set("downloadId", opts.DownloadID)
	}
	if opts.AuthorID != 0 {
		q.Set("authorId", strconv.Itoa(opts.AuthorID))
	}
	q.Set("filterExistingFiles", strconv.FormatBool(opts.FilterExistingFiles))
	q.Set("replaceExistingFiles", strconv.FormatBool(opts.ReplaceExistingFiles))

	return arr.GetList[ManualImportCandidate](ctx, c.base, "/api/v1/manualimport?"+q.Encode())
}

// SendManualImport queues the ManualImport command, which imports the given files.
// The command runs asynchronously: the reply only proves it was queued.
func (c *Client) SendManualImport(ctx context.Context, mode ImportMode, files []ManualImportFile) (*arr.CommandResponse, error) {
	return arr.SendManualImport(ctx, c.base, "/api/v1/command", mode, files, ErrNoImportFiles, ErrInvalidImportMode)
}
