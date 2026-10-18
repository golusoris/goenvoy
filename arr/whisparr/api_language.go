package whisparr

import (
	"context"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetLocalization returns localization strings.
func (c *api) GetLocalization(ctx context.Context) (map[string]string, error) {
	var out map[string]string
	if err := c.base.Get(ctx, "/api/v3/localization", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetLocalizationLanguage returns the current localization language.
func (c *api) GetLocalizationLanguage(ctx context.Context) (*LocalizationLanguageResource, error) {
	return arr.GetOne[LocalizationLanguageResource](ctx, c.base, "/api/v3/localization/language")
}
