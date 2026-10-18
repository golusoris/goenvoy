package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// SendCommand sends a command to the instance.
func (c *api) SendCommand(ctx context.Context, cmd arr.CommandRequest) (*arr.CommandResponse, error) {
	return arr.PostOne[arr.CommandResponse](ctx, c.base, "/api/v3/command", cmd)
}

// GetCommands returns all current commands.
func (c *api) GetCommands(ctx context.Context) ([]arr.CommandResponse, error) {
	return arr.GetList[arr.CommandResponse](ctx, c.base, "/api/v3/command")
}

// GetCommand returns a command by ID.
func (c *api) GetCommand(ctx context.Context, id int) (*arr.CommandResponse, error) {
	return arr.GetOne[arr.CommandResponse](ctx, c.base, fmt.Sprintf("/api/v3/command/%d", id))
}

// DeleteCommand cancels a command by ID.
func (c *api) DeleteCommand(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/command/%d", id), nil, nil)
}
