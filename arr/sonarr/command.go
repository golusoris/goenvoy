package sonarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// SendCommand triggers a named command (e.g. "RefreshSeries", "EpisodeSearch").
func (c *Client) SendCommand(ctx context.Context, cmd arr.CommandRequest) (*arr.CommandResponse, error) {
	return arr.PostOne[arr.CommandResponse](ctx, c.base, "/api/v3/command", cmd)
}

// GetCommands returns all currently queued or running commands.
func (c *Client) GetCommands(ctx context.Context) ([]arr.CommandResponse, error) {
	return arr.GetList[arr.CommandResponse](ctx, c.base, "/api/v3/command")
}

// GetCommand returns the status of a single command by its ID.
func (c *Client) GetCommand(ctx context.Context, id int) (*arr.CommandResponse, error) {
	var out arr.CommandResponse
	path := fmt.Sprintf("/api/v3/command/%d", id)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteCommand cancels/deletes a pending command by ID.
func (c *Client) DeleteCommand(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/command/%d", id), nil, nil)
}

// ---------- Episode File ----------.
