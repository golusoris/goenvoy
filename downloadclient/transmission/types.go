package transmission

import "encoding/json"

// Torrent represents a torrent managed by Transmission.
type Torrent struct {
	ID                 int           `json:"id"`
	Name               string        `json:"name"`
	HashString         string        `json:"hashString"`
	Status             int           `json:"status"`
	Error              int           `json:"error"`
	ErrorString        string        `json:"errorString"`
	AddedDate          int64         `json:"addedDate"`
	DoneDate           int64         `json:"doneDate"`
	ETA                int64         `json:"eta"`
	IsFinished         bool          `json:"isFinished"`
	IsStalled          bool          `json:"isStalled"`
	PercentDone        float64       `json:"percentDone"`
	SeedRatioMode      int           `json:"seedRatioMode"`
	SeedRatioLimit     float64       `json:"seedRatioLimit"`
	TotalSize          int64         `json:"totalSize"`
	SizeWhenDone       int64         `json:"sizeWhenDone"`
	LeftUntilDone      int64         `json:"leftUntilDone"`
	DownloadedEver     int64         `json:"downloadedEver"`
	UploadedEver       int64         `json:"uploadedEver"`
	UploadRatio        float64       `json:"uploadRatio"`
	RateDownload       int64         `json:"rateDownload"`
	RateUpload         int64         `json:"rateUpload"`
	DownloadDir        string        `json:"downloadDir"`
	PeersConnected     int           `json:"peersConnected"`
	PeersSendingToUs   int           `json:"peersSendingToUs"`
	PeersGettingFromUs int           `json:"peersGettingFromUs"`
	QueuePosition      int           `json:"queuePosition"`
	Labels             []string      `json:"labels"`
	Trackers           []TrackerInfo `json:"trackers"`
	Files              []File        `json:"files"`
	FileStats          []FileStat    `json:"fileStats"`
}

// TrackerInfo represents a tracker for a torrent.
type TrackerInfo struct {
	ID       int    `json:"id"`
	Announce string `json:"announce"`
	Scrape   string `json:"scrape"`
	Tier     int    `json:"tier"`
}

// File represents a file within a torrent.
type File struct {
	Name           string `json:"name"`
	Length         int64  `json:"length"`
	BytesCompleted int64  `json:"bytesCompleted"`
}

// FileStat contains per-file priority and wanted status.
type FileStat struct {
	BytesCompleted int64 `json:"bytesCompleted"`
	Wanted         bool  `json:"wanted"`
	Priority       int   `json:"priority"`
}

// SessionStats holds Transmission session statistics.
type SessionStats struct {
	ActiveTorrentCount int         `json:"activeTorrentCount"`
	DownloadSpeed      int64       `json:"downloadSpeed"`
	UploadSpeed        int64       `json:"uploadSpeed"`
	PausedTorrentCount int         `json:"pausedTorrentCount"`
	TorrentCount       int         `json:"torrentCount"`
	CumulativeStats    *CountStats `json:"cumulative-stats"`
	CurrentStats       *CountStats `json:"current-stats"`
}

// CountStats holds cumulative or current transfer statistics.
type CountStats struct {
	UploadedBytes   int64 `json:"uploadedBytes"`
	DownloadedBytes int64 `json:"downloadedBytes"`
	FilesAdded      int   `json:"filesAdded"`
	SessionCount    int   `json:"sessionCount"`
	SecondsActive   int64 `json:"secondsActive"`
}

// Session holds Transmission session/configuration information.
type Session struct {
	Version               string  `json:"version"`
	RPCVersion            int     `json:"rpc-version"`
	RPCVersionMinimum     int     `json:"rpc-version-minimum"`
	DownloadDir           string  `json:"download-dir"`
	IncompleteDir         string  `json:"incomplete-dir"`
	IncompleteDirEnabled  bool    `json:"incomplete-dir-enabled"`
	SpeedLimitDown        int64   `json:"speed-limit-down"`
	SpeedLimitDownEnabled bool    `json:"speed-limit-down-enabled"`
	SpeedLimitUp          int64   `json:"speed-limit-up"`
	SpeedLimitUpEnabled   bool    `json:"speed-limit-up-enabled"`
	AltSpeedDown          int64   `json:"alt-speed-down"`
	AltSpeedUp            int64   `json:"alt-speed-up"`
	AltSpeedEnabled       bool    `json:"alt-speed-enabled"`
	PeerPort              int     `json:"peer-port"`
	DHTEnabled            bool    `json:"dht-enabled"`
	PEXEnabled            bool    `json:"pex-enabled"`
	LPDEnabled            bool    `json:"lpd-enabled"`
	SeedRatioLimit        float64 `json:"seedRatioLimit"`
	SeedRatioLimited      bool    `json:"seedRatioLimited"`
	DownloadQueueSize     int     `json:"download-queue-size"`
	DownloadQueueEnabled  bool    `json:"download-queue-enabled"`
}

// FreeSpace holds the result of a free-space query.
type FreeSpace struct {
	Path      string `json:"path"`
	SizeBytes int64  `json:"size-bytes"`
}

// AddTorrentOptions holds parameters for adding a torrent.
type AddTorrentOptions struct {
	DownloadDir       string   `json:"download-dir,omitempty"`
	Paused            bool     `json:"paused,omitempty"`
	PeerLimit         int      `json:"peer-limit,omitempty"`
	BandwidthPriority int      `json:"bandwidthPriority,omitempty"`
	Labels            []string `json:"labels,omitempty"`
}

// AddTorrentResult holds the response from adding a torrent.
type AddTorrentResult struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	HashString string `json:"hashString"`
}

// rpcRequest is the JSON-RPC request envelope.
type rpcRequest struct {
	Method    string `json:"method"`
	Arguments any    `json:"arguments,omitempty"`
}

// rpcResponse is the JSON-RPC response envelope.
type rpcResponse struct {
	Result    string          `json:"result"`
	Arguments json.RawMessage `json:"arguments"`
}
