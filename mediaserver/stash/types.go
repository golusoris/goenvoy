package stash

// Scene represents a media scene.
type Scene struct {
	ID           string        `json:"id"`
	Title        string        `json:"title,omitempty"`
	Code         string        `json:"code,omitempty"`
	Details      string        `json:"details,omitempty"`
	Director     string        `json:"director,omitempty"`
	Date         string        `json:"date,omitempty"`
	Rating100    *int          `json:"rating100,omitempty"`
	OCounter     int           `json:"o_counter"`
	Organized    bool          `json:"organized"`
	PlayCount    int           `json:"play_count"`
	PlayDuration float64       `json:"play_duration"`
	ResumeTime   float64       `json:"resume_time"`
	URLs         []string      `json:"urls,omitempty"`
	Studio       *Studio       `json:"studio,omitempty"`
	Performers   []Performer   `json:"performers,omitempty"`
	Tags         []Tag         `json:"tags,omitempty"`
	Galleries    []Gallery     `json:"galleries,omitempty"`
	Groups       []SceneGroup  `json:"groups,omitempty"`
	Markers      []SceneMarker `json:"scene_markers,omitempty"`
	Files        []VideoFile   `json:"files,omitempty"`
	StashIDs     []RemoteID    `json:"stash_ids,omitempty"`
	CreatedAt    string        `json:"created_at"`
	UpdatedAt    string        `json:"updated_at"`
}

// VideoFile holds file metadata for a scene.
type VideoFile struct {
	Path         string            `json:"path"`
	Size         int64             `json:"size"`
	Duration     float64           `json:"duration"`
	VideoCodec   string            `json:"video_codec,omitempty"`
	AudioCodec   string            `json:"audio_codec,omitempty"`
	Width        int               `json:"width"`
	Height       int               `json:"height"`
	FrameRate    float64           `json:"frame_rate"`
	BitRate      int64             `json:"bit_rate"`
	Format       string            `json:"format,omitempty"`
	Fingerprints []FileFingerprint `json:"fingerprints,omitempty"`
}

// FileFingerprint holds a file hash.
type FileFingerprint struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// Performer represents a performer in the local library.
type Performer struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Disambiguation string     `json:"disambiguation,omitempty"`
	Gender         string     `json:"gender,omitempty"`
	Birthdate      string     `json:"birthdate,omitempty"`
	DeathDate      string     `json:"death_date,omitempty"`
	Country        string     `json:"country,omitempty"`
	Ethnicity      string     `json:"ethnicity,omitempty"`
	HairColor      string     `json:"hair_color,omitempty"`
	EyeColor       string     `json:"eye_color,omitempty"`
	Height         *int       `json:"height_cm,omitempty"`
	Weight         *int       `json:"weight,omitempty"`
	Measurements   string     `json:"measurements,omitempty"`
	FakeTits       string     `json:"fake_tits,omitempty"`
	Tattoos        string     `json:"tattoos,omitempty"`
	Piercings      string     `json:"piercings,omitempty"`
	CareerLength   string     `json:"career_length,omitempty"`
	URLs           []string   `json:"urls,omitempty"`
	Details        string     `json:"details,omitempty"`
	Aliases        []string   `json:"aliases,omitempty"`
	ImagePath      string     `json:"image_path,omitempty"`
	Tags           []Tag      `json:"tags,omitempty"`
	SceneCount     int        `json:"scene_count"`
	ImageCount     int        `json:"image_count"`
	GalleryCount   int        `json:"gallery_count"`
	Rating100      *int       `json:"rating100,omitempty"`
	Favorite       bool       `json:"favorite"`
	StashIDs       []RemoteID `json:"stash_ids,omitempty"`
	CreatedAt      string     `json:"created_at"`
	UpdatedAt      string     `json:"updated_at"`
}

