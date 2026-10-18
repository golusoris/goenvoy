package whisparr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetCalendar returns movies/scenes releasing between start and end dates.
func (c *ClientV3) GetCalendar(ctx context.Context, start, end string, unmonitored bool) ([]Movie, error) {
	path := fmt.Sprintf("/api/v3/calendar?start=%s&end=%s&unmonitored=%t",
		url.QueryEscape(start), url.QueryEscape(end), unmonitored)
	return arr.GetList[Movie](ctx, c.base, path)
}

// Parse parses a title string and returns matched movie info.
func (c *ClientV3) Parse(ctx context.Context, title string) (*ParseResultV3, error) {
	path := "/api/v3/parse?title=" + url.QueryEscape(title)
	return arr.GetOne[ParseResultV3](ctx, c.base, path)
}

// GetCalendarByID returns a single calendar entry by ID.
func (c *ClientV3) GetCalendarByID(ctx context.Context, id int) (*Movie, error) {
	return arr.GetOne[Movie](ctx, c.base, fmt.Sprintf("/api/v3/calendar/%d", id))
}
