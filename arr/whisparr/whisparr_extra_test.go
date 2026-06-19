package whisparr_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golusoris/goenvoy/arr/v2"
	"github.com/golusoris/goenvoy/arr/whisparr"
)

func newV3JSONEndpointServer(t *testing.T, method, wantURI, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			t.Errorf("method = %s, want %s", r.Method, method)
		}
		if got := r.URL.RequestURI(); got != wantURI {
			t.Errorf("uri = %s, want %s", got, wantURI)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if body != "" {
			_, _ = w.Write([]byte(body))
		}
	}))
}

func TestV3ProviderEndpointWrappers(t *testing.T) {
	t.Parallel()

	provider := &arr.ProviderResource{ID: 1, Name: "provider"}
	bulk := &arr.ProviderBulkResource{IDs: []int{1}}
	cases := []struct {
		name   string
		method string
		uri    string
		body   string
		call   func(*whisparr.ClientV3) error
	}{
		{
			name: "get download clients", method: http.MethodGet, uri: "/api/v3/downloadclient", body: `[{"id":1,"name":"qbit"}]`,
			call: func(c *whisparr.ClientV3) error { _, err := c.GetDownloadClients(context.Background()); return err },
		},
		{
			name: "get download client", method: http.MethodGet, uri: "/api/v3/downloadclient/1", body: `{"id":1,"name":"qbit"}`,
			call: func(c *whisparr.ClientV3) error { _, err := c.GetDownloadClient(context.Background(), 1); return err },
		},
		{
			name: "create download client", method: http.MethodPost, uri: "/api/v3/downloadclient", body: `{"id":1,"name":"qbit"}`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.CreateDownloadClient(context.Background(), provider)
				return err
			},
		},
		{
			name: "update download client", method: http.MethodPut, uri: "/api/v3/downloadclient/1", body: `{"id":1,"name":"qbit"}`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.UpdateDownloadClient(context.Background(), provider)
				return err
			},
		},
		{
			name: "download client schema", method: http.MethodGet, uri: "/api/v3/downloadclient/schema", body: `[{"id":1}]`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.GetDownloadClientSchema(context.Background())
				return err
			},
		},
		{
			name: "test download client", method: http.MethodPost, uri: "/api/v3/downloadclient/test", body: `{}`,
			call: func(c *whisparr.ClientV3) error { return c.TestDownloadClient(context.Background(), provider) },
		},
		{
			name: "test all download clients", method: http.MethodPost, uri: "/api/v3/downloadclient/testall", body: `{}`,
			call: func(c *whisparr.ClientV3) error { return c.TestAllDownloadClients(context.Background()) },
		},
		{
			name: "bulk update download clients", method: http.MethodPut, uri: "/api/v3/downloadclient/bulk", body: `[{"id":1}]`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.BulkUpdateDownloadClients(context.Background(), bulk)
				return err
			},
		},
		{
			name: "bulk delete download clients", method: http.MethodDelete, uri: "/api/v3/downloadclient/bulk", body: `{}`,
			call: func(c *whisparr.ClientV3) error { return c.BulkDeleteDownloadClients(context.Background(), []int{1}) },
		},
		{
			name: "get indexer", method: http.MethodGet, uri: "/api/v3/indexer/1", body: `{"id":1,"name":"indexer"}`,
			call: func(c *whisparr.ClientV3) error { _, err := c.GetIndexer(context.Background(), 1); return err },
		},
		{
			name: "update indexer", method: http.MethodPut, uri: "/api/v3/indexer/1", body: `{"id":1,"name":"indexer"}`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.UpdateIndexer(context.Background(), provider)
				return err
			},
		},
		{
			name: "indexer schema", method: http.MethodGet, uri: "/api/v3/indexer/schema", body: `[{"id":1}]`,
			call: func(c *whisparr.ClientV3) error { _, err := c.GetIndexerSchema(context.Background()); return err },
		},
		{
			name: "test indexer", method: http.MethodPost, uri: "/api/v3/indexer/test", body: `{}`,
			call: func(c *whisparr.ClientV3) error { return c.TestIndexer(context.Background(), provider) },
		},
		{
			name: "bulk update indexers", method: http.MethodPut, uri: "/api/v3/indexer/bulk", body: `[{"id":1}]`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.BulkUpdateIndexers(context.Background(), bulk)
				return err
			},
		},
		{
			name: "bulk delete indexers", method: http.MethodDelete, uri: "/api/v3/indexer/bulk", body: `{}`,
			call: func(c *whisparr.ClientV3) error { return c.BulkDeleteIndexers(context.Background(), []int{1}) },
		},
		{
			name: "get import list", method: http.MethodGet, uri: "/api/v3/importlist/1", body: `{"id":1,"name":"list"}`,
			call: func(c *whisparr.ClientV3) error { _, err := c.GetImportList(context.Background(), 1); return err },
		},
		{
			name: "update import list", method: http.MethodPut, uri: "/api/v3/importlist/1", body: `{"id":1,"name":"list"}`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.UpdateImportList(context.Background(), provider)
				return err
			},
		},
		{
			name: "import list schema", method: http.MethodGet, uri: "/api/v3/importlist/schema", body: `[{"id":1}]`,
			call: func(c *whisparr.ClientV3) error { _, err := c.GetImportListSchema(context.Background()); return err },
		},
		{
			name: "test import list", method: http.MethodPost, uri: "/api/v3/importlist/test", body: `{}`,
			call: func(c *whisparr.ClientV3) error { return c.TestImportList(context.Background(), provider) },
		},
		{
			name: "bulk update import lists", method: http.MethodPut, uri: "/api/v3/importlist/bulk", body: `[{"id":1}]`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.BulkUpdateImportLists(context.Background(), bulk)
				return err
			},
		},
		{
			name: "bulk delete import lists", method: http.MethodDelete, uri: "/api/v3/importlist/bulk", body: `{}`,
			call: func(c *whisparr.ClientV3) error { return c.BulkDeleteImportLists(context.Background(), []int{1}) },
		},
		{
			name: "import list movies", method: http.MethodGet, uri: "/api/v3/importlist/movie", body: `[{"id":1,"title":"Scene"}]`,
			call: func(c *whisparr.ClientV3) error { _, err := c.GetImportListMovies(context.Background()); return err },
		},
		{
			name: "get metadata by id", method: http.MethodGet, uri: "/api/v3/metadata/1", body: `{"id":1,"name":"metadata"}`,
			call: func(c *whisparr.ClientV3) error { _, err := c.GetMetadataByID(context.Background(), 1); return err },
		},
		{
			name: "update metadata", method: http.MethodPut, uri: "/api/v3/metadata/1", body: `{"id":1,"name":"metadata"}`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.UpdateMetadata(context.Background(), provider)
				return err
			},
		},
		{
			name: "metadata schema", method: http.MethodGet, uri: "/api/v3/metadata/schema", body: `[{"id":1}]`,
			call: func(c *whisparr.ClientV3) error { _, err := c.GetMetadataSchema(context.Background()); return err },
		},
		{
			name: "test metadata", method: http.MethodPost, uri: "/api/v3/metadata/test", body: `{}`,
			call: func(c *whisparr.ClientV3) error { return c.TestMetadata(context.Background(), provider) },
		},
		{
			name: "test all metadata", method: http.MethodPost, uri: "/api/v3/metadata/testall", body: `{}`,
			call: func(c *whisparr.ClientV3) error { return c.TestAllMetadata(context.Background()) },
		},
		{
			name: "get notification", method: http.MethodGet, uri: "/api/v3/notification/1", body: `{"id":1,"name":"notify"}`,
			call: func(c *whisparr.ClientV3) error { _, err := c.GetNotification(context.Background(), 1); return err },
		},
		{
			name: "update notification", method: http.MethodPut, uri: "/api/v3/notification/1", body: `{"id":1,"name":"notify"}`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.UpdateNotification(context.Background(), provider)
				return err
			},
		},
		{
			name: "notification schema", method: http.MethodGet, uri: "/api/v3/notification/schema", body: `[{"id":1}]`,
			call: func(c *whisparr.ClientV3) error { _, err := c.GetNotificationSchema(context.Background()); return err },
		},
		{
			name: "test notification", method: http.MethodPost, uri: "/api/v3/notification/test", body: `{}`,
			call: func(c *whisparr.ClientV3) error { return c.TestNotification(context.Background(), provider) },
		},
		{
			name: "test all notifications", method: http.MethodPost, uri: "/api/v3/notification/testall", body: `{}`,
			call: func(c *whisparr.ClientV3) error { return c.TestAllNotifications(context.Background()) },
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			srv := newV3JSONEndpointServer(t, tc.method, tc.uri, tc.body)
			defer srv.Close()
			c, err := whisparr.NewV3(srv.URL, "test-key")
			if err != nil {
				t.Fatalf("NewV3: %v", err)
			}
			if err := tc.call(c); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
		})
	}
}

