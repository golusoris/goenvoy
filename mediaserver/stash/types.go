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
