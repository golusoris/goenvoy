package whisparr

import (
	"context"
	"fmt"
)

// GetPerformers returns all performers.
func (c *ClientV3) GetPerformers(ctx context.Context) ([]Performer, error) {
	var out []Performer
	if err := c.base.Get(ctx, "/api/v3/performer", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetPerformer returns a single performer by ID.
func (c *ClientV3) GetPerformer(ctx context.Context, id int) (*Performer, error) {
	var out Performer
	path := fmt.Sprintf("/api/v3/performer/%d", id)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AddPerformer adds a new performer to the instance.
func (c *ClientV3) AddPerformer(ctx context.Context, performer *Performer) (*Performer, error) {
	var out Performer
	if err := c.base.Post(ctx, "/api/v3/performer", performer, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdatePerformer updates an existing performer.
func (c *ClientV3) UpdatePerformer(ctx context.Context, performer *Performer) (*Performer, error) {
	var out Performer
	path := fmt.Sprintf("/api/v3/performer/%d", performer.ID)
	if err := c.base.Put(ctx, path, performer, &out); err != nil {
		return nil, err
	}
	return &out, nil
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

// ---------- Studio Editor ----------.
