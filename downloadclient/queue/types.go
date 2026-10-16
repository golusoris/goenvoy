package queue

// Kind identifies a supported download client.
type Kind string

const (
	// KindQBittorrent identifies qBittorrent.
	KindQBittorrent Kind = "qbittorrent"
	// KindTransmission identifies Transmission.
	KindTransmission Kind = "transmission"
	// KindDeluge identifies Deluge.
	KindDeluge Kind = "deluge"
	// KindRTorrent identifies rTorrent.
	KindRTorrent Kind = "rtorrent"
	// KindSABnzbd identifies SABnzbd.
	KindSABnzbd Kind = "sabnzbd"
	// KindNZBGet identifies NZBGet.
	KindNZBGet Kind = "nzbget"
)

// Target specifies the connection parameters for observing a download client.
type Target struct {
	Kind     Kind   `json:"kind"`
	BaseURL  string `json:"baseUrl"`
	Username string `json:"username,omitempty"`
	// Secret is the password, API key or RPC secret; never serialized.
	Secret string `json:"-"`
}

// Snapshot represents a point-in-time observation of a download client queue.
type Snapshot struct {
	Kind         Kind   `json:"kind"`
	Paused       bool   `json:"paused"`
	Active       int    `json:"active"`
	Queued       int    `json:"queued"`
	DownloadRate int64  `json:"downloadRate"`
	UploadRate   int64  `json:"uploadRate"`
	Version      string `json:"version"`
}
