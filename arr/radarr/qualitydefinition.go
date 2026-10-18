package radarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetQualityDefinitions returns all quality definitions with size limits.
func (c *Client) GetQualityDefinitions(ctx context.Context) ([]arr.QualityDefinitionResource, error) {
	return arr.GetList[arr.QualityDefinitionResource](ctx, c.base, "/api/v3/qualitydefinition")
}

// GetQualityDefinition returns a single quality definition by ID.
func (c *Client) GetQualityDefinition(ctx context.Context, id int) (*arr.QualityDefinitionResource, error) {
	return arr.GetOne[arr.QualityDefinitionResource](ctx, c.base, fmt.Sprintf("/api/v3/qualitydefinition/%d", id))
}

// UpdateQualityDefinition updates a quality definition.
func (c *Client) UpdateQualityDefinition(ctx context.Context, qd *arr.QualityDefinitionResource) (*arr.QualityDefinitionResource, error) {
	return arr.PutOne[arr.QualityDefinitionResource](ctx, c.base, fmt.Sprintf("/api/v3/qualitydefinition/%d", qd.ID), qd)
}

// ---------- Tags (full CRUD) ----------.

// UpdateQualityDefinitions performs a bulk update of quality definitions.
func (c *Client) UpdateQualityDefinitions(ctx context.Context, defs []arr.QualityDefinitionResource) ([]arr.QualityDefinitionResource, error) {
	return arr.PutList[arr.QualityDefinitionResource](ctx, c.base, "/api/v3/qualitydefinition/update", defs)
}

// ---------- Quality Profile Schema ----------.

// GetQualityDefinitionLimits returns the min/max limits for quality definitions.
func (c *Client) GetQualityDefinitionLimits(ctx context.Context) (*QualityDefinitionLimitsResource, error) {
	return arr.GetOne[QualityDefinitionLimitsResource](ctx, c.base, "/api/v3/qualitydefinition/limits")
}

// ---------- System Routes Duplicate ----------.
