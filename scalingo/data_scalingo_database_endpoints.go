package scalingo

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v11"
)

const databaseIDDescription = "ID of the database"

// dataSourceScDatabaseEndpoints lists endpoints, optionally filtering by type:
//
//	data "scalingo_database_endpoint" "public" {
//	  database_id = scalingo_database.db.id
//	  type        = "public-rw"
//	}
//
// The matching endpoints are available in the endpoints attribute, for example
// data.scalingo_database_endpoint.public.endpoints[0].hostname. Omitting type
// returns every endpoint for the database:
//
//	data "scalingo_database_endpoints" "all" {
//	  database_id = scalingo_database.db.id
//	}
func dataSourceScDatabaseEndpoints() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScDatabaseEndpointsRead,
		Description: "Database endpoints retrieved from the Database API",

		Schema: map[string]*schema.Schema{
			"database_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: databaseIDDescription,
			},
			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Type of the database endpoint",
			},
			"include_default_credentials": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to include the default endpoint credentials",
			},
			"endpoints": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Database endpoints matching the filters",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the database endpoint",
						},
						"database_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: databaseIDDescription,
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of the database endpoint",
						},
						"hostname": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Hostname of the database endpoint",
						},
						"port": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Port of the database endpoint",
						},
						"username": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Username of the database endpoint",
						},
						"password": {
							Type:        schema.TypeString,
							Computed:    true,
							Sensitive:   true,
							Description: "Password of the database endpoint",
						},
					},
				},
			},
		},
	}
}

func dataSourceScDatabaseEndpointsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	databaseID, _ := d.Get("database_id").(string)
	if databaseID == "" {
		return diag.Errorf("database ID must be provided")
	}

	endpointType, _ := d.Get("type").(string)

	includeDefaultCredentials, _ := d.Get("include_default_credentials").(bool)
	previewClient := scalingo.NewPreviewClient(client)
	endpoints, err := previewClient.DatabaseEndpointsList(ctx, databaseID, scalingo.DatabaseEndpointsListParams{
		IncludeDefaultCredentials: includeDefaultCredentials,
	})
	if err != nil {
		return diag.Errorf("list database endpoints: %v", err)
	}

	selectedEndpoints := endpoints
	if endpointType != "" {
		selectedEndpoints = keepIf(endpoints, func(endpoint scalingo.DatabaseEndpoint) bool {
			return string(endpoint.Type) == endpointType
		})
	}
	if len(selectedEndpoints) == 0 {
		if endpointType != "" {
			return diag.Errorf("no endpoint found for database %q with type %q", databaseID, endpointType)
		}
		return diag.Errorf("no endpoints found for database %q", databaseID)
	}

	endpointState := func(endpoint scalingo.DatabaseEndpoint) map[string]any {
		username, password := "", ""
		if endpoint.Credentials != nil {
			username = endpoint.Credentials.Username
			password = endpoint.Credentials.Password
		}
		endpointDatabaseID := endpoint.DatabaseID
		if endpointDatabaseID == "" {
			endpointDatabaseID = databaseID
		}

		return map[string]any{
			"id":          endpoint.ID,
			"database_id": endpointDatabaseID,
			"type":        string(endpoint.Type),
			"hostname":    endpoint.Hostname,
			"port":        endpoint.Port,
			"username":    username,
			"password":    password,
		}
	}

	endpointsState := make([]map[string]any, 0, len(selectedEndpoints))
	for _, endpoint := range selectedEndpoints {
		endpointsState = append(endpointsState, endpointState(endpoint))
	}

	err = SetAll(d, map[string]any{
		"database_id": databaseID,
		"endpoints":   endpointsState,
	})
	if err != nil {
		return diag.Errorf("store database endpoint information: %v", err)
	}

	d.SetId(databaseID)

	return nil
}
