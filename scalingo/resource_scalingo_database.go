package scalingo

import (
	"context"
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
				Description: "technology of the Database NG",
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

	technology, _ := d.Get("technology").(string)
	planName, _ := d.Get("plan").(string)

	planID, err := addonPlanID(ctx, client, technology, planName)
	if err != nil {
		return diag.Errorf("fail to get addon plan id: %v", err)
	}

	if err := d.Set("plan_id", planID); err != nil {
		return diag.FromErr(err)
	}

	res, err := previewClient.DatabaseCreate(ctx, scalingo.DatabaseCreateParams{
		AddonProviderID: technology,
		PlanID:          planID,
		Name:            d.Get("name").(string),
		ProjectID:       d.Get("project_id").(string),
	})

	if err != nil {
		return diag.Errorf("fail to provision addon: %v", err)
	}

	res, err = waitUntilProvisionedPreview(ctx, client, res)
	if err != nil {
		return diag.Errorf("fail to wait for the addon to be provisioned: %v", err)
	}

	d.SetId(res.ID)
	if err := d.Set("app_id", res.App.ID); err != nil {
		return diag.Errorf("fail to store app id: %v", err)
	}
	if err := d.Set("database_id", res.Database.ID); err != nil {
		return diag.Errorf("fail to store database id: %v", err)
	}

	return nil
}

func resourceDatabaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)
	previewClient := scalingo.NewPreviewClient(client)

	database, err := previewClient.DatabaseShow(ctx, d.Id())
	if err != nil {
		return diag.Errorf("fail to get database details: %v", err)
	}

	d.SetId(database.ID)

	planID, err := addonPlanID(ctx, client, database.Technology, database.Plan)
	if err != nil {
		return diag.Errorf("fail to get addon plan id: %v", err)
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
		return diag.Errorf("fail to store database information: %v", err)
	}

	return nil
}

func resourceDatabaseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)
	previewClient := scalingo.NewPreviewClient(client)

	database, err := previewClient.DatabaseShow(ctx, d.Id())
	if err != nil {
		return diag.Errorf("fail to get addon information for %v: %v", d.Id(), err)
	}

	if d.HasChange("name") {
		app, err := client.AppsShow(ctx, database.App.ID)
		if err != nil {
			return diag.Errorf("fetch database application: %v", err)
		}

		_, err = client.AppsRename(ctx, app.Name, d.Get("name").(string))
		if err != nil {
			return diag.Errorf("fail to rename database app: %v", err)
		}
	}

	if d.HasChange("project_id") {
		_, stackID := d.GetChange("project_id")
		_, err := client.AppsSetProject(ctx, database.App.ID, stackID.(string))
		if err != nil {
			return diag.Errorf("fail to set project ID: %v", err)
		}
	}

	return nil
}

func resourceDatabaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	id := d.Get("app_id").(string)
	name, _ := d.Get("name").(string)

	err := client.AppsDestroy(ctx, id, name)
	if err != nil {
		return diag.Errorf("destroy app: %v", err)
	}
	return nil
}

func resourceDatabaseImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client, _ := meta.(*scalingo.Client)
	previewClient := scalingo.NewPreviewClient(client)

	databaseName := d.Id()

	databases, err := previewClient.DatabasesList(ctx)
	if err != nil {
		return nil, fmt.Errorf("fail to list databases: %v", err)
	}

	var foundDatabase *scalingo.DatabaseNG
	for _, db := range databases {
		if db.Name == databaseName {
			foundDatabase = &db
			break
		}
	}

	if foundDatabase == nil {
		return nil, fmt.Errorf("database with name %s not found", databaseName)
	}

	d.SetId(foundDatabase.ID)

	diags := resourceDatabaseRead(ctx, d, meta)
	if diags.HasError() {
		return nil, fmt.Errorf("fail to read database: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

func waitUntilProvisionedPreview(ctx context.Context, client *scalingo.Client, scalingo_database scalingo.DatabaseNG) (scalingo.DatabaseNG, error) {
	var err error
	previewClient := scalingo.NewPreviewClient(client)

	timer := time.NewTimer(5 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)
	defer timer.Stop()
	defer ticker.Stop()

	for scalingo_database.Database.Status != scalingo.DatabaseStatusRunning {
		scalingo_database, err = previewClient.DatabaseShow(ctx, scalingo_database.App.ID)
		if err != nil {
			// Database might not be available immediately after creation, retry
			if !strings.Contains(err.Error(), "not found") {
				return scalingo_database, fmt.Errorf("fail to get the database: %w", err)
			}
			// Continue waiting if database not found yet
		} else if scalingo_database.Database.Status == scalingo.DatabaseStatusRunning {
			return scalingo_database, nil
		}
		select {
		case <-timer.C:
			return scalingo_database, fmt.Errorf("database provisioning timed out")
		case <-ticker.C:
		}
	}
	return scalingo_database, nil
}