// Studio represents a production studio.
type Studio struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	URL          string     `json:"url,omitempty"`
	Details      string     `json:"details,omitempty"`
	ImagePath    string     `json:"image_path,omitempty"`
	ParentStudio *Studio    `json:"parent_studio,omitempty"`
	ChildStudios []Studio   `json:"child_studios,omitempty"`
	Aliases      []string   `json:"aliases,omitempty"`
	SceneCount   int        `json:"scene_count"`
	StashIDs     []RemoteID `json:"stash_ids,omitempty"`
	Rating100    *int       `json:"rating100,omitempty"`
	CreatedAt    string     `json:"created_at"`
	UpdatedAt    string     `json:"updated_at"`
}

// Tag represents a category tag.
type Tag struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description,omitempty"`
	ImagePath      string   `json:"image_path,omitempty"`
	Aliases        []string `json:"aliases,omitempty"`
	SceneCount     int      `json:"scene_count"`
	PerformerCount int      `json:"performer_count"`
	GalleryCount   int      `json:"gallery_count"`
	ImageCount     int      `json:"image_count"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
}

// Gallery represents an image gallery.
type Gallery struct {
	ID         string      `json:"id"`
	Title      string      `json:"title,omitempty"`
	Details    string      `json:"details,omitempty"`
	Date       string      `json:"date,omitempty"`
	Rating100  *int        `json:"rating100,omitempty"`
	Organized  bool        `json:"organized"`
	URLs       []string    `json:"urls,omitempty"`
	Studio     *Studio     `json:"studio,omitempty"`
	Performers []Performer `json:"performers,omitempty"`
	Tags       []Tag       `json:"tags,omitempty"`
	ImageCount int         `json:"image_count"`
	CreatedAt  string      `json:"created_at"`
	UpdatedAt  string      `json:"updated_at"`
}

// Image represents a single image.
type Image struct {
	ID         string      `json:"id"`
	Title      string      `json:"title,omitempty"`
	Date       string      `json:"date,omitempty"`
	Rating100  *int        `json:"rating100,omitempty"`
	OCounter   int         `json:"o_counter"`
	Organized  bool        `json:"organized"`
	Studio     *Studio     `json:"studio,omitempty"`
	Performers []Performer `json:"performers,omitempty"`
	Tags       []Tag       `json:"tags,omitempty"`
	Galleries  []Gallery   `json:"galleries,omitempty"`
	Files      []ImageFile `json:"visual_files,omitempty"`
	CreatedAt  string      `json:"created_at"`
	UpdatedAt  string      `json:"updated_at"`
}

// ImageFile holds image file metadata.
type ImageFile struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// Group represents a group (formerly movie).
type Group struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Aliases   string   `json:"aliases,omitempty"`
	Duration  *int     `json:"duration,omitempty"`
	Date      string   `json:"date,omitempty"`
	Rating100 *int     `json:"rating100,omitempty"`
	Director  string   `json:"director,omitempty"`
	Synopsis  string   `json:"synopsis,omitempty"`
	Studio    *Studio  `json:"studio,omitempty"`
	URLs      []string `json:"urls,omitempty"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

// SceneGroup links a scene to a group.
type SceneGroup struct {
	Group      Group `json:"group"`
	SceneIndex *int  `json:"scene_index,omitempty"`
}

