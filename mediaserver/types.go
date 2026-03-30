package mediaserver

// ServerInfo holds basic information about a media server.
type ServerInfo struct {
	// Name is the friendly name of the server.
	Name string `json:"name"`
	// Version is the server software version.
	Version string `json:"version"`
	// MachineID is the unique machine identifier.
	MachineID string `json:"machineId"`
	// Platform is the OS or platform the server runs on.
	Platform string `json:"platform,omitempty"`
}

// Library represents a media library section.
type Library struct {
	// ID is the section identifier.
	ID string `json:"id"`
	// Name is the display name of the library.
	Name string `json:"name"`
	// Type classifies the library (movie, show, music, photo).
	Type string `json:"type"`
	// ItemCount is the total number of items in the library.
	ItemCount int `json:"itemCount,omitempty"`
}

// Session represents an active playback session.
type Session struct {
	// ID is the session identifier.
	ID string `json:"id"`
	// UserName is the name of the user playing content.
	UserName string `json:"userName"`
	// Title is the title of the playing media.
	Title string `json:"title"`
	// State is the playback state (playing, paused, buffering).
	State string `json:"state"`
}
