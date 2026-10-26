package audiobookshelf

// Library represents an Audiobookshelf library.
type Library struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Folders    []Folder        `json:"folders,omitempty"`
	Icon       string          `json:"icon,omitempty"`
	MediaType  string          `json:"mediaType"`
	Provider   string          `json:"provider,omitempty"`
	CreatedAt  int64           `json:"createdAt,omitempty"`
	LastUpdate int64           `json:"lastUpdate,omitempty"`
	Settings   LibrarySettings `json:"settings,omitempty"`
}

// LibrarySettings represents library-level settings.
type LibrarySettings struct {
	CoverAspectRatio          int    `json:"coverAspectRatio,omitempty"`
	DisableWatcher            bool   `json:"disableWatcher,omitempty"`
	SkipMatchingMediaWithASIN bool   `json:"skipMatchingMediaWithAsin,omitempty"`
	SkipMatchingMediaWithISBN bool   `json:"skipMatchingMediaWithIsbn,omitempty"`
	AutoScanCronExpression    string `json:"autoScanCronExpression,omitempty"`
}

// Folder represents a folder in a library.
type Folder struct {
	ID       string `json:"id"`
	FullPath string `json:"fullPath"`
}

// LibraryItem represents an item (audiobook or podcast) in a library.
type LibraryItem struct {
	ID          string `json:"id"`
	INO         string `json:"ino,omitempty"`
	LibraryID   string `json:"libraryId"`
	FolderID    string `json:"folderId,omitempty"`
	Path        string `json:"path,omitempty"`
	RelPath     string `json:"relPath,omitempty"`
	IsFile      bool   `json:"isFile,omitempty"`
	MTIMEMS     int64  `json:"mtimeMs,omitempty"`
	CTIMEMS     int64  `json:"ctimeMs,omitempty"`
	BirthtimeMS int64  `json:"birthtimeMs,omitempty"`
	AddedAt     int64  `json:"addedAt,omitempty"`
	UpdatedAt   int64  `json:"updatedAt,omitempty"`
	LastScan    int64  `json:"lastScan,omitempty"`
	ScanVersion string `json:"scanVersion,omitempty"`
	IsMissing   bool   `json:"isMissing,omitempty"`
	IsInvalid   bool   `json:"isInvalid,omitempty"`
	MediaType   string `json:"mediaType,omitempty"`
	Media       Media  `json:"media,omitempty"`
	NumFiles    int    `json:"numFiles,omitempty"`
	Size        int64  `json:"size,omitempty"`
}

// Media represents the media content of a library item.
type Media struct {
	ID                       string           `json:"id,omitempty"`
	LibraryItemID            string           `json:"libraryItemId,omitempty"`
	Metadata                 Metadata         `json:"metadata,omitempty"`
	CoverPath                string           `json:"coverPath,omitempty"`
	Tags                     []string         `json:"tags,omitempty"`
	AudioFiles               []AudioFile      `json:"audioFiles,omitempty"`
	Chapters                 []Chapter        `json:"chapters,omitempty"`
	Duration                 float64          `json:"duration,omitempty"`
	Size                     int64            `json:"size,omitempty"`
	NumTracks                int              `json:"numTracks,omitempty"`
	NumChapters              int              `json:"numChapters,omitempty"`
	Episodes                 []PodcastEpisode `json:"episodes,omitempty"`
	NumEpisodes              int              `json:"numEpisodes,omitempty"`
	AutoDownloadEpisodes     bool             `json:"autoDownloadEpisodes,omitempty"`
	AutoDownloadSchedule     string           `json:"autoDownloadSchedule,omitempty"`
	LastEpisodeCheck         int64            `json:"lastEpisodeCheck,omitempty"`
	MaxEpisodesToKeep        int              `json:"maxEpisodesToKeep,omitempty"`
	MaxNewEpisodesToDownload int              `json:"maxNewEpisodesToDownload,omitempty"`
}

