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
