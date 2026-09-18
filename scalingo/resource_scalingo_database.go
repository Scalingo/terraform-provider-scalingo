package scalingo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v11"
)

// provisioningTimeout is set to 40 minutes as a safe timeout because
// database provisioning and plan changes can take more than 30 minutes.
const provisioningTimeout = 40 * time.Minute

func resourceScalingoDatabase() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDatabaseCreate,
		ReadContext:   resourceDatabaseRead,
		UpdateContext: resourceDatabaseUpdate,
		DeleteContext: resourceDatabaseDelete,
		Description:   "Resource representing a Database NG on Scalingo",
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(provisioningTimeout),
			Update: schema.DefaultTimeout(provisioningTimeout),
		},

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
			"database_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "Full database connection URL, including scheme, credentials, host, port, database name and connection options",
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

	// Keep the created resource in state even if provisioning fails or is interrupted.
	d.SetId(res.ID)

	res, err = waitUntilDatabaseProvisioned(ctx, previewClient, res)
	if err != nil {
		return diag.Errorf("wait for the addon to be provisioned: %v", err)
	}

	err = d.Set("database_id", res.Database.ID)
	if err != nil {
		return diag.Errorf("store database id: %v", err)
	}

	return resourceDatabaseRead(ctx, d, meta)
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

	variables, err := client.VariablesList(ctx, database.ID)
	if err != nil {
		return diag.Errorf("get database environment variables: %v", err)
	}

	dbTypeName, err := toDatabaseTypeName(ctx, database)
	if err != nil {
		return diag.Errorf("to database type name: %v", err)
	}

	variableName := "SCALINGO_" + dbTypeName + "_URL"
	databaseURL, ok := variables.Contains(variableName)
	if !ok || databaseURL.Value == "" {
		return diag.Errorf("database connection URL variable %s is missing or empty", variableName)
	}

	d.SetId(database.ID)

	err = SetAll(d, map[string]interface{}{
		"name":         database.Name,
		"technology":   database.Technology,
		"plan":         database.Plan,
		"plan_id":      planID,
		"project_id":   database.ProjectID,
		"database_id":  database.Database.ID,
		"database_url": databaseURL.Value,
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

	if d.HasChange("plan") {
		technology, _ := d.Get("technology").(string)
		planName, _ := d.Get("plan").(string)
		planID, err := addonPlanID(ctx, client, technology, planName)
		if err != nil {
			return diag.Errorf("get addon plan id: %v", err)
		}

		addons, err := client.AddonsList(ctx, database.App.ID)
		if err != nil {
			return diag.Errorf("list addons: %v", err)
		}

		if len(addons) == 0 {
			return diag.Errorf("no addons found for database application %v", database.App.ID)
		}

		addonID := addons[0].ID

		_, err = client.AddonUpgrade(ctx, database.App.ID, addonID, scalingo.AddonUpgradeParams{
			PlanID: planID,
		})
		if err != nil {
			return diag.Errorf("upgrade database: %v", err)
		}

		database, err = waitUntilDatabasePlanChanged(ctx, client, database)
		if err != nil {
			return diag.Errorf("wait for database provisioning: %v", err)
		}

		if err := d.Set("plan_id", planID); err != nil {
			return diag.Errorf("store plan id: %v", err)
		}
	}

	return resourceDatabaseRead(ctx, d, meta)
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

func waitUntilDatabasePlanChanged(ctx context.Context, client *scalingo.Client, scalingoDatabase scalingo.DatabaseNG) (scalingo.DatabaseNG, error) {
	previewClient := scalingo.NewPreviewClient(client)

	// First, wait for the database to start updating (status != running)
	if scalingoDatabase.Database.Status == scalingo.DatabaseStatusRunning {
		var err error
		err = waitUntil(ctx, waitOptions{
			timeout:    provisioningTimeout,
			timeoutErr: errors.New("database plan change timed out waiting for update to start"),
		}, func() (bool, error) {
			scalingoDatabase, err = previewClient.DatabaseShow(ctx, scalingoDatabase.ID)
			if err != nil {
				return false, fmt.Errorf("get the database: %w", err)
			}
			return scalingoDatabase.Database.Status != scalingo.DatabaseStatusRunning, nil
		})
		if err != nil {
			return scalingoDatabase, err
		}
	}

	// Then wait for the database to be running again
	return waitUntilDatabaseProvisioned(ctx, previewClient, scalingoDatabase)
}

func waitUntilDatabaseProvisioned(ctx context.Context, previewClient scalingo.DatabasesPreviewService, scalingoDatabase scalingo.DatabaseNG) (scalingo.DatabaseNG, error) {
	// DatabaseNG.ID is currently the app ID expected by the preview API.
	// Creation responses do not populate App, so App.ID cannot be used here.
	appID := scalingoDatabase.ID
	err := waitUntil(ctx, waitOptions{
		timeout:    provisioningTimeout,
		timeoutErr: errors.New("database provisioning timed out"),
	}, func() (bool, error) {
		database, err := previewClient.DatabaseShow(ctx, appID)
		if err != nil {
			// Database might not be available immediately after creation, retry.
			if !errors.Is(err, scalingo.ErrDatabaseNotFound) {
				return false, fmt.Errorf("get the database: %w", err)
			}
			return false, nil
		}
		// Only replace the saved resource after a successful lookup. On a temporary
		// not-found error, DatabaseShow returns an empty DatabaseNG; assigning it
		// would erase the saved ID. Previously, subsequent lookups read that ID
		// and kept searching with an empty ID. Retaining the resource and capturing
		// appID before polling prevents this.
		scalingoDatabase = database
		return database.Database.Status == scalingo.DatabaseStatusRunning, nil
	})
	return scalingoDatabase, err
}