func TestV3ActionEndpointWrappers(t *testing.T) {
	t.Parallel()

	provider := &arr.ProviderResource{ID: 1, Name: "provider"}
	cases := []struct {
		name string
		uri  string
		call func(*whisparr.ClientV3) error
	}{
		{"download client action", "/api/v3/downloadclient/action/testAction", func(c *whisparr.ClientV3) error {
			return c.DownloadClientAction(context.Background(), "testAction", provider)
		}},
		{"import list action", "/api/v3/importlist/action/testAction", func(c *whisparr.ClientV3) error {
			return c.ImportListAction(context.Background(), "testAction", provider)
		}},
		{"indexer action", "/api/v3/indexer/action/testAction", func(c *whisparr.ClientV3) error {
			return c.IndexerAction(context.Background(), "testAction", provider)
		}},
		{"metadata action", "/api/v3/metadata/action/testAction", func(c *whisparr.ClientV3) error {
			return c.MetadataAction(context.Background(), "testAction", provider)
		}},
		{"notification action", "/api/v3/notification/action/testAction", func(c *whisparr.ClientV3) error {
			return c.NotificationAction(context.Background(), "testAction", provider)
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			srv := newV3JSONEndpointServer(t, http.MethodPost, tc.uri, `{}`)
			defer srv.Close()
			c, err := whisparr.NewV3(srv.URL, "test-key")
			if err != nil {
				t.Fatalf("NewV3: %v", err)
			}
			if err := tc.call(c); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
		})
	}
}

