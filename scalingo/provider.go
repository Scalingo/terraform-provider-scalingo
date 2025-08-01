// Package scalingo is the main package of the Terraform Provider for Scalingo.
// It aims at defining all resources and data providers from the plugin.
package scalingo

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v8"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SCALINGO_API_TOKEN", nil),
				Description: "API Token to authenticate requests with. Can also be sourced from `SCALINGO_API_TOKEN`.",
			},
			"api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SCALINGO_API_URL", nil),
				Description: "URL of the Scalingo Application API to use (Override region default). Can also be sourced from `SCALINGO_API_URL`.",
			},
			"db_api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SCALINGO_DB_API_URL", nil),
				Description: "URL of the Scalingo Database API to use (Override region default). Can also be sourced from `SCALINGO_DB_API_URL`.",
			},
			"auth_api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SCALINGO_AUTH_URL", nil),
				Description: "URL of the Scalingo Authentication API to use (Override region default). Can also be sourced from `SCALINGO_AUTH_URL`.",
			},
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SCALINGO_REGION", "osc-fr1"),
				Description: "Region to use with the provider. Can also be sourced from `SCALINGO_REGION`. (default: `osc-fr1`)",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"scalingo_notification_platform": dataSourceScNotificationPlatform(),
			"scalingo_region":                dataSourceScRegion(),
			"scalingo_stack":                 dataSourceScStack(),
			"scalingo_container_size":        dataSourceScContainerSize(),
			"scalingo_addon_providers":       dataSourceScAddonProvider(),
			"scalingo_scm_integration":       dataSourceScScmIntegration(),
			"scalingo_invoices":              dataSourceScInvoice(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"scalingo_addon":           resourceScalingoAddon(),
			"scalingo_alert":           resourceScalingoAlert(),
			"scalingo_app":             resourceScalingoApp(),
			"scalingo_autoscaler":      resourceScalingoAutoscaler(),
			"scalingo_collaborator":    resourceScalingoCollaborator(),
			"scalingo_container_type":  resourceScalingoContainerType(),
			"scalingo_domain":          resourceScalingoDomain(),
			"scalingo_log_drain":       resourceScalingoLogDrain(),
			"scalingo_notifier":        resourceScalingoNotifier(),
			"scalingo_project":         resourceScalingoProject(),
			"scalingo_scm_integration": resourceScalingoScmIntegration(),
			"scalingo_scm_repo_link":   resourceScalingoScmRepoLink(),
			"scalingo_ssh_key":         resourceScalingoSSHKey(),
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
