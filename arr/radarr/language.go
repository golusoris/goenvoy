package radarr

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

// GetLocalizationLanguages returns the list of available localization languages.
func (c *Client) GetLocalizationLanguages(ctx context.Context) ([]LocalizationLanguageResource, error) {
	return arr.GetList[LocalizationLanguageResource](ctx, c.base, "/api/v3/localization/language")
}

// ---------- Metadata Config ----------.