func TestV3ConfigEndpointWrappers(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		method string
		uri    string
		body   string
		call   func(*whisparr.ClientV3) error
	}{
		{
			name: "update download client config", method: http.MethodPut, uri: "/api/v3/config/downloadclient/1", body: `{"id":1}`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.UpdateDownloadClientConfig(context.Background(), &arr.DownloadClientConfigResource{ID: 1})
				return err
			},
		},
		{
			name: "update host config", method: http.MethodPut, uri: "/api/v3/config/host/1", body: `{"id":1}`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.UpdateHostConfig(context.Background(), &arr.HostConfigResource{ID: 1})
				return err
			},
		},
		{
			name: "update indexer config", method: http.MethodPut, uri: "/api/v3/config/indexer/1", body: `{"id":1}`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.UpdateIndexerConfig(context.Background(), &arr.IndexerConfigResource{ID: 1})
				return err
			},
		},
		{
			name: "update media management config", method: http.MethodPut, uri: "/api/v3/config/mediamanagement/1", body: `{"id":1}`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.UpdateMediaManagementConfig(context.Background(), &arr.MediaManagementConfigResource{ID: 1})
				return err
			},
		},
		{
			name: "update naming config", method: http.MethodPut, uri: "/api/v3/config/naming/1", body: `{"id":1}`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.UpdateNamingConfig(context.Background(), &arr.NamingConfigResource{ID: 1})
				return err
			},
		},
		{
			name: "naming examples", method: http.MethodGet, uri: "/api/v3/config/naming/examples", body: `{"movie":"Example"}`,
			call: func(c *whisparr.ClientV3) error { _, err := c.GetNamingExamples(context.Background()); return err },
		},
		{
			name: "update ui config", method: http.MethodPut, uri: "/api/v3/config/ui/1", body: `{"id":1}`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.UpdateUIConfig(context.Background(), &arr.UIConfigResource{ID: 1})
				return err
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			srv := newV3JSONEndpointServer(t, tc.method, tc.uri, tc.body)
			defer srv.Close()
			c, err := whisparr.NewV3(srv.URL, "test-key")
			if err != nil {
				t.Fatalf("NewV3: %v", err)
			}
			if err := tc.call(c); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
		})
	}
}

