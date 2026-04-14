package audiodb

// Artist represents a TheAudioDB artist.
type Artist struct {
	IDArtist           string `json:"idArtist"`
	StrArtist          string `json:"strArtist"`
	StrArtistAlternate string `json:"strArtistAlternate"`
	IntBornYear        string `json:"intBornYear"`
	IntFormedYear      string `json:"intFormedYear"`
	StrGenre           string `json:"strGenre"`
	StrMood            string `json:"strMood"`
	StrStyle           string `json:"strStyle"`
	StrBiographyEN     string `json:"strBiographyEN"`
	StrCountry         string `json:"strCountry"`
	StrArtistThumb     string `json:"strArtistThumb"`
	StrArtistFanart    string `json:"strArtistFanart"`
	StrArtistBanner    string `json:"strArtistBanner"`
	StrArtistLogo      string `json:"strArtistLogo"`
	StrMusicBrainzID   string `json:"strMusicBrainzID"`
}

// Album represents a TheAudioDB album.
type Album struct {
	IDAlbum          string `json:"idAlbum"`
	IDArtist         string `json:"idArtist"`
	StrAlbum         string `json:"strAlbum"`
	StrArtist        string `json:"strArtist"`
	IntYearReleased  string `json:"intYearReleased"`
	StrGenre         string `json:"strGenre"`
	StrMood          string `json:"strMood"`
	StrDescription   string `json:"strDescription"`
	StrAlbumThumb    string `json:"strAlbumThumb"`
	StrAlbumCDart    string `json:"strAlbumCDart"`
	StrMusicBrainzID string `json:"strMusicBrainzID"`
	IntScore         string `json:"intScore"`
	StrLabel         string `json:"strLabel"`
}

// Track represents a TheAudioDB track.
type Track struct {
	IDTrack          string `json:"idTrack"`
	IDAlbum          string `json:"idAlbum"`
	IDArtist         string `json:"idArtist"`
	StrTrack         string `json:"strTrack"`
	StrAlbum         string `json:"strAlbum"`
	StrArtist        string `json:"strArtist"`
	IntDuration      string `json:"intDuration"`
	IntTrackNumber   string `json:"intTrackNumber"`
	StrGenre         string `json:"strGenre"`
	StrMusicVideo    string `json:"strMusicVideo"`
	StrMusicBrainzID string `json:"strMusicBrainzID"`
}

// MusicVideo represents a TheAudioDB music video.
type MusicVideo struct {
	IDTrack          string `json:"idTrack"`
	StrTrack         string `json:"strTrack"`
	StrMusicVideo    string `json:"strMusicVideo"`
	StrDescriptionEN string `json:"strDescriptionEN"`
}

// Discography represents a discography entry.
type Discography struct {
	StrAlbum        string `json:"strAlbum"`
	IntYearReleased string `json:"intYearReleased"`
}

// Trending represents a trending entry from TheAudioDB.
type Trending struct {
	IDArtist      string `json:"idArtist"`
	StrArtist     string `json:"strArtist"`
	IDAlbum       string `json:"idAlbum"`
	StrAlbum      string `json:"strAlbum"`
	StrAlbumThumb string `json:"strAlbumThumb"`
	IntChartPlace string `json:"intChartPlace"`
}