// Metadata represents metadata for a media item (book or podcast).
type Metadata struct {
	Title             string   `json:"title,omitempty"`
	TitleIgnorePrefix string   `json:"titleIgnorePrefix,omitempty"`
	Subtitle          string   `json:"subtitle,omitempty"`
	Author            string   `json:"author,omitempty"`
	Authors           []Author `json:"authors,omitempty"`
	Narrators         []string `json:"narrators,omitempty"`
	Series            []Series `json:"series,omitempty"`
	Genres            []string `json:"genres,omitempty"`
	PublishedYear     string   `json:"publishedYear,omitempty"`
	ReleaseDate       string   `json:"releaseDate,omitempty"`
	Publisher         string   `json:"publisher,omitempty"`
	Description       string   `json:"description,omitempty"`
	ISBN              string   `json:"isbn,omitempty"`
	ASIN              string   `json:"asin,omitempty"`
	Language          string   `json:"language,omitempty"`
	Explicit          bool     `json:"explicit,omitempty"`
	FeedURL           string   `json:"feedUrl,omitempty"`
	ImageURL          string   `json:"imageUrl,omitempty"`
	ITunesPageURL     string   `json:"itunesPageUrl,omitempty"`
	ITunesID          string   `json:"itunesId,omitempty"`
	ITunesArtistID    string   `json:"itunesArtistId,omitempty"`
	Type              string   `json:"type,omitempty"`
}

// Author represents an audiobook author.
type Author struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Series represents a book series.
type Series struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Sequence string `json:"sequence,omitempty"`
}

// AudioFile represents an audio file in a library item.
type AudioFile struct {
	Index    int     `json:"index"`
	INO      string  `json:"ino,omitempty"`
	Metadata any     `json:"metadata,omitempty"`
	AddedAt  int64   `json:"addedAt,omitempty"`
	Duration float64 `json:"duration,omitempty"`
	MimeType string  `json:"mimeType,omitempty"`
}