func TestV3QueueReleaseAndQualityWrappers(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		method string
		uri    string
		body   string
		call   func(*whisparr.ClientV3) error
	}{
		{
			name: "custom format", method: http.MethodGet, uri: "/api/v3/customformat/1", body: `{"id":1,"name":"format"}`,
			call: func(c *whisparr.ClientV3) error { _, err := c.GetCustomFormat(context.Background(), 1); return err },
		},
		{
			name: "update custom format", method: http.MethodPut, uri: "/api/v3/customformat/1", body: `{"id":1,"name":"format"}`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.UpdateCustomFormat(context.Background(), &arr.CustomFormatResource{ID: 1})
				return err
			},
		},
		{
			name: "custom format schema", method: http.MethodGet, uri: "/api/v3/customformat/schema", body: `[{"name":"size"}]`,
			call: func(c *whisparr.ClientV3) error { _, err := c.GetCustomFormatSchema(context.Background()); return err },
		},
		{
			name: "import movie", method: http.MethodPost, uri: "/api/v3/movie/import", body: `{}`,
			call: func(c *whisparr.ClientV3) error {
				return c.ImportMovie(context.Background(), []whisparr.Movie{{ID: 1}})
			},
		},
		{
			name: "movie list", method: http.MethodGet, uri: "/api/v3/movie?movieIds=1&movieIds=2", body: `[{"id":1},{"id":2}]`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.GetMovieList(context.Background(), []int{1, 2})
				return err
			},
		},
		{
			name: "bulk delete queue", method: http.MethodDelete, uri: "/api/v3/queue/bulk", body: `{}`,
			call: func(c *whisparr.ClientV3) error {
				return c.BulkDeleteQueue(context.Background(), &arr.QueueBulkResource{IDs: []int{1}})
			},
		},
		{
			name: "bulk grab queue", method: http.MethodPost, uri: "/api/v3/queue/grab/bulk", body: `{}`,
			call: func(c *whisparr.ClientV3) error { return c.BulkGrabQueue(context.Background(), []int{1}) },
		},
		{
			name: "queue details", method: http.MethodGet, uri: "/api/v3/queue/details", body: `[{"id":1,"title":"Queued"}]`,
			call: func(c *whisparr.ClientV3) error { _, err := c.GetQueueDetails(context.Background()); return err },
		},
		{
			name: "queue details by movie", method: http.MethodGet, uri: "/api/v3/queue/details?movieIds=1&movieIds=2", body: `[{"id":1,"title":"Queued"}]`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.GetQueueDetailsByMovieID(context.Background(), []int{1, 2})
				return err
			},
		},
		{
			name: "queue status", method: http.MethodGet, uri: "/api/v3/queue/status", body: `{"totalCount":1}`,
			call: func(c *whisparr.ClientV3) error { _, err := c.GetQueueStatus(context.Background()); return err },
		},
		{
			name: "push release", method: http.MethodPost, uri: "/api/v3/release/push", body: `[{"guid":"guid-1"}]`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.PushRelease(context.Background(), &arr.ReleasePushResource{Title: "Release", DownloadURL: "https://example.test/file"})
				return err
			},
		},
		{
			name: "bulk update quality definitions", method: http.MethodPut, uri: "/api/v3/qualitydefinition/update", body: `[{"id":1}]`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.BulkUpdateQualityDefinitions(context.Background(), []arr.QualityDefinitionResource{{ID: 1}})
				return err
			},
		},
		{
			name: "quality definition limits", method: http.MethodGet, uri: "/api/v3/qualitydefinition/limits", body: `{}`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.GetQualityDefinitionLimits(context.Background())
				return err
			},
		},
		{
			name: "update quality profile", method: http.MethodPut, uri: "/api/v3/qualityprofile/1", body: `{"id":1,"name":"Any"}`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.UpdateQualityProfile(context.Background(), &arr.QualityProfile{ID: 1})
				return err
			},
		},
		{
			name: "quality profile schema", method: http.MethodGet, uri: "/api/v3/qualityprofile/schema", body: `{"id":1,"name":"Any"}`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.GetQualityProfileSchema(context.Background())
				return err
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			srv := newV3JSONEndpointServer(t, tc.method, tc.uri, tc.body)
			defer srv.Close()
			c, err := whisparr.NewV3(srv.URL, "test-key")
			if err != nil {
				t.Fatalf("NewV3: %v", err)
			}
			if err := tc.call(c); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
		})
	}
}

