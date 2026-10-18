package lidarr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetLanguages returns all available languages.
func (c *Client) GetLanguages(ctx context.Context) ([]arr.LanguageResource, error) {
	return arr.GetList[arr.LanguageResource](ctx, c.base, "/api/v1/language")
}

// GetLanguage returns a single language by ID.
func (c *Client) GetLanguage(ctx context.Context, id int) (*arr.LanguageResource, error) {
	return arr.GetOne[arr.LanguageResource](ctx, c.base, fmt.Sprintf("/api/v1/language/%d", id))
}

// GetLocalization returns the localization strings for the current locale.
func (c *Client) GetLocalization(ctx context.Context) (*LocalizationResource, error) {
	return arr.GetOne[LocalizationResource](ctx, c.base, "/api/v1/localization")
}
