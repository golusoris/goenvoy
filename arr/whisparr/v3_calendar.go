package whisparr

import (
	"context"
	"fmt"
	"net/url"
)

// GetCalendar returns movies/scenes releasing between start and end dates.
func (c *ClientV3) GetCalendar(ctx context.Context, start, end string, unmonitored bool) ([]Movie, error) {
	var out []Movie
	path := fmt.Sprintf("/api/v3/calendar?start=%s&end=%s&unmonitored=%t",
		url.QueryEscape(start), url.QueryEscape(end), unmonitored)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Parse parses a title string and returns matched movie info.
func (c *ClientV3) Parse(ctx context.Context, title string) (*ParseResultV3, error) {
	var out ParseResultV3
	path := "/api/v3/parse?title=" + url.QueryEscape(title)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetCalendarByID returns a single calendar entry by ID.
func (c *ClientV3) GetCalendarByID(ctx context.Context, id int) (*Movie, error) {
	var out Movie
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v3/calendar/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------- Command Extended ----------.
