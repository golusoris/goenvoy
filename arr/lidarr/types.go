package lidarr

// Artist represents an artist in Lidarr.
type Artist struct {
	ID                int               `json:"id"`
	Status            string            `json:"status"`
	Ended             bool              `json:"ended,omitempty"`
	ArtistName        string            `json:"artistName"`
	ForeignArtistID   string            `json:"foreignArtistId"`
	MBID              string            `json:"mbId,omitempty"`
	TADBID            int               `json:"tadbId,omitempty"`
	DiscogsID         int               `json:"discogsId,omitempty"`
	AllMusicID        string            `json:"allMusicId,omitempty"`
	Overview          string            `json:"overview,omitempty"`
	ArtistType        string            `json:"artistType,omitempty"`
	Disambiguation    string            `json:"disambiguation,omitempty"`
	Links             []Link            `json:"links,omitempty"`
	Images            []Image           `json:"images,omitempty"`
	Members           []Member          `json:"members,omitempty"`
	RemotePoster      string            `json:"remotePoster,omitempty"`
	Path              string            `json:"path"`
	QualityProfileID  int               `json:"qualityProfileId"`
	MetadataProfileID int               `json:"metadataProfileId"`
	Monitored         bool              `json:"monitored"`
	MonitorNewItems   string            `json:"monitorNewItems,omitempty"`
	RootFolderPath    string            `json:"rootFolderPath,omitempty"`
	Folder            string            `json:"folder,omitempty"`
	Genres            []string          `json:"genres,omitempty"`
	CleanName         string            `json:"cleanName,omitempty"`
	SortName          string            `json:"sortName,omitempty"`
	Tags              []int             `json:"tags"`
	Added             string            `json:"added"`
	AddOptions        *AddArtistOptions `json:"addOptions,omitempty"`
	Ratings           Ratings           `json:"ratings,omitempty"`
	Statistics        *ArtistStatistics `json:"statistics,omitempty"`
}

// Album represents an album in Lidarr.
type Album struct {
	ID             int              `json:"id"`
	Title          string           `json:"title"`
	Disambiguation string           `json:"disambiguation,omitempty"`
	Overview       string           `json:"overview,omitempty"`
	ArtistID       int              `json:"artistId"`
	ForeignAlbumID string           `json:"foreignAlbumId"`
	Monitored      bool             `json:"monitored"`
	AnyReleaseOk   bool             `json:"anyReleaseOk"`
	ProfileID      int              `json:"profileId"`
	Duration       int              `json:"duration"`
	AlbumType      string           `json:"albumType,omitempty"`
	SecondaryTypes []string         `json:"secondaryTypes,omitempty"`
	MediumCount    int              `json:"mediumCount,omitempty"`
	Ratings        Ratings          `json:"ratings,omitempty"`
	ReleaseDate    string           `json:"releaseDate,omitempty"`
	Releases       []AlbumRelease   `json:"releases,omitempty"`
	Genres         []string         `json:"genres,omitempty"`
	Media          []Medium         `json:"media,omitempty"`
	Artist         *Artist          `json:"artist,omitempty"`
	Images         []Image          `json:"images,omitempty"`
	Links          []Link           `json:"links,omitempty"`
	Statistics     *AlbumStatistics `json:"statistics,omitempty"`
	AddOptions     *AddAlbumOptions `json:"addOptions,omitempty"`
	RemoteCover    string           `json:"remoteCover,omitempty"`
}

// AlbumRelease represents a specific release of an album.
type AlbumRelease struct {
	ID               int      `json:"id"`
	AlbumID          int      `json:"albumId"`
	ForeignReleaseID string   `json:"foreignReleaseId"`
	Title            string   `json:"title"`
	Status           string   `json:"status,omitempty"`
	Duration         int      `json:"duration"`
	TrackCount       int      `json:"trackCount"`
	Media            []Medium `json:"media,omitempty"`
	MediumCount      int      `json:"mediumCount,omitempty"`
	Disambiguation   string   `json:"disambiguation,omitempty"`
	Country          []string `json:"country,omitempty"`
	Label            []string `json:"label,omitempty"`
	Format           string   `json:"format,omitempty"`
	Monitored        bool     `json:"monitored"`
}

