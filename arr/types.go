package arr

import "time"

// StatusResponse holds the system status returned by /api/v3/system/status.
type StatusResponse struct {
	AppName                string `json:"appName"`
	InstanceName           string `json:"instanceName"`
	Version                string `json:"version"`
	BuildTime              string `json:"buildTime"`
	IsDebug                bool   `json:"isDebug"`
	IsProduction           bool   `json:"isProduction"`
	IsAdmin                bool   `json:"isAdmin"`
	IsUserInteractive      bool   `json:"isUserInteractive"`
	StartupPath            string `json:"startupPath"`
	AppData                string `json:"appData"`
	OsName                 string `json:"osName"`
	OsVersion              string `json:"osVersion"`
	IsMonoRuntime          bool   `json:"isMonoRuntime"`
	IsMono                 bool   `json:"isMono"`
	IsLinux                bool   `json:"isLinux"`
	IsOsx                  bool   `json:"isOsx"`
	IsWindows              bool   `json:"isWindows"`
	IsDocker               bool   `json:"isDocker"`
	Branch                 string `json:"branch"`
	Authentication         string `json:"authentication"`
	SqliteVersion          string `json:"sqliteVersion"`
	MigrationVersion       int    `json:"migrationVersion"`
	URLBase                string `json:"urlBase"`
	RuntimeVersion         string `json:"runtimeVersion"`
	RuntimeName            string `json:"runtimeName"`
	StartTime              string `json:"startTime"`
	PackageVersion         string `json:"packageVersion"`
	PackageAuthor          string `json:"packageAuthor"`
	PackageUpdateMechanism string `json:"packageUpdateMechanism"`
}

// HealthCheck represents a single health-check entry from /api/v3/health.
type HealthCheck struct {
	Source  string `json:"source"`
	Type    string `json:"type"`
	Message string `json:"message"`
	WikiURL string `json:"wikiUrl"`
}

// StatusMessage is an embedded status note inside a queue record.
type StatusMessage struct {
	Title    string   `json:"title"`
	Messages []string `json:"messages"`
}

// QueueRecord represents an item in the download queue.
type QueueRecord struct {
	ID                      int             `json:"id"`
	Title                   string          `json:"title"`
	Size                    float64         `json:"size"`
	SizeLeft                float64         `json:"sizeleft"`
	Status                  string          `json:"status"`
	TrackedDownloadStatus   string          `json:"trackedDownloadStatus"`
	TrackedDownloadState    string          `json:"trackedDownloadState"`
	StatusMessages          []StatusMessage `json:"statusMessages"`
	DownloadID              string          `json:"downloadId"`
	Protocol                string          `json:"protocol"`
	DownloadClient          string          `json:"downloadClient"`
	Indexer                 string          `json:"indexer"`
	OutputPath              string          `json:"outputPath"`
	TimeleftEstimation      time.Duration   `json:"-"`
	EstimatedCompletionTime string          `json:"estimatedCompletionTime"`
	Added                   string          `json:"added"`
}

// PagingResource wraps a paginated response from an *arr API.
type PagingResource[T any] struct {
	Page          int    `json:"page"`
	PageSize      int    `json:"pageSize"`
	SortKey       string `json:"sortKey"`
	SortDirection string `json:"sortDirection"`
	TotalRecords  int    `json:"totalRecords"`
	Records       []T    `json:"records"`
}

// CommandRequest is the payload sent to /api/v3/command.
type CommandRequest struct {
	Name string `json:"name"`
}

// CommandResponse is the reply from /api/v3/command.
type CommandResponse struct {
	ID                  int    `json:"id"`
	Name                string `json:"name"`
	Message             string `json:"message"`
	Started             string `json:"started"`
	Ended               string `json:"ended"`
	Status              string `json:"status"`
	Priority            string `json:"priority"`
	Trigger             string `json:"trigger"`
	StateChangeTime     string `json:"stateChangeTime"`
	SendUpdatesToClient bool   `json:"sendUpdatesToClient"`
	UpdateScheduledTask bool   `json:"updateScheduledTask"`
}

// QualityProfile describes a quality profile configured in the *arr app.
type QualityProfile struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	UpgradeAllowed    bool   `json:"upgradeAllowed"`
	CutoffFormatScore int    `json:"cutoffFormatScore"`
	MinFormatScore    int    `json:"minFormatScore"`
}

// Tag is a simple label used for organizing items in *arr applications.
type Tag struct {
	ID    int    `json:"id"`
	Label string `json:"label"`
}

// UnmappedFolder represents a folder not yet mapped to a root folder.
type UnmappedFolder struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// RootFolder represents a configured root folder in an *arr application.
type RootFolder struct {
	ID              int              `json:"id"`
	Path            string           `json:"path"`
	Accessible      bool             `json:"accessible"`
	FreeSpace       int64            `json:"freeSpace"`
	UnmappedFolders []UnmappedFolder `json:"unmappedFolders"`
}

// DiskSpace contains disk usage information.
type DiskSpace struct {
	Path       string `json:"path"`
	Label      string `json:"label"`
	FreeSpace  int64  `json:"freeSpace"`
	TotalSpace int64  `json:"totalSpace"`
}
