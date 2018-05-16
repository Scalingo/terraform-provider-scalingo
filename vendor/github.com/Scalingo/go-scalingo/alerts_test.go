package scalingo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlertsClient(t *testing.T) {
	appName := "my-app"
	alertID := "my-id"

	tests := []struct {
		action           string
		testedClientCall func(c AlertsService) error
		expectedEndpoint string
		expectedMethod   string
		response         interface{}
		responseStatus   int
		noBody           bool
	}{
		{
			action: "list",
			testedClientCall: func(c AlertsService) error {
				_, err := c.AlertsList(appName)
				return err
			},
			expectedEndpoint: "/v1/apps/my-app/alerts",
			expectedMethod:   "GET",
			response:         AlertsRes{},
		},
		{
			action: "add",
			testedClientCall: func(c AlertsService) error {
				_, err := c.AlertAdd(appName, AlertAddParams{})
				return err
			},
			expectedEndpoint: "/v1/apps/my-app/alerts",
			expectedMethod:   "POST",
			response:         AlertsRes{},
			responseStatus:   201,
		},
		{
			action: "show",
			testedClientCall: func(c AlertsService) error {
				_, err := c.AlertShow(appName, alertID)
				return err
			},
			expectedEndpoint: "/v1/apps/my-app/alerts/my-id",
			expectedMethod:   "GET",
			response:         AlertsRes{},
		},
		{
			action: "update",
			testedClientCall: func(c AlertsService) error {
				_, err := c.AlertUpdate(appName, alertID, AlertUpdateParams{})
				return err
			},
			expectedEndpoint: "/v1/apps/my-app/alerts/my-id",
			expectedMethod:   "PATCH",
			response:         AlertsRes{},
		},
		{
			action: "remove",
			testedClientCall: func(c AlertsService) error {
				return c.AlertRemove(appName, alertID)
			},
			expectedEndpoint: "/v1/apps/my-app/alerts/my-id",
			expectedMethod:   "DELETE",
			responseStatus:   204,
		},
	}

	for _, test := range tests {
		for msg, run := range map[string]struct {
			invalidResponse bool
		}{
			"it should fail if it fails to " + test.action + "the subresource": {
				invalidResponse: true,
			},
			"it should succeed if it succeeds to " + test.action + " the subresource": {
				invalidResponse: false,
			},
		} {
			t.Run(msg, func(t *testing.T) {
				handler := func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, test.expectedMethod, r.Method)
					assert.Equal(t, test.expectedEndpoint, r.URL.Path)
					if run.invalidResponse {
						w.WriteHeader(500)
						w.Write([]byte("INVALID"))
					} else {
						if test.responseStatus != 0 {
							w.WriteHeader(test.responseStatus)
						}
						if test.response != nil {
							err := json.NewEncoder(w).Encode(&test.response)
							assert.NoError(t, err)
						}
					}
				}
				ts := httptest.NewServer(http.HandlerFunc(handler))
				defer ts.Close()

				c := NewClient(ClientConfig{
					Endpoint:       ts.URL,
					TokenGenerator: NewStaticTokenGenerator("test"),
				})

				err := test.testedClientCall(c)
				if run.invalidResponse {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			})
		}
	}
}
