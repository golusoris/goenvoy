package downloadclient

import "context"

// TransferState represents the current state of a download transfer.
type TransferState string

const (
	// TransferStateDownloading means the transfer is actively downloading.
	TransferStateDownloading TransferState = "downloading"
	// TransferStatePaused means the transfer is paused.
	TransferStatePaused TransferState = "paused"
	// TransferStateSeeding means the transfer has completed and is seeding.
	TransferStateSeeding TransferState = "seeding"
	// TransferStateCompleted means the transfer is finished.
	TransferStateCompleted TransferState = "completed"
	// TransferStateError means the transfer encountered an error.
	TransferStateError TransferState = "error"
	// TransferStateQueued means the transfer is queued and waiting.
	TransferStateQueued TransferState = "queued"
)

// TransferStatus holds the current status of a download transfer.
type TransferStatus struct {
	// ID is the client-specific identifier for this transfer.
	ID string `json:"id"`
	// Name is the display name of the transfer (usually the torrent/NZB name).
	Name string `json:"name"`
	// State is the current transfer state.
	State TransferState `json:"state"`
	// Progress is the download progress from 0.0 to 1.0.
	Progress float64 `json:"progress"`
	// SizeBytes is the total size in bytes.
	SizeBytes int64 `json:"sizeBytes"`
	// DownloadedBytes is the number of bytes downloaded so far.
	DownloadedBytes int64 `json:"downloadedBytes"`
	// DownloadRate is the current download speed in bytes per second.
	DownloadRate int64 `json:"downloadRate"`
	// UploadRate is the current upload speed in bytes per second.
	UploadRate int64 `json:"uploadRate"`
	// SavePath is the destination directory for completed files.
	SavePath string `json:"savePath"`
	// Category assigned to this transfer.
	Category string `json:"category,omitempty"`
}

// ClientInfo holds basic information about the download client application.
type ClientInfo struct {
	// Name is the application name (e.g. "qBittorrent", "SABnzbd").
	Name string `json:"name"`
	// Version is the application version string.
	Version string `json:"version"`
}

// Downloader is the common interface that all download client implementations
// must satisfy. Methods accept a context for cancellation and timeout control.
type Downloader interface {
	// Info returns basic information about the download client.
	Info(ctx context.Context) (*ClientInfo, error)
	// List returns all current transfers.
	List(ctx context.Context) ([]TransferStatus, error)
	// Pause pauses the transfer identified by id.
	Pause(ctx context.Context, id string) error
	// Resume resumes a paused transfer identified by id.
	Resume(ctx context.Context, id string) error
	// Remove removes the transfer identified by id. If deleteFiles is true,
	// downloaded data is also deleted from disk.
	Remove(ctx context.Context, id string, deleteFiles bool) error
}
