package scalingo

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/Scalingo/go-scalingo/v11"
	scalingohttp "github.com/Scalingo/go-scalingo/v11/http"
	"github.com/Scalingo/go-scalingo/v11/scalingomock"
)

func TestProjectCreateInOrganization(t *testing.T) {
	ctx := t.Context()
	client := scalingomock.NewMockAPI(gomock.NewController(t))
	gomock.InOrder(
		client.EXPECT().OrganizationsList(ctx).Return([]scalingo.Organization{
			{ID: "org-staging-uuid", Slug: "my-org-staging"},
			{ID: "org-uuid", Slug: "my-org"},
		}, nil),
		client.EXPECT().ProjectAdd(ctx, scalingo.ProjectAddParams{
			Name: "my-project", Default: true, OrganizationID: "org-uuid",
		}).Return(scalingo.Project{ID: "prj-1", Name: "my-project", Default: true, OrganizationID: "org-uuid"}, nil),
	)
	data := schema.TestResourceDataRaw(t, resourceScalingoProject().Schema, map[string]any{
		"name": "my-project", "default": true, "organization": "my-org",
	})

	diags := resourceProjectCreate(ctx, data, client)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, "prj-1", data.Id())
	assert.Equal(t, "my-org", data.Get("organization"))
}

func TestProjectCreatePersonal(t *testing.T) {
	ctx := t.Context()
	client := scalingomock.NewMockAPI(gomock.NewController(t))
	client.EXPECT().ProjectAdd(ctx, scalingo.ProjectAddParams{Name: "my-project"}).
		Return(scalingo.Project{ID: "prj-1", Name: "my-project"}, nil)
	data := schema.TestResourceDataRaw(t, resourceScalingoProject().Schema, map[string]any{
		"name": "my-project",
	})

	diags := resourceProjectCreate(ctx, data, client)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, "prj-1", data.Id())
	assert.Empty(t, data.Get("organization"))
}

func TestProjectCreateOrganizationResolutionFailure(t *testing.T) {
	tests := map[string]struct {
		organizations []scalingo.Organization
		listErr       error
		wantErr       string
	}{
		"organizations list fails": {
			listErr: &scalingohttp.RequestFailedError{Code: http.StatusUnauthorized, APIError: errors.New("unauthorized")},
			wantErr: "list organizations",
		},
		"slug not found": {
			organizations: []scalingo.Organization{{ID: "org-staging-uuid", Slug: "my-org-staging"}},
			wantErr:       `organization "my-org" not found`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := t.Context()
			// No ProjectAdd expectation: the project must not be created when the slug cannot be resolved.
			client := scalingomock.NewMockAPI(gomock.NewController(t))
			client.EXPECT().OrganizationsList(ctx).Return(test.organizations, test.listErr)
			data := schema.TestResourceDataRaw(t, resourceScalingoProject().Schema, map[string]any{
				"name": "my-project", "organization": "my-org",
			})

			diags := resourceProjectCreate(ctx, data, client)
			require.True(t, diags.HasError())
			assert.Contains(t, diags[0].Summary, test.wantErr)
			assert.Empty(t, data.Id())
		})
	}
}

// Terraform skips the plan check when the organization is unknown at plan time, so Update must refuse the change itself.
func TestProjectApplyRefusesOrganizationChangeWithoutUpdate(t *testing.T) {
	ctx := t.Context()
	// No ProjectUpdate expectation: the API must not be called.
	client := scalingomock.NewMockAPI(gomock.NewController(t))
	state := &terraform.InstanceState{
		ID: "prj-1",
		Attributes: map[string]string{
			"id": "prj-1", "name": "my-project", "default": "false", "organization": "my-org",
		},
	}
	diff := &terraform.InstanceDiff{
		Attributes: map[string]*terraform.ResourceAttrDiff{
			"name":         {Old: "my-project", New: "renamed"},
			"organization": {Old: "my-org", New: "other-org"},
		},
	}

	newState, diags := resourceScalingoProject().Apply(ctx, state, diff, client)
	require.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary, "organization can only be set when the project is created")
	assert.Equal(t, "prj-1", newState.ID)
	assert.Equal(t, "my-org", newState.Attributes["organization"])
	assert.Equal(t, "my-project", newState.Attributes["name"])
}

// fakeProjectAPI serves the organizations and projects endpoints used by the plan test.
type fakeProjectAPI struct {
	t        *testing.T
	mu       sync.Mutex
	projects map[string]map[string]any
}

func (api *fakeProjectAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.mu.Lock()
	defer api.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	projectID := strings.TrimPrefix(r.URL.Path, "/v1/projects/")
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/v1/organizations":
		api.writeJSON(w, http.StatusOK, map[string]any{"organizations": []map[string]any{
			{"id": "org-uuid", "slug": "my-org"},
		}})
	case r.Method == http.MethodPost && r.URL.Path == "/v1/projects":
		var body struct {
			Project map[string]any `json:"project"`
		}
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			api.t.Errorf("decode project creation: %v", err)
		}
		project := body.Project
		project["id"] = "prj-1"
		api.projects["prj-1"] = project
		api.writeJSON(w, http.StatusCreated, map[string]any{"project": project})
	case api.projects[projectID] == nil:
		api.writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
	case r.Method == http.MethodGet:
		api.writeJSON(w, http.StatusOK, map[string]any{"project": api.projects[projectID]})
	case r.Method == http.MethodDelete:
		delete(api.projects, projectID)
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func (api *fakeProjectAPI) writeJSON(w http.ResponseWriter, status int, value any) {
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(value)
	if err != nil {
		api.t.Errorf("encode response: %v", err)
	}
}

func testProjectConfig(serverURL, organizationLine string) string {
	return `
provider "scalingo" {
  region       = ""
  api_url      = "` + serverURL + `"
  auth_api_url = "` + serverURL + `"
  db_api_url   = "` + serverURL + `"
}

resource "scalingo_project" "test" {
  name = "my-project"
  ` + organizationLine + `
}
`
}

func TestProjectPlanInOrganization(t *testing.T) {
	// The SDK harness would otherwise download Terraform.
	_, err := exec.LookPath("terraform")
	if err != nil {
		t.Skip("terraform CLI not found; add it to PATH to run plan tests")
	}
	for _, env := range []string{"SCALINGO_API_TOKEN", "SCALINGO_API_URL", "SCALINGO_DB_API_URL", "SCALINGO_AUTH_URL", "SCALINGO_REGION"} {
		t.Setenv(env, "")
	}
	server := httptest.NewServer(&fakeProjectAPI{t: t, projects: map[string]map[string]any{}})
	t.Cleanup(server.Close)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				// The harness also checks that the plan following the apply is empty.
				Config: testProjectConfig(server.URL, `organization = "my-org"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalingo_project.test", "id", "prj-1"),
					resource.TestCheckResourceAttr("scalingo_project.test", "organization", "my-org"),
				),
			},
			{
				Config:   testProjectConfig(server.URL, ""),
				PlanOnly: true,
			},
			{
				Config:      testProjectConfig(server.URL, `organization = "other-org"`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`organization can only be set when the project is created`),
			},
		},
	})
}
