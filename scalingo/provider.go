package scalingo

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	scalingo "github.com/Scalingo/go-scalingo/v5"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SCALINGO_API_TOKEN", nil),
			},
			"api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SCALINGO_API_URL", nil),
			},
			"db_api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SCALINGO_DB_API_URL", nil),
			},
			"auth_api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SCALINGO_AUTH_URL", nil),
			},
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SCALINGO_REGION", "osc-fr1"),
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"scalingo_notification_platform": dataSourceScNotificationPlatform(),
			"scalingo_region":                dataSourceScRegion(),
			"scalingo_stack":                 dataSourceScStack(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"scalingo_addon":           resourceScalingoAddon(),
			"scalingo_app":             resourceScalingoApp(),
			"scalingo_collaborator":    resourceScalingoCollaborator(),
			"scalingo_container_type":  resourceScalingoContainerType(),
			"scalingo_autoscaler":      resourceScalingoAutoscaler(),
			"scalingo_domain":          resourceScalingoDomain(),
			"scalingo_github_link":     resourceScalingoGithubLink(),
			"scalingo_scm_integration": resourceScalingoScmIntegration(),
			"scalingo_scm_repo_link":   resourceScalingoScmRepoLink(),
			"scalingo_notifier":        resourceScalingoNotifier(),
			"scalingo_run":             resourceScalingoRun(),
			"scalingo_log_drain":       resourceScalingoLogDrain(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
	apiURL, _ := data.Get("api_url").(string)
	authAPIURL, _ := data.Get("auth_api_url").(string)
	dbAPIURL, _ := data.Get("db_api_url").(string)
	apiToken, _ := data.Get("api_token").(string)
	region, _ := data.Get("region").(string)

	client, err := scalingo.New(ctx, scalingo.ClientConfig{
		Region:              region,
		APIToken:            apiToken,
		APIEndpoint:         apiURL,
		DatabaseAPIEndpoint: dbAPIURL,
		AuthEndpoint:        authAPIURL,
	})
	if err != nil {
		return nil, diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("fail to initialize Scalingo client: %v", err),
			},
		}
	}

	return client, nil
}
