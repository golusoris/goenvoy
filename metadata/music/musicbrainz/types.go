package musicbrainz

// Artist represents a MusicBrainz artist entity.
type Artist struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	SortName       string         `json:"sort-name"`
	Type           string         `json:"type,omitempty"`
	TypeID         string         `json:"type-id,omitempty"`
	Gender         string         `json:"gender,omitempty"`
	GenderID       string         `json:"gender-id,omitempty"`
	Country        string         `json:"country,omitempty"`
	Area           *Area          `json:"area,omitempty"`
	BeginArea      *Area          `json:"begin-area,omitempty"`
	EndArea        *Area          `json:"end-area,omitempty"`
	Disambiguation string         `json:"disambiguation,omitempty"`
	LifeSpan       *LifeSpan      `json:"life-span,omitempty"`
	Aliases        []Alias        `json:"aliases,omitempty"`
	Tags           []Tag          `json:"tags,omitempty"`
	Genres         []Genre        `json:"genres,omitempty"`
	Releases       []Release      `json:"releases,omitempty"`
	ReleaseGroups  []ReleaseGroup `json:"release-groups,omitempty"`
	Recordings     []Recording    `json:"recordings,omitempty"`
	Works          []Work         `json:"works,omitempty"`
	Relations      []Relation     `json:"relations,omitempty"`
}

// Release represents a MusicBrainz release (album, single, etc.).
type Release struct {
	ID                 string              `json:"id"`
	Title              string              `json:"title"`
	Status             string              `json:"status,omitempty"`
	StatusID           string              `json:"status-id,omitempty"`
	Quality            string              `json:"quality,omitempty"`
	Date               string              `json:"date,omitempty"`
	Country            string              `json:"country,omitempty"`
	Barcode            string              `json:"barcode,omitempty"`
	Disambiguation     string              `json:"disambiguation,omitempty"`
	Packaging          string              `json:"packaging,omitempty"`
	PackagingID        string              `json:"packaging-id,omitempty"`
	ArtistCredit       []ArtistCredit      `json:"artist-credit,omitempty"`
	ReleaseGroup       *ReleaseGroup       `json:"release-group,omitempty"`
	Media              []Medium            `json:"media,omitempty"`
	LabelInfo          []LabelInfo         `json:"label-info,omitempty"`
	TextRepresentation *TextRepresentation `json:"text-representation,omitempty"`
	CoverArtArchive    *CoverArtArchive    `json:"cover-art-archive,omitempty"`
	Tags               []Tag               `json:"tags,omitempty"`
	Genres             []Genre             `json:"genres,omitempty"`
	Relations          []Relation          `json:"relations,omitempty"`
}

// ReleaseGroup represents a MusicBrainz release group (album, single, EP, etc.).
type ReleaseGroup struct {
	ID               string         `json:"id"`
	Title            string         `json:"title"`
	PrimaryType      string         `json:"primary-type,omitempty"`
	PrimaryTypeID    string         `json:"primary-type-id,omitempty"`
	SecondaryTypes   []string       `json:"secondary-types,omitempty"`
	SecondaryTypeIDs []string       `json:"secondary-type-ids,omitempty"`
	FirstReleaseDate string         `json:"first-release-date,omitempty"`
	Disambiguation   string         `json:"disambiguation,omitempty"`
	ArtistCredit     []ArtistCredit `json:"artist-credit,omitempty"`
	Releases         []Release      `json:"releases,omitempty"`
	Tags             []Tag          `json:"tags,omitempty"`
	Genres           []Genre        `json:"genres,omitempty"`
	Relations        []Relation     `json:"relations,omitempty"`
}

// Recording represents a MusicBrainz recording (track).
type Recording struct {
	ID               string         `json:"id"`
	Title            string         `json:"title"`
	Length           int            `json:"length,omitempty"`
	Disambiguation   string         `json:"disambiguation,omitempty"`
	Video            bool           `json:"video,omitempty"`
	FirstReleaseDate string         `json:"first-release-date,omitempty"`
	ArtistCredit     []ArtistCredit `json:"artist-credit,omitempty"`
	Releases         []Release      `json:"releases,omitempty"`
	ISRCs            []string       `json:"isrcs,omitempty"`
	Tags             []Tag          `json:"tags,omitempty"`
	Genres           []Genre        `json:"genres,omitempty"`
	Relations        []Relation     `json:"relations,omitempty"`
}

