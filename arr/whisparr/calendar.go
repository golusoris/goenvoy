package whisparr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetCalendar returns episodes airing between start and end dates.
func (c *Client) GetCalendar(ctx context.Context, start, end string, unmonitored bool) ([]Episode, error) {
	path := fmt.Sprintf("/api/v3/calendar?start=%s&end=%s&unmonitored=%t",
		url.QueryEscape(start), url.QueryEscape(end), unmonitored)
	return arr.GetList[Episode](ctx, c.base, path)
}

// Parse parses a title string and returns matched series and episodes.
func (c *Client) Parse(ctx context.Context, title string) (*V2ParseResult, error) {
	path := "/api/v3/parse?title=" + url.QueryEscape(title)
	return arr.GetOne[V2ParseResult](ctx, c.base, path)
}

// GetCalendarByID returns a single calendar entry by ID.
func (c *Client) GetCalendarByID(ctx context.Context, id int) (*Episode, error) {
	return arr.GetOne[Episode](ctx, c.base, fmt.Sprintf("/api/v3/calendar/%d", id))
}