// Chapter represents a chapter in an audiobook or podcast episode.
type Chapter struct {
	ID    int     `json:"id"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Title string  `json:"title"`
}

// UserPermissions represents the permission flags on an Audiobookshelf [User]
// (server/models/User.js:169-183, 603-638).
type UserPermissions struct {
	Download                  bool `json:"download"`
	Update                    bool `json:"update"`
	Delete                    bool `json:"delete"`
	Upload                    bool `json:"upload"`
	CreateEreader             bool `json:"createEreader"`
	AccessAllLibraries        bool `json:"accessAllLibraries"`
	AccessAllTags             bool `json:"accessAllTags"`
	AccessExplicitContent     bool `json:"accessExplicitContent"`
	SelectedTagsNotAccessible bool `json:"selectedTagsNotAccessible"`
}

// AudioBookmark represents a user bookmark on a library item
// (server/models/User.js:59-64, 623).
type AudioBookmark struct {
	LibraryItemID string  `json:"libraryItemId"`
	Title         string  `json:"title"`
	Time          float64 `json:"time"`
	CreatedAt     int64   `json:"createdAt"`
}

// User represents an Audiobookshelf user
// (server/models/User.js:603-638, server/Auth.js:296-312).
type User struct {
	ID                              string          `json:"id"`
	Username                        string          `json:"username"`
	Email                           string          `json:"email,omitempty"`
	Type                            string          `json:"type"`
	Token                           string          `json:"token,omitempty"`
	AccessToken                     string          `json:"accessToken,omitempty"`
	RefreshToken                    string          `json:"refreshToken,omitempty"`
	IsOldToken                      bool            `json:"isOldToken,omitempty"`
	MediaProgress                   []MediaProgress `json:"mediaProgress,omitempty"`
	SeriesHideFromContinueListening []string        `json:"seriesHideFromContinueListening,omitempty"`
	Bookmarks                       []AudioBookmark `json:"bookmarks,omitempty"`
	IsActive                        bool            `json:"isActive"`
	IsLocked                        bool            `json:"isLocked"`
	LastSeen                        int64           `json:"lastSeen,omitempty"`
	CreatedAt                       int64           `json:"createdAt,omitempty"`
	Permissions                     UserPermissions `json:"permissions,omitempty"`
	LibrariesAccessible             []string        `json:"librariesAccessible,omitempty"`
	ItemTagsSelected                []string        `json:"itemTagsSelected,omitempty"`
	HasOpenIDLink                   bool            `json:"hasOpenIDLink,omitempty"`
}

// PlaybackSession represents an active listening session.
type PlaybackSession struct {
	ID            string  `json:"id"`
	UserID        string  `json:"userId"`
	LibraryItemID string  `json:"libraryItemId"`
	EpisodeID     string  `json:"episodeId,omitempty"`
	MediaType     string  `json:"mediaType"`
	DisplayTitle  string  `json:"displayTitle,omitempty"`
	DisplayAuthor string  `json:"displayAuthor,omitempty"`
	Duration      float64 `json:"duration"`
	PlayMethod    int     `json:"playMethod"`
	StartTime     float64 `json:"startTime"`
	CurrentTime   float64 `json:"currentTime"`
	StartedAt     int64   `json:"startedAt,omitempty"`
	UpdatedAt     int64   `json:"updatedAt,omitempty"`
}

// ServerInfo represents legacy Audiobookshelf server information.
type ServerInfo struct {
	IsInit        bool   `json:"isInit"`
	Language      string `json:"language,omitempty"`
	ConfigPath    string `json:"ConfigPath,omitempty"`
	MetadataPath  string `json:"MetadataPath,omitempty"`
	Source        string `json:"Source,omitempty"`
	Version       string `json:"version"`
	LatestVersion string `json:"latestVersion,omitempty"`
	HasUpdate     bool   `json:"hasUpdate"`
}

// AuthFormData represents login form settings included in [ServerStatus]
// (server/objects/settings/ServerSettings.js:348-357).
type AuthFormData struct {
	AuthLoginCustomMessage string `json:"authLoginCustomMessage,omitempty"`
	AuthOpenIDButtonText   string `json:"authOpenIDButtonText,omitempty"`
	AuthOpenIDAutoLaunch   bool   `json:"authOpenIDAutoLaunch,omitempty"`
}

// ServerStatus represents the Audiobookshelf GET /status response
// (server/Server.js:384-400).
type ServerStatus struct {
	App           string        `json:"app"`
	ServerVersion string        `json:"serverVersion"`
	IsInit        bool          `json:"isInit"`
	Language      string        `json:"language,omitempty"`
	AuthMethods   []string      `json:"authMethods,omitempty"`
	AuthFormData  *AuthFormData `json:"authFormData,omitempty"`
	ConfigPath    string        `json:"ConfigPath,omitempty"`
	MetadataPath  string        `json:"MetadataPath,omitempty"`
}

// LibraryItemsOptions configures query parameters for [Client.GetLibraryItemsWithOptions]
// (server/controllers/LibraryController.js:604-624, 1487-1495).
type LibraryItemsOptions struct {
	Limit          int
	Page           int
	Sort           string
	Desc           bool
	Filter         string
	Minified       bool
	CollapseSeries bool
	Include        string
}

// LibraryItemsResponse wraps a paginated library items response
// (server/controllers/LibraryController.js:610-638).
type LibraryItemsResponse struct {
	Results        []LibraryItem `json:"results"`
	Total          int           `json:"total"`
	Limit          int           `json:"limit"`
	Page           int           `json:"page"`
	SortBy         string        `json:"sortBy,omitempty"`
	SortDesc       bool          `json:"sortDesc,omitempty"`
	FilterBy       string        `json:"filterBy,omitempty"`
	MediaType      string        `json:"mediaType,omitempty"`
	Minified       bool          `json:"minified,omitempty"`
	CollapseSeries bool          `json:"collapseseries,omitempty"`
	Include        string        `json:"include,omitempty"`
	Offset         int           `json:"offset,omitempty"`
}

// Collection represents a collection of library items.
type Collection struct {
	ID          string        `json:"id"`
	LibraryID   string        `json:"libraryId"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Books       []LibraryItem `json:"books,omitempty"`
	CreatedAt   int64         `json:"createdAt,omitempty"`
	UpdatedAt   int64         `json:"updatedAt,omitempty"`
}

// MediaProgress represents user progress for a media item
// (server/models/MediaProgress.js:155-176, server/controllers/UserController.js:86-98).
type MediaProgress struct {
	ID               string  `json:"id"`
	UserID           string  `json:"userId,omitempty"`
	LibraryItemID    string  `json:"libraryItemId"`
	EpisodeID        string  `json:"episodeId,omitempty"`
	MediaItemID      string  `json:"mediaItemId,omitempty"`
	MediaItemType    string  `json:"mediaItemType,omitempty"`
	Duration         float64 `json:"duration"`
	Progress         float64 `json:"progress"`
	CurrentTime      float64 `json:"currentTime"`
	IsFinished       bool    `json:"isFinished"`
	HideFromContinue bool    `json:"hideFromContinueListening"`
	EbookLocation    string  `json:"ebookLocation,omitempty"`
	EbookProgress    float64 `json:"ebookProgress,omitempty"`
	LastUpdate       int64   `json:"lastUpdate,omitempty"`
	StartedAt        int64   `json:"startedAt,omitempty"`
	FinishedAt       int64   `json:"finishedAt,omitempty"`
	DisplayTitle     string  `json:"displayTitle,omitempty"`
	DisplaySubtitle  string  `json:"displaySubtitle,omitempty"`
	CoverPath        string  `json:"coverPath,omitempty"`
}