// Track represents a track in Lidarr.
type Track struct {
	ID                  int     `json:"id"`
	ArtistID            int     `json:"artistId"`
	ForeignTrackID      string  `json:"foreignTrackId,omitempty"`
	ForeignRecordingID  string  `json:"foreignRecordingId,omitempty"`
	TrackFileID         int     `json:"trackFileId"`
	AlbumID             int     `json:"albumId"`
	Explicit            bool    `json:"explicit"`
	AbsoluteTrackNumber int     `json:"absoluteTrackNumber"`
	TrackNumber         string  `json:"trackNumber,omitempty"`
	Title               string  `json:"title"`
	Duration            int     `json:"duration"`
	MediumNumber        int     `json:"mediumNumber"`
	HasFile             bool    `json:"hasFile"`
	Ratings             Ratings `json:"ratings,omitempty"`
}

// TrackFile represents a downloaded track file on disk.
type TrackFile struct {
	ID                  int          `json:"id"`
	ArtistID            int          `json:"artistId"`
	AlbumID             int          `json:"albumId"`
	Path                string       `json:"path"`
	Size                int64        `json:"size"`
	DateAdded           string       `json:"dateAdded"`
	Quality             QualityModel `json:"quality"`
	QualityCutoffNotMet bool         `json:"qualityCutoffNotMet"`
}

// Medium represents a physical medium (disc) within an album release.
type Medium struct {
	MediumNumber int    `json:"mediumNumber"`
	MediumName   string `json:"mediumName,omitempty"`
	MediumFormat string `json:"mediumFormat,omitempty"`
}

// Image represents a cover image for a media item.
type Image struct {
	CoverType string `json:"coverType"`
	URL       string `json:"url"`
	RemoteURL string `json:"remoteUrl,omitempty"`
}

// Link represents a web link for an artist or album.
type Link struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

// Member represents a member of a band or group.
type Member struct {
	Name       string  `json:"name"`
	Instrument string  `json:"instrument,omitempty"`
	Images     []Image `json:"images,omitempty"`
}

// Ratings holds community rating data.
type Ratings struct {
	Votes int     `json:"votes"`
	Value float64 `json:"value"`
}

// AddArtistOptions controls behavior when adding a new artist.
type AddArtistOptions struct {
	Monitor                string   `json:"monitor"`
	AlbumsToMonitor        []string `json:"albumsToMonitor,omitempty"`
	Monitored              bool     `json:"monitored"`
	SearchForMissingAlbums bool     `json:"searchForMissingAlbums"`
}

// AddAlbumOptions controls behavior when adding a new album.
type AddAlbumOptions struct {
	AddType           string `json:"addType,omitempty"`
	SearchForNewAlbum bool   `json:"searchForNewAlbum"`
}

// AlbumStatistics contains file counts and size information for an album.
type AlbumStatistics struct {
	TrackFileCount  int     `json:"trackFileCount"`
	TrackCount      int     `json:"trackCount"`
	TotalTrackCount int     `json:"totalTrackCount"`
	SizeOnDisk      int64   `json:"sizeOnDisk"`
	PercentOfTracks float64 `json:"percentOfTracks"`
}

// ArtistStatistics contains aggregate statistics for an artist.
type ArtistStatistics struct {
	AlbumCount      int     `json:"albumCount"`
	TrackFileCount  int     `json:"trackFileCount"`
	TrackCount      int     `json:"trackCount"`
	TotalTrackCount int     `json:"totalTrackCount"`
	SizeOnDisk      int64   `json:"sizeOnDisk"`
	PercentOfTracks float64 `json:"percentOfTracks"`
}

// QualityModel pairs a quality definition with its revision.
type QualityModel struct {
	Quality  Quality  `json:"quality"`
	Revision Revision `json:"revision"`
}

// Quality identifies a quality tier.
type Quality struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Revision tracks repack and version information for a release.
type Revision struct {
	Version  int  `json:"version"`
	Real     int  `json:"real"`
	IsRepack bool `json:"isRepack"`
}

