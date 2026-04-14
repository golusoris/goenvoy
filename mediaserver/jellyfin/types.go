package jellyfin

// AuthenticationResult is returned after successful authentication.
type AuthenticationResult struct {
	User        *UserDto        `json:"User"`
	SessionInfo *SessionInfoDto `json:"SessionInfo"`
	AccessToken string          `json:"AccessToken"`
	ServerID    string          `json:"ServerId"`
}

// UserDto represents a user.
type UserDto struct {
	Name                  string      `json:"Name"`
	ServerID              string      `json:"ServerId"`
	ID                    string      `json:"Id"`
	HasPassword           bool        `json:"HasPassword"`
	HasConfiguredPassword bool        `json:"HasConfiguredPassword"`
	EnableAutoLogin       bool        `json:"EnableAutoLogin"`
	LastLoginDate         string      `json:"LastLoginDate,omitempty"`
	LastActivityDate      string      `json:"LastActivityDate,omitempty"`
	Configuration         *UserConfig `json:"Configuration,omitempty"`
	Policy                *UserPolicy `json:"Policy,omitempty"`
}

// UserConfig holds user configuration.
type UserConfig struct {
	AudioLanguagePreference  string `json:"AudioLanguagePreference,omitempty"`
	MaxParentalRating        int    `json:"MaxParentalRating,omitempty"`
	RemoteClientBitrateLimit int    `json:"RemoteClientBitrateLimit,omitempty"`
}

// UserPolicy holds user permissions.
type UserPolicy struct {
	IsAdministrator       bool     `json:"IsAdministrator"`
	IsHidden              bool     `json:"IsHidden"`
	IsDisabled            bool     `json:"IsDisabled"`
	BlockedTags           []string `json:"BlockedTags,omitempty"`
	EnableContentDeletion bool     `json:"EnableContentDeletion"`
}

// SessionInfoDto represents an active session.
type SessionInfoDto struct {
	ID                   string        `json:"Id"`
	UserID               string        `json:"UserId"`
	UserName             string        `json:"UserName"`
	DeviceID             string        `json:"DeviceId"`
	DeviceName           string        `json:"DeviceName"`
	Client               string        `json:"Client"`
	LastActivityDate     string        `json:"LastActivityDate"`
	LastPlaybackCheckIn  string        `json:"LastPlaybackCheckIn,omitempty"`
	IsActive             bool          `json:"IsActive"`
	SupportsMediaControl bool          `json:"SupportsMediaControl"`
	NowPlayingItem       *BaseItemDto  `json:"NowPlayingItem,omitempty"`
	PlayState            *PlayStateDto `json:"PlayState,omitempty"`
}

// PlayStateDto holds playback state.
type PlayStateDto struct {
	CanSeek       bool   `json:"CanSeek"`
	IsPaused      bool   `json:"IsPaused"`
	IsMuted       bool   `json:"IsMuted"`
	PlayMethod    string `json:"PlayMethod,omitempty"`
	PositionTicks int64  `json:"PositionTicks,omitempty"`
}

// BaseItemDto represents a media item.
type BaseItemDto struct {
	Name              string            `json:"Name"`
	ServerID          string            `json:"ServerId"`
	ID                string            `json:"Id"`
	Type              string            `json:"Type"`
	Overview          string            `json:"Overview,omitempty"`
	Genres            []string          `json:"Genres,omitempty"`
	RunTimeTicks      int64             `json:"RunTimeTicks,omitempty"`
	PremiereDate      string            `json:"PremiereDate,omitempty"`
	ParentID          string            `json:"ParentId,omitempty"`
	IsFolder          bool              `json:"IsFolder"`
	ImageTags         map[string]string `json:"ImageTags,omitempty"`
	BackdropImageTags []string          `json:"BackdropImageTags,omitempty"`
	UserData          *UserItemDataDto  `json:"UserData,omitempty"`
	SeriesName        string            `json:"SeriesName,omitempty"`
	SeriesID          string            `json:"SeriesId,omitempty"`
	SeasonName        string            `json:"SeasonName,omitempty"`
	SeasonID          string            `json:"SeasonId,omitempty"`
	IndexNumber       int               `json:"IndexNumber,omitempty"`
	ParentIndexNumber int               `json:"ParentIndexNumber,omitempty"`
}

// UserItemDataDto holds user-specific item data.
type UserItemDataDto struct {
	PlayCount             int    `json:"PlayCount"`
	IsFavorite            bool   `json:"IsFavorite"`
	PlaybackPositionTicks int64  `json:"PlaybackPositionTicks"`
	LastPlayedDate        string `json:"LastPlayedDate,omitempty"`
}

// ItemsResult is a paginated list of items.
type ItemsResult struct {
	Items            []BaseItemDto `json:"Items"`
	TotalRecordCount int           `json:"TotalRecordCount"`
	StartIndex       int           `json:"StartIndex"`
}

// SystemInfo holds server information.
type SystemInfo struct {
	OperatingSystemDisplayName string `json:"OperatingSystemDisplayName"`
	ID                         string `json:"Id"`
	ServerName                 string `json:"ServerName"`
	Version                    string `json:"Version"`
	OperatingSystem            string `json:"OperatingSystem"`
}

// PublicSystemInfo is returned without authentication.
type PublicSystemInfo struct {
	LocalAddress    string `json:"LocalAddress"`
	ServerName      string `json:"ServerName"`
	Version         string `json:"Version"`
	ID              string `json:"Id"`
	OperatingSystem string `json:"OperatingSystem,omitempty"`
}

// authRequest is the body for AuthenticateByName.
type authRequest struct {
	Username string `json:"Username"`
	Pw       string `json:"Pw"`
}
