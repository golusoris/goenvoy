package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetQualityDefinitions returns all quality definitions.
func (c *ClientV3) GetQualityDefinitions(ctx context.Context) ([]arr.QualityDefinitionResource, error) {
	var out []arr.QualityDefinitionResource
	if err := c.base.Get(ctx, "/api/v3/qualitydefinition", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetQualityDefinition returns a quality definition by ID.
func (c *ClientV3) GetQualityDefinition(ctx context.Context, id int) (*arr.QualityDefinitionResource, error) {
	var out arr.QualityDefinitionResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v3/qualitydefinition/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateQualityDefinition updates a quality definition.
func (c *ClientV3) UpdateQualityDefinition(ctx context.Context, qd *arr.QualityDefinitionResource) (*arr.QualityDefinitionResource, error) {
	var out arr.QualityDefinitionResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v3/qualitydefinition/%d", qd.ID), qd, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// BulkUpdateQualityDefinitions updates multiple quality definitions.
func (c *ClientV3) BulkUpdateQualityDefinitions(ctx context.Context, defs []arr.QualityDefinitionResource) ([]arr.QualityDefinitionResource, error) {
	var out []arr.QualityDefinitionResource
	if err := c.base.Put(ctx, "/api/v3/qualitydefinition/update", defs, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetQualityDefinitionLimits returns the quality definition limits.
func (c *ClientV3) GetQualityDefinitionLimits(ctx context.Context) (*QualityDefinitionLimitsResource, error) {
	var out QualityDefinitionLimitsResource
	if err := c.base.Get(ctx, "/api/v3/qualitydefinition/limits", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------- Quality Profile Extended ----------.
