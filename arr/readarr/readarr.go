package readarr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/lusoris/goenvoy/arr"
)

// Client is a Readarr API client.
type Client struct {
	base *arr.BaseClient
}

// New creates a Readarr [Client] for the instance at baseURL.
func New(baseURL, apiKey string, opts ...arr.Option) (*Client, error) {
	base, err := arr.NewBaseClient(baseURL, apiKey, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{base: base}, nil
}

// GetAllAuthors returns every author configured in Readarr.
func (c *Client) GetAllAuthors(ctx context.Context) ([]Author, error) {
	var out []Author
	if err := c.base.Get(ctx, "/api/v1/author", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetAuthor returns a single author by its database ID.
func (c *Client) GetAuthor(ctx context.Context, id int) (*Author, error) {
	var out Author
	path := fmt.Sprintf("/api/v1/author/%d", id)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AddAuthor adds a new author to Readarr.
func (c *Client) AddAuthor(ctx context.Context, author *Author) (*Author, error) {
	var out Author
	if err := c.base.Post(ctx, "/api/v1/author", author, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateAuthor updates an existing author. Set moveFiles to true to relocate
// files when the author path changes.
func (c *Client) UpdateAuthor(ctx context.Context, author *Author, moveFiles bool) (*Author, error) {
	var out Author
	path := fmt.Sprintf("/api/v1/author/%d?moveFiles=%t", author.ID, moveFiles)
	if err := c.base.Put(ctx, path, author, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteAuthor removes an author. Set deleteFiles to true to also delete
// downloaded files from disk.
func (c *Client) DeleteAuthor(ctx context.Context, id int, deleteFiles, addImportListExclusion bool) error {
	path := fmt.Sprintf("/api/v1/author/%d?deleteFiles=%t&addImportListExclusion=%t", id, deleteFiles, addImportListExclusion)
	return c.base.Delete(ctx, path, nil)
}

// LookupAuthor searches for an author by name.
func (c *Client) LookupAuthor(ctx context.Context, term string) ([]Author, error) {
	var out []Author
	path := "/api/v1/author/lookup?term=" + url.QueryEscape(term)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetBooks returns books for the given author.
func (c *Client) GetBooks(ctx context.Context, authorID int) ([]Book, error) {
	var out []Book
	path := fmt.Sprintf("/api/v1/book?authorId=%d", authorID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetBook returns a single book by its database ID.
func (c *Client) GetBook(ctx context.Context, id int) (*Book, error) {
	var out Book
	path := fmt.Sprintf("/api/v1/book/%d", id)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AddBook adds a new book to Readarr.
func (c *Client) AddBook(ctx context.Context, book *Book) (*Book, error) {
	var out Book
	if err := c.base.Post(ctx, "/api/v1/book", book, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateBook updates an existing book.
func (c *Client) UpdateBook(ctx context.Context, book *Book) (*Book, error) {
	var out Book
	path := fmt.Sprintf("/api/v1/book/%d", book.ID)
	if err := c.base.Put(ctx, path, book, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteBook removes a book.
func (c *Client) DeleteBook(ctx context.Context, id int, deleteFiles, addImportListExclusion bool) error {
	path := fmt.Sprintf("/api/v1/book/%d?deleteFiles=%t&addImportListExclusion=%t", id, deleteFiles, addImportListExclusion)
	return c.base.Delete(ctx, path, nil)
}

// LookupBook searches for a book by term.
func (c *Client) LookupBook(ctx context.Context, term string) ([]Book, error) {
	var out []Book
	path := "/api/v1/book/lookup?term=" + url.QueryEscape(term)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// MonitorBooks sets the monitored status for a list of books.
func (c *Client) MonitorBooks(ctx context.Context, req *BooksMonitoredResource) error {
	return c.base.Put(ctx, "/api/v1/book/monitor", req, nil)
}

// GetBookFiles returns all book files for the given author.
func (c *Client) GetBookFiles(ctx context.Context, authorID int) ([]BookFile, error) {
	var out []BookFile
	path := fmt.Sprintf("/api/v1/bookfile?authorId=%d", authorID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetBookFile returns a single book file by its database ID.
func (c *Client) GetBookFile(ctx context.Context, id int) (*BookFile, error) {
	var out BookFile
	path := fmt.Sprintf("/api/v1/bookfile/%d", id)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteBookFile removes a single book file by its database ID.
func (c *Client) DeleteBookFile(ctx context.Context, id int) error {
	path := fmt.Sprintf("/api/v1/bookfile/%d", id)
	return c.base.Delete(ctx, path, nil)
}

// DeleteBookFiles removes multiple book files by their IDs.
func (c *Client) DeleteBookFiles(ctx context.Context, ids []int) error {
	body := BookFileListResource{BookFileIDs: ids}
	return c.base.Delete(ctx, "/api/v1/bookfile/bulk", &body)
}

// GetEditions returns all editions for the given book IDs.
func (c *Client) GetEditions(ctx context.Context, bookID int) ([]Edition, error) {
	var out []Edition
	path := fmt.Sprintf("/api/v1/edition?bookId=%d", bookID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetCalendar returns books with releases between start and end (RFC 3339 timestamps).
func (c *Client) GetCalendar(ctx context.Context, start, end string, unmonitored bool) ([]Book, error) {
	var out []Book
	path := fmt.Sprintf("/api/v1/calendar?start=%s&end=%s&unmonitored=%t",
		url.QueryEscape(start), url.QueryEscape(end), unmonitored)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SendCommand triggers a named command (e.g. "RefreshAuthor", "BookSearch").
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

// GetSystemStatus returns Readarr system information.
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
	return c.base.Delete(ctx, path, nil)
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

// GetWantedMissing returns books with missing files (paginated).
func (c *Client) GetWantedMissing(ctx context.Context, page, pageSize int) (*arr.PagingResource[Book], error) {
	var out arr.PagingResource[Book]
	path := fmt.Sprintf("/api/v1/wanted/missing?page=%d&pageSize=%d", page, pageSize)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetWantedCutoff returns books not meeting quality cutoff (paginated).
func (c *Client) GetWantedCutoff(ctx context.Context, page, pageSize int) (*arr.PagingResource[Book], error) {
	var out arr.PagingResource[Book]
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

// GetSeries returns book series for the given author.
func (c *Client) GetSeries(ctx context.Context, authorID int) ([]Series, error) {
	var out []Series
	path := fmt.Sprintf("/api/v1/series?authorId=%d", authorID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// EditAuthors performs a batch update on multiple authors.
func (c *Client) EditAuthors(ctx context.Context, editor *AuthorEditorResource) error {
	return c.base.Put(ctx, "/api/v1/author/editor", editor, nil)
}

// DeleteAuthors performs a batch delete of multiple authors.
func (c *Client) DeleteAuthors(ctx context.Context, editor *AuthorEditorResource) error {
	return c.base.Delete(ctx, "/api/v1/author/editor", editor)
}

// EditBooks performs a batch update on multiple books.
func (c *Client) EditBooks(ctx context.Context, editor *BookEditorResource) error {
	return c.base.Put(ctx, "/api/v1/book/editor", editor, nil)
}

// DeleteBooks performs a batch delete of multiple books.
func (c *Client) DeleteBooks(ctx context.Context, editor *BookEditorResource) error {
	return c.base.Delete(ctx, "/api/v1/book/editor", editor)
}
