package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetQualityDefinitions returns all quality definitions.
func (c *api) GetQualityDefinitions(ctx context.Context) ([]arr.QualityDefinitionResource, error) {
	return arr.GetList[arr.QualityDefinitionResource](ctx, c.base, "/api/v3/qualitydefinition")
}

// GetQualityDefinition returns a quality definition by ID.
func (c *api) GetQualityDefinition(ctx context.Context, id int) (*arr.QualityDefinitionResource, error) {
	return arr.GetOne[arr.QualityDefinitionResource](ctx, c.base, fmt.Sprintf("/api/v3/qualitydefinition/%d", id))
}

// UpdateQualityDefinition updates a quality definition.
func (c *api) UpdateQualityDefinition(ctx context.Context, qd *arr.QualityDefinitionResource) (*arr.QualityDefinitionResource, error) {
	return arr.PutOne[arr.QualityDefinitionResource](ctx, c.base, fmt.Sprintf("/api/v3/qualitydefinition/%d", qd.ID), qd)
}

// BulkUpdateQualityDefinitions updates multiple quality definitions.
func (c *api) BulkUpdateQualityDefinitions(ctx context.Context, defs []arr.QualityDefinitionResource) ([]arr.QualityDefinitionResource, error) {
	return arr.PutList[arr.QualityDefinitionResource](ctx, c.base, "/api/v3/qualitydefinition/update", defs)
}

// GetQualityDefinitionLimits returns the quality definition limits.
func (c *api) GetQualityDefinitionLimits(ctx context.Context) (*QualityDefinitionLimitsResource, error) {
	return arr.GetOne[QualityDefinitionLimitsResource](ctx, c.base, "/api/v3/qualitydefinition/limits")
}