// LoginResponse represents the response from POST /login
// (server/Auth.js:96-105, 296-312, 322-328).
type LoginResponse struct {
	User                 User           `json:"user"`
	UserDefaultLibraryID string         `json:"userDefaultLibraryId,omitempty"`
	ServerSettings       map[string]any `json:"serverSettings,omitempty"`
	EReaderDevices       []any          `json:"ereaderDevices,omitempty"`
	Source               string         `json:"Source,omitempty"`
}

// LogoutResponse represents the response from POST /logout
// (server/Auth.js:486-524).
type LogoutResponse struct {
	RedirectURL string `json:"redirect_url,omitempty"`
}

// PodcastEnclosure represents a podcast episode enclosure
// (server/utils/podcastUtils.js:31, server/models/PodcastEpisode.js:198-205).
type PodcastEnclosure struct {
	URL    string `json:"url"`
	Type   string `json:"type,omitempty"`
	Length string `json:"length,omitempty"`
}

// PodcastEpisode represents a stored podcast episode in a library item
// (server/models/PodcastEpisode.js:193-238).
type PodcastEpisode struct {
	ID            string            `json:"id"`
	LibraryItemID string            `json:"libraryItemId,omitempty"`
	PodcastID     string            `json:"podcastId,omitempty"`
	OldEpisodeID  string            `json:"oldEpisodeId,omitempty"`
	Index         int               `json:"index,omitempty"`
	Season        string            `json:"season,omitempty"`
	Episode       string            `json:"episode,omitempty"`
	EpisodeType   string            `json:"episodeType,omitempty"`
	Title         string            `json:"title,omitempty"`
	Subtitle      string            `json:"subtitle,omitempty"`
	Description   string            `json:"description,omitempty"`
	Enclosure     *PodcastEnclosure `json:"enclosure,omitempty"`
	GUID          string            `json:"guid,omitempty"`
	PubDate       string            `json:"pubDate,omitempty"`
	Chapters      []Chapter         `json:"chapters,omitempty"`
	AudioFile     *AudioFile        `json:"audioFile,omitempty"`
	PublishedAt   int64             `json:"publishedAt,omitempty"`
	AddedAt       int64             `json:"addedAt,omitempty"`
	UpdatedAt     int64             `json:"updatedAt,omitempty"`
	Size          int64             `json:"size,omitempty"`
	Duration      float64           `json:"duration,omitempty"`
}

// PodcastFeedMetadata represents metadata parsed from a podcast RSS feed
// (server/utils/podcastUtils.js:38-52, 121-151).
type PodcastFeedMetadata struct {
	Title            string   `json:"title,omitempty"`
	Language         string   `json:"language,omitempty"`
	Explicit         string   `json:"explicit,omitempty"`
	Author           string   `json:"author,omitempty"`
	PubDate          string   `json:"pubDate,omitempty"`
	Link             string   `json:"link,omitempty"`
	Image            string   `json:"image,omitempty"`
	Categories       []string `json:"categories,omitempty"`
	FeedURL          string   `json:"feedUrl,omitempty"`
	Description      string   `json:"description,omitempty"`
	DescriptionPlain string   `json:"descriptionPlain,omitempty"`
	Type             string   `json:"type,omitempty"`
}

// PodcastFeedEpisode represents an episode parsed from a podcast RSS feed
// (server/utils/podcastUtils.js:16-36, 267-291).
type PodcastFeedEpisode struct {
	Title            string            `json:"title"`
	Subtitle         string            `json:"subtitle,omitempty"`
	Description      string            `json:"description,omitempty"`
	DescriptionPlain string            `json:"descriptionPlain,omitempty"`
	PubDate          string            `json:"pubDate,omitempty"`
	EpisodeType      string            `json:"episodeType,omitempty"`
	Season           string            `json:"season,omitempty"`
	Episode          string            `json:"episode,omitempty"`
	Author           string            `json:"author,omitempty"`
	Duration         string            `json:"duration,omitempty"`
	DurationSeconds  float64           `json:"durationSeconds,omitempty"`
	Explicit         string            `json:"explicit,omitempty"`
	PublishedAt      int64             `json:"publishedAt,omitempty"`
	Enclosure        *PodcastEnclosure `json:"enclosure,omitempty"`
	GUID             string            `json:"guid,omitempty"`
	ChaptersURL      string            `json:"chaptersUrl,omitempty"`
	ChaptersType     string            `json:"chaptersType,omitempty"`
	Chapters         []Chapter         `json:"chapters,omitempty"`
}

