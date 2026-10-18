package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetQualityProfiles returns all quality profiles.
func (c *api) GetQualityProfiles(ctx context.Context) ([]arr.QualityProfile, error) {
	return arr.GetList[arr.QualityProfile](ctx, c.base, "/api/v3/qualityprofile")
}

// GetQualityProfile returns a quality profile by ID.
func (c *api) GetQualityProfile(ctx context.Context, id int) (*arr.QualityProfile, error) {
	return arr.GetOne[arr.QualityProfile](ctx, c.base, fmt.Sprintf("/api/v3/qualityprofile/%d", id))
}

// CreateQualityProfile creates a new quality profile.
func (c *api) CreateQualityProfile(ctx context.Context, qp *arr.QualityProfile) (*arr.QualityProfile, error) {
	return arr.PostOne[arr.QualityProfile](ctx, c.base, "/api/v3/qualityprofile", qp)
}

// UpdateQualityProfile updates an existing quality profile.
func (c *api) UpdateQualityProfile(ctx context.Context, qp *arr.QualityProfile) (*arr.QualityProfile, error) {
	return arr.PutOne[arr.QualityProfile](ctx, c.base, fmt.Sprintf("/api/v3/qualityprofile/%d", qp.ID), qp)
}

// DeleteQualityProfile deletes a quality profile by ID.
func (c *api) DeleteQualityProfile(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/qualityprofile/%d", id), nil, nil)
}

// GetQualityProfileSchema returns the quality profile schema.
func (c *api) GetQualityProfileSchema(ctx context.Context) (*arr.QualityProfile, error) {
	return arr.GetOne[arr.QualityProfile](ctx, c.base, "/api/v3/qualityprofile/schema")
}