// SceneMarker is a chapter point within a scene.
type SceneMarker struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	Seconds    float64 `json:"seconds"`
	PrimaryTag Tag     `json:"primary_tag"`
	Tags       []Tag   `json:"tags,omitempty"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

// RemoteID links a local entity to a remote StashBox entry.
type RemoteID struct {
	Endpoint string `json:"endpoint"`
	StashID  string `json:"stash_id"`
}

// FindFilter specifies pagination and sorting.
type FindFilter struct {
	Query     string `json:"q,omitempty"`
	Page      int    `json:"page,omitempty"`
	PerPage   int    `json:"per_page,omitempty"`
	Sort      string `json:"sort,omitempty"`
	Direction string `json:"direction,omitempty"`
}

// FindResult holds paginated query results.
type FindResult[T any] struct {
	Count int `json:"count"`
	Items []T `json:"-"`
}

// Stats holds system statistics.
type Stats struct {
	SceneCount     int     `json:"scene_count"`
	ScenesSize     float64 `json:"scenes_size"`
	ScenesDuration float64 `json:"scenes_duration"`
	ImageCount     int     `json:"image_count"`
	ImagesSize     float64 `json:"images_size"`
	GalleryCount   int     `json:"gallery_count"`
	PerformerCount int     `json:"performer_count"`
	StudioCount    int     `json:"studio_count"`
	TagCount       int     `json:"tag_count"`
	TotalOCount    int     `json:"total_o_count"`
	TotalPlayCount int     `json:"total_play_count"`
}

// Version holds version information.
type Version struct {
	Version   string `json:"version"`
	Hash      string `json:"hash"`
	BuildType string `json:"build_type"`
}

// SystemStatus holds system status info.
type SystemStatus struct {
	DatabaseSchema *int   `json:"databaseSchema"`
	DatabasePath   string `json:"databasePath"`
	AppSchema      int    `json:"appSchema"`
	Status         string `json:"status"`
	OS             string `json:"os"`
}

// GenderEnum represents a performer gender.
// Stash v0.31.1 graphql/schema/types/performer.graphql:1.
type GenderEnum string

const (
	// GenderMale indicates male gender.
	// Stash v0.31.1 graphql/schema/types/performer.graphql:2.
	GenderMale GenderEnum = "MALE"
	// GenderFemale indicates female gender.
	// Stash v0.31.1 graphql/schema/types/performer.graphql:3.
	GenderFemale GenderEnum = "FEMALE"
	// GenderTransgenderMale indicates transgender male gender.
	// Stash v0.31.1 graphql/schema/types/performer.graphql:4.
	GenderTransgenderMale GenderEnum = "TRANSGENDER_MALE"
	// GenderTransgenderFemale indicates transgender female gender.
	// Stash v0.31.1 graphql/schema/types/performer.graphql:5.
	GenderTransgenderFemale GenderEnum = "TRANSGENDER_FEMALE"
	// GenderIntersex indicates intersex gender.
	// Stash v0.31.1 graphql/schema/types/performer.graphql:6.
	GenderIntersex GenderEnum = "INTERSEX"
	// GenderNonBinary indicates non-binary gender.
	// Stash v0.31.1 graphql/schema/types/performer.graphql:7.
	GenderNonBinary GenderEnum = "NON_BINARY"
)

// CircumcisedEnum represents performer circumcision status.
// Stash v0.31.1 graphql/schema/types/performer.graphql:10.
type CircumcisedEnum string

const (
	// CircumcisedCut indicates circumcised status.
	// Stash v0.31.1 graphql/schema/types/performer.graphql:11.
	CircumcisedCut CircumcisedEnum = "CUT"
	// CircumcisedUncut indicates uncircumcised status.
	// Stash v0.31.1 graphql/schema/types/performer.graphql:12.
	CircumcisedUncut CircumcisedEnum = "UNCUT"
)

// StashIDInput specifies an external StashBox ID mapping input.
// Stash v0.31.1 graphql/schema/types/stash-box.graphql:22.
type StashIDInput struct { //nolint:revive // StashIDInput mirrors upstream GraphQL type name verbatim.
	Endpoint  string  `json:"endpoint"`
	StashID   string  `json:"stash_id"`
	UpdatedAt *string `json:"updated_at,omitempty"`
}

// CustomFieldsInput specifies mutation inputs for custom fields.
// Stash v0.31.1 graphql/schema/types/metadata.graphql:369.
type CustomFieldsInput struct {
	Full    map[string]any `json:"full,omitempty"`
	Partial map[string]any `json:"partial,omitempty"`
	Remove  []string       `json:"remove,omitzero"`
}

// SceneMovieInput specifies a movie assignment input for a scene.
// Stash v0.31.1 graphql/schema/types/scene.graphql:88.
type SceneMovieInput struct {
	MovieID    string `json:"movie_id"`
	SceneIndex *int   `json:"scene_index,omitempty"`
}

// SceneGroupInput specifies a group assignment input for a scene.
// Stash v0.31.1 graphql/schema/types/scene.graphql:93.
type SceneGroupInput struct {
	GroupID    string `json:"group_id"`
	SceneIndex *int   `json:"scene_index,omitempty"`
}

// PerformerCreateInput defines input fields for creating a performer.
// Stash v0.31.1 graphql/schema/types/performer.graphql:67.
type PerformerCreateInput struct {
	Name           string  `json:"name"`
	Disambiguation *string `json:"disambiguation,omitempty"`
	// Deprecated: use URLs.
	URL          *string          `json:"url,omitempty"`
	URLs         []string         `json:"urls,omitzero"`
	Gender       *GenderEnum      `json:"gender,omitempty"`
	Birthdate    *string          `json:"birthdate,omitempty"`
	Ethnicity    *string          `json:"ethnicity,omitempty"`
	Country      *string          `json:"country,omitempty"`
	EyeColor     *string          `json:"eye_color,omitempty"`
	HeightCM     *int             `json:"height_cm,omitempty"`
	Measurements *string          `json:"measurements,omitempty"`
	FakeTits     *string          `json:"fake_tits,omitempty"`
	PenisLength  *float64         `json:"penis_length,omitempty"`
	Circumcised  *CircumcisedEnum `json:"circumcised,omitempty"`
	// Deprecated: use CareerStart and CareerEnd.
	CareerLength *string  `json:"career_length,omitempty"`
	CareerStart  *string  `json:"career_start,omitempty"`
	CareerEnd    *string  `json:"career_end,omitempty"`
	Tattoos      *string  `json:"tattoos,omitempty"`
	Piercings    *string  `json:"piercings,omitempty"`
	AliasList    []string `json:"alias_list,omitzero"`
	// Deprecated: use URLs.
	Twitter *string `json:"twitter,omitempty"`
	// Deprecated: use URLs.
	Instagram     *string        `json:"instagram,omitempty"`
	Favorite      *bool          `json:"favorite,omitempty"`
	TagIDs        []string       `json:"tag_ids,omitzero"`
	Image         *string        `json:"image,omitempty"`
	StashIDs      []StashIDInput `json:"stash_ids,omitzero"`
	Rating100     *int           `json:"rating100,omitempty"`
	Details       *string        `json:"details,omitempty"`
	DeathDate     *string        `json:"death_date,omitempty"`
	HairColor     *string        `json:"hair_color,omitempty"`
	Weight        *int           `json:"weight,omitempty"`
	IgnoreAutoTag *bool          `json:"ignore_auto_tag,omitempty"`
	CustomFields  map[string]any `json:"custom_fields,omitempty"`
}

// PerformerUpdateInput defines input fields for updating a performer.
// Stash v0.31.1 graphql/schema/types/performer.graphql:107.
// A nil slice leaves the field unchanged; an empty non-nil slice clears it.
type PerformerUpdateInput struct {
	ID             string  `json:"id"`
	Name           *string `json:"name,omitempty"`
	Disambiguation *string `json:"disambiguation,omitempty"`
	// Deprecated: use URLs.
	URL          *string          `json:"url,omitempty"`
	URLs         []string         `json:"urls,omitzero"`
	Gender       *GenderEnum      `json:"gender,omitempty"`
	Birthdate    *string          `json:"birthdate,omitempty"`
	Ethnicity    *string          `json:"ethnicity,omitempty"`
	Country      *string          `json:"country,omitempty"`
	EyeColor     *string          `json:"eye_color,omitempty"`
	HeightCM     *int             `json:"height_cm,omitempty"`
	Measurements *string          `json:"measurements,omitempty"`
	FakeTits     *string          `json:"fake_tits,omitempty"`
	PenisLength  *float64         `json:"penis_length,omitempty"`
	Circumcised  *CircumcisedEnum `json:"circumcised,omitempty"`
	// Deprecated: use CareerStart and CareerEnd.
	CareerLength *string  `json:"career_length,omitempty"`
	CareerStart  *string  `json:"career_start,omitempty"`
	CareerEnd    *string  `json:"career_end,omitempty"`
	Tattoos      *string  `json:"tattoos,omitempty"`
	Piercings    *string  `json:"piercings,omitempty"`
	AliasList    []string `json:"alias_list,omitzero"`
	// Deprecated: use URLs.
	Twitter *string `json:"twitter,omitempty"`
	// Deprecated: use URLs.
	Instagram     *string            `json:"instagram,omitempty"`
	Favorite      *bool              `json:"favorite,omitempty"`
	TagIDs        []string           `json:"tag_ids,omitzero"`
	Image         *string            `json:"image,omitempty"`
	StashIDs      []StashIDInput     `json:"stash_ids,omitzero"`
	Rating100     *int               `json:"rating100,omitempty"`
	Details       *string            `json:"details,omitempty"`
	DeathDate     *string            `json:"death_date,omitempty"`
	HairColor     *string            `json:"hair_color,omitempty"`
	Weight        *int               `json:"weight,omitempty"`
	IgnoreAutoTag *bool              `json:"ignore_auto_tag,omitempty"`
	CustomFields  *CustomFieldsInput `json:"custom_fields,omitempty"`
}

// SceneUpdateInput defines input fields for updating a scene.
// Stash v0.31.1 graphql/schema/types/scene.graphql:129.
// A nil slice leaves the field unchanged; an empty non-nil slice clears it.
type SceneUpdateInput struct {
	ClientMutationID *string `json:"clientMutationId,omitempty"`
	ID               string  `json:"id"`
	Title            *string `json:"title,omitempty"`
	Code             *string `json:"code,omitempty"`
	Details          *string `json:"details,omitempty"`
	Director         *string `json:"director,omitempty"`
	// Deprecated: use URLs.
	URL       *string  `json:"url,omitempty"`
	URLs      []string `json:"urls,omitzero"`
	Date      *string  `json:"date,omitempty"`
	Rating100 *int     `json:"rating100,omitempty"`
	// Deprecated: unsupported upstream; Stash has sceneIncrementO/sceneDecrementO.
	OCounter     *int              `json:"o_counter,omitempty"`
	Organized    *bool             `json:"organized,omitempty"`
	StudioID     *string           `json:"studio_id,omitempty"`
	GalleryIDs   []string          `json:"gallery_ids,omitzero"`
	PerformerIDs []string          `json:"performer_ids,omitzero"`
	Groups       []SceneGroupInput `json:"groups,omitzero"`
	// Deprecated: use Groups.
	Movies       []SceneMovieInput `json:"movies,omitzero"`
	TagIDs       []string          `json:"tag_ids,omitzero"`
	CoverImage   *string           `json:"cover_image,omitempty"`
	StashIDs     []StashIDInput    `json:"stash_ids,omitzero"`
	ResumeTime   *float64          `json:"resume_time,omitempty"`
	PlayDuration *float64          `json:"play_duration,omitempty"`
	// Deprecated: unsupported upstream; Stash has sceneIncrementPlayCount/sceneDecrementPlayCount.
	PlayCount     *int               `json:"play_count,omitempty"`
	PrimaryFileID *string            `json:"primary_file_id,omitempty"`
	CustomFields  *CustomFieldsInput `json:"custom_fields,omitempty"`
}
