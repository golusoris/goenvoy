package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetStudios returns all studios.
func (c *ClientV3) GetStudios(ctx context.Context) ([]Studio, error) {
	return arr.GetList[Studio](ctx, c.base, "/api/v3/studio")
}

// GetStudio returns a single studio by ID.
func (c *ClientV3) GetStudio(ctx context.Context, id int) (*Studio, error) {
	path := fmt.Sprintf("/api/v3/studio/%d", id)
	return arr.GetOne[Studio](ctx, c.base, path)
}

// AddStudio adds a new studio to the instance.
func (c *ClientV3) AddStudio(ctx context.Context, studio *Studio) (*Studio, error) {
	return arr.PostOne[Studio](ctx, c.base, "/api/v3/studio", studio)
}

// UpdateStudio updates an existing studio.
func (c *ClientV3) UpdateStudio(ctx context.Context, studio *Studio) (*Studio, error) {
	path := fmt.Sprintf("/api/v3/studio/%d", studio.ID)
	return arr.PutOne[Studio](ctx, c.base, path, studio)
}

// DeleteStudio removes a studio by ID.
func (c *ClientV3) DeleteStudio(ctx context.Context, id int, deleteFiles bool) error {
	path := fmt.Sprintf("/api/v3/studio/%d?deleteFiles=%t", id, deleteFiles)
	return c.base.Delete(ctx, path, nil, nil)
}

// EditStudios applies bulk edits to multiple studios.
func (c *ClientV3) EditStudios(ctx context.Context, editor *StudioEditorResource) error {
	return c.base.Put(ctx, "/api/v3/studio/editor", editor, nil)
}

// DeleteStudiosBulk deletes multiple studios according to the editor payload.
func (c *ClientV3) DeleteStudiosBulk(ctx context.Context, editor *StudioEditorResource) error {
	return c.base.Delete(ctx, "/api/v3/studio/editor", editor, nil)
}
