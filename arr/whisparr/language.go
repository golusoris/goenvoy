package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetLanguages returns all available languages.
func (c *Client) GetLanguages(ctx context.Context) ([]arr.LanguageResource, error) {
	var out []arr.LanguageResource
	if err := c.base.Get(ctx, "/api/v3/language", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetLanguage returns a language by ID.
func (c *Client) GetLanguage(ctx context.Context, id int) (*arr.LanguageResource, error) {
	var out arr.LanguageResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v3/language/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------- Language Profiles ----------.

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
func (c *Client) CreateLanguageProfile(ctx context.Context, lp *LanguageProfileResource) (*LanguageProfileResource, error) {
	var out LanguageProfileResource
	if err := c.base.Post(ctx, "/api/v3/languageprofile", lp, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateLanguageProfile updates an existing language profile.
func (c *Client) UpdateLanguageProfile(ctx context.Context, lp *LanguageProfileResource) (*LanguageProfileResource, error) {
	var out LanguageProfileResource
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v3/languageprofile/%d", lp.ID), lp, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteLanguageProfile deletes a language profile by ID.
func (c *Client) DeleteLanguageProfile(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/languageprofile/%d", id), nil, nil)
}

// GetLanguageProfileSchema returns the language profile schema.
func (c *Client) GetLanguageProfileSchema(ctx context.Context) (*LanguageProfileResource, error) {
	var out LanguageProfileResource
	if err := c.base.Get(ctx, "/api/v3/languageprofile/schema", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------- Localization ----------.

// GetLocalization returns localization strings.
func (c *Client) GetLocalization(ctx context.Context) (map[string]string, error) {
	var out map[string]string
	if err := c.base.Get(ctx, "/api/v3/localization", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetLocalizationLanguage returns the current localization language.
func (c *Client) GetLocalizationLanguage(ctx context.Context) (*LocalizationLanguageResource, error) {
	var out LocalizationLanguageResource
	if err := c.base.Get(ctx, "/api/v3/localization/language", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------- Logs ----------.
