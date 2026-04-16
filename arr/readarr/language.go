package readarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetLanguages returns all languages.
func (c *Client) GetLanguages(ctx context.Context) ([]arr.LanguageResource, error) {
	var out []arr.LanguageResource
	if err := c.base.Get(ctx, "/api/v1/language", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetLanguage returns a single language by its ID.
func (c *Client) GetLanguage(ctx context.Context, id int) (*arr.LanguageResource, error) {
	var out arr.LanguageResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v1/language/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------- Localization ----------.

// GetLocalization returns the localization strings.
func (c *Client) GetLocalization(ctx context.Context) (*LocalizationResource, error) {
	var out LocalizationResource
	if err := c.base.Get(ctx, "/api/v1/localization", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------- Ping ----------.
