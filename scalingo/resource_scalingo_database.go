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
			"app_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the application automatically created with the Database NG",
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

	technology, ok := d.Get("technology").(string)
	if !ok {
		return diag.Errorf("technology must be a string")
	}
	planName, ok := d.Get("plan").(string)
	if !ok {
		return diag.Errorf("plan must be a string")
	}
	name, ok := d.Get("name").(string)
	if !ok {
		return diag.Errorf("name must be a string")
	}
	projectID, ok := d.Get("project_id").(string)
	if !ok {
		return diag.Errorf("project_id must be a string")
	}

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
		return diag.Errorf("provision addon: %v", err)
	}

	res, err = waitUntilDatabaseProvisioned(ctx, client, res)
	if err != nil {
		return diag.Errorf("wait for the addon to be provisioned: %v", err)
	}

	d.SetId(res.ID)

	err = d.Set("app_id", res.App.ID)
	if err != nil {
		return diag.Errorf("store app id: %v", err)
	}

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

	d.SetId(database.ID)

	planID, err := addonPlanID(ctx, client, database.Technology, database.Plan)
	if err != nil {
		return diag.Errorf("get addon plan id: %v", err)
	}

	err = SetAll(d, map[string]interface{}{
		"name":        database.Name,
		"technology":  database.Technology,
		"plan":        database.Plan,
		"plan_id":     planID,
		"project_id":  database.ProjectID,
		"app_id":      database.App.ID,
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
		return diag.Errorf("get addon information for %v: %v", d.Id(), err)
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
			return diag.Errorf("rename database app: %v", err)
		}
	}

	if d.HasChange("project_id") {
		_, stackID := d.GetChange("project_id")
		var err error
		_, err = client.AppsSetProject(ctx, database.App.ID, stackID.(string))
		if err != nil {
			return diag.Errorf("set project ID: %v", err)
		}
	}

	return nil
}

func resourceDatabaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	id, ok := d.Get("app_id").(string)
	if !ok {
		return diag.Errorf("app_id must be a string")
	}
	name, ok := d.Get("name").(string)
	if !ok {
		return diag.Errorf("name must be a string")
	}

	err := client.AppsDestroy(ctx, id, name)
	if err != nil {
		return diag.Errorf("destroy app: %v", err)
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

func waitUntilDatabaseProvisioned(ctx context.Context, client *scalingo.Client, scalingo_database scalingo.DatabaseNG) (scalingo.DatabaseNG, error) {
	var err error
	previewClient := scalingo.NewPreviewClient(client)

	timer := time.NewTimer(20 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)
	defer timer.Stop()
	defer ticker.Stop()

	for scalingo_database.Database.Status != scalingo.DatabaseStatusRunning {
		scalingo_database, err = previewClient.DatabaseShow(ctx, scalingo_database.App.ID)
		if err != nil {
			// Database might not be available immediately after creation, retry
			if !strings.Contains(err.Error(), "not found") {
				return scalingo_database, fmt.Errorf("get the database: %w", err)
			}
			// Continue waiting if database not found yet
		} else if scalingo_database.Database.Status == scalingo.DatabaseStatusRunning {
			return scalingo_database, nil
		}
		select {
		case <-timer.C:
			return scalingo_database, errors.New("database provisioning timed out")
		case <-ticker.C:
		}
	}
	return scalingo_database, nil
}
