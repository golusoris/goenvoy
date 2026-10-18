package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetPerformers returns all performers.
func (c *ClientV3) GetPerformers(ctx context.Context) ([]Performer, error) {
	return arr.GetList[Performer](ctx, c.base, "/api/v3/performer")
}

// GetPerformer returns a single performer by ID.
func (c *ClientV3) GetPerformer(ctx context.Context, id int) (*Performer, error) {
	path := fmt.Sprintf("/api/v3/performer/%d", id)
	return arr.GetOne[Performer](ctx, c.base, path)
}

// AddPerformer adds a new performer to the instance.
func (c *ClientV3) AddPerformer(ctx context.Context, performer *Performer) (*Performer, error) {
	return arr.PostOne[Performer](ctx, c.base, "/api/v3/performer", performer)
}

// UpdatePerformer updates an existing performer.
func (c *ClientV3) UpdatePerformer(ctx context.Context, performer *Performer) (*Performer, error) {
	path := fmt.Sprintf("/api/v3/performer/%d", performer.ID)
	return arr.PutOne[Performer](ctx, c.base, path, performer)
}

// DeletePerformer removes a performer by ID.
func (c *ClientV3) DeletePerformer(ctx context.Context, id int, deleteFiles bool) error {
	path := fmt.Sprintf("/api/v3/performer/%d?deleteFiles=%t", id, deleteFiles)
	return c.base.Delete(ctx, path, nil, nil)
}

// EditPerformers applies bulk edits to multiple performers.
func (c *ClientV3) EditPerformers(ctx context.Context, editor *PerformerEditorResource) error {
	return c.base.Put(ctx, "/api/v3/performer/editor", editor, nil)
}

// DeletePerformersBulk deletes multiple performers according to the editor payload.
func (c *ClientV3) DeletePerformersBulk(ctx context.Context, editor *PerformerEditorResource) error {
	return c.base.Delete(ctx, "/api/v3/performer/editor", editor, nil)
}
