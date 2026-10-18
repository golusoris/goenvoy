package sonarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetQualityProfiles returns all configured quality profiles.
func (c *Client) GetQualityProfiles(ctx context.Context) ([]arr.QualityProfile, error) {
	return arr.GetList[arr.QualityProfile](ctx, c.base, "/api/v3/qualityprofile")
}

// GetQualityProfile returns a single quality profile by ID.
func (c *Client) GetQualityProfile(ctx context.Context, id int) (*arr.QualityProfile, error) {
	return arr.GetOne[arr.QualityProfile](ctx, c.base, fmt.Sprintf("/api/v3/qualityprofile/%d", id))
}

// CreateQualityProfile creates a new quality profile.
func (c *Client) CreateQualityProfile(ctx context.Context, qp *arr.QualityProfile) (*arr.QualityProfile, error) {
	return arr.PostOne[arr.QualityProfile](ctx, c.base, "/api/v3/qualityprofile", qp)
}

// UpdateQualityProfile updates an existing quality profile.
func (c *Client) UpdateQualityProfile(ctx context.Context, qp *arr.QualityProfile) (*arr.QualityProfile, error) {
	return arr.PutOne[arr.QualityProfile](ctx, c.base, fmt.Sprintf("/api/v3/qualityprofile/%d", qp.ID), qp)
}

// DeleteQualityProfile removes a quality profile.
func (c *Client) DeleteQualityProfile(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/qualityprofile/%d", id), nil, nil)
}

// ---------- Quality Definitions ----------.

// GetQualityProfileSchema returns the schema for quality profiles.
func (c *Client) GetQualityProfileSchema(ctx context.Context) (*arr.QualityProfile, error) {
	return arr.GetOne[arr.QualityProfile](ctx, c.base, "/api/v3/qualityprofile/schema")
}

// ---------- Root Folder ----------.
