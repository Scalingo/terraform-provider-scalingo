package scalingo

import (
	"fmt"

	scalingo "github.com/Scalingo/go-scalingo"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func Provider() terraform.ResourceProvider {
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
		},
		ResourcesMap: map[string]*schema.Resource{
			"scalingo_addon":          resourceScalingoAddon(),
			"scalingo_app":            resourceScalingoApp(),
			"scalingo_collaborator":   resourceScalingoCollaborator(),
			"scalingo_container_type": resourceScalingoContainerType(),
			"scalingo_autoscaler":     resourceScalingoAutoscaler(),
			"scalingo_domain":         resourceScalingoDomain(),
			"scalingo_github_link":    resourceScalingoGithubLink(),
			"scalingo_notifier":       resourceScalingoNotifier(),
			"scalingo_run":            resourceScalingoRun(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(data *schema.ResourceData) (interface{}, error) {
	apiURL := data.Get("api_url").(string)
	authAPIURL := data.Get("auth_api_url").(string)
	dbAPIURL := data.Get("db_api_url").(string)
	apiToken := data.Get("api_token").(string)
	region := data.Get("region").(string)

	client, err := scalingo.New(scalingo.ClientConfig{
		Region:              region,
		APIToken:            apiToken,
		APIEndpoint:         apiURL,
		DatabaseAPIEndpoint: dbAPIURL,
		AuthEndpoint:        authAPIURL,
	})
	if err != nil {
		return nil, fmt.Errorf("fail to initialize Scalingo client: %v", err)
	}

	return client, nil
}
