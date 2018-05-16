package scalingo

import (
	scalingo "github.com/Scalingo/go-scalingo"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_token": {
				Type:     schema.TypeString,
				Required: true,
			},
			"api_url": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "https://api.scalingo.com/",
			},
			"auth_api_url": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "https://auth.scalingo.com/",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"scalingo_app":    resourceScalingoApp(),
			"scalingo_domain": resourceScalingoDomain(),
			"scalingo_addon":  resourceScalingoAddon(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(data *schema.ResourceData) (interface{}, error) {
	api_url := data.Get("api_url").(string)
	auth_api_url := data.Get("auth_api_url").(string)
	api_token := data.Get("api_token").(string)

	client := scalingo.NewClient(scalingo.ClientConfig{
		APIToken:     api_token,
		Endpoint:     api_url,
		AuthEndpoint: auth_api_url,
	})

	return client, nil
}