func TestV2AdditionalEndpointWrappers(t *testing.T) {
	t.Parallel()

	provider := &arr.ProviderResource{ID: 1, Name: "provider"}
	bulk := &arr.ProviderBulkResource{IDs: []int{1}}
	cases := []struct {
		name   string
		method string
		uri    string
		body   string
		call   func(*whisparr.Client) error
	}{
		{
			name: "import list action", method: http.MethodPost, uri: "/api/v3/importlist/action/testAction", body: `{}`,
			call: func(c *whisparr.Client) error {
				return c.ImportListAction(context.Background(), "testAction", provider)
			},
		},
		{
			name: "indexer action", method: http.MethodPost, uri: "/api/v3/indexer/action/testAction", body: `{}`,
			call: func(c *whisparr.Client) error { return c.IndexerAction(context.Background(), "testAction", provider) },
		},
		{
			name: "metadata action", method: http.MethodPost, uri: "/api/v3/metadata/action/testAction", body: `{}`,
			call: func(c *whisparr.Client) error { return c.MetadataAction(context.Background(), "testAction", provider) },
		},
		{
			name: "notification action", method: http.MethodPost, uri: "/api/v3/notification/action/testAction", body: `{}`,
			call: func(c *whisparr.Client) error {
				return c.NotificationAction(context.Background(), "testAction", provider)
			},
		},
		{
			name: "naming examples", method: http.MethodGet, uri: "/api/v3/config/naming/examples", body: `{"series":"Example"}`,
			call: func(c *whisparr.Client) error { _, err := c.GetNamingExamples(context.Background()); return err },
		},
		{
			name: "bulk update import lists", method: http.MethodPut, uri: "/api/v3/importlist/bulk", body: `[{"id":1}]`,
			call: func(c *whisparr.Client) error {
				_, err := c.BulkUpdateImportLists(context.Background(), bulk)
				return err
			},
		},
		{
			name: "bulk delete import lists", method: http.MethodDelete, uri: "/api/v3/importlist/bulk", body: `{}`,
			call: func(c *whisparr.Client) error { return c.BulkDeleteImportLists(context.Background(), []int{1}) },
		},
		{
			name: "bulk update indexers", method: http.MethodPut, uri: "/api/v3/indexer/bulk", body: `[{"id":1}]`,
			call: func(c *whisparr.Client) error { _, err := c.BulkUpdateIndexers(context.Background(), bulk); return err },
		},
		{
			name: "bulk delete indexers", method: http.MethodDelete, uri: "/api/v3/indexer/bulk", body: `{}`,
			call: func(c *whisparr.Client) error { return c.BulkDeleteIndexers(context.Background(), []int{1}) },
		},
		{
			name: "language profile schema", method: http.MethodGet, uri: "/api/v3/languageprofile/schema", body: `{"id":1,"name":"English"}`,
			call: func(c *whisparr.Client) error { _, err := c.GetLanguageProfileSchema(context.Background()); return err },
		},
		{
			name: "localization language", method: http.MethodGet, uri: "/api/v3/localization/language", body: `{"identifier":"en"}`,
			call: func(c *whisparr.Client) error { _, err := c.GetLocalizationLanguage(context.Background()); return err },
		},
		{
			name: "update log files", method: http.MethodGet, uri: "/api/v3/log/file/update", body: `[{"filename":"update.txt"}]`,
			call: func(c *whisparr.Client) error { _, err := c.GetUpdateLogFiles(context.Background()); return err },
		},
		{
			name: "update metadata", method: http.MethodPut, uri: "/api/v3/metadata/1", body: `{"id":1,"name":"metadata"}`,
			call: func(c *whisparr.Client) error { _, err := c.UpdateMetadata(context.Background(), provider); return err },
		},
		{
			name: "metadata schema", method: http.MethodGet, uri: "/api/v3/metadata/schema", body: `[{"id":1}]`,
			call: func(c *whisparr.Client) error { _, err := c.GetMetadataSchema(context.Background()); return err },
		},
		{
			name: "test metadata", method: http.MethodPost, uri: "/api/v3/metadata/test", body: `{}`,
			call: func(c *whisparr.Client) error { return c.TestMetadata(context.Background(), provider) },
		},
		{
			name: "test all metadata", method: http.MethodPost, uri: "/api/v3/metadata/testall", body: `{}`,
			call: func(c *whisparr.Client) error { return c.TestAllMetadata(context.Background()) },
		},
		{
			name: "update notification", method: http.MethodPut, uri: "/api/v3/notification/1", body: `{"id":1,"name":"notify"}`,
			call: func(c *whisparr.Client) error {
				_, err := c.UpdateNotification(context.Background(), provider)
				return err
			},
		},
		{
			name: "notification schema", method: http.MethodGet, uri: "/api/v3/notification/schema", body: `[{"id":1}]`,
			call: func(c *whisparr.Client) error { _, err := c.GetNotificationSchema(context.Background()); return err },
		},
		{
			name: "test notification", method: http.MethodPost, uri: "/api/v3/notification/test", body: `{}`,
			call: func(c *whisparr.Client) error { return c.TestNotification(context.Background(), provider) },
		},
		{
			name: "test all notifications", method: http.MethodPost, uri: "/api/v3/notification/testall", body: `{}`,
			call: func(c *whisparr.Client) error { return c.TestAllNotifications(context.Background()) },
		},
		{
			name: "bulk update quality definitions", method: http.MethodPut, uri: "/api/v3/qualitydefinition/update", body: `[{"id":1}]`,
			call: func(c *whisparr.Client) error {
				_, err := c.BulkUpdateQualityDefinitions(context.Background(), []arr.QualityDefinitionResource{{ID: 1}})
				return err
			},
		},
		{
			name: "quality definition limits", method: http.MethodGet, uri: "/api/v3/qualitydefinition/limits", body: `{"min":0,"max":100}`,
			call: func(c *whisparr.Client) error {
				_, err := c.GetQualityDefinitionLimits(context.Background())
				return err
			},
		},
		{
			name: "update quality profile", method: http.MethodPut, uri: "/api/v3/qualityprofile/1", body: `{"id":1,"name":"Any"}`,
			call: func(c *whisparr.Client) error {
				_, err := c.UpdateQualityProfile(context.Background(), &arr.QualityProfile{ID: 1})
				return err
			},
		},
		{
			name: "quality profile schema", method: http.MethodGet, uri: "/api/v3/qualityprofile/schema", body: `{"id":1,"name":"Any"}`,
			call: func(c *whisparr.Client) error { _, err := c.GetQualityProfileSchema(context.Background()); return err },
		},
		{
			name: "bulk delete queue", method: http.MethodDelete, uri: "/api/v3/queue/bulk", body: `{}`,
			call: func(c *whisparr.Client) error {
				return c.BulkDeleteQueue(context.Background(), &arr.QueueBulkResource{IDs: []int{1}})
			},
		},
		{
			name: "bulk grab queue", method: http.MethodPost, uri: "/api/v3/queue/grab/bulk", body: `{}`,
			call: func(c *whisparr.Client) error { return c.BulkGrabQueue(context.Background(), []int{1}) },
		},
		{
			name: "push release", method: http.MethodPost, uri: "/api/v3/release/push", body: `[{"guid":"guid-1"}]`,
			call: func(c *whisparr.Client) error {
				_, err := c.PushRelease(context.Background(), &arr.ReleasePushResource{Title: "Release", DownloadURL: "https://example.test/file"})
				return err
			},
		},
		{
			name: "update release profile", method: http.MethodPut, uri: "/api/v3/releaseprofile/1", body: `{"id":1,"name":"Profile"}`,
			call: func(c *whisparr.Client) error {
				_, err := c.UpdateReleaseProfile(context.Background(), &arr.ReleaseProfileResource{ID: 1})
				return err
			},
		},
		{
			name: "update remote path mapping", method: http.MethodPut, uri: "/api/v3/remotepathmapping/1", body: `{"id":1}`,
			call: func(c *whisparr.Client) error {
				_, err := c.UpdateRemotePathMapping(context.Background(), &arr.RemotePathMappingResource{ID: 1})
				return err
			},
		},
		{
			name: "process manual import", method: http.MethodPost, uri: "/api/v3/manualimport", body: `{}`,
			call: func(c *whisparr.Client) error {
				return c.ProcessManualImport(context.Background(), []arr.ManualImportReprocessResource{{}})
			},
		},
		{
			name: "filesystem type", method: http.MethodGet, uri: "/api/v3/filesystem/type?path=%2Fdata", body: `"folder"`,
			call: func(c *whisparr.Client) error {
				_, err := c.GetFileSystemType(context.Background(), "/data")
				return err
			},
		},
		{
			name: "filesystem media files", method: http.MethodGet, uri: "/api/v3/filesystem/mediafiles?path=%2Fdata", body: `[{"path":"/data/file.mkv"}]`,
			call: func(c *whisparr.Client) error {
				_, err := c.GetFileSystemMediaFiles(context.Background(), "/data")
				return err
			},
		},
		{
			name: "duplicate routes", method: http.MethodGet, uri: "/api/v3/system/routes/duplicate", body: `[]`,
			call: func(c *whisparr.Client) error { _, err := c.GetSystemRoutesDuplicate(context.Background()); return err },
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			srv := newV3JSONEndpointServer(t, tc.method, tc.uri, tc.body)
			defer srv.Close()
			c, err := whisparr.New(srv.URL, "test-key")
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			if err := tc.call(c); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
		})
	}
}

