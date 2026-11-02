package sonarr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// ErrInvalidSeriesID is returned when a series ID is non-positive.
var ErrInvalidSeriesID = errors.New("sonarr: invalid series ID")

// ErrEmptyFileName is returned when a media cover filename is empty.
var ErrEmptyFileName = errors.New("sonarr: empty file name")

// GetMediaCover retrieves the image bytes for a series media cover by series ID and filename.
func (c *Client) GetMediaCover(ctx context.Context, seriesID int, filename string) ([]byte, error) {
	if seriesID <= 0 {
		return nil, ErrInvalidSeriesID
	}
	if filename == "" {
		return nil, ErrEmptyFileName
	}
	path := fmt.Sprintf("/api/v3/mediacover/%d/%s", seriesID, url.PathEscape(filename))
	return c.base.GetRaw(ctx, path)
}
