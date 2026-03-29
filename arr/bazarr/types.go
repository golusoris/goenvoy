package bazarr

// PagedResponse wraps list responses from the Bazarr API.
type PagedResponse[T any] struct {
	Data  []T `json:"data"`
	Total int `json:"total"`
}

// Series represents a TV series tracked by Bazarr.
type Series struct {
	AlternativeTitles   []string        `json:"alternativeTitles,omitempty"`
	AudioLanguage       []AudioLanguage `json:"audio_language,omitempty"`
	EpisodeFileCount    int             `json:"episodeFileCount"`
	Ended               bool            `json:"ended"`
	EpisodeMissingCount int             `json:"episodeMissingCount"`
	Fanart              string          `json:"fanart,omitempty"`
	ImdbID              string          `json:"imdbId,omitempty"`
	LastAired           string          `json:"lastAired,omitempty"`
	Monitored           bool            `json:"monitored"`
	Overview            string          `json:"overview,omitempty"`
	Path                string          `json:"path,omitempty"`
	Poster              string          `json:"poster,omitempty"`
	ProfileID           *int            `json:"profileId,omitempty"`
	SeriesType          string          `json:"seriesType,omitempty"`
	SonarrSeriesID      int             `json:"sonarrSeriesId"`
	Tags                []string        `json:"tags,omitempty"`
	Title               string          `json:"title,omitempty"`
	TvdbID              int             `json:"tvdbId"`
	Year                string          `json:"year,omitempty"`
}

// Episode represents a TV episode tracked by Bazarr.
type Episode struct {
	AudioLanguage    []AudioLanguage    `json:"audio_language,omitempty"`
	EpisodeNumber    int                `json:"episode"`
	MissingSubtitles []SubtitleLanguage `json:"missing_subtitles,omitempty"`
	Monitored        bool               `json:"monitored"`
	Path             string             `json:"path,omitempty"`
	Season           int                `json:"season"`
	SonarrEpisodeID  int                `json:"sonarrEpisodeId"`
	SonarrSeriesID   int                `json:"sonarrSeriesId"`
	Subtitles        []Subtitle         `json:"subtitles,omitempty"`
	Title            string             `json:"title,omitempty"`
	SceneName        string             `json:"sceneName,omitempty"`
}

// Movie represents a movie tracked by Bazarr.
type Movie struct {
	AlternativeTitles []string           `json:"alternativeTitles,omitempty"`
	AudioLanguage     []AudioLanguage    `json:"audio_language,omitempty"`
	Fanart            string             `json:"fanart,omitempty"`
	ImdbID            string             `json:"imdbId,omitempty"`
	MissingSubtitles  []SubtitleLanguage `json:"missing_subtitles,omitempty"`
	Monitored         bool               `json:"monitored"`
	Overview          string             `json:"overview,omitempty"`
	Path              string             `json:"path,omitempty"`
	Poster            string             `json:"poster,omitempty"`
	ProfileID         *int               `json:"profileId,omitempty"`
	RadarrID          int                `json:"radarrId"`
	SceneName         string             `json:"sceneName,omitempty"`
	Subtitles         []Subtitle         `json:"subtitles,omitempty"`
	Tags              []string           `json:"tags,omitempty"`
	Title             string             `json:"title,omitempty"`
	Year              string             `json:"year,omitempty"`
}

// WantedEpisode represents an episode missing subtitles.
type WantedEpisode struct {
	SeriesTitle      string             `json:"seriesTitle,omitempty"`
	EpisodeNumber    string             `json:"episode_number,omitempty"`
	EpisodeTitle     string             `json:"episodeTitle,omitempty"`
	MissingSubtitles []SubtitleLanguage `json:"missing_subtitles,omitempty"`
	SonarrSeriesID   int                `json:"sonarrSeriesId"`
	SonarrEpisodeID  int                `json:"sonarrEpisodeId"`
	SceneName        string             `json:"sceneName,omitempty"`
	Tags             []string           `json:"tags,omitempty"`
	SeriesType       string             `json:"seriesType,omitempty"`
}

// WantedMovie represents a movie missing subtitles.
type WantedMovie struct {
	Title            string             `json:"title,omitempty"`
	MissingSubtitles []SubtitleLanguage `json:"missing_subtitles,omitempty"`
	RadarrID         int                `json:"radarrId"`
	SceneName        string             `json:"sceneName,omitempty"`
	Tags             []string           `json:"tags,omitempty"`
}