// Label represents a MusicBrainz record label.
type Label struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	SortName       string     `json:"sort-name,omitempty"`
	Type           string     `json:"type,omitempty"`
	TypeID         string     `json:"type-id,omitempty"`
	LabelCode      int        `json:"label-code,omitempty"`
	Country        string     `json:"country,omitempty"`
	Area           *Area      `json:"area,omitempty"`
	Disambiguation string     `json:"disambiguation,omitempty"`
	LifeSpan       *LifeSpan  `json:"life-span,omitempty"`
	Aliases        []Alias    `json:"aliases,omitempty"`
	Releases       []Release  `json:"releases,omitempty"`
	Tags           []Tag      `json:"tags,omitempty"`
	Genres         []Genre    `json:"genres,omitempty"`
	Relations      []Relation `json:"relations,omitempty"`
}

// Work represents a MusicBrainz work (composition).
type Work struct {
	ID             string     `json:"id"`
	Title          string     `json:"title"`
	Type           string     `json:"type,omitempty"`
	TypeID         string     `json:"type-id,omitempty"`
	Language       string     `json:"language,omitempty"`
	Disambiguation string     `json:"disambiguation,omitempty"`
	ISWCs          []string   `json:"iswcs,omitempty"`
	Aliases        []Alias    `json:"aliases,omitempty"`
	Relations      []Relation `json:"relations,omitempty"`
	Tags           []Tag      `json:"tags,omitempty"`
	Genres         []Genre    `json:"genres,omitempty"`
}

// Area represents a geographic area.
type Area struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	SortName       string    `json:"sort-name,omitempty"`
	Type           string    `json:"type,omitempty"`
	TypeID         string    `json:"type-id,omitempty"`
	ISO31661Codes  []string  `json:"iso-3166-1-codes,omitempty"`
	ISO31662Codes  []string  `json:"iso-3166-2-codes,omitempty"`
	Disambiguation string    `json:"disambiguation,omitempty"`
	LifeSpan       *LifeSpan `json:"life-span,omitempty"`
}

// Event represents a MusicBrainz event.
type Event struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Type           string     `json:"type,omitempty"`
	TypeID         string     `json:"type-id,omitempty"`
	Canceled       bool       `json:"cancelled,omitempty"` //nolint:misspell // API field name is "cancelled"       `json:"cancelled,omitempty"`
	Disambiguation string     `json:"disambiguation,omitempty"`
	LifeSpan       *LifeSpan  `json:"life-span,omitempty"`
	Time           string     `json:"time,omitempty"`
	Setlist        string     `json:"setlist,omitempty"`
	Relations      []Relation `json:"relations,omitempty"`
	Tags           []Tag      `json:"tags,omitempty"`
	Genres         []Genre    `json:"genres,omitempty"`
}

// Genre represents a MusicBrainz genre.
type Genre struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Count          int    `json:"count,omitempty"`
	Disambiguation string `json:"disambiguation,omitempty"`
}

// Instrument represents a musical instrument.
type Instrument struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Type           string `json:"type,omitempty"`
	TypeID         string `json:"type-id,omitempty"`
	Description    string `json:"description,omitempty"`
	Disambiguation string `json:"disambiguation,omitempty"`
}

// Place represents a physical place (venue, studio, etc.).
type Place struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	Type           string       `json:"type,omitempty"`
	TypeID         string       `json:"type-id,omitempty"`
	Address        string       `json:"address,omitempty"`
	Area           *Area        `json:"area,omitempty"`
	Coordinates    *Coordinates `json:"coordinates,omitempty"`
	Disambiguation string       `json:"disambiguation,omitempty"`
	LifeSpan       *LifeSpan    `json:"life-span,omitempty"`
}

