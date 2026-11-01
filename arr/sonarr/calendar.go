package sonarr

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/golusoris/goenvoy/arr/v2"
)

// CalendarFeedOptions configures optional query parameters for [Client.GetCalendarFeed].
type CalendarFeedOptions struct {
	PastDays      *int
	FutureDays    *int
	Tags          []string
	Unmonitored   bool
	PremieresOnly bool
	AsAllDay      bool
}

func (opts *CalendarFeedOptions) values() url.Values {
	if opts == nil {
		return nil
	}
	q := make(url.Values)
	if opts.PastDays != nil {
		q.Set("pastDays", strconv.Itoa(*opts.PastDays))
	}
	if opts.FutureDays != nil {
		q.Set("futureDays", strconv.Itoa(*opts.FutureDays))
	}
	if len(opts.Tags) > 0 {
		q.Set("tags", strings.Join(opts.Tags, ","))
	}
	if opts.Unmonitored {
		q.Set("unmonitored", "true")
	}
	if opts.PremieresOnly {
		q.Set("premieresOnly", "true")
	}
	if opts.AsAllDay {
		q.Set("asAllDay", "true")
	}
	return q
}

// GetCalendarFeed retrieves the Sonarr iCalendar feed in iCal format.
func (c *Client) GetCalendarFeed(ctx context.Context, opts *CalendarFeedOptions) (string, error) {
	path := "/feed/v3/calendar/sonarr.ics"
	if q := opts.values(); len(q) > 0 {
		path += "?" + q.Encode()
	}
	return c.base.GetText(ctx, path)
}

// GetCalendar returns episodes airing between start and end (RFC 3339 timestamps).
func (c *Client) GetCalendar(ctx context.Context, start, end string, unmonitored bool) ([]Episode, error) {
	path := fmt.Sprintf("/api/v3/calendar?start=%s&end=%s&unmonitored=%t",
		url.QueryEscape(start), url.QueryEscape(end), unmonitored)
	return arr.GetList[Episode](ctx, c.base, path)
}

// Parse parses a release title and returns the extracted information.
func (c *Client) Parse(ctx context.Context, title string) (*ParseResult, error) {
	path := "/api/v3/parse?title=" + url.QueryEscape(title)
	return arr.GetOne[ParseResult](ctx, c.base, path)
}

// GetCalendarByID returns a single calendar entry by its ID.
func (c *Client) GetCalendarByID(ctx context.Context, id int) (*Episode, error) {
	return arr.GetOne[Episode](ctx, c.base, fmt.Sprintf("/api/v3/calendar/%d", id))
}

// ---------- Wanted: Cutoff By ID ----------.
