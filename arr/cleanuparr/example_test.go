package cleanuparr_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/golusoris/goenvoy/arr/cleanuparr"
)

func Example() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"application":{"version":"2.10.11","startTime":"2026-09-20T08:00:00Z"},"mediaManagers":{"Sonarr":{"instanceCount":2}}}`)
	}))
	defer server.Close()

	client, err := cleanuparr.New(server.URL, "example-api-key")
	if err != nil {
		return
	}

	status, err := client.GetStatus(context.Background())
	if err != nil {
		return
	}

	fmt.Println(status.Application.Version)
	fmt.Println(status.MediaManagers["Sonarr"].InstanceCount)
	// Output:
	// 2.10.11
	// 2
}

func ExampleClient_GetSeekerConfig() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"searchEnabled":true,"searchInterval":15,"proactiveSearchEnabled":true,"selectionStrategy":"BalancedWeighted","useRoundRobin":true,"postReleaseGraceHours":4,"instances":[{"arrInstanceId":"11111111-1111-1111-1111-111111111111","instanceName":"sonarr-main","instanceType":"Sonarr","enabled":true,"skipTags":["archive"],"lastProcessedAt":"2026-09-20T08:00:00Z","arrInstanceEnabled":true,"activeDownloadLimit":10,"ignoreStruckDownloads":true,"minCycleTimeDays":2,"monitoredOnly":true,"useCutoff":true,"useCustomFormatScore":false}]}`)
	}))
	defer server.Close()

	client, err := cleanuparr.New(server.URL, "example-api-key")
	if err != nil {
		return
	}

	cfg, err := client.GetSeekerConfig(context.Background())
	if err != nil {
		return
	}

	fmt.Println(cfg.SearchEnabled)
	fmt.Println(cfg.SelectionStrategy)
	fmt.Println(cfg.Instances[0].InstanceName)
	// Output:
	// true
	// BalancedWeighted
	// sonarr-main
}