// Coordinates represents geographic coordinates.
type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Series represents a MusicBrainz series.
type Series struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Type           string     `json:"type,omitempty"`
	TypeID         string     `json:"type-id,omitempty"`
	Disambiguation string     `json:"disambiguation,omitempty"`
	Relations      []Relation `json:"relations,omitempty"`
	Tags           []Tag      `json:"tags,omitempty"`
	Genres         []Genre    `json:"genres,omitempty"`
}

// LifeSpan holds begin/end date and an ended flag.
type LifeSpan struct {
	Begin string `json:"begin,omitempty"`
	End   string `json:"end,omitempty"`
	Ended bool   `json:"ended,omitempty"`
}

// Alias represents an alternative name.
type Alias struct {
	Name     string `json:"name"`
	SortName string `json:"sort-name,omitempty"`
	Type     string `json:"type,omitempty"`
	TypeID   string `json:"type-id,omitempty"`
	Locale   string `json:"locale,omitempty"`
	Primary  bool   `json:"primary,omitempty"`
}

// Tag represents a folksonomy tag.
type Tag struct {
	Name  string `json:"name"`
	Count int    `json:"count,omitempty"`
}

// ArtistCredit represents an artist credit on a release or recording.
type ArtistCredit struct {
	Name       string `json:"name,omitempty"`
	JoinPhrase string `json:"joinphrase,omitempty"`
	Artist     Artist `json:"artist"`
}

// Medium represents a physical medium (disc, side, etc.).
type Medium struct {
	Position   int     `json:"position"`
	Format     string  `json:"format,omitempty"`
	FormatID   string  `json:"format-id,omitempty"`
	Title      string  `json:"title,omitempty"`
	TrackCount int     `json:"track-count"`
	Tracks     []Track `json:"tracks,omitempty"`
}

// Track represents a track on a medium.
type Track struct {
	ID        string    `json:"id"`
	Number    string    `json:"number"`
	Title     string    `json:"title"`
	Length    int       `json:"length,omitempty"`
	Position  int       `json:"position"`
	Recording Recording `json:"recording,omitempty"`
}

// LabelInfo associates a label with a catalog number.
type LabelInfo struct {
	CatalogNumber string `json:"catalog-number,omitempty"`
	Label         *Label `json:"label,omitempty"`
}

// TextRepresentation holds language and script info for a release.
type TextRepresentation struct {
	Language string `json:"language,omitempty"`
	Script   string `json:"script,omitempty"`
}

// CoverArtArchive indicates cover art availability.
type CoverArtArchive struct {
	Artwork  bool `json:"artwork"`
	Front    bool `json:"front"`
	Back     bool `json:"back"`
	Count    int  `json:"count"`
	Darkened bool `json:"darkened"`
}

// Relation represents a relationship between entities.
type Relation struct {
	Type       string     `json:"type"`
	TypeID     string     `json:"type-id,omitempty"`
	TargetType string     `json:"target-type,omitempty"`
	Direction  string     `json:"direction,omitempty"`
	Attributes []string   `json:"attributes,omitempty"`
	Begin      string     `json:"begin,omitempty"`
	End        string     `json:"end,omitempty"`
	Ended      bool       `json:"ended,omitempty"`
	Artist     *Artist    `json:"artist,omitempty"`
	Release    *Release   `json:"release,omitempty"`
	Recording  *Recording `json:"recording,omitempty"`
	Label      *Label     `json:"label,omitempty"`
	Work       *Work      `json:"work,omitempty"`
	URL        *URLEntity `json:"url,omitempty"`
}

// URLEntity represents a MusicBrainz URL entity.
type URLEntity struct {
	ID        string     `json:"id"`
	Resource  string     `json:"resource"`
	Relations []Relation `json:"relations,omitempty"`
}

// BrowseResult is returned by browse endpoints and includes paging info.
type BrowseResult[T any] struct {
	Entities []T `json:"-"`
	Count    int `json:"-"`
	Offset   int `json:"-"`
}
