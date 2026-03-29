package metadata

// Rating represents a rating from a metadata provider.
type Rating struct {
	// Source identifies the provider (e.g. "imdb", "tmdb", "trakt").
	Source string `json:"source"`
	// Value is the rating value, typically 0–10 or 0–100 depending on provider.
	Value float64 `json:"value"`
	// Votes is the number of votes that make up the rating.
	Votes int `json:"votes"`
}

// ExternalID links a media item to an identifier on another service.
type ExternalID struct {
	// Source identifies the provider (e.g. "imdb", "tvdb", "tmdb").
	Source string `json:"source"`
	// ID is the provider-specific identifier.
	ID string `json:"id"`
}

// ImageType classifies what an image depicts.
type ImageType string

const (
	// ImageTypePoster is a vertical poster image.
	ImageTypePoster ImageType = "poster"
	// ImageTypeBackdrop is a wide background/banner image.
	ImageTypeBackdrop ImageType = "backdrop"
	// ImageTypeLogo is a transparent logo or clearart.
	ImageTypeLogo ImageType = "logo"
	// ImageTypeStill is a screenshot or episode still.
	ImageTypeStill ImageType = "still"
)

// Image represents a media image with its URL and type.
type Image struct {
	// URL is the full URL to the image resource.
	URL string `json:"url"`
	// Type classifies the image (poster, backdrop, logo, still).
	Type ImageType `json:"type"`
	// Language is the ISO 639-1 language code, empty if language-neutral.
	Language string `json:"language,omitempty"`
	// Width is the image width in pixels, zero if unknown.
	Width int `json:"width,omitempty"`
	// Height is the image height in pixels, zero if unknown.
	Height int `json:"height,omitempty"`
}

// Person represents a cast or crew member.
type Person struct {
	// Name is the person's display name.
	Name string `json:"name"`
	// Role describes the person's role (e.g. character name or crew position).
	Role string `json:"role"`
	// ImageURL is an optional portrait/headshot URL.
	ImageURL string `json:"imageUrl,omitempty"`
	// ExternalIDs links this person to provider-specific identifiers.
	ExternalIDs []ExternalID `json:"externalIds,omitempty"`
}

// MediaType identifies the kind of media item.
type MediaType string

const (
	// MediaTypeMovie is a film.
	MediaTypeMovie MediaType = "movie"
	// MediaTypeSeries is a TV series or anime show.
	MediaTypeSeries MediaType = "series"
	// MediaTypePerson is a person (actor, director, etc.).
	MediaTypePerson MediaType = "person"
)

// SearchResult is a normalised search hit returned by metadata providers.
type SearchResult struct {
	// Title is the primary title of the media item.
	Title string `json:"title"`
	// Year is the release or first-air year, zero if unknown.
	Year int `json:"year,omitempty"`
	// Type classifies the result (movie, series, person).
	Type MediaType `json:"type"`
	// ExternalIDs links this result to provider-specific identifiers.
	ExternalIDs []ExternalID `json:"externalIds,omitempty"`
	// PosterURL is an optional poster image URL.
	PosterURL string `json:"posterUrl,omitempty"`
	// Overview is a short synopsis.
	Overview string `json:"overview,omitempty"`
}
