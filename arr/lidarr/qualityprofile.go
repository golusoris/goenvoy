package lidarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetQualityProfiles returns all configured quality profiles.
func (c *Client) GetQualityProfiles(ctx context.Context) ([]arr.QualityProfile, error) {
	return arr.GetList[arr.QualityProfile](ctx, c.base, "/api/v1/qualityprofile")
}

// GetQualityProfile returns a single quality profile by ID.
func (c *Client) GetQualityProfile(ctx context.Context, id int) (*arr.QualityProfile, error) {
	return arr.GetOne[arr.QualityProfile](ctx, c.base, fmt.Sprintf("/api/v1/qualityprofile/%d", id))
}

// CreateQualityProfile creates a new quality profile.
func (c *Client) CreateQualityProfile(ctx context.Context, p *arr.QualityProfile) (*arr.QualityProfile, error) {
	return arr.PostOne[arr.QualityProfile](ctx, c.base, "/api/v1/qualityprofile", p)
}

// UpdateQualityProfile updates an existing quality profile.
func (c *Client) UpdateQualityProfile(ctx context.Context, p *arr.QualityProfile) (*arr.QualityProfile, error) {
	return arr.PutOne[arr.QualityProfile](ctx, c.base, fmt.Sprintf("/api/v1/qualityprofile/%d", p.ID), p)
}

// DeleteQualityProfile deletes a quality profile by ID.
func (c *Client) DeleteQualityProfile(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v1/qualityprofile/%d", id), nil, nil)
}

// GetQualityProfileSchema returns the quality profile schema.
func (c *Client) GetQualityProfileSchema(ctx context.Context) (*arr.QualityProfile, error) {
	return arr.GetOne[arr.QualityProfile](ctx, c.base, "/api/v1/qualityprofile/schema")
}

// ---------- Quality Definitions ----------.
