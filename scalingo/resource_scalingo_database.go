package scalingo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v8"
)

const PROVISIONING_TIMEOUT = 20 * time.Minute
const PROVISIONING_POLL_INTERVAL = 5 * time.Second

func resourceScalingoDatabase() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDatabaseCreate,
		ReadContext:   resourceDatabaseRead,
		UpdateContext: resourceDatabaseUpdate,
		DeleteContext: resourceDatabaseDelete,
		Description:   "Resource representing a Database NG on Scalingo",

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Database NG",
			},
			"project_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "ID of the project to which the Database NG belongs to",
			},
			"technology": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "Technology of the Database NG",
			},
			"plan": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the plan of the Database NG to provision",
			},
			"plan_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the plan of the Database NG to provision",
			},
			"database_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the Database NG on DBAPI side",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceDatabaseImport,
		},
	}
}

func resourceDatabaseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)
	previewClient := scalingo.NewPreviewClient(client)

	//nolint:errcheck // type assertions cannot fail it's defined in the schema.
	var (
		technology = d.Get("technology").(string)
		planName   = d.Get("plan").(string)
		name       = d.Get("name").(string)
		projectID  = d.Get("project_id").(string)
	)

	planID, err := addonPlanID(ctx, client, technology, planName)
	if err != nil {
		return diag.Errorf("get addon plan id: %v", err)
	}

	err = d.Set("plan_id", planID)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := previewClient.DatabaseCreate(ctx, scalingo.DatabaseCreateParams{
		AddonProviderID: technology,
		PlanID:          planID,
		Name:            name,
		ProjectID:       projectID,
	})
	if err != nil {
		return diag.Errorf("provision database: %v", err)
	}

	res, err = waitUntilDatabaseProvisioned(ctx, client, res)
	if err != nil {
		return diag.Errorf("wait for the addon to be provisioned: %v", err)
	}

	d.SetId(res.ID)

	err = d.Set("database_id", res.Database.ID)
	if err != nil {
		return diag.Errorf("store database id: %v", err)
	}

	return nil
}

func resourceDatabaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)
	previewClient := scalingo.NewPreviewClient(client)

	database, err := previewClient.DatabaseShow(ctx, d.Id())
	if err != nil {
		return diag.Errorf("get database details: %v", err)
	}

	planID, err := addonPlanID(ctx, client, database.Technology, database.Plan)
	if err != nil {
		return diag.Errorf("get addon plan id: %v", err)
	}

	d.SetId(database.ID)

	err = SetAll(d, map[string]interface{}{
		"name":        database.Name,
		"technology":  database.Technology,
		"plan":        database.Plan,
		"plan_id":     planID,
		"project_id":  database.ProjectID,
		"database_id": database.Database.ID,
	})
	if err != nil {
		return diag.Errorf("store database information: %v", err)
	}

	return nil
}

func resourceDatabaseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)
	previewClient := scalingo.NewPreviewClient(client)

	database, err := previewClient.DatabaseShow(ctx, d.Id())
	if err != nil {
		return diag.Errorf("get database information for %v: %v", d.Id(), err)
	}

	if d.HasChange("name") {
		newName, ok := d.Get("name").(string)
		if !ok {
			return diag.Errorf("name must be a string")
		}

		app, err := client.AppsShow(ctx, database.App.ID)
		if err != nil {
			return diag.Errorf("fetch database application: %v", err)
		}

		_, err = client.AppsRename(ctx, app.Name, newName)
		if err != nil {
			return diag.Errorf("rename database application: %v", err)
		}
	}

	if d.HasChange("project_id") {
		_, rawProjectID := d.GetChange("project_id")
		projectID, ok := rawProjectID.(string)
		if !ok {
			return diag.Errorf("cast project ID")
		}
		_, err := client.AppsSetProject(ctx, database.App.ID, projectID)
		if err != nil {
			return diag.Errorf("set project ID: %v", err)
		}
	}

	return nil
}

func resourceDatabaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	name, ok := d.Get("name").(string)
	if !ok {
		return diag.Errorf("name must be a string")
	}

	err := client.AppsDestroy(ctx, d.Id(), name)
	if err != nil {
		return diag.Errorf("destroy database: %v", err)
	}
	return nil
}

func resourceDatabaseImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client, _ := meta.(*scalingo.Client)
	previewClient := scalingo.NewPreviewClient(client)

	// d.Id() contains the database name provided by the user during import corresponding to the app name
	// We need to find the App ID associated with this app name to retrieve the database

	appName := d.Id()

	app, err := client.AppsShow(ctx, appName)
	if err != nil {
		return nil, fmt.Errorf("search database: %v", err)
	}

	database, err := previewClient.DatabaseShow(ctx, app.ID)
	if err != nil {
		return nil, fmt.Errorf("get database details: %v", err)
	}

	// Set the ID to the database ID for subsequent read operation
	d.SetId(database.ID)

	diags := resourceDatabaseRead(ctx, d, meta)
	if diags.HasError() {
		return nil, fmt.Errorf("read database: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

func waitUntilDatabaseProvisioned(ctx context.Context, client *scalingo.Client, scalingoDatabase scalingo.DatabaseNG) (scalingo.DatabaseNG, error) {
	var err error
	previewClient := scalingo.NewPreviewClient(client)

	timer := time.NewTimer(PROVISIONING_TIMEOUT)
	ticker := time.NewTicker(PROVISIONING_POLL_INTERVAL)
	defer timer.Stop()
	defer ticker.Stop()

	for scalingoDatabase.Database.Status != scalingo.DatabaseStatusRunning {
		scalingoDatabase, err = previewClient.DatabaseShow(ctx, scalingoDatabase.App.ID)
		if err != nil {
			// Database might not be available immediately after creation, retry
			if !strings.Contains(err.Error(), "not found") {
				return scalingoDatabase, fmt.Errorf("get the database: %w", err)
			}
			// Continue waiting if database not found yet
		} else if scalingoDatabase.Database.Status == scalingo.DatabaseStatusRunning {
			return scalingoDatabase, nil
		}
		select {
		case <-timer.C:
			return scalingoDatabase, errors.New("database provisioning timed out")
		case <-ticker.C:
		}
	}
	return scalingoDatabase, nil
}
