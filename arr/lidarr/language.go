package lidarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetLanguages returns all available languages.
func (c *Client) GetLanguages(ctx context.Context) ([]arr.LanguageResource, error) {
	var out []arr.LanguageResource
	if err := c.base.Get(ctx, "/api/v1/language", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetLanguage returns a single language by ID.
func (c *Client) GetLanguage(ctx context.Context, id int) (*arr.LanguageResource, error) {
	var out arr.LanguageResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v1/language/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetLocalization returns the localization strings for the current locale.
func (c *Client) GetLocalization(ctx context.Context) (*LocalizationResource, error) {
	var out LocalizationResource
	if err := c.base.Get(ctx, "/api/v1/localization", &out); err != nil {
		return nil, err
	}
	return &out, nil
}
