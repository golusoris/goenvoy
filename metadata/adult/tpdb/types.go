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

// MediaSizeMap represents image URLs keyed by size.
type MediaSizeMap struct {
	Full   string `json:"full,omitempty"`
	Large  string `json:"large,omitempty"`
	Medium string `json:"medium,omitempty"`
	Small  string `json:"small,omitempty"`
}

// MediaEntry represents an image/poster entry.
type MediaEntry struct {
	ID    int    `json:"id"`
	URL   string `json:"url"`
	Type  string `json:"type"`
	Size  int    `json:"size,omitempty"`
	Order int    `json:"order,omitempty"`
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
	UUID string `json:"uuid"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// DirectorRef is a director reference.
type DirectorRef struct {
	ID   int    `json:"id"`
	UUID string `json:"uuid"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// PerformerRef is a minimal performer reference embedded in scenes.
type PerformerRef struct {
	// ID is the numeric ThePornDB id (live key `_id`).
	ID int `json:"_id"`
	// PID is the UUID (live key `id`).
	PID    string        `json:"id"`
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
	// ID is the numeric ThePornDB id (live key `_id`).
	ID int `json:"_id"`
	// PID is the UUID (live key `id`).
	PID         string  `json:"id"`
	Title       string  `json:"title"`
	Type        string  `json:"type"`
	Slug        string  `json:"slug"`
	ExternalID  string  `json:"external_id"`
	Description string  `json:"description"`
	Rating      float64 `json:"rating"`
	SiteID      int     `json:"site_id"`
	Date        string  `json:"date"`
	URL         string  `json:"url"`
	Image       string  `json:"image"`
	BackImage   string  `json:"back_image"`
	PosterImage string  `json:"poster_image"`
	Poster      string  `json:"poster"`
	Trailer     string  `json:"trailer"`
	Duration    int     `json:"duration"`
	Format      string  `json:"format"`
	SKU         string  `json:"sku"`
	// Deprecated: no longer populated; the live key is "background", a size map. Use Background.
	Backgrounds []MediaEntry `json:"-"`
	// Deprecated: no longer populated; the live API returns a size map. Use PosterSizes.
	Posters     []MediaEntry   `json:"-"`
	PosterSizes MediaSizeMap   `json:"posters"`
	Background  MediaSizeMap   `json:"background"`
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
	// ID is the numeric ThePornDB id (live key `_id`).
	ID int `json:"_id"`
	// PID is the UUID (live key `id`).
	PID            string         `json:"id"`
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
	UUID string `json:"uuid"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// Director represents a director resource.
type Director struct {
	ID   int    `json:"id"`
	UUID string `json:"uuid"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// Movie represents a movie/DVD resource from the API.
type Movie struct {
	// ID is the numeric ThePornDB id (live key `_id`).
	ID int `json:"_id"`
	// PID is the UUID (live key `id`).
	PID         string         `json:"id"`
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
	// ID is the numeric ThePornDB id (live key `_id`).
	ID int `json:"_id"`
	// PID is the UUID (live key `id`).
	PID         string         `json:"id"`
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

// ChangeRecord represents an entity change record.
type ChangeRecord struct {
	ID   int    `json:"id"`
	UUID string `json:"uuid"`
}

// ChangeEntity specifies the type of entity for change tracking.
type ChangeEntity string

const (
	// ChangeScenes tracks scene changes.
	ChangeScenes ChangeEntity = "scenes"
	// ChangeMovies tracks movie changes.
	ChangeMovies ChangeEntity = "movies"
	// ChangeJav tracks JAV changes.
	ChangeJav ChangeEntity = "jav"
	// ChangePerformers tracks performer changes.
	ChangePerformers ChangeEntity = "performers"
)

// FavouriteType specifies the type of entity for favorites.
type FavouriteType string

const (
	// FavouritePerformer indicates a performer favorite.
	FavouritePerformer FavouriteType = "performer"
	// FavouriteSite indicates a site favorite.
	FavouriteSite FavouriteType = "site"
)

// PerformerSite represents a performer-site relationship from the API.
type PerformerSite struct {
	ID        string     `json:"id"`
	NumericID int        `json:"_id"`
	SiteID    int        `json:"site_id"`
	Name      string     `json:"name"`
	Bio       string     `json:"bio"`
	IsParent  bool       `json:"is_parent"`
	Extras    ExtraData  `json:"extra"`
	Image     string     `json:"image"`
	Thumbnail string     `json:"thumbnail"`
	Face      string     `json:"face"`
	Scenes    []Scene    `json:"scenes,omitempty"`
	Parent    *Performer `json:"parent,omitempty"`
	Site      *Site      `json:"site,omitempty"`
}