// EpisodeHistoryRecord represents a subtitle history event for an episode.
type EpisodeHistoryRecord struct {
	SeriesTitle     string            `json:"seriesTitle,omitempty"`
	Monitored       bool              `json:"monitored"`
	EpisodeNumber   string            `json:"episode_number,omitempty"`
	EpisodeTitle    string            `json:"episodeTitle,omitempty"`
	Timestamp       string            `json:"timestamp,omitempty"`
	SubsID          string            `json:"subs_id,omitempty"`
	Description     string            `json:"description,omitempty"`
	SonarrSeriesID  int               `json:"sonarrSeriesId"`
	Language        *SubtitleLanguage `json:"language,omitempty"`
	Score           string            `json:"score,omitempty"`
	Tags            []string          `json:"tags,omitempty"`
	Action          int               `json:"action"`
	SubtitlesPath   string            `json:"subtitles_path,omitempty"`
	SonarrEpisodeID int               `json:"sonarrEpisodeId"`
	Provider        string            `json:"provider,omitempty"`
	Upgradable      bool              `json:"upgradable"`
	ParsedTimestamp string            `json:"parsed_timestamp,omitempty"`
	Blacklisted     bool              `json:"blacklisted"`
	Matches         []string          `json:"matches,omitempty"`
	DontMatches     []string          `json:"dont_matches,omitempty"`
}

// MovieHistoryRecord represents a subtitle history event for a movie.
type MovieHistoryRecord struct {
	Title           string            `json:"title,omitempty"`
	Monitored       bool              `json:"monitored"`
	Timestamp       string            `json:"timestamp,omitempty"`
	SubsID          string            `json:"subs_id,omitempty"`
	Description     string            `json:"description,omitempty"`
	RadarrID        int               `json:"radarrId"`
	Language        *SubtitleLanguage `json:"language,omitempty"`
	Score           string            `json:"score,omitempty"`
	Tags            []string          `json:"tags,omitempty"`
	Action          int               `json:"action"`
	SubtitlesPath   string            `json:"subtitles_path,omitempty"`
	Provider        string            `json:"provider,omitempty"`
	Upgradable      bool              `json:"upgradable"`
	ParsedTimestamp string            `json:"parsed_timestamp,omitempty"`
	Blacklisted     bool              `json:"blacklisted"`
	Matches         []string          `json:"matches,omitempty"`
	DontMatches     []string          `json:"dont_matches,omitempty"`
}

// Subtitle represents a subtitle file attached to a media item.
type Subtitle struct {
	Code2    string `json:"code2,omitempty"`
	Code3    string `json:"code3,omitempty"`
	Forced   bool   `json:"forced"`
	Hi       bool   `json:"hi"`
	Name     string `json:"name,omitempty"`
	Path     string `json:"path,omitempty"`
	FileSize int64  `json:"file_size"`
}

// SubtitleLanguage describes a language requirement for subtitles.
type SubtitleLanguage struct {
	Code2  string `json:"code2,omitempty"`
	Code3  string `json:"code3,omitempty"`
	Forced bool   `json:"forced"`
	Hi     bool   `json:"hi"`
	Name   string `json:"name,omitempty"`
}

// AudioLanguage describes an audio track language.
type AudioLanguage struct {
	Code2 string `json:"code2,omitempty"`
	Code3 string `json:"code3,omitempty"`
	Name  string `json:"name,omitempty"`
}

// Provider represents a subtitle provider and its status.
type Provider struct {
	Name   string `json:"name,omitempty"`
	Status string `json:"status,omitempty"`
	Retry  string `json:"retry,omitempty"`
}

// BadgeCounts contains the counts displayed in the Bazarr UI badges.
type BadgeCounts struct {
	Episodes      int    `json:"episodes"`
	Movies        int    `json:"movies"`
	Providers     int    `json:"providers"`
	Status        int    `json:"status"`
	SonarrSignalR string `json:"sonarr_signalr,omitempty"`
	RadarrSignalR string `json:"radarr_signalr,omitempty"`
	Announcements int    `json:"announcements"`
}

// SystemStatus contains Bazarr environment and version information.
type SystemStatus struct {
	BazarrVersion     string `json:"bazarr_version,omitempty"`
	PackageVersion    string `json:"package_version,omitempty"`
	SonarrVersion     string `json:"sonarr_version,omitempty"`
	RadarrVersion     string `json:"radarr_version,omitempty"`
	OperatingSystem   string `json:"operating_system,omitempty"`
	PythonVersion     string `json:"python_version,omitempty"`
	DatabaseEngine    string `json:"database_engine,omitempty"`
	DatabaseMigration string `json:"database_migration,omitempty"`
	BazarrDirectory   string `json:"bazarr_directory,omitempty"`
	BazarrConfigDir   string `json:"bazarr_config_directory,omitempty"`
	StartTime         int64  `json:"start_time"`
	Timezone          string `json:"timezone,omitempty"`
	CPUCores          int    `json:"cpu_cores"`
}

// LanguageProfile represents a language profile configured in Bazarr.
type LanguageProfile struct {
	ProfileID int    `json:"profileId"`
	Name      string `json:"name,omitempty"`
}

// Language represents a language available in Bazarr.
type Language struct {
	Code2   string `json:"code2,omitempty"`
	Code3   string `json:"code3,omitempty"`
	Name    string `json:"name,omitempty"`
	Enabled bool   `json:"enabled"`
}
