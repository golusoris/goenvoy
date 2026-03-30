package shoko

// SeriesIDs holds cross-referenced identifiers for a series.
type SeriesIDs struct {
	ID            int      `json:"ID"`
	ParentGroup   int      `json:"ParentGroup,omitempty"`
	TopLevelGroup int      `json:"TopLevelGroup,omitempty"`
	AniDB         int      `json:"AniDB,omitempty"`
	TvDB          []int    `json:"TvDB,omitempty"`
	IMDB          []string `json:"IMDB,omitempty"`
	TMDB          *TMDBIDs `json:"TMDB,omitempty"`
	MAL           []int    `json:"MAL,omitempty"`
}

// TMDBIDs contains TMDB movie and show identifiers.
type TMDBIDs struct {
	Movie []int `json:"Movie,omitempty"`
	Show  []int `json:"Show,omitempty"`
}

// Rating holds a rating value with its source.
type Rating struct {
	Value    float64 `json:"Value"`
	MaxValue float64 `json:"MaxValue"`
	Votes    int     `json:"Votes,omitempty"`
	Type     string  `json:"Type,omitempty"`
	Source   string  `json:"Source,omitempty"`
}

// Title is a localized title entry.
type Title struct {
	Name     string `json:"Name"`
	Language string `json:"Language"`
	Type     string `json:"Type"`
	Default  bool   `json:"Default"`
	Source   string `json:"Source,omitempty"`
}

// Sizes holds file count statistics for a series.
type Sizes struct {
	FileSources map[string]FileCounts `json:"FileSources,omitempty"`
	Total       FileCounts            `json:"Total,omitempty"`
	Local       FileCounts            `json:"Local,omitempty"`
	Watched     FileCounts            `json:"Watched,omitempty"`
	Missing     FileCounts            `json:"Missing,omitempty"`
}

// FileCounts holds episode counts by type.
type FileCounts struct {
	Episodes int `json:"Episodes,omitempty"`
	Specials int `json:"Specials,omitempty"`
	Credits  int `json:"Credits,omitempty"`
	Trailers int `json:"Trailers,omitempty"`
	Parodies int `json:"Parodies,omitempty"`
	Others   int `json:"Others,omitempty"`
}

// Series represents a Shoko series with cross-referenced IDs.
type Series struct {
	IDs           SeriesIDs `json:"IDs"`
	Name          string    `json:"Name"`
	HasCustomName bool      `json:"HasCustomName"`
	Description   string    `json:"Description,omitempty"`
	IsFavorite    bool      `json:"IsFavorite"`
	UserRating    *Rating   `json:"UserRating,omitempty"`
	AirsOn        []string  `json:"AirsOn,omitempty"`
	Sizes         *Sizes    `json:"Sizes,omitempty"`
	Size          int       `json:"Size"`
	Created       string    `json:"Created,omitempty"`
	Updated       string    `json:"Updated,omitempty"`
}

// AniDBAnime holds AniDB-specific anime metadata.
type AniDBAnime struct {
	ID          int     `json:"ID"`
	ShokoID     int     `json:"ShokoID,omitempty"`
	Type        string  `json:"Type"`
	Title       string  `json:"Title"`
	Titles      []Title `json:"Titles,omitempty"`
	Description string  `json:"Description,omitempty"`
	AirDate     string  `json:"AirDate,omitempty"`
	EndDate     string  `json:"EndDate,omitempty"`
	Restricted  bool    `json:"Restricted"`
	Rating      *Rating `json:"Rating,omitempty"`
}

// AniDBRelation represents a relation between two AniDB anime.
type AniDBRelation struct {
	RelatedID int         `json:"RelatedID"`
	Type      string      `json:"Type"`
	Related   *AniDBAnime `json:"Related,omitempty"`
}

// EpisodeIDs holds cross-referenced identifiers for an episode.
type EpisodeIDs struct {
	ID           int      `json:"ID"`
	ParentSeries int      `json:"ParentSeries,omitempty"`
	AniDB        int      `json:"AniDB,omitempty"`
	TvDB         []int    `json:"TvDB,omitempty"`
	IMDB         []string `json:"IMDB,omitempty"`
	TMDB         *TMDBIDs `json:"TMDB,omitempty"`
}

// Episode represents a Shoko episode.
type Episode struct {
	IDs  EpisodeIDs `json:"IDs"`
	Name string     `json:"Name,omitempty"`
	Size int        `json:"Size,omitempty"`
}

// AniDBEpisode holds AniDB-specific episode metadata.
type AniDBEpisode struct {
	ID            int     `json:"ID"`
	AnimeID       int     `json:"AnimeID"`
	Type          string  `json:"Type"`
	EpisodeNumber int     `json:"EpisodeNumber"`
	AirDate       string  `json:"AirDate,omitempty"`
	Description   string  `json:"Description,omitempty"`
	Rating        *Rating `json:"Rating,omitempty"`
	Title         string  `json:"Title,omitempty"`
	Titles        []Title `json:"Titles,omitempty"`
}

// File represents a media file managed by Shoko.
type File struct {
	ID        int            `json:"ID"`
	Size      int64          `json:"Size"`
	Hashes    *FileHashes    `json:"Hashes,omitempty"`
	Locations []FileLocation `json:"Locations,omitempty"`
	Created   string         `json:"Created,omitempty"`
	Updated   string         `json:"Updated,omitempty"`
}

// FileHashes holds hash values for a file.
type FileHashes struct {
	ED2K  string `json:"ED2K,omitempty"`
	MD5   string `json:"MD5,omitempty"`
	SHA1  string `json:"SHA1,omitempty"`
	CRC32 string `json:"CRC32,omitempty"`
}

// FileLocation represents a physical file location.
type FileLocation struct {
	ID              int    `json:"ID"`
	ManagedFolderID int    `json:"ImportFolderID"`
	RelativePath    string `json:"RelativePath"`
	IsAccessible    bool   `json:"IsAccessible"`
}

// ManagedFolder represents an import folder configuration.
type ManagedFolder struct {
	ID             int    `json:"ID"`
	Name           string `json:"Name,omitempty"`
	Path           string `json:"Path"`
	DropFolderType int    `json:"DropFolderType"`
	Size           int64  `json:"Size,omitempty"`
	FileCount      int    `json:"FileCount,omitempty"`
}

// Tag represents a tag attached to a series.
type Tag struct {
	Name        string `json:"Name"`
	Description string `json:"Description,omitempty"`
	Weight      int    `json:"Weight,omitempty"`
	Source      string `json:"Source,omitempty"`
}

// DashboardStats holds server dashboard statistics.
type DashboardStats struct {
	SeriesCount     int `json:"SeriesCount"`
	GroupCount      int `json:"GroupCount"`
	FileCount       int `json:"FileCount"`
	FinishedSeries  int `json:"FinishedSeries"`
	WatchedEpisodes int `json:"WatchedEpisodes"`
	MissingEpisodes int `json:"MissingEpisodes,omitempty"`
}

// ListResponse wraps paginated list responses.
type ListResponse[T any] struct {
	List  []T `json:"List"`
	Total int `json:"Total"`
}

// loginRequest is the JSON body for POST /api/auth.
type loginRequest struct {
	User   string `json:"user"`
	Pass   string `json:"pass"`
	Device string `json:"device"`
}

// loginResponse is the JSON body returned by POST /api/auth.
type loginResponse struct {
	APIKey string `json:"apikey"`
}
