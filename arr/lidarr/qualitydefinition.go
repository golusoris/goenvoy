package lidarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetQualityDefinitions returns all quality definitions.
func (c *Client) GetQualityDefinitions(ctx context.Context) ([]arr.QualityDefinitionResource, error) {
	return arr.GetList[arr.QualityDefinitionResource](ctx, c.base, "/api/v1/qualitydefinition")
}

// GetQualityDefinition returns a single quality definition by ID.
func (c *Client) GetQualityDefinition(ctx context.Context, id int) (*arr.QualityDefinitionResource, error) {
	return arr.GetOne[arr.QualityDefinitionResource](ctx, c.base, fmt.Sprintf("/api/v1/qualitydefinition/%d", id))
}

// UpdateQualityDefinition updates a single quality definition.
func (c *Client) UpdateQualityDefinition(ctx context.Context, def *arr.QualityDefinitionResource) (*arr.QualityDefinitionResource, error) {
	return arr.PutOne[arr.QualityDefinitionResource](ctx, c.base, fmt.Sprintf("/api/v1/qualitydefinition/%d", def.ID), def)
}

// UpdateQualityDefinitions performs a bulk update of quality definitions.
func (c *Client) UpdateQualityDefinitions(ctx context.Context, defs []arr.QualityDefinitionResource) ([]arr.QualityDefinitionResource, error) {
	return arr.PutList[arr.QualityDefinitionResource](ctx, c.base, "/api/v1/qualitydefinition/update", defs)
}

// ---------- Metadata Profiles ----------.
