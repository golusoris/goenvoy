package sonarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetLanguages returns all available languages.
func (c *Client) GetLanguages(ctx context.Context) ([]Language, error) {
	return arr.GetList[Language](ctx, c.base, "/api/v3/language")
}

// ---------- System extras ----------.

// GetLanguage returns a single language by ID.
func (c *Client) GetLanguage(ctx context.Context, id int) (*Language, error) {
	return arr.GetOne[Language](ctx, c.base, fmt.Sprintf("/api/v3/language/%d", id))
}

// ---------- Localization ----------.

// GetLocalization returns the localization strings for the current locale.
func (c *Client) GetLocalization(ctx context.Context) (*LocalizationResource, error) {
	return arr.GetOne[LocalizationResource](ctx, c.base, "/api/v3/localization")
}

// ---------- Log File ----------.

// GetLanguageProfiles returns all language profiles.
func (c *Client) GetLanguageProfiles(ctx context.Context) ([]LanguageProfileResource, error) {
	return arr.GetList[LanguageProfileResource](ctx, c.base, "/api/v3/languageprofile")
}

// GetLanguageProfile returns a single language profile by ID.
func (c *Client) GetLanguageProfile(ctx context.Context, id int) (*LanguageProfileResource, error) {
	return arr.GetOne[LanguageProfileResource](ctx, c.base, fmt.Sprintf("/api/v3/languageprofile/%d", id))
}

// CreateLanguageProfile creates a new language profile.
func (c *Client) CreateLanguageProfile(ctx context.Context, profile *LanguageProfileResource) (*LanguageProfileResource, error) {
	return arr.PostOne[LanguageProfileResource](ctx, c.base, "/api/v3/languageprofile", profile)
}

// UpdateLanguageProfile updates an existing language profile.
func (c *Client) UpdateLanguageProfile(ctx context.Context, profile *LanguageProfileResource) (*LanguageProfileResource, error) {
	return arr.PutOne[LanguageProfileResource](ctx, c.base, fmt.Sprintf("/api/v3/languageprofile/%d", profile.ID), profile)
}

// DeleteLanguageProfile deletes a language profile by ID.
func (c *Client) DeleteLanguageProfile(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/languageprofile/%d", id), nil, nil)
}

// GetLanguageProfileSchema returns the available language profile schema.
func (c *Client) GetLanguageProfileSchema(ctx context.Context) (*LanguageProfileResource, error) {
	return arr.GetOne[LanguageProfileResource](ctx, c.base, "/api/v3/languageprofile/schema")
}

// ---------- Localization extras ----------.

// GetLocalizationByID returns localization strings by localization ID.
func (c *Client) GetLocalizationByID(ctx context.Context, id int) (*LocalizationResource, error) {
	return arr.GetOne[LocalizationResource](ctx, c.base, fmt.Sprintf("/api/v3/localization/%d", id))
}

// GetLocalizationLanguages returns the list of available localization languages.
func (c *Client) GetLocalizationLanguages(ctx context.Context) ([]LocalizationLanguageResource, error) {
	return arr.GetList[LocalizationLanguageResource](ctx, c.base, "/api/v3/localization/language")
}

// ---------- Naming Examples ----------.
