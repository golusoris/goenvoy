package rtorrent

// Torrent represents a torrent managed by rTorrent.
type Torrent struct {
	Hash           string
	Name           string
	SizeBytes      int64
	CompletedBytes int64
	DownRate       int64
	UpRate         int64
	UpTotal        int64
	Ratio          int64
	IsOpen         bool
	IsActive       bool
	IsComplete     bool
	IsHashChecking bool
	BasePath       string
	Directory      string
	Label          string
	AddedTime      int64
	Message        string
}

// File represents a file within a torrent.
type File struct {
	Path            string
	SizeBytes       int64
	CompletedChunks int64
	SizeChunks      int64
	Priority        int
}

// Tracker represents a tracker for a torrent.
type Tracker struct {
	URL  string
	Type int
}

// SystemInfo holds rTorrent system information.
type SystemInfo struct {
	ClientVersion  string
	LibraryVersion string
	Hostname       string
	PID            int
}
