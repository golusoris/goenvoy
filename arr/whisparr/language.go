package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetLanguages returns all available languages.
func (c *Client) GetLanguages(ctx context.Context) ([]arr.LanguageResource, error) {
	return arr.GetList[arr.LanguageResource](ctx, c.base, "/api/v3/language")
}

// GetLanguage returns a language by ID.
func (c *Client) GetLanguage(ctx context.Context, id int) (*arr.LanguageResource, error) {
	return arr.GetOne[arr.LanguageResource](ctx, c.base, fmt.Sprintf("/api/v3/language/%d", id))
}

// GetLanguageProfiles returns all language profiles.
func (c *Client) GetLanguageProfiles(ctx context.Context) ([]LanguageProfileResource, error) {
	return arr.GetList[LanguageProfileResource](ctx, c.base, "/api/v3/languageprofile")
}

// GetLanguageProfile returns a single language profile by ID.
func (c *Client) GetLanguageProfile(ctx context.Context, id int) (*LanguageProfileResource, error) {
	return arr.GetOne[LanguageProfileResource](ctx, c.base, fmt.Sprintf("/api/v3/languageprofile/%d", id))
}

// CreateLanguageProfile creates a new language profile.
func (c *Client) CreateLanguageProfile(ctx context.Context, lp *LanguageProfileResource) (*LanguageProfileResource, error) {
	return arr.PostOne[LanguageProfileResource](ctx, c.base, "/api/v3/languageprofile", lp)
}

// UpdateLanguageProfile updates an existing language profile.
func (c *Client) UpdateLanguageProfile(ctx context.Context, lp *LanguageProfileResource) (*LanguageProfileResource, error) {
	return arr.PutOne[LanguageProfileResource](ctx, c.base, fmt.Sprintf("/api/v3/languageprofile/%d", lp.ID), lp)
}

// DeleteLanguageProfile deletes a language profile by ID.
func (c *Client) DeleteLanguageProfile(ctx context.Context, id int) error {
	return c.base.Delete(ctx, fmt.Sprintf("/api/v3/languageprofile/%d", id), nil, nil)
}

// GetLanguageProfileSchema returns the language profile schema.
func (c *Client) GetLanguageProfileSchema(ctx context.Context) (*LanguageProfileResource, error) {
	return arr.GetOne[LanguageProfileResource](ctx, c.base, "/api/v3/languageprofile/schema")
}