// PodcastFeed represents a parsed RSS podcast feed returned by POST /api/podcasts/feed
// (server/controllers/PodcastController.js:191-216, server/utils/podcastUtils.js:54-59).
type PodcastFeed struct {
	Metadata    PodcastFeedMetadata  `json:"metadata"`
	Episodes    []PodcastFeedEpisode `json:"episodes,omitempty"`
	NumEpisodes int                  `json:"numEpisodes,omitempty"`
}

// PodcastFeedRequest is the request body for POST /api/podcasts/feed
// (server/controllers/PodcastController.js:194-216).
type PodcastFeedRequest struct {
	RSSFeed string `json:"rssFeed"`
}

// PodcastFeedResponse is the response envelope from POST /api/podcasts/feed
// (server/controllers/PodcastController.js:215).
type PodcastFeedResponse struct {
	Podcast PodcastFeed `json:"podcast"`
}

// PodcastMetadata represents metadata in a [CreatePodcastRequest]
// (server/models/Podcast.js:79-121, 402-418).
type PodcastMetadata struct {
	Title          string   `json:"title"`
	Author         string   `json:"author,omitempty"`
	Description    string   `json:"description,omitempty"`
	ReleaseDate    string   `json:"releaseDate,omitempty"`
	Genres         []string `json:"genres,omitempty"`
	FeedURL        string   `json:"feedUrl,omitempty"`
	ImageURL       string   `json:"imageUrl,omitempty"`
	ITunesPageURL  string   `json:"itunesPageUrl,omitempty"`
	ITunesID       string   `json:"itunesId,omitempty"`
	ITunesArtistID string   `json:"itunesArtistId,omitempty"`
	Language       string   `json:"language,omitempty"`
	Type           string   `json:"type,omitempty"`
	Explicit       bool     `json:"explicit,omitempty"`
}

// CreatePodcastMedia represents the media payload inside [CreatePodcastRequest]
// (server/controllers/PodcastController.js:46-54, server/models/Podcast.js:79-121).
type CreatePodcastMedia struct {
	Metadata             PodcastMetadata `json:"metadata"`
	AutoDownloadEpisodes bool            `json:"autoDownloadEpisodes,omitempty"`
	AutoDownloadSchedule string          `json:"autoDownloadSchedule,omitempty"`
	Tags                 []string        `json:"tags,omitempty"`
}

// CreatePodcastRequest represents the request body for POST /api/podcasts
// (server/controllers/PodcastController.js:31-189).
type CreatePodcastRequest struct {
	LibraryID string             `json:"libraryId"`
	FolderID  string             `json:"folderId"`
	Path      string             `json:"path"`
	Media     CreatePodcastMedia `json:"media"`
}

// CheckNewEpisodesResponse is the response envelope from GET /api/podcasts/:id/checknew
// (server/controllers/PodcastController.js:298-300).
type CheckNewEpisodesResponse struct {
	Episodes []PodcastFeedEpisode `json:"episodes"`
}

// PodcastSearchResult represents an iTunes podcast search result from GET /api/search/podcast
// (server/controllers/SearchController.js:135-156, server/providers/iTunes.js:14-28, 139-155).
type PodcastSearchResult struct {
	ID               int64    `json:"id"`
	ArtistID         int64    `json:"artistId,omitempty"`
	Title            string   `json:"title"`
	ArtistName       string   `json:"artistName,omitempty"`
	Description      string   `json:"description,omitempty"`
	DescriptionPlain string   `json:"descriptionPlain,omitempty"`
	ReleaseDate      string   `json:"releaseDate,omitempty"`
	Genres           []string `json:"genres,omitempty"`
	Cover            string   `json:"cover,omitempty"`
	TrackCount       int      `json:"trackCount,omitempty"`
	FeedURL          string   `json:"feedUrl,omitempty"`
	PageURL          string   `json:"pageUrl,omitempty"`
	Explicit         bool     `json:"explicit,omitempty"`
}
