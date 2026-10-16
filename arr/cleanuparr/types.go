package cleanuparr

import (
	"fmt"
	"time"

	"github.com/golusoris/goenvoy/arr/v2"
)

// APIError represents an HTTP error response from the Cleanuparr API.
type APIError = arr.APIError

// MediaManager reports instance counts for a specific *arr service type.
type MediaManager struct {
	InstanceCount int `json:"instanceCount"` //nolint:tagliatelle // Cleanuparr speaks camelCase.
}

// ApplicationStatus holds application runtime metadata from Cleanuparr.
type ApplicationStatus struct {
	Version       string  `json:"version"`                 //nolint:tagliatelle // Cleanuparr speaks camelCase.
	StartTime     string  `json:"startTime"`               //nolint:tagliatelle // Cleanuparr speaks camelCase.
	UpTime        string  `json:"upTime,omitempty"`        //nolint:tagliatelle // Cleanuparr speaks camelCase.
	MemoryUsageMB float64 `json:"memoryUsageMB,omitempty"` //nolint:tagliatelle // Cleanuparr speaks camelCase.
	ProcessorTime string  `json:"processorTime,omitempty"` //nolint:tagliatelle // Cleanuparr speaks camelCase.
}

// StartTimeParsed parses StartTime as time.Time trying [time.RFC3339Nano]
// followed by a zone-less .NET DateTime layout ("2006-01-02T15:04:05.9999999"),
// returning the local-naive value in UTC. It returns an error only if both formats fail.
func (a ApplicationStatus) StartTimeParsed() (time.Time, error) {
	if parsed, err := time.Parse(time.RFC3339Nano, a.StartTime); err == nil {
		return parsed, nil
	}
	parsed, err := time.Parse("2006-01-02T15:04:05.9999999", a.StartTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("cleanuparr: parse start time %q: %w", a.StartTime, err)
	}
	return parsed.UTC(), nil
}

// Status represents the system status response from GET /api/status.
type Status struct {
	Application   ApplicationStatus       `json:"application"`   //nolint:tagliatelle // Cleanuparr speaks camelCase.
	MediaManagers map[string]MediaManager `json:"mediaManagers"` //nolint:tagliatelle // Cleanuparr speaks camelCase.
}

// Strike represents an individual recent strike returned by GET /api/strikes/recent.
type Strike struct {
	ID         string    `json:"id"`         //nolint:tagliatelle // Cleanuparr speaks camelCase.
	Type       string    `json:"type"`       //nolint:tagliatelle // Cleanuparr speaks camelCase.
	CreatedAt  time.Time `json:"createdAt"`  //nolint:tagliatelle // Cleanuparr speaks camelCase.
	DownloadID string    `json:"downloadId"` //nolint:tagliatelle // Cleanuparr speaks camelCase.
	Title      string    `json:"title"`      //nolint:tagliatelle // Cleanuparr speaks camelCase.
	IsDryRun   bool      `json:"isDryRun"`   //nolint:tagliatelle // Cleanuparr speaks camelCase.
}

// StrikeDetail represents individual strike details inside a download item.
type StrikeDetail struct {
	ID                  string    `json:"id"`                            //nolint:tagliatelle // Cleanuparr speaks camelCase.
	Type                string    `json:"type"`                          //nolint:tagliatelle // Cleanuparr speaks camelCase.
	CreatedAt           time.Time `json:"createdAt"`                     //nolint:tagliatelle // Cleanuparr speaks camelCase.
	LastDownloadedBytes *int64    `json:"lastDownloadedBytes,omitempty"` //nolint:tagliatelle // Cleanuparr speaks camelCase.
	JobRunID            string    `json:"jobRunId"`                      //nolint:tagliatelle // Cleanuparr speaks camelCase.
	IsDryRun            bool      `json:"isDryRun"`                      //nolint:tagliatelle // Cleanuparr speaks camelCase.
}

// DownloadItemStrikes represents a download item grouped with its strikes and removal status.
type DownloadItemStrikes struct {
	DownloadItemID     string         `json:"downloadItemId"`          //nolint:tagliatelle // Cleanuparr speaks camelCase.
	DownloadID         string         `json:"downloadId"`              //nolint:tagliatelle // Cleanuparr speaks camelCase.
	Title              string         `json:"title"`                   //nolint:tagliatelle // Cleanuparr speaks camelCase.
	TotalStrikes       int            `json:"totalStrikes"`            //nolint:tagliatelle // Cleanuparr speaks camelCase.
	StrikesByType      map[string]int `json:"strikesByType,omitempty"` //nolint:tagliatelle // Cleanuparr speaks camelCase.
	LatestStrikeAt     time.Time      `json:"latestStrikeAt"`          //nolint:tagliatelle // Cleanuparr speaks camelCase.
	FirstStrikeAt      time.Time      `json:"firstStrikeAt"`           //nolint:tagliatelle // Cleanuparr speaks camelCase.
	IsMarkedForRemoval bool           `json:"isMarkedForRemoval"`      //nolint:tagliatelle // Cleanuparr speaks camelCase.
	IsRemoved          bool           `json:"isRemoved"`               //nolint:tagliatelle // Cleanuparr speaks camelCase.
	IsReturning        bool           `json:"isReturning"`             //nolint:tagliatelle // Cleanuparr speaks camelCase.
	HasDryRunStrikes   bool           `json:"hasDryRunStrikes"`        //nolint:tagliatelle // Cleanuparr speaks camelCase.
	Strikes            []StrikeDetail `json:"strikes,omitempty"`       //nolint:tagliatelle // Cleanuparr speaks camelCase.
}

// StrikePage represents a paginated collection of download item strikes.
type StrikePage struct {
	Items      []DownloadItemStrikes `json:"items"`      //nolint:tagliatelle // Cleanuparr speaks camelCase.
	Page       int                   `json:"page"`       //nolint:tagliatelle // Cleanuparr speaks camelCase.
	PageSize   int                   `json:"pageSize"`   //nolint:tagliatelle // Cleanuparr speaks camelCase.
	TotalCount int                   `json:"totalCount"` //nolint:tagliatelle // Cleanuparr speaks camelCase.
	TotalPages int                   `json:"totalPages"` //nolint:tagliatelle // Cleanuparr speaks camelCase.
}

// StrikeSummary aggregates counts over struck downloads.
type StrikeSummary struct {
	StruckDownloads    int  `json:"struckDownloads"`    //nolint:tagliatelle // Cleanuparr speaks camelCase.
	MarkedForRemoval   int  `json:"markedForRemoval"`   //nolint:tagliatelle // Cleanuparr speaks camelCase.
	RemovedDownloads   int  `json:"removedDownloads"`   //nolint:tagliatelle // Cleanuparr speaks camelCase.
	ReturningDownloads int  `json:"returningDownloads"` //nolint:tagliatelle // Cleanuparr speaks camelCase.
	Partial            bool `json:"partial"`            //nolint:tagliatelle // Cleanuparr speaks camelCase.
}
