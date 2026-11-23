package scalingo

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
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

func TestProjectCreateAPIErrorWithoutPersonalFallback(t *testing.T) {
	ctx := t.Context()
	client := scalingomock.NewMockAPI(gomock.NewController(t))
	client.EXPECT().OrganizationsList(ctx).Return([]scalingo.Organization{{ID: "org-uuid", Slug: "my-org"}}, nil)
	client.EXPECT().ProjectAdd(ctx, scalingo.ProjectAddParams{Name: "my-project", OrganizationID: "org-uuid"}).
		Return(scalingo.Project{}, &scalingohttp.RequestFailedError{
			Code: http.StatusForbidden, APIError: errors.New("you are not allowed to create a project in this organization"),
		})
	data := schema.TestResourceDataRaw(t, resourceScalingoProject().Schema, map[string]any{
		"name": "my-project", "organization": "my-org",
	})

	diags := resourceProjectCreate(ctx, data, client)
	require.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary, "you are not allowed to create a project in this organization")
	assert.Empty(t, data.Id())
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

type fakeProjectAPI struct {
	t        *testing.T
	mu       sync.Mutex
	projects map[string]map[string]any
	requests []fakeProjectAPIRequest
	nextID   int
}

type fakeProjectAPIRequest struct {
	method string
	path   string
	body   map[string]any
}

func newFakeProjectAPI(t *testing.T, projects map[string]map[string]any) *httptest.Server {
	t.Helper()
	if projects == nil {
		projects = map[string]map[string]any{}
	}
	api := &fakeProjectAPI{t: t, projects: projects}
	server := httptest.NewServer(api)
	t.Cleanup(server.Close)
	fakeProjectAPIs.Store(server.URL, api)
	return server
}

var fakeProjectAPIs sync.Map

func fakeProjectAPIFor(serverURL string) *fakeProjectAPI {
	value, _ := fakeProjectAPIs.Load(serverURL)
	api, _ := value.(*fakeProjectAPI)
	return api
}

func (api *fakeProjectAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.mu.Lock()
	defer api.mu.Unlock()

	var body map[string]any
	if r.Body != nil && r.ContentLength != 0 {
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			api.t.Errorf("decode %s %s body: %v", r.Method, r.URL.Path, err)
		}
	}
	api.requests = append(api.requests, fakeProjectAPIRequest{method: r.Method, path: r.URL.Path, body: body})

	w.Header().Set("Content-Type", "application/json")
	projectID := strings.TrimPrefix(r.URL.Path, "/v1/projects/")
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/v1/organizations":
		api.writeJSON(w, http.StatusOK, map[string]any{"organizations": []map[string]any{
			{"id": "org-staging-uuid", "slug": "my-org-staging"},
			{"id": "org-uuid", "slug": "my-org"},
		}})
	case r.Method == http.MethodPost && r.URL.Path == "/v1/projects":
		params, _ := body["project"].(map[string]any)
		api.nextID++
		newProjectID := fmt.Sprintf("prj-%d", api.nextID)
		project := map[string]any{
			"id":              newProjectID,
			"name":            params["name"],
			"default":         params["default"],
			"organization_id": params["organization_id"],
		}
		api.projects[newProjectID] = project
		api.writeJSON(w, http.StatusCreated, map[string]any{"project": project})
	case api.projects[projectID] == nil:
		api.writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
	case r.Method == http.MethodGet:
		api.writeJSON(w, http.StatusOK, map[string]any{"project": api.projects[projectID]})
	case r.Method == http.MethodPatch:
		params, _ := body["project"].(map[string]any)
		for _, attribute := range []string{"name", "default"} {
			if value, ok := params[attribute]; ok {
				api.projects[projectID][attribute] = value
			}
		}
		api.writeJSON(w, http.StatusOK, map[string]any{"project": api.projects[projectID]})
	case r.Method == http.MethodDelete:
		delete(api.projects, projectID)
		w.WriteHeader(http.StatusNoContent)
	default:
		api.t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
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

