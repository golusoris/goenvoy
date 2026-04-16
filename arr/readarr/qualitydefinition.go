package readarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetQualityDefinitions returns all quality definitions.
func (c *Client) GetQualityDefinitions(ctx context.Context) ([]arr.QualityDefinitionResource, error) {
	var out []arr.QualityDefinitionResource
	if err := c.base.Get(ctx, "/api/v1/qualitydefinition", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetQualityDefinition returns a single quality definition by its ID.
func (c *Client) GetQualityDefinition(ctx context.Context, id int) (*arr.QualityDefinitionResource, error) {
	var out arr.QualityDefinitionResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v1/qualitydefinition/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateQualityDefinition updates a single quality definition.
func (c *Client) UpdateQualityDefinition(ctx context.Context, def *arr.QualityDefinitionResource) (*arr.QualityDefinitionResource, error) {
	var out arr.QualityDefinitionResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v1/qualitydefinition/%d", def.ID), def, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// BulkUpdateQualityDefinitions updates multiple quality definitions at once.
func (c *Client) BulkUpdateQualityDefinitions(ctx context.Context, defs []arr.QualityDefinitionResource) ([]arr.QualityDefinitionResource, error) {
	var out []arr.QualityDefinitionResource
	if err := c.base.Put(ctx, "/api/v1/qualitydefinition/update", defs, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ---------- Metadata Profiles extended ----------.
