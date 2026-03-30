package tpdb

// Pagination contains paging metadata from the API response.
type Pagination struct {
	Total       int `json:"total"`
	PerPage     int `json:"per_page"`
	CurrentPage int `json:"current_page"`
	LastPage    int `json:"last_page"`
}

// Hash represents a content hash for scene identification.
type Hash struct {
	Algorithm string `json:"algorithm"`
	Hash      string `json:"hash"`
}

// Marker represents a scene chapter/marker.
type Marker struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Seconds int    `json:"seconds"`
}

// MediaEntry represents an image/poster entry.
type MediaEntry struct {
	ID   int    `json:"id"`
	URL  string `json:"url"`
	Type string `json:"type"`
}

// SiteRef is a minimal site reference embedded in scenes.
type SiteRef struct {
	UUID      string `json:"uuid"`
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"short_name"`
	URL       string `json:"url"`
}

// TagRef is a tag reference embedded in resources.
type TagRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// DirectorRef is a director reference.
type DirectorRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// PerformerRef is a minimal performer reference embedded in scenes.
type PerformerRef struct {
	ID     int           `json:"id"`
	PID    string        `json:"_id"`
	Name   string        `json:"name"`
	Slug   string        `json:"slug"`
	Image  string        `json:"image"`
	Extras ExtraData     `json:"extras,omitempty"`
	Parent *PerformerRef `json:"parent,omitempty"`
}

// Link represents an external URL link.
type Link struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}

// Scene represents a scene resource from the API.
type Scene struct {
	ID          int            `json:"id"`
	PID         string         `json:"_id"`
	Title       string         `json:"title"`
	Type        string         `json:"type"`
	Slug        string         `json:"slug"`
	ExternalID  string         `json:"external_id"`
	Description string         `json:"description"`
	Rating      float64        `json:"rating"`
	SiteID      int            `json:"site_id"`
	Date        string         `json:"date"`
	URL         string         `json:"url"`
	Image       string         `json:"image"`
	BackImage   string         `json:"back_image"`
	PosterImage string         `json:"poster_image"`
	Poster      string         `json:"poster"`
	Trailer     string         `json:"trailer"`
	Duration    int            `json:"duration"`
	Format      string         `json:"format"`
	SKU         string         `json:"sku"`
	Backgrounds []MediaEntry   `json:"backgrounds"`
	Posters     []MediaEntry   `json:"posters"`
	Media       []MediaEntry   `json:"media"`
	Performers  []PerformerRef `json:"performers"`
	Site        *SiteRef       `json:"site"`
	Tags        []TagRef       `json:"tags"`
	Hashes      []Hash         `json:"hashes"`
	Markers     []Marker       `json:"markers"`
	Directors   []DirectorRef  `json:"directors"`
	Links       []Link         `json:"links"`
	CreatedAt   string         `json:"created_at"`
	UpdatedAt   string         `json:"updated_at"`
}

// ExtraData holds extended performer metadata.
type ExtraData struct {
	Astrology      string `json:"astrology,omitempty"`
	Birthday       string `json:"birthday,omitempty"`
	Birthplace     string `json:"birthplace,omitempty"`
	BirthplaceCode string `json:"birthplace_code,omitempty"`
	CupSize        string `json:"cupsize,omitempty"`
	Ethnicity      string `json:"ethnicity,omitempty"`
	EyeColor       string `json:"eye_colour,omitempty"` //nolint:misspell // API field name is "eye_colour"
	FakeBoobs      string `json:"fakeboobs,omitempty"`
	Gender         string `json:"gender,omitempty"`
	HairColor      string `json:"haircolor,omitempty"`
	Height         string `json:"height,omitempty"`
	Measurements   string `json:"measurements,omitempty"`
	Nationality    string `json:"nationality,omitempty"`
	Piercings      string `json:"piercings,omitempty"`
	Tattoos        string `json:"tattoos,omitempty"`
	Weight         string `json:"weight,omitempty"`
}

// Performer represents a performer resource from the API.
type Performer struct {
	ID             int            `json:"id"`
	PID            string         `json:"_id"`
	Slug           string         `json:"slug"`
	Name           string         `json:"name"`
	FullName       string         `json:"full_name"`
	Disambiguation string         `json:"disambiguation"`
	Bio            string         `json:"bio"`
	Rating         float64        `json:"rating"`
	IsParent       bool           `json:"is_parent"`
	Extras         ExtraData      `json:"extras"`
	Aliases        []string       `json:"aliases"`
	Image          string         `json:"image"`
	Thumbnail      string         `json:"thumbnail"`
	Face           string         `json:"face"`
	Posters        []MediaEntry   `json:"posters"`
	SitePerformers []PerformerRef `json:"site_performers"`
	CreatedAt      string         `json:"created_at"`
	UpdatedAt      string         `json:"updated_at"`
}

// Site represents a site/studio resource from the API.
type Site struct {
	UUID        string   `json:"uuid"`
	ID          int      `json:"id"`
	ParentID    *int     `json:"parent_id,omitempty"`
	NetworkID   *int     `json:"network_id,omitempty"`
	Name        string   `json:"name"`
	ShortName   string   `json:"short_name"`
	URL         string   `json:"url"`
	Description string   `json:"description"`
	Rating      float64  `json:"rating"`
	Logo        string   `json:"logo"`
	Favicon     string   `json:"favicon"`
	Poster      string   `json:"poster"`
	Network     *SiteRef `json:"network,omitempty"`
	Parent      *SiteRef `json:"parent,omitempty"`
}

// Tag represents a tag resource.
type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// Director represents a director resource.
type Director struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// Movie represents a movie/DVD resource from the API.
type Movie struct {
	ID          int            `json:"id"`
	PID         string         `json:"_id"`
	Title       string         `json:"title"`
	Slug        string         `json:"slug"`
	Description string         `json:"description"`
	Date        string         `json:"date"`
	URL         string         `json:"url"`
	Image       string         `json:"image"`
	Poster      string         `json:"poster"`
	Duration    int            `json:"duration"`
	SKU         string         `json:"sku"`
	Performers  []PerformerRef `json:"performers"`
	Site        *SiteRef       `json:"site"`
	Tags        []TagRef       `json:"tags"`
	Scenes      []Scene        `json:"scenes"`
	Directors   []DirectorRef  `json:"directors"`
	CreatedAt   string         `json:"created_at"`
	UpdatedAt   string         `json:"updated_at"`
}

// Jav represents a JAV (Japanese Adult Video) resource.
type Jav struct {
	ID          int            `json:"id"`
	PID         string         `json:"_id"`
	Title       string         `json:"title"`
	Slug        string         `json:"slug"`
	Description string         `json:"description"`
	Date        string         `json:"date"`
	URL         string         `json:"url"`
	Image       string         `json:"image"`
	Poster      string         `json:"poster"`
	Duration    int            `json:"duration"`
	SKU         string         `json:"sku"`
	Performers  []PerformerRef `json:"performers"`
	Site        *SiteRef       `json:"site"`
	Tags        []TagRef       `json:"tags"`
	Directors   []DirectorRef  `json:"directors"`
	CreatedAt   string         `json:"created_at"`
	UpdatedAt   string         `json:"updated_at"`
}
