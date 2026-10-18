package lidarr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetCalendar returns albums with releases between start and end (RFC 3339 timestamps).
func (c *Client) GetCalendar(ctx context.Context, start, end string, unmonitored bool) ([]Album, error) {
	path := fmt.Sprintf("/api/v1/calendar?start=%s&end=%s&unmonitored=%t",
		url.QueryEscape(start), url.QueryEscape(end), unmonitored)
	return arr.GetList[Album](ctx, c.base, path)
}

// Parse parses a release title and returns the extracted information.
func (c *Client) Parse(ctx context.Context, title string) (*ParseResult, error) {
	path := "/api/v1/parse?title=" + url.QueryEscape(title)
	return arr.GetOne[ParseResult](ctx, c.base, path)
}

// GetCalendarByID returns a single calendar entry by its ID.
func (c *Client) GetCalendarByID(ctx context.Context, id int) (*Album, error) {
	return arr.GetOne[Album](ctx, c.base, fmt.Sprintf("/api/v1/calendar/%d", id))
}

// ---------- Wanted By ID ----------.