func TestV3RemainingEndpointWrappers(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		method string
		uri    string
		body   string
		call   func(*whisparr.ClientV3) error
	}{
		{
			name: "test all import lists", method: http.MethodPost, uri: "/api/v3/importlist/testall", body: `{}`,
			call: func(c *whisparr.ClientV3) error { return c.TestAllImportLists(context.Background()) },
		},
		{
			name: "test all indexers", method: http.MethodPost, uri: "/api/v3/indexer/testall", body: `{}`,
			call: func(c *whisparr.ClientV3) error { return c.TestAllIndexers(context.Background()) },
		},
		{
			name: "localization language", method: http.MethodGet, uri: "/api/v3/localization/language", body: `{"identifier":"en"}`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.GetLocalizationLanguage(context.Background())
				return err
			},
		},
		{
			name: "update log files", method: http.MethodGet, uri: "/api/v3/log/file/update", body: `[{"filename":"update.txt"}]`,
			call: func(c *whisparr.ClientV3) error { _, err := c.GetUpdateLogFiles(context.Background()); return err },
		},
		{
			name: "quality definition", method: http.MethodGet, uri: "/api/v3/qualitydefinition/1", body: `{"id":1}`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.GetQualityDefinition(context.Background(), 1)
				return err
			},
		},
		{
			name: "update quality definition", method: http.MethodPut, uri: "/api/v3/qualitydefinition/1", body: `{"id":1}`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.UpdateQualityDefinition(context.Background(), &arr.QualityDefinitionResource{ID: 1})
				return err
			},
		},
		{
			name: "remote path mapping", method: http.MethodGet, uri: "/api/v3/remotepathmapping/1", body: `{"id":1}`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.GetRemotePathMapping(context.Background(), 1)
				return err
			},
		},
		{
			name: "update remote path mapping", method: http.MethodPut, uri: "/api/v3/remotepathmapping/1", body: `{"id":1}`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.UpdateRemotePathMapping(context.Background(), &arr.RemotePathMappingResource{ID: 1})
				return err
			},
		},
		{
			name: "restore backup", method: http.MethodPost, uri: "/api/v3/system/backup/restore/1", body: `{}`,
			call: func(c *whisparr.ClientV3) error { return c.RestoreBackup(context.Background(), 1) },
		},
		{
			name: "filesystem type", method: http.MethodGet, uri: "/api/v3/filesystem/type?path=%2Fdata", body: `"folder"`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.GetFileSystemType(context.Background(), "/data")
				return err
			},
		},
		{
			name: "filesystem media files", method: http.MethodGet, uri: "/api/v3/filesystem/mediafiles?path=%2Fdata", body: `[{"path":"/data/file.mkv"}]`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.GetFileSystemMediaFiles(context.Background(), "/data")
				return err
			},
		},
		{
			name: "duplicate routes", method: http.MethodGet, uri: "/api/v3/system/routes/duplicate", body: `[]`,
			call: func(c *whisparr.ClientV3) error {
				_, err := c.GetSystemRoutesDuplicate(context.Background())
				return err
			},
		},
		{
			name: "task", method: http.MethodGet, uri: "/api/v3/system/task/1", body: `{"id":1,"name":"Refresh"}`,
			call: func(c *whisparr.ClientV3) error { _, err := c.GetTask(context.Background(), 1); return err },
		},
		{
			name: "tag detail", method: http.MethodGet, uri: "/api/v3/tag/detail/1", body: `{"id":1,"label":"tag"}`,
			call: func(c *whisparr.ClientV3) error { _, err := c.GetTagDetail(context.Background(), 1); return err },
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			srv := newV3JSONEndpointServer(t, tc.method, tc.uri, tc.body)
			defer srv.Close()
			c, err := whisparr.NewV3(srv.URL, "test-key")
			if err != nil {
				t.Fatalf("NewV3: %v", err)
			}
			if err := tc.call(c); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
		})
	}
}