func (api *fakeProjectAPI) requestsMatching(method, path string) []fakeProjectAPIRequest {
	api.mu.Lock()
	defer api.mu.Unlock()

	var requests []fakeProjectAPIRequest
	for _, request := range api.requests {
		if request.method == method && strings.HasPrefix(request.path, path) {
			requests = append(requests, request)
		}
	}
	return requests
}

// skipWithoutTerraform keeps plan tests hermetic: the SDK harness would otherwise download Terraform.
func skipWithoutTerraform(t *testing.T) {
	t.Helper()
	if os.Getenv("TF_ACC_TERRAFORM_PATH") != "" {
		return
	}
	_, err := exec.LookPath("terraform")
	if err != nil {
		t.Skip("terraform CLI not found; set TF_ACC_TERRAFORM_PATH or add terraform to PATH to run plan tests")
	}
}

func setupProjectPlanTest(t *testing.T, projects map[string]map[string]any) string {
	t.Helper()
	skipWithoutTerraform(t)
	for _, env := range []string{"SCALINGO_API_TOKEN", "SCALINGO_API_URL", "SCALINGO_DB_API_URL", "SCALINGO_AUTH_URL", "SCALINGO_REGION"} {
		t.Setenv(env, "")
	}
	return newFakeProjectAPI(t, projects).URL
}

func testProjectProviderConfig(serverURL string) string {
	return fmt.Sprintf(`
provider "scalingo" {
  region       = ""
  api_url      = %[1]q
  auth_api_url = %[1]q
  db_api_url   = %[1]q
}
`, serverURL)
}

func testProjectConfig(serverURL, name, organization string) string {
	organizationLine := ""
	if organization != "" {
		organizationLine = fmt.Sprintf("  organization = %s\n", organization)
	}
	return testProjectProviderConfig(serverURL) + fmt.Sprintf(`
resource "scalingo_project" "test" {
  name = %q
%s}
`, name, organizationLine)
}

func testCheckProjectRequests(serverURL string, check func(api *fakeProjectAPI) error) resource.TestCheckFunc {
	return func(*terraform.State) error {
		return check(fakeProjectAPIFor(serverURL))
	}
}