// CustomFormat describes a custom format definition.
type CustomFormat struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// HistoryRecord represents an event in the download history.
type HistoryRecord struct {
	ID                  int               `json:"id"`
	AlbumID             int               `json:"albumId"`
	ArtistID            int               `json:"artistId"`
	TrackID             int               `json:"trackId,omitempty"`
	SourceTitle         string            `json:"sourceTitle"`
	Quality             QualityModel      `json:"quality"`
	CustomFormats       []CustomFormat    `json:"customFormats,omitempty"`
	CustomFormatScore   int               `json:"customFormatScore"`
	QualityCutoffNotMet bool              `json:"qualityCutoffNotMet"`
	Date                string            `json:"date"`
	DownloadID          string            `json:"downloadId,omitempty"`
	EventType           string            `json:"eventType"`
	Data                map[string]string `json:"data,omitempty"`
}

// ParseResult contains the result of parsing a release title.
type ParseResult struct {
	ID                int              `json:"id"`
	Title             string           `json:"title"`
	ParsedAlbumInfo   *ParsedAlbumInfo `json:"parsedAlbumInfo,omitempty"`
	Artist            *Artist          `json:"artist,omitempty"`
	Albums            []Album          `json:"albums,omitempty"`
	CustomFormats     []CustomFormat   `json:"customFormats,omitempty"`
	CustomFormatScore int              `json:"customFormatScore"`
}

// ParsedAlbumInfo holds the structured data extracted from a release title.
type ParsedAlbumInfo struct {
	ReleaseTitle     string       `json:"releaseTitle,omitempty"`
	AlbumTitle       string       `json:"albumTitle,omitempty"`
	ArtistName       string       `json:"artistName,omitempty"`
	AlbumType        string       `json:"albumType,omitempty"`
	Quality          QualityModel `json:"quality"`
	ReleaseDate      string       `json:"releaseDate,omitempty"`
	Discography      bool         `json:"discography"`
	DiscographyStart int          `json:"discographyStart,omitempty"`
	DiscographyEnd   int          `json:"discographyEnd,omitempty"`
	ReleaseGroup     string       `json:"releaseGroup,omitempty"`
	ReleaseHash      string       `json:"releaseHash,omitempty"`
	ReleaseVersion   string       `json:"releaseVersion,omitempty"`
}

// SearchResult represents a result from the search endpoint.
type SearchResult struct {
	ID        int     `json:"id"`
	ForeignID string  `json:"foreignId,omitempty"`
	Artist    *Artist `json:"artist,omitempty"`
	Album     *Album  `json:"album,omitempty"`
}

// ArtistEditorResource is used for batch editing or deleting artists.
type ArtistEditorResource struct {
	ArtistIDs              []int  `json:"artistIds"`
	Monitored              *bool  `json:"monitored,omitempty"`
	MonitorNewItems        string `json:"monitorNewItems,omitempty"`
	QualityProfileID       *int   `json:"qualityProfileId,omitempty"`
	MetadataProfileID      *int   `json:"metadataProfileId,omitempty"`
	RootFolderPath         string `json:"rootFolderPath,omitempty"`
	Tags                   []int  `json:"tags,omitempty"`
	ApplyTags              string `json:"applyTags,omitempty"`
	MoveFiles              bool   `json:"moveFiles"`
	DeleteFiles            bool   `json:"deleteFiles"`
	AddImportListExclusion bool   `json:"addImportListExclusion"`
}

// AlbumsMonitoredResource is used to set the monitored status of albums.
type AlbumsMonitoredResource struct {
	AlbumIDs  []int `json:"albumIds"`
	Monitored bool  `json:"monitored"`
}

// TrackFileListResource is the request body for bulk track file operations.
type TrackFileListResource struct {
	TrackFileIDs []int `json:"trackFileIds"`
}

// ImportListExclusion represents an artist excluded from import lists.
type ImportListExclusion struct {
	ID         int    `json:"id"`
	ForeignID  string `json:"foreignId,omitempty"`
	ArtistName string `json:"artistName,omitempty"`
}

// MetadataProfile describes a metadata profile for filtering album types.
type MetadataProfile struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
