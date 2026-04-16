package sonarr

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

// GetLanguageProfiles returns all language profiles.
func (c *Client) GetLanguageProfiles(ctx context.Context) ([]LanguageProfileResource, error) {
	var out []LanguageProfileResource
	if err := c.base.Get(ctx, "/api/v3/languageprofile", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetLanguageProfile returns a single language profile by ID.
func (c *Client) GetLanguageProfile(ctx context.Context, id int) (*LanguageProfileResource, error) {
	var out LanguageProfileResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v3/languageprofile/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateLanguageProfile creates a new language profile.
func (c *Client) CreateLanguageProfile(ctx context.Context, profile *LanguageProfileResource) (*LanguageProfileResource, error) {
	var out LanguageProfileResource
	if err := c.base.Post(ctx, "/api/v3/languageprofile", profile, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateLanguageProfile updates an existing language profile.
func (c *Client) UpdateLanguageProfile(ctx context.Context, profile *LanguageProfileResource) (*LanguageProfileResource, error) {
	var out LanguageProfileResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v3/languageprofile/%d", profile.ID), profile, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteLanguageProfile deletes a language profile by ID.
func (c *Client) DeleteLanguageProfile(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/languageprofile/%d", id), nil, nil)
}

// GetLanguageProfileSchema returns the available language profile schema.
func (c *Client) GetLanguageProfileSchema(ctx context.Context) (*LanguageProfileResource, error) {
	var out LanguageProfileResource
	if err := c.base.Get(ctx, "/api/v3/languageprofile/schema", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------- Localization extras ----------.

// GetLocalizationByID returns localization strings by localization ID.
func (c *Client) GetLocalizationByID(ctx context.Context, id int) (*LocalizationResource, error) {
	var out LocalizationResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v3/localization/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetLocalizationLanguages returns the list of available localization languages.
func (c *Client) GetLocalizationLanguages(ctx context.Context) ([]LocalizationLanguageResource, error) {
	var out []LocalizationLanguageResource
	if err := c.base.Get(ctx, "/api/v3/localization/language", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ---------- Naming Examples ----------.