func TestProjectPlanInOrganization(t *testing.T) {
	serverURL := setupProjectPlanTest(t, nil)
	const resourceName = "scalingo_project.test"
	organizationChangeError := regexp.MustCompile(`organization can only be set when the project is created`)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		CheckDestroy: testCheckProjectRequests(serverURL, func(api *fakeProjectAPI) error {
			if creations := api.requestsMatching(http.MethodPost, "/v1/projects"); len(creations) != 1 {
				return fmt.Errorf("got %d project creations, want 1", len(creations))
			}
			if updates := api.requestsMatching(http.MethodPatch, "/v1/projects/"); len(updates) != 2 {
				return fmt.Errorf("got %d project updates, want only the 2 renames", len(updates))
			}
			return nil
		}),
		Steps: []resource.TestStep{
			{
				// The harness also checks that the plan following the apply is empty.
				Config: testProjectConfig(serverURL, "my-project", `"my-org"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "prj-1"),
					resource.TestCheckResourceAttr(resourceName, "organization", "my-org"),
					testCheckProjectRequests(serverURL, func(api *fakeProjectAPI) error {
						creations := api.requestsMatching(http.MethodPost, "/v1/projects")
						if len(creations) != 1 {
							return fmt.Errorf("got %d project creations, want 1", len(creations))
						}
						project, _ := creations[0].body["project"].(map[string]any)
						if project["organization_id"] != "org-uuid" {
							return fmt.Errorf("organization_id = %v, want org-uuid", project["organization_id"])
						}
						return nil
					}),
				),
			},
			{
				Config:   testProjectConfig(serverURL, "my-project", ""),
				PlanOnly: true,
			},
			{
				Config:      testProjectConfig(serverURL, "my-project", `"my-org-staging"`),
				PlanOnly:    true,
				ExpectError: organizationChangeError,
			},
			{
				Config:      testProjectConfig(serverURL, "my-project", `" "`),
				ExpectError: regexp.MustCompile(`organization must not be empty`),
			},
			{
				Config: testProjectConfig(serverURL, "renamed-project", `"my-org"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "prj-1"),
					resource.TestCheckResourceAttr(resourceName, "name", "renamed-project"),
					resource.TestCheckResourceAttr(resourceName, "organization", "my-org"),
					testCheckProjectRequests(serverURL, func(api *fakeProjectAPI) error {
						updates := api.requestsMatching(http.MethodPatch, "/v1/projects/")
						if len(updates) != 1 {
							return fmt.Errorf("got %d project updates, want 1", len(updates))
						}
						project, _ := updates[0].body["project"].(map[string]any)
						if _, ok := project["organization_id"]; ok {
							return fmt.Errorf("update payload contains organization_id: %v", project)
						}
						if len(api.requestsMatching(http.MethodPost, "/v1/projects")) != 1 {
							return errors.New("project was recreated")
						}
						return nil
					}),
				),
			},
			{
				// The organization is unknown at plan time: the refusal happens during the apply.
				Config: testProjectProviderConfig(serverURL) + `
resource "terraform_data" "organization" {
  input = "my-org-staging"
}

resource "scalingo_project" "test" {
  name         = "renamed-project"
  organization = terraform_data.organization.output
}
`,
				ExpectError: organizationChangeError,
			},
			{
				// An unknown organization resolving to the stored slug must not block a rename.
				Config: testProjectProviderConfig(serverURL) + `
resource "terraform_data" "organization" {
  input = "my-org"
}

resource "scalingo_project" "test" {
  name         = "renamed-again"
  organization = terraform_data.organization.output
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "prj-1"),
					resource.TestCheckResourceAttr(resourceName, "name", "renamed-again"),
					resource.TestCheckResourceAttr(resourceName, "organization", "my-org"),
				),
			},
			{
				Config: testProjectProviderConfig(serverURL) + `
resource "terraform_data" "organization" {
  input = "my-org"
}

resource "scalingo_project" "test" {
  name = "renamed-again"
}
`,
				PlanOnly: true,
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"organization"},
			},
		},
	})
}

// Each case uses a single Scalingo resource: concurrent first calls race in the lazy client initialization of go-scalingo.
func TestProjectPlanPersonal(t *testing.T) {
	serverURL := setupProjectPlanTest(t, nil)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testProjectConfig(serverURL, "my-project", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalingo_project.test", "organization", ""),
					testCheckProjectRequests(serverURL, func(api *fakeProjectAPI) error {
						if calls := api.requestsMatching(http.MethodGet, "/v1/organizations"); len(calls) != 0 {
							return fmt.Errorf("got %d organizations calls for a personal project, want 0", len(calls))
						}
						return nil
					}),
				),
			},
			{
				Config:      testProjectConfig(serverURL, "my-project", `"my-org"`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`organization can only be set when the project is created`),
			},
		},
	})
}

func TestProjectPlanImportWithoutOrganization(t *testing.T) {
	serverURL := setupProjectPlanTest(t, map[string]map[string]any{
		"prj-existing": {"id": "prj-existing", "name": "existing-project", "default": false},
	})
	importConfig := testProjectProviderConfig(serverURL) + `
import {
  to = scalingo_project.existing
  id = "prj-existing"
}
`

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				// The harness also checks that the plan following the import is empty.
				Config: importConfig + `
resource "scalingo_project" "existing" {
  name = "existing-project"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scalingo_project.existing", "id", "prj-existing"),
					resource.TestCheckNoResourceAttr("scalingo_project.existing", "organization"),
				),
			},
			{
				Config: importConfig + `
resource "scalingo_project" "existing" {
  name         = "existing-project"
  organization = "my-org"
}
`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`organization can only be set when the project is created`),
			},
		},
	})
}
