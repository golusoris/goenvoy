package nzbget

import "encoding/json"

// Group represents a download in the NZBGet queue.
type Group struct {
	ID               int    `json:"NZBID"`
	Name             string `json:"NZBName"`
	Filename         string `json:"NZBFilename"`
	Status           string `json:"Status"`
	Category         string `json:"Category"`
	FileSizeLo       int64  `json:"FileSizeLo"`
	FileSizeHi       int64  `json:"FileSizeHi"`
	RemainingSizeLo  int64  `json:"RemainingSizeLo"`
	RemainingSizeHi  int64  `json:"RemainingSizeHi"`
	DownloadedSizeLo int64  `json:"DownloadedSizeLo"`
	DownloadedSizeHi int64  `json:"DownloadedSizeHi"`
	PausedSizeLo     int64  `json:"PausedSizeLo"`
	PausedSizeHi     int64  `json:"PausedSizeHi"`
	MaxPriority      int    `json:"MaxPriority"`
	ActiveDownloads  int    `json:"ActiveDownloads"`
	TotalArticles    int    `json:"TotalArticles"`
	SuccessArticles  int    `json:"SuccessArticles"`
	FailedArticles   int    `json:"FailedArticles"`
	Health           int    `json:"Health"`
	DestDir          string `json:"DestDir"`
	FinalDir         string `json:"FinalDir"`
	PostTotalTimeSec int    `json:"PostTotalTimeSec"`
	MinPostTime      int64  `json:"MinPostTime"`
	MaxPostTime      int64  `json:"MaxPostTime"`
}

// HistoryItem represents a completed download in NZBGet history.
type HistoryItem struct {
	ID              int    `json:"NZBID"`
	Name            string `json:"Name"`
	Status          string `json:"Status"`
	Category        string `json:"Category"`
	FileSizeLo      int64  `json:"FileSizeLo"`
	FileSizeHi      int64  `json:"FileSizeHi"`
	DestDir         string `json:"DestDir"`
	FinalDir        string `json:"FinalDir"`
	DownloadTimeSec int    `json:"DownloadTimeSec"`
	Health          int    `json:"Health"`
	DeleteStatus    string `json:"DeleteStatus"`
	MarkStatus      string `json:"MarkStatus"`
	ParStatus       string `json:"ParStatus"`
	UnpackStatus    string `json:"UnpackStatus"`
	ScriptStatuses  []struct {
		Name   string `json:"Name"`
		Status string `json:"Status"`
	} `json:"ScriptStatuses"`
}

// Status holds NZBGet server status.
type Status struct {
	ServerPaused     bool  `json:"ServerPaused"`
	DownloadPaused   bool  `json:"DownloadPaused"`
	DownloadRate     int64 `json:"DownloadRate"`
	RemainingSizeLo  int64 `json:"RemainingSizeLo"`
	RemainingSizeHi  int64 `json:"RemainingSizeHi"`
	DownloadedSizeLo int64 `json:"DownloadedSizeLo"`
	DownloadedSizeHi int64 `json:"DownloadedSizeHi"`
	ArticleCacheLo   int64 `json:"ArticleCacheLo"`
	ArticleCacheHi   int64 `json:"ArticleCacheHi"`
	FreeDiskSpaceLo  int64 `json:"FreeDiskSpaceLo"`
	FreeDiskSpaceHi  int64 `json:"FreeDiskSpaceHi"`
	DownloadLimit    int64 `json:"DownloadLimit"`
	UpTimeSec        int   `json:"UpTimeSec"`
	DownloadTimeSec  int   `json:"DownloadTimeSec"`
	ThreadCount      int   `json:"ThreadCount"`
	NewsServers      []struct {
		ID     int  `json:"ID"`
		Active bool `json:"Active"`
	} `json:"NewsServers"`
}

// ConfigEntry is a single NZBGet configuration entry.
type ConfigEntry struct {
	Name  string `json:"Name"`
	Value string `json:"Value"`
}

// LogEntry is a single NZBGet log entry.
type LogEntry struct {
	ID   int    `json:"ID"`
	Kind string `json:"Kind"`
	Time int64  `json:"Time"`
	Text string `json:"Text"`
}

// rpcRequest is the JSON-RPC request envelope.
type rpcRequest struct {
	Version string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
	ID      int    `json:"id"`
}

// rpcResponse is the JSON-RPC response envelope.
type rpcResponse struct {
	Version string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *rpcError       `json:"error"`
	ID      int             `json:"id"`
}

// rpcError represents a JSON-RPC error.
type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
