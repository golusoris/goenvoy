package stashbox

// Performer represents an adult performer.
type Performer struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Disambiguation  string         `json:"disambiguation,omitempty"`
	Gender          string         `json:"gender,omitempty"`
	Birthdate       string         `json:"birthdate,omitempty"`
	Deathdate       string         `json:"deathdate,omitempty"`
	Height          *int           `json:"height,omitempty"`
	Weight          *int           `json:"weight,omitempty"`
	HairColor       string         `json:"hair_color,omitempty"`
	EyeColor        string         `json:"eye_color,omitempty"`
	Ethnicity       string         `json:"ethnicity,omitempty"`
	BreastType      string         `json:"breast_type,omitempty"`
	Country         string         `json:"country,omitempty"`
	Measurements    string         `json:"measurements,omitempty"`
	CareerStartYear *int           `json:"career_start_year,omitempty"`
	CareerEndYear   *int           `json:"career_end_year,omitempty"`
	Aliases         []string       `json:"aliases,omitempty"`
	URLs            []URLEntry     `json:"urls,omitempty"`
	Images          []Image        `json:"images,omitempty"`
	Tattoos         []BodyLocation `json:"tattoos,omitempty"`
	Piercings       []BodyLocation `json:"piercings,omitempty"`
	IsFavorite      bool           `json:"is_favorite"`
	Deleted         bool           `json:"deleted"`
	Created         string         `json:"created"`
	Updated         string         `json:"updated"`
}

// Scene represents a filmed scene.
type Scene struct {
	ID             string                `json:"id"`
	Title          string                `json:"title,omitempty"`
	Details        string                `json:"details,omitempty"`
	ReleaseDate    string                `json:"release_date,omitempty"`
	ProductionDate string                `json:"production_date,omitempty"`
	Duration       *int                  `json:"duration,omitempty"`
	Director       string                `json:"director,omitempty"`
	Code           string                `json:"code,omitempty"`
	Studio         *Studio               `json:"studio,omitempty"`
	Performers     []PerformerAppearance `json:"performers,omitempty"`
	Tags           []Tag                 `json:"tags,omitempty"`
	Images         []Image               `json:"images,omitempty"`
	URLs           []URLEntry            `json:"urls,omitempty"`
	Fingerprints   []Fingerprint         `json:"fingerprints,omitempty"`
	Deleted        bool                  `json:"deleted"`
	Created        string                `json:"created"`
	Updated        string                `json:"updated"`
}

// Studio represents a production entity.
type Studio struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Aliases      []string   `json:"aliases,omitempty"`
	Parent       *Studio    `json:"parent,omitempty"`
	ChildStudios []Studio   `json:"child_studios,omitempty"`
	URLs         []URLEntry `json:"urls,omitempty"`
	Images       []Image    `json:"images,omitempty"`
	Deleted      bool       `json:"deleted"`
	Created      string     `json:"created"`
	Updated      string     `json:"updated"`
}

// Tag represents a category label.
type Tag struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Aliases     []string     `json:"aliases,omitempty"`
	Category    *TagCategory `json:"category,omitempty"`
	Deleted     bool         `json:"deleted"`
	Created     string       `json:"created"`
	Updated     string       `json:"updated"`
}

// TagCategory groups related tags.
type TagCategory struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// Site represents a content site known to StashBox.
type Site struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	URL         string   `json:"url,omitempty"`
	Regex       string   `json:"regex,omitempty"`
	ValidTypes  []string `json:"valid_types,omitempty"`
	Icon        string   `json:"icon,omitempty"`
}

// Image represents a hosted image.
type Image struct {
	ID           string `json:"id"`
	URL          string `json:"url"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	HasThumbnail bool   `json:"has_thumbnail,omitempty"`
}

// URLEntry is a URL linked to a site.
type URLEntry struct {
	URL  string `json:"url"`
	Site Site   `json:"site"`
}

// PerformerAppearance links a performer to a scene.
type PerformerAppearance struct {
	Performer Performer `json:"performer"`
	As        string    `json:"as,omitempty"`
}

// Fingerprint is a file hash for scene matching.
type Fingerprint struct {
	Hash        string `json:"hash"`
	Algorithm   string `json:"algorithm"`
	Duration    int    `json:"duration"`
	Submissions int    `json:"submissions"`
}

// BodyLocation describes tattoos or piercings.
type BodyLocation struct {
	Location    string `json:"location"`
	Description string `json:"description,omitempty"`
}

// QueryResult holds paginated query results.
type QueryResult[T any] struct {
	Count int `json:"count"`
	Items []T `json:"-"`
}

// QueryInput specifies pagination and sorting for list queries.
type QueryInput struct {
	Page    int    `json:"page,omitempty"`
	PerPage int    `json:"per_page,omitempty"`
	Sort    string `json:"sort,omitempty"`
	Dir     string `json:"direction,omitempty"`
}

// FingerprintInput is used for scene fingerprint lookups.
type FingerprintInput struct {
	Hash      string `json:"hash"`
	Algorithm string `json:"algorithm"`
}
