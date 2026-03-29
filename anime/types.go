package anime

// AniDBInfo holds AniDB-specific metadata for a series or episode.
type AniDBInfo struct {
	// AniDBID is the AniDB anime identifier.
	AniDBID int `json:"anidbId"`
	// Type classifies the anime (TV, Movie, OVA, etc.).
	Type string `json:"type"`
	// Title is the preferred title.
	Title string `json:"title"`
	// AirDate is the original air date in YYYY-MM-DD format.
	AirDate string `json:"airDate,omitempty"`
	// EndDate is the final episode air date in YYYY-MM-DD format.
	EndDate string `json:"endDate,omitempty"`
	// EpisodeCount is the total number of episodes.
	EpisodeCount int `json:"episodeCount"`
	// Rating is the AniDB community rating (0–10).
	Rating float64 `json:"rating"`
}

// Series represents an anime series in a media management system.
type Series struct {
	// ID is the service-specific series identifier.
	ID int `json:"id"`
	// Name is the series display name.
	Name string `json:"name"`
	// AniDB contains AniDB-specific metadata when available.
	AniDB *AniDBInfo `json:"anidb,omitempty"`
	// Size is the total number of files in the series.
	Size int `json:"size"`
}

// Episode represents a single anime episode.
type Episode struct {
	// ID is the service-specific episode identifier.
	ID int `json:"id"`
	// SeriesID links this episode back to its parent series.
	SeriesID int `json:"seriesId"`
	// Name is the episode title.
	Name string `json:"name"`
	// Number is the episode number within its season or series.
	Number int `json:"number"`
	// Type classifies the episode (Normal, Special, Credit, Trailer, etc.).
	Type string `json:"type"`
	// AirDate is the original air date in YYYY-MM-DD format.
	AirDate string `json:"airDate,omitempty"`
}
