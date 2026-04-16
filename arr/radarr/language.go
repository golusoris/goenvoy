package radarr

import (
	"context"
	"fmt"
)

// GetLanguages returns all available languages.
func (c *Client) GetLanguages(ctx context.Context) ([]Language, error) {
	var out []Language
	if err := c.base.Get(ctx, "/api/v3/language", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ---------- System extras ----------.

// GetLanguage returns a single language by ID.
func (c *Client) GetLanguage(ctx context.Context, id int) (*Language, error) {
	var out Language
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v3/language/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------- Localization ----------.

// GetLocalization returns the localization strings for the current locale.
func (c *Client) GetLocalization(ctx context.Context) (*LocalizationResource, error) {
	var out LocalizationResource
	if err := c.base.Get(ctx, "/api/v3/localization", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------- Log File ----------.

// GetLocalizationLanguages returns the list of available localization languages.
func (c *Client) GetLocalizationLanguages(ctx context.Context) ([]LocalizationLanguageResource, error) {
	var out []LocalizationLanguageResource
	if err := c.base.Get(ctx, "/api/v3/localization/language", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ---------- Metadata Config ----------.
