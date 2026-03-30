package plex

// MediaContainer is the top-level wrapper for all Plex API responses.
type MediaContainer struct {
	Size              int    `json:"size"`
	TotalSize         int    `json:"totalSize,omitempty"`
	Offset            int    `json:"offset,omitempty"`
	FriendlyName      string `json:"friendlyName,omitempty"`
	MachineIdentifier string `json:"machineIdentifier,omitempty"`
	Version           string `json:"version,omitempty"`
	Platform          string `json:"platform,omitempty"`
	PlatformVersion   string `json:"platformVersion,omitempty"`

	Directory []Directory `json:"Directory,omitempty"`
	Metadata  []Metadata  `json:"Metadata,omitempty"`
	Video     []Metadata  `json:"Video,omitempty"`
}

// Directory represents a library section or navigation entry.
type Directory struct {
	Key        string `json:"key"`
	Title      string `json:"title"`
	Type       string `json:"type,omitempty"`
	Agent      string `json:"agent,omitempty"`
	Scanner    string `json:"scanner,omitempty"`
	Language   string `json:"language,omitempty"`
	Thumb      string `json:"thumb,omitempty"`
	Art        string `json:"art,omitempty"`
	Composite  string `json:"composite,omitempty"`
	Refreshing bool   `json:"refreshing,omitempty"`
	CreatedAt  int64  `json:"createdAt,omitempty"`
	UpdatedAt  int64  `json:"updatedAt,omitempty"`
}

// Metadata represents a media item (movie, episode, show, album, track, etc.).
type Metadata struct {
	RatingKey             string  `json:"ratingKey"`
	Key                   string  `json:"key"`
	GUID                  string  `json:"guid,omitempty"`
	Title                 string  `json:"title"`
	Type                  string  `json:"type,omitempty"`
	Summary               string  `json:"summary,omitempty"`
	Year                  int     `json:"year,omitempty"`
	Duration              int64   `json:"duration,omitempty"`
	Rating                float64 `json:"rating,omitempty"`
	AudienceRating        float64 `json:"audienceRating,omitempty"`
	ContentRating         string  `json:"contentRating,omitempty"`
	OriginallyAvailableAt string  `json:"originallyAvailableAt,omitempty"`
	Thumb                 string  `json:"thumb,omitempty"`
	Art                   string  `json:"art,omitempty"`
	ViewCount             int     `json:"viewCount,omitempty"`
	ViewOffset            int64   `json:"viewOffset,omitempty"`
	AddedAt               int64   `json:"addedAt,omitempty"`
	UpdatedAt             int64   `json:"updatedAt,omitempty"`

	ParentRatingKey string `json:"parentRatingKey,omitempty"`
	ParentTitle     string `json:"parentTitle,omitempty"`
	ParentIndex     int    `json:"parentIndex,omitempty"`
	Index           int    `json:"index,omitempty"`

	Media  []Media `json:"Media,omitempty"`
	Genre  []Tag   `json:"Genre,omitempty"`
	Role   []Role  `json:"Role,omitempty"`
	Writer []Tag   `json:"Writer,omitempty"`

	User             *SessionUser   `json:"User,omitempty"`
	Player           *Player        `json:"Player,omitempty"`
	Session          *SessionInfo   `json:"Session,omitempty"`
	TranscodeSession *TranscodeInfo `json:"TranscodeSession,omitempty"`
}

// Media holds codec and quality information for a media item.
type Media struct {
	ID              int    `json:"id"`
	VideoCodec      string `json:"videoCodec,omitempty"`
	AudioCodec      string `json:"audioCodec,omitempty"`
	VideoResolution string `json:"videoResolution,omitempty"`
	Bitrate         int    `json:"bitrate,omitempty"`
	Width           int    `json:"width,omitempty"`
	Height          int    `json:"height,omitempty"`
	Container       string `json:"container,omitempty"`
	Duration        int64  `json:"duration,omitempty"`
	Part            []Part `json:"Part,omitempty"`
}

// Part represents a media file part.
type Part struct {
	ID       int      `json:"id"`
	Key      string   `json:"key"`
	File     string   `json:"file,omitempty"`
	Size     int64    `json:"size,omitempty"`
	Duration int64    `json:"duration,omitempty"`
	Stream   []Stream `json:"Stream,omitempty"`
}

// Stream represents a single audio, video, or subtitle stream.
type Stream struct {
	ID          int    `json:"id"`
	StreamType  int    `json:"streamType"`
	Codec       string `json:"codec,omitempty"`
	Language    string `json:"language,omitempty"`
	LanguageTag string `json:"languageTag,omitempty"`
	Channels    int    `json:"channels,omitempty"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
	Bitrate     int    `json:"bitrate,omitempty"`
}

// Tag is a generic tag (genre, director, writer, etc.).
type Tag struct {
	ID     int    `json:"id,omitempty"`
	Tag    string `json:"tag"`
	Filter string `json:"filter,omitempty"`
}

// Role represents a cast member.
type Role struct {
	ID    int    `json:"id,omitempty"`
	Tag   string `json:"tag"`
	Role  string `json:"role,omitempty"`
	Thumb string `json:"thumb,omitempty"`
}

// SessionUser holds user info for a playback session.
type SessionUser struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Thumb string `json:"thumb,omitempty"`
}

// Player holds client player info for a session.
type Player struct {
	Title             string `json:"title"`
	Platform          string `json:"platform,omitempty"`
	Device            string `json:"device,omitempty"`
	State             string `json:"state"`
	Local             bool   `json:"local,omitempty"`
	MachineIdentifier string `json:"machineIdentifier,omitempty"`
	Address           string `json:"address,omitempty"`
}

// SessionInfo holds session metadata.
type SessionInfo struct {
	ID        string `json:"id"`
	Bandwidth int    `json:"bandwidth,omitempty"`
	Location  string `json:"location,omitempty"`
}

// TranscodeInfo holds transcode session info.
type TranscodeInfo struct {
	Key           string  `json:"key,omitempty"`
	Progress      float64 `json:"progress,omitempty"`
	Speed         float64 `json:"speed,omitempty"`
	VideoDecision string  `json:"videoDecision,omitempty"`
	AudioDecision string  `json:"audioDecision,omitempty"`
}

// Identity is the response from the /identity endpoint.
type Identity struct {
	MachineIdentifier string `json:"machineIdentifier"`
	Version           string `json:"version"`
}
