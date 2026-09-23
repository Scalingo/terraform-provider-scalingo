package scalingo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	goscalingo "github.com/Scalingo/go-scalingo/v11"
)

func TestReadDatabaseEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("request method = %s, want %s", r.Method, http.MethodGet)
		}
		if r.URL.Path != "/v1/databases/database-id/endpoints" {
			t.Errorf("request path = %s, want /v1/databases/database-id/endpoints", r.URL.Path)
		}
		if got := r.URL.Query().Get("include_default_credentials"); got != "true" {
			t.Errorf("include_default_credentials = %q, want true", got)
		}
		if _, ok := r.URL.Query()["type"]; ok {
			t.Error("request unexpectedly included the endpoint type filter")
		}

		response := goscalingo.DatabaseEndpointsResponse{
			Endpoints: []goscalingo.DatabaseEndpoint{
				{
					ID:         "private-endpoint-id",
					DatabaseID: "database-id",
					Hostname:   "private.example.com",
					Port:       5433,
					Type:       goscalingo.DatabaseEndpointTypePrivatePeeringRW,
				},
				{
					ID:         "public-endpoint-id",
					DatabaseID: "database-id",
					Hostname:   "public.example.com",
					Port:       5432,
					Type:       goscalingo.DatabaseEndpointTypePublicRW,
					Credentials: &goscalingo.DatabaseEndpointCredentials{
						Username: "user",
						Password: "password",
					},
				},
			},
		}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer server.Close()

	client, err := goscalingo.New(t.Context(), goscalingo.ClientConfig{APIEndpoint: server.URL})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	data := schema.TestResourceDataRaw(t, dataSourceScDatabaseEndpoints().Schema, map[string]interface{}{
		"database_id":                 "database-id",
		"type":                        "public-rw",
		"include_default_credentials": true,
	})
	diagnostics := dataSourceScDatabaseEndpointsRead(t.Context(), data, client)
	err = DiagnosticError(diagnostics)
	if err != nil {
		t.Fatalf("read database endpoint: %v", err)
	}

	if data.Id() != "database-id" {
		t.Errorf("data source ID = %q, want database-id", data.Id())
	}
	endpoints, ok := data.Get("endpoints").([]interface{})
	if !ok {
		t.Fatalf("endpoints has type %T, want []interface{}", data.Get("endpoints"))
	}
	if len(endpoints) != 1 {
		t.Fatalf("endpoints count = %d, want 1", len(endpoints))
	}
	endpoint, ok := endpoints[0].(map[string]interface{})
	if !ok {
		t.Fatalf("endpoint has type %T, want map[string]interface{}", endpoints[0])
	}
	for field, want := range map[string]interface{}{
		"id":       "public-endpoint-id",
		"hostname": "public.example.com",
		"port":     5432,
		"username": "user",
		"password": "password",
	} {
		if got := endpoint[field]; got != want {
			t.Errorf("endpoint.%s = %#v, want %#v", field, got, want)
		}
	}
}

func TestReadDatabaseEndpointsReturnsMultipleMatches(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := goscalingo.DatabaseEndpointsResponse{
			Endpoints: []goscalingo.DatabaseEndpoint{
				{ID: "public-endpoint-id-1", Type: goscalingo.DatabaseEndpointTypePublicRW},
				{ID: "public-endpoint-id-2", Type: goscalingo.DatabaseEndpointTypePublicRW},
			},
		}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer server.Close()

	client, err := goscalingo.New(t.Context(), goscalingo.ClientConfig{APIEndpoint: server.URL})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	for _, endpointType := range []string{"", "public-rw"} {
		config := map[string]interface{}{"database_id": "database-id"}
		if endpointType != "" {
			config["type"] = endpointType
		}
		data := schema.TestResourceDataRaw(t, dataSourceScDatabaseEndpoints().Schema, config)
		diagnostics := dataSourceScDatabaseEndpointsRead(t.Context(), data, client)
		err := DiagnosticError(diagnostics)
		if err != nil {
			t.Fatalf("read database endpoints with type %q: %v", endpointType, err)
		}

		if data.Id() != "database-id" {
			t.Errorf("data source ID = %q, want database-id", data.Id())
		}
		endpoints, ok := data.Get("endpoints").([]interface{})
		if !ok {
			t.Fatalf("endpoints has type %T, want []interface{}", data.Get("endpoints"))
		}
		if len(endpoints) != 2 {
			t.Errorf("endpoints count with type %q = %d, want 2", endpointType, len(endpoints))
		}
	}
}

func TestReadDatabaseEndpointWithoutType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := goscalingo.DatabaseEndpointsResponse{
			Endpoints: []goscalingo.DatabaseEndpoint{
				{
					ID:       "endpoint-id",
					Hostname: "database.example.com",
					Port:     5432,
					Type:     goscalingo.DatabaseEndpointTypePublicRW,
				},
			},
		}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer server.Close()

	client, err := goscalingo.New(t.Context(), goscalingo.ClientConfig{APIEndpoint: server.URL})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	data := schema.TestResourceDataRaw(t, dataSourceScDatabaseEndpoints().Schema, map[string]interface{}{
		"database_id": "database-id",
	})
	diagnostics := dataSourceScDatabaseEndpointsRead(t.Context(), data, client)
	err = DiagnosticError(diagnostics)
	if err != nil {
		t.Fatalf("read database endpoint: %v", err)
	}

	if data.Id() != "database-id" {
		t.Errorf("data source ID = %q, want database-id", data.Id())
	}
	endpoints, ok := data.Get("endpoints").([]interface{})
	if !ok {
		t.Fatalf("endpoints has type %T, want []interface{}", data.Get("endpoints"))
	}
	if len(endpoints) != 1 {
		t.Fatalf("endpoints count = %d, want 1", len(endpoints))
	}
	endpoint, ok := endpoints[0].(map[string]interface{})
	if !ok {
		t.Fatalf("endpoint has type %T, want map[string]interface{}", endpoints[0])
	}
	if got := endpoint["type"]; got != string(goscalingo.DatabaseEndpointTypePublicRW) {
		t.Errorf("endpoint type = %#v, want %q", got, goscalingo.DatabaseEndpointTypePublicRW)
	}
}
