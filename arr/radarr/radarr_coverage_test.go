package radarr_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/golusoris/goenvoy/arr/radarr"
	arr "github.com/golusoris/goenvoy/arr/v2"
)

// TestAutoTagging covers GetAutoTagging, GetAutoTag, CreateAutoTag, UpdateAutoTag, DeleteAutoTag, GetAutoTagSchema.
func TestAutoTagging(t *testing.T) {
	t.Parallel()

	t.Run("GetAutoTagging", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/autotagging", []arr.AutoTaggingResource{{ID: 1}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetAutoTagging(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})

	t.Run("GetAutoTag", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/autotagging/1", arr.AutoTaggingResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetAutoTag(context.Background(), 1)
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("CreateAutoTag", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPost, "/api/v3/autotagging", arr.AutoTaggingResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.CreateAutoTag(context.Background(), &arr.AutoTaggingResource{})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("UpdateAutoTag", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPut, "/api/v3/autotagging/1", arr.AutoTaggingResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.UpdateAutoTag(context.Background(), &arr.AutoTaggingResource{ID: 1})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("DeleteAutoTag", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodDelete, "/api/v3/autotagging/1", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.DeleteAutoTag(context.Background(), 1); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("GetAutoTagSchema", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/autotagging/schema", []arr.AutoTaggingSpecification{{Name: "x"}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetAutoTagSchema(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})
}

// TestBlocklist covers GetBlocklist, DeleteBlocklistItem, DeleteBlocklistBulk.
func TestBlocklist(t *testing.T) {
	t.Parallel()

	t.Run("GetBlocklist", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/blocklist?page=1&pageSize=10",
			arr.PagingResource[arr.BlocklistResource]{TotalRecords: 1, Records: []arr.BlocklistResource{{ID: 1}}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetBlocklist(context.Background(), 1, 10)
		if err != nil {
			t.Fatal(err)
		}
		if got.TotalRecords != 1 {
			t.Errorf("TotalRecords = %d", got.TotalRecords)
		}
	})

	t.Run("DeleteBlocklistItem", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodDelete, "/api/v3/blocklist/5", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.DeleteBlocklistItem(context.Background(), 5); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("DeleteBlocklistBulk", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodDelete, "/api/v3/blocklist/bulk", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.DeleteBlocklistBulk(context.Background(), []int{1, 2}); err != nil {
			t.Fatal(err)
		}
	})
}

// TestCustomFilters covers GetCustomFilters, GetCustomFilter, CreateCustomFilter, UpdateCustomFilter, DeleteCustomFilter.
func TestCustomFilters(t *testing.T) {
	t.Parallel()

	t.Run("GetCustomFilters", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/customfilter", []arr.CustomFilterResource{{ID: 1}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetCustomFilters(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})

	t.Run("GetCustomFilter", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/customfilter/1", arr.CustomFilterResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetCustomFilter(context.Background(), 1)
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("CreateCustomFilter", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPost, "/api/v3/customfilter", arr.CustomFilterResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.CreateCustomFilter(context.Background(), &arr.CustomFilterResource{})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("UpdateCustomFilter", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPut, "/api/v3/customfilter/1", arr.CustomFilterResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.UpdateCustomFilter(context.Background(), &arr.CustomFilterResource{ID: 1})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("DeleteCustomFilter", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodDelete, "/api/v3/customfilter/1", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.DeleteCustomFilter(context.Background(), 1); err != nil {
			t.Fatal(err)
		}
	})
}

// TestCustomFormats covers GetCustomFormats, GetCustomFormat, CreateCustomFormat, UpdateCustomFormat, DeleteCustomFormat, GetCustomFormatSchema.
func TestCustomFormats(t *testing.T) {
	t.Parallel()

	t.Run("GetCustomFormats", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/customformat", []arr.CustomFormatResource{{ID: 1, Name: "x"}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetCustomFormats(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})

	t.Run("GetCustomFormat", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/customformat/1", arr.CustomFormatResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetCustomFormat(context.Background(), 1)
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("CreateCustomFormat", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPost, "/api/v3/customformat", arr.CustomFormatResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.CreateCustomFormat(context.Background(), &arr.CustomFormatResource{})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("UpdateCustomFormat", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPut, "/api/v3/customformat/1", arr.CustomFormatResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.UpdateCustomFormat(context.Background(), &arr.CustomFormatResource{ID: 1})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("DeleteCustomFormat", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodDelete, "/api/v3/customformat/1", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.DeleteCustomFormat(context.Background(), 1); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("GetCustomFormatSchema", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/customformat/schema", []arr.CustomFormatSpecification{{Name: "x"}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetCustomFormatSchema(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})
}

// TestDelayProfiles covers GetDelayProfiles, GetDelayProfile, CreateDelayProfile, UpdateDelayProfile, DeleteDelayProfile, ReorderDelayProfile.
func TestDelayProfiles(t *testing.T) {
	t.Parallel()

	t.Run("GetDelayProfiles", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/delayprofile", []arr.DelayProfileResource{{ID: 1}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetDelayProfiles(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})

	t.Run("GetDelayProfile", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/delayprofile/1", arr.DelayProfileResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetDelayProfile(context.Background(), 1)
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("CreateDelayProfile", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPost, "/api/v3/delayprofile", arr.DelayProfileResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.CreateDelayProfile(context.Background(), &arr.DelayProfileResource{})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("UpdateDelayProfile", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPut, "/api/v3/delayprofile/1", arr.DelayProfileResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.UpdateDelayProfile(context.Background(), &arr.DelayProfileResource{ID: 1})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("DeleteDelayProfile", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodDelete, "/api/v3/delayprofile/1", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.DeleteDelayProfile(context.Background(), 1); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ReorderDelayProfile", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPut, "/api/v3/delayprofile/reorder/1?after=2", []arr.DelayProfileResource{{ID: 1}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.ReorderDelayProfile(context.Background(), 1, 2)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})
}

// TestDownloadClients covers GetDownloadClients, GetDownloadClient, CreateDownloadClient, UpdateDownloadClient, DeleteDownloadClient, GetDownloadClientSchema, TestDownloadClient.
func TestDownloadClients(t *testing.T) {
	t.Parallel()

	t.Run("GetDownloadClients", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/downloadclient", []arr.ProviderResource{{ID: 1}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetDownloadClients(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})

	t.Run("GetDownloadClient", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/downloadclient/1", arr.ProviderResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetDownloadClient(context.Background(), 1)
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("CreateDownloadClient", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPost, "/api/v3/downloadclient", arr.ProviderResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.CreateDownloadClient(context.Background(), &arr.ProviderResource{})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("UpdateDownloadClient", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPut, "/api/v3/downloadclient/1", arr.ProviderResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.UpdateDownloadClient(context.Background(), &arr.ProviderResource{ID: 1})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("DeleteDownloadClient", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodDelete, "/api/v3/downloadclient/1", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.DeleteDownloadClient(context.Background(), 1); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("GetDownloadClientSchema", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/downloadclient/schema", []arr.ProviderResource{{ID: 1}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetDownloadClientSchema(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})

	t.Run("TestDownloadClient", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPost, "/api/v3/downloadclient/test", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.TestDownloadClient(context.Background(), &arr.ProviderResource{ID: 1}); err != nil {
			t.Fatal(err)
		}
	})
}

// TestNotifications covers GetNotifications, GetNotification, CreateNotification, UpdateNotification, DeleteNotification, GetNotificationSchema, TestNotification.
func TestNotifications(t *testing.T) {
	t.Parallel()

	t.Run("GetNotifications", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/notification", []arr.ProviderResource{{ID: 1}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetNotifications(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})

	t.Run("GetNotification", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/notification/1", arr.ProviderResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetNotification(context.Background(), 1)
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("CreateNotification", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPost, "/api/v3/notification", arr.ProviderResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.CreateNotification(context.Background(), &arr.ProviderResource{})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("UpdateNotification", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPut, "/api/v3/notification/1", arr.ProviderResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.UpdateNotification(context.Background(), &arr.ProviderResource{ID: 1})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("DeleteNotification", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodDelete, "/api/v3/notification/1", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.DeleteNotification(context.Background(), 1); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("GetNotificationSchema", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/notification/schema", []arr.ProviderResource{{ID: 1}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetNotificationSchema(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})

	t.Run("TestNotification", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPost, "/api/v3/notification/test", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.TestNotification(context.Background(), &arr.ProviderResource{ID: 1}); err != nil {
			t.Fatal(err)
		}
	})
}

// TestMetadataConsumers covers GetMetadataConsumers, GetMetadataConsumer, CreateMetadataConsumer, UpdateMetadataConsumer, DeleteMetadataConsumer, GetMetadataSchema, TestMetadataConsumer.
func TestMetadataConsumers(t *testing.T) {
	t.Parallel()

	t.Run("GetMetadataConsumers", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/metadata", []arr.ProviderResource{{ID: 1}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetMetadataConsumers(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})

	t.Run("GetMetadataConsumer", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/metadata/1", arr.ProviderResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetMetadataConsumer(context.Background(), 1)
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("CreateMetadataConsumer", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPost, "/api/v3/metadata", arr.ProviderResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.CreateMetadataConsumer(context.Background(), &arr.ProviderResource{})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("UpdateMetadataConsumer", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPut, "/api/v3/metadata/1", arr.ProviderResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.UpdateMetadataConsumer(context.Background(), &arr.ProviderResource{ID: 1})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("DeleteMetadataConsumer", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodDelete, "/api/v3/metadata/1", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.DeleteMetadataConsumer(context.Background(), 1); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("GetMetadataSchema", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/metadata/schema", []arr.ProviderResource{{ID: 1}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetMetadataSchema(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})

	t.Run("TestMetadataConsumer", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPost, "/api/v3/metadata/test", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.TestMetadataConsumer(context.Background(), &arr.ProviderResource{ID: 1}); err != nil {
			t.Fatal(err)
		}
	})
}

// TestIndexers covers GetIndexers, GetIndexer, CreateIndexer, UpdateIndexer, DeleteIndexer, GetIndexerSchema, TestIndexer, GetIndexerFlags.
func TestIndexers(t *testing.T) {
	t.Parallel()

	t.Run("GetIndexers", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/indexer", []arr.ProviderResource{{ID: 1}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetIndexers(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})

	t.Run("GetIndexer", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/indexer/1", arr.ProviderResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetIndexer(context.Background(), 1)
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("CreateIndexer", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPost, "/api/v3/indexer", arr.ProviderResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.CreateIndexer(context.Background(), &arr.ProviderResource{})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("UpdateIndexer", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPut, "/api/v3/indexer/1", arr.ProviderResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.UpdateIndexer(context.Background(), &arr.ProviderResource{ID: 1})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("DeleteIndexer", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodDelete, "/api/v3/indexer/1", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.DeleteIndexer(context.Background(), 1); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("GetIndexerSchema", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/indexer/schema", []arr.ProviderResource{{ID: 1}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetIndexerSchema(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})

	t.Run("TestIndexer", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPost, "/api/v3/indexer/test", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.TestIndexer(context.Background(), &arr.ProviderResource{ID: 1}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("GetIndexerFlags", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/indexerflag", []arr.IndexerFlagResource{{ID: 1, Name: "test"}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetIndexerFlags(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})
}

// TestImportLists covers GetImportLists, GetImportList, CreateImportList, UpdateImportList, DeleteImportList, GetImportListSchema, TestImportList.
func TestImportLists(t *testing.T) {
	t.Parallel()

	t.Run("GetImportLists", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/importlist", []arr.ProviderResource{{ID: 1}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetImportLists(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})

	t.Run("GetImportList", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/importlist/1", arr.ProviderResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetImportList(context.Background(), 1)
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("CreateImportList", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPost, "/api/v3/importlist", arr.ProviderResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.CreateImportList(context.Background(), &arr.ProviderResource{})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("UpdateImportList", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPut, "/api/v3/importlist/1", arr.ProviderResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.UpdateImportList(context.Background(), &arr.ProviderResource{ID: 1})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("DeleteImportList", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodDelete, "/api/v3/importlist/1", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.DeleteImportList(context.Background(), 1); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("GetImportListSchema", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/importlist/schema", []arr.ProviderResource{{ID: 1}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetImportListSchema(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})

	t.Run("TestImportList", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPost, "/api/v3/importlist/test", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.TestImportList(context.Background(), &arr.ProviderResource{ID: 1}); err != nil {
			t.Fatal(err)
		}
	})
}

// TestImportListExclusionExtras covers CreateImportListExclusion, DeleteImportListExclusion.
func TestImportListExclusionExtras(t *testing.T) {
	t.Parallel()

	t.Run("CreateImportListExclusion", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPost, "/api/v3/exclusions", radarr.ImportListExclusion{ID: 1, TmdbID: 550})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.CreateImportListExclusion(context.Background(), &radarr.ImportListExclusion{TmdbID: 550})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("DeleteImportListExclusion", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodDelete, "/api/v3/exclusions/1", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.DeleteImportListExclusion(context.Background(), 1); err != nil {
			t.Fatal(err)
		}
	})
}

// TestQualityProfiles covers GetQualityProfile, CreateQualityProfile, UpdateQualityProfile, DeleteQualityProfile.
func TestQualityProfiles(t *testing.T) {
	t.Parallel()

	t.Run("GetQualityProfile", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/qualityprofile/1", arr.QualityProfile{ID: 1, Name: "Any"})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetQualityProfile(context.Background(), 1)
		if err != nil {
			t.Fatal(err)
		}
		if got.Name != "Any" {
			t.Errorf("Name = %q", got.Name)
		}
	})

	t.Run("CreateQualityProfile", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPost, "/api/v3/qualityprofile", arr.QualityProfile{ID: 1, Name: "HD"})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.CreateQualityProfile(context.Background(), &arr.QualityProfile{Name: "HD"})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("UpdateQualityProfile", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPut, "/api/v3/qualityprofile/1", arr.QualityProfile{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.UpdateQualityProfile(context.Background(), &arr.QualityProfile{ID: 1})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("DeleteQualityProfile", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodDelete, "/api/v3/qualityprofile/1", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.DeleteQualityProfile(context.Background(), 1); err != nil {
			t.Fatal(err)
		}
	})
}

// TestQualityDefinitions covers GetQualityDefinitions, GetQualityDefinition, UpdateQualityDefinition.
func TestQualityDefinitions(t *testing.T) {
	t.Parallel()

	t.Run("GetQualityDefinitions", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/qualitydefinition", []arr.QualityDefinitionResource{{ID: 1}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetQualityDefinitions(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})

	t.Run("GetQualityDefinition", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/qualitydefinition/1", arr.QualityDefinitionResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetQualityDefinition(context.Background(), 1)
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("UpdateQualityDefinition", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPut, "/api/v3/qualitydefinition/1", arr.QualityDefinitionResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.UpdateQualityDefinition(context.Background(), &arr.QualityDefinitionResource{ID: 1})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})
}

// TestReleaseProfiles covers GetReleaseProfiles, GetReleaseProfile, CreateReleaseProfile, UpdateReleaseProfile, DeleteReleaseProfile.
func TestReleaseProfiles(t *testing.T) {
	t.Parallel()

	t.Run("GetReleaseProfiles", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/releaseprofile", []arr.ReleaseProfileResource{{ID: 1}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetReleaseProfiles(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})

	t.Run("GetReleaseProfile", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/releaseprofile/1", arr.ReleaseProfileResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetReleaseProfile(context.Background(), 1)
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("CreateReleaseProfile", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPost, "/api/v3/releaseprofile", arr.ReleaseProfileResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.CreateReleaseProfile(context.Background(), &arr.ReleaseProfileResource{})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("UpdateReleaseProfile", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPut, "/api/v3/releaseprofile/1", arr.ReleaseProfileResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.UpdateReleaseProfile(context.Background(), &arr.ReleaseProfileResource{ID: 1})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("DeleteReleaseProfile", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodDelete, "/api/v3/releaseprofile/1", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.DeleteReleaseProfile(context.Background(), 1); err != nil {
			t.Fatal(err)
		}
	})
}

// TestRemotePathMappings covers GetRemotePathMappings, GetRemotePathMapping, CreateRemotePathMapping, UpdateRemotePathMapping, DeleteRemotePathMapping.
func TestRemotePathMappings(t *testing.T) {
	t.Parallel()

	t.Run("GetRemotePathMappings", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/remotepathmapping", []arr.RemotePathMappingResource{{ID: 1}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetRemotePathMappings(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})

	t.Run("GetRemotePathMapping", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/remotepathmapping/1", arr.RemotePathMappingResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetRemotePathMapping(context.Background(), 1)
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("CreateRemotePathMapping", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPost, "/api/v3/remotepathmapping", arr.RemotePathMappingResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.CreateRemotePathMapping(context.Background(), &arr.RemotePathMappingResource{})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("UpdateRemotePathMapping", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPut, "/api/v3/remotepathmapping/1", arr.RemotePathMappingResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.UpdateRemotePathMapping(context.Background(), &arr.RemotePathMappingResource{ID: 1})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("DeleteRemotePathMapping", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodDelete, "/api/v3/remotepathmapping/1", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.DeleteRemotePathMapping(context.Background(), 1); err != nil {
			t.Fatal(err)
		}
	})
}

// TestRootFolderExtras covers GetRootFolder, CreateRootFolder, DeleteRootFolder.
func TestRootFolderExtras(t *testing.T) {
	t.Parallel()

	t.Run("GetRootFolder", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/rootfolder/1", arr.RootFolder{ID: 1, Path: "/movies"})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetRootFolder(context.Background(), 1)
		if err != nil {
			t.Fatal(err)
		}
		if got.Path != "/movies" {
			t.Errorf("Path = %q", got.Path)
		}
	})

	t.Run("CreateRootFolder", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPost, "/api/v3/rootfolder", arr.RootFolder{ID: 1, Path: "/movies"})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.CreateRootFolder(context.Background(), &arr.RootFolder{Path: "/movies"})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("DeleteRootFolder", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodDelete, "/api/v3/rootfolder/1", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.DeleteRootFolder(context.Background(), 1); err != nil {
			t.Fatal(err)
		}
	})
}

// TestTagExtras covers GetTag, UpdateTag, DeleteTag, GetTagDetails, GetTagDetail.
func TestTagExtras(t *testing.T) {
	t.Parallel()

	t.Run("GetTag", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/tag/1", arr.Tag{ID: 1, Label: "4k"})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetTag(context.Background(), 1)
		if err != nil {
			t.Fatal(err)
		}
		if got.Label != "4k" {
			t.Errorf("Label = %q", got.Label)
		}
	})

	t.Run("UpdateTag", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPut, "/api/v3/tag/1", arr.Tag{ID: 1, Label: "updated"})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.UpdateTag(context.Background(), &arr.Tag{ID: 1, Label: "updated"})
		if err != nil {
			t.Fatal(err)
		}
		if got.Label != "updated" {
			t.Errorf("Label = %q", got.Label)
		}
	})

	t.Run("DeleteTag", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodDelete, "/api/v3/tag/1", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.DeleteTag(context.Background(), 1); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("GetTagDetails", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/tag/detail", []arr.TagDetail{{ID: 1, Label: "4k"}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetTagDetails(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})

	t.Run("GetTagDetail", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/tag/detail/1", arr.TagDetail{ID: 1, Label: "4k"})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetTagDetail(context.Background(), 1)
		if err != nil {
			t.Fatal(err)
		}
		if got.Label != "4k" {
			t.Errorf("Label = %q", got.Label)
		}
	})
}

// TestConfigs covers DownloadClientConfig, IndexerConfig, NamingConfig, HostConfig, UIConfig, MediaManagementConfig.
func TestConfigs(t *testing.T) {
	t.Parallel()

	t.Run("GetDownloadClientConfig", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/config/downloadclient", arr.DownloadClientConfigResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetDownloadClientConfig(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("UpdateDownloadClientConfig", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPut, "/api/v3/config/downloadclient/1", arr.DownloadClientConfigResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.UpdateDownloadClientConfig(context.Background(), &arr.DownloadClientConfigResource{ID: 1})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("GetIndexerConfig", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/config/indexer", arr.IndexerConfigResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetIndexerConfig(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("UpdateIndexerConfig", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPut, "/api/v3/config/indexer/1", arr.IndexerConfigResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.UpdateIndexerConfig(context.Background(), &arr.IndexerConfigResource{ID: 1})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("GetNamingConfig", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/config/naming", arr.NamingConfigResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetNamingConfig(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("UpdateNamingConfig", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPut, "/api/v3/config/naming/1", arr.NamingConfigResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.UpdateNamingConfig(context.Background(), &arr.NamingConfigResource{ID: 1})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("GetHostConfig", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/config/host", arr.HostConfigResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetHostConfig(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("UpdateHostConfig", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPut, "/api/v3/config/host/1", arr.HostConfigResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.UpdateHostConfig(context.Background(), &arr.HostConfigResource{ID: 1})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("GetUIConfig", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/config/ui", arr.UIConfigResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetUIConfig(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("UpdateUIConfig", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPut, "/api/v3/config/ui/1", arr.UIConfigResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.UpdateUIConfig(context.Background(), &arr.UIConfigResource{ID: 1})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("GetMediaManagementConfig", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/config/mediamanagement", arr.MediaManagementConfigResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetMediaManagementConfig(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("UpdateMediaManagementConfig", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPut, "/api/v3/config/mediamanagement/1", arr.MediaManagementConfigResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.UpdateMediaManagementConfig(context.Background(), &arr.MediaManagementConfigResource{ID: 1})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})
}

// TestQueueExtras covers DeleteQueueItems, GrabQueueItem, GrabQueueItems, GetQueueDetails, GetQueueStatus.
func TestQueueExtras(t *testing.T) {
	t.Parallel()

	t.Run("DeleteQueueItems", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodDelete, "/api/v3/queue/bulk?removeFromClient=true&blocklist=false", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.DeleteQueueItems(context.Background(), []int{1, 2}, true, false); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("GrabQueueItem", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPost, "/api/v3/queue/grab/1", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.GrabQueueItem(context.Background(), 1); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("GrabQueueItems", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPost, "/api/v3/queue/grab/bulk", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.GrabQueueItems(context.Background(), []int{1, 2}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("GetQueueDetails", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/queue/details", []arr.QueueRecord{{ID: 1}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetQueueDetails(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})

	t.Run("GetQueueStatus", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/queue/status", arr.QueueStatusResource{TotalCount: 5})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetQueueStatus(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if got.TotalCount != 5 {
			t.Errorf("TotalCount = %d", got.TotalCount)
		}
	})
}

// TestHistoryExtras covers GetHistoryMovie, GetHistorySince, MarkHistoryFailed.
func TestHistoryExtras(t *testing.T) {
	t.Parallel()

	t.Run("GetHistoryMovie", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/history/movie?movieId=1", []radarr.HistoryRecord{{ID: 1}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetHistoryMovie(context.Background(), 1)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})

	t.Run("GetHistorySince", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/history/since?date=2026-01-01", []radarr.HistoryRecord{{ID: 1}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetHistorySince(context.Background(), "2026-01-01")
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})

	t.Run("MarkHistoryFailed", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPost, "/api/v3/history/failed/1", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.MarkHistoryFailed(context.Background(), 1); err != nil {
			t.Fatal(err)
		}
	})
}

// TestLanguages covers GetLanguages.
func TestLanguages(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodGet, "/api/v3/language", []radarr.Language{{ID: 1, Name: "English"}})
	defer srv.Close()
	c, _ := radarr.New(srv.URL, "k")
	got, err := c.GetLanguages(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d", len(got))
	}
}

// TestSystemExtras covers GetBackups, DeleteBackup, RestoreBackup, GetSystemRoutes, Shutdown, Restart, GetTasks, GetTask, GetUpdates.
func TestSystemExtras(t *testing.T) {
	t.Parallel()

	t.Run("GetBackups", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/system/backup", []arr.Backup{{ID: 1}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetBackups(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})

	t.Run("DeleteBackup", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodDelete, "/api/v3/system/backup/1", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.DeleteBackup(context.Background(), 1); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("RestoreBackup", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPost, "/api/v3/system/backup/restore/1", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.RestoreBackup(context.Background(), 1); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("GetSystemRoutes", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/system/routes", []arr.SystemRouteResource{{Path: "/api/v3/movie"}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetSystemRoutes(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})

	t.Run("Shutdown", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPost, "/api/v3/system/shutdown", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.Shutdown(context.Background()); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Restart", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPost, "/api/v3/system/restart", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.Restart(context.Background()); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("GetTasks", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/system/task", []arr.TaskResource{{ID: 1, Name: "RefreshMovie"}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetTasks(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})

	t.Run("GetTask", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/system/task/1", arr.TaskResource{ID: 1, Name: "RefreshMovie"})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetTask(context.Background(), 1)
		if err != nil {
			t.Fatal(err)
		}
		if got.Name != "RefreshMovie" {
			t.Errorf("Name = %q", got.Name)
		}
	})

	t.Run("GetUpdates", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/update", []arr.UpdateResource{{Version: "5.0.0"}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetUpdates(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})
}

// TestFileSystemExtras covers GetFileSystemType, GetFileSystemMediaFiles.
func TestFileSystemExtras(t *testing.T) {
	t.Parallel()

	t.Run("GetFileSystemType", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/filesystem/type?path=%2Fdata", "local")
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetFileSystemType(context.Background(), "/data")
		if err != nil {
			t.Fatal(err)
		}
		if got != "local" {
			t.Errorf("type = %q", got)
		}
	})

	t.Run("GetFileSystemMediaFiles", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/filesystem/mediafiles?path=%2Fdata", []string{"movie.mkv"})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetFileSystemMediaFiles(context.Background(), "/data")
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})
}

// TestLogExtras covers GetLogs, GetLogFiles, GetUpdateLogFiles.
func TestLogExtras(t *testing.T) {
	t.Parallel()

	t.Run("GetLogs", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/log?page=1&pageSize=10",
			arr.PagingResource[arr.LogRecord]{TotalRecords: 1, Records: []arr.LogRecord{{ID: 1}}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetLogs(context.Background(), 1, 10)
		if err != nil {
			t.Fatal(err)
		}
		if got.TotalRecords != 1 {
			t.Errorf("TotalRecords = %d", got.TotalRecords)
		}
	})

	t.Run("GetLogFiles", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/log/file", []arr.LogFileResource{{ID: 1, Filename: "radarr.log"}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetLogFiles(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})

	t.Run("GetUpdateLogFiles", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/log/file/update", []arr.LogFileResource{{ID: 1}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetUpdateLogFiles(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})
}

// TestWanted covers GetWantedMissing, GetWantedCutoff.
func TestWanted(t *testing.T) {
	t.Parallel()

	t.Run("GetWantedMissing", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/wanted/missing?page=1&pageSize=10",
			arr.PagingResource[radarr.Movie]{TotalRecords: 1, Records: []radarr.Movie{{ID: 1}}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetWantedMissing(context.Background(), 1, 10)
		if err != nil {
			t.Fatal(err)
		}
		if got.TotalRecords != 1 {
			t.Errorf("TotalRecords = %d", got.TotalRecords)
		}
	})

	t.Run("GetWantedCutoff", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/wanted/cutoff?page=1&pageSize=10",
			arr.PagingResource[radarr.Movie]{TotalRecords: 2, Records: []radarr.Movie{{ID: 1}, {ID: 2}}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetWantedCutoff(context.Background(), 1, 10)
		if err != nil {
			t.Fatal(err)
		}
		if got.TotalRecords != 2 {
			t.Errorf("TotalRecords = %d", got.TotalRecords)
		}
	})
}

// TestRename covers GetRenameList.
func TestRename(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodGet, "/api/v3/rename?movieId=1", []arr.RenameMovieResource{{MovieID: 1}})
	defer srv.Close()
	c, _ := radarr.New(srv.URL, "k")
	got, err := c.GetRenameList(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d", len(got))
	}
}

// TestManualImport covers GetManualImport, ProcessManualImport.
func TestManualImport(t *testing.T) {
	t.Parallel()

	t.Run("GetManualImport", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/manualimport?folder=%2Fdata&downloadId=abc",
			[]arr.ManualImportResource{{ID: 1}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetManualImport(context.Background(), "/data", "abc")
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})

	t.Run("ProcessManualImport", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPost, "/api/v3/manualimport", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.ProcessManualImport(context.Background(), []arr.ManualImportReprocessResource{{ID: 1}}); err != nil {
			t.Fatal(err)
		}
	})
}

// TestReleases covers SearchReleases, PushRelease, GrabRelease.
func TestReleases(t *testing.T) {
	t.Parallel()

	t.Run("SearchReleases", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/release?movieId=1", []arr.ReleaseResource{{GUID: "abc"}})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.SearchReleases(context.Background(), 1)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}
	})

	t.Run("PushRelease", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPost, "/api/v3/release/push", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.PushRelease(context.Background(), &arr.ReleasePushResource{Title: "test"}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("GrabRelease", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPost, "/api/v3/release", nil)
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		if err := c.GrabRelease(context.Background(), "guid123", 1); err != nil {
			t.Fatal(err)
		}
	})
}

// TestCollectionExtras covers UpdateCollections.
func TestCollectionExtras(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodPut, "/api/v3/collection", nil)
	defer srv.Close()
	c, _ := radarr.New(srv.URL, "k")
	if err := c.UpdateCollections(context.Background(), []radarr.Collection{{ID: 1}}); err != nil {
		t.Fatal(err)
	}
}

// TestMovieExtras covers ImportMovies.
func TestMovieExtras(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodPost, "/api/v3/movie/import", []radarr.Movie{{ID: 1}})
	defer srv.Close()
	c, _ := radarr.New(srv.URL, "k")
	got, err := c.ImportMovies(context.Background(), []radarr.Movie{{Title: "test"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d", len(got))
	}
}

// TestMetadataConfigExtras covers GetMetadataConfig, UpdateMetadataConfig.
func TestMetadataConfigExtras(t *testing.T) {
	t.Parallel()

	t.Run("GetMetadataConfig", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodGet, "/api/v3/config/metadata", radarr.MetadataConfigResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.GetMetadataConfig(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})

	t.Run("UpdateMetadataConfig", func(t *testing.T) {
		t.Parallel()
		srv := newTestServer(t, http.MethodPut, "/api/v3/config/metadata/1", radarr.MetadataConfigResource{ID: 1})
		defer srv.Close()
		c, _ := radarr.New(srv.URL, "k")
		got, err := c.UpdateMetadataConfig(context.Background(), &radarr.MetadataConfigResource{ID: 1})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d", got.ID)
		}
	})
}

// TestUploadBackupCoverage covers UploadBackup.
func TestUploadBackupCoverage(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, http.MethodPost, "", nil)
	defer srv.Close()
	c, _ := radarr.New(srv.URL, "k")
	err := c.UploadBackup(context.Background(), "backup.zip", strings.NewReader("data"))
	if err != nil {
		t.Fatal(err)
	}
}
