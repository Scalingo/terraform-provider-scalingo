package scalingo

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	scalingo "github.com/Scalingo/go-scalingo/v6"
)

func resourceScalingoApp() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAppCreate,
		ReadContext:   resourceAppRead,
		UpdateContext: resourceAppUpdate,
		DeleteContext: resourceAppDelete,
		Description:   "Resource representing an application",

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the application",
			},
			"environment": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Key-value map of environment variables attached to the application",
			},
			"all_environment": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Computed key-value map containing environment in read-only",
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Base URL (https://*) to access the application",
			},
			"git_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Hostname to use to deploy code with Git + SSH",
			},
			"force_https": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Redirect HTTP traffic to HTTPS + HSTS header if enabled",
			},
			"router_logs": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable Router Logs into your application logs for a deeper understanding of your application",
			},
			"sticky_session": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable the Sticky Session feature, which associate alls HTTP requests from an end-user to a single `web` application container.",
			},
			"stack_id": {
				Type: schema.TypeString,
				// Either set by the user, either set automatically by server if
				// no value is provided
				Optional:    true,
				Computed:    true,
				Description: "ID of the base stack to use (scalingo-18/scalingo-20/scalingo-22)",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceAppCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	appName, _ := d.Get("name").(string)

	createOpts := scalingo.AppsCreateOpts{
		Name: appName,
	}

	stackID, _ := d.Get("stack_id").(string)
	if stackID != "" {
		createOpts.StackID = stackID
	}

	tflog.Info(ctx, "Creating Scalingo application", map[string]interface{}{
		"name":     createOpts.Name,
		"stack_id": createOpts.StackID,
	})
	app, err := client.AppsCreate(ctx, createOpts)
	if err != nil {
		return diag.Errorf("fail to create app: %v", err)
	}

	d.SetId(app.ID)
	err = SetAll(d, map[string]interface{}{
		"url":      app.URL,
		"git_url":  app.GitURL,
		"stack_id": app.StackID,
	})

	if err != nil {
		return diag.Errorf("fail to store application attributes: %v", err)
	}

	if d.Get("environment") != nil {
		environment, _ := d.Get("environment").(map[string]interface{})
		var variables scalingo.Variables
		for name, value := range environment {
			variables = append(variables, &scalingo.Variable{
				Name:  name,
				Value: value.(string),
			})
		}

		_, _, err := client.VariableMultipleSet(ctx, d.Id(), variables)
		if err != nil {
			return diag.Errorf("fail to set environment variables: %v", err)
		}

		allEnvironment, err := appEnvironment(ctx, client, app.ID)
		if err != nil {
			return diag.Errorf("fail to fetch application environment: %v", err)
		}
		err = d.Set("all_environment", allEnvironment)
		if err != nil {
			return diag.Errorf("fail to store application environment: %v", err)
		}
	}

	if forceHTTPS, _ := d.Get("force_https").(bool); forceHTTPS {
		_, err := client.AppsForceHTTPS(ctx, app.ID, forceHTTPS)
		if err != nil {
			return diag.Errorf("fail to force HTTPS: %v", err)
		}
	}

	if routerLogs, _ := d.Get("router_logs").(bool); routerLogs {
		_, err := client.AppsRouterLogs(ctx, app.ID, routerLogs)
		if err != nil {
			return diag.Errorf("fail to enable Router Logs: %v", err)
		}
	}

	if stickySession, _ := d.Get("sticky_session").(bool); stickySession {
		_, err := client.AppsStickySession(ctx, app.ID, stickySession)
		if err != nil {
			return diag.Errorf("fail to enable StickySession: %v", err)
		}
	}

	return nil
}

func resourceAppRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	id := d.Id()

	app, err := client.AppsShow(ctx, id)
	if err != nil {
		return diag.Errorf("fail to fetch application: %v", err)
	}

	d.SetId(app.ID)
	err = SetAll(d, map[string]interface{}{
		"name":           app.Name,
		"url":            app.URL,
		"git_url":        app.GitURL,
		"force_https":    app.ForceHTTPS,
		"router_logs":    app.RouterLogs,
		"sticky_session": app.StickySession,
		"stack_id":       app.StackID,
	})
	if err != nil {
		return diag.Errorf("fail to store application information: %v", err)
	}

	variables, err := client.VariablesList(ctx, d.Id())
	if err != nil {
		return diag.Errorf("fail to list application variables: %v", err)
	}

	currentEnvironment, _ := d.Get("environment").(map[string]interface{})

	environment := make(map[string]interface{})
	allEnvironment := make(map[string]interface{})

	for _, variable := range variables {
		if _, ok := currentEnvironment[variable.Name]; ok {
			environment[variable.Name] = variable.Value
		}
		allEnvironment[variable.Name] = variable.Value
	}

	err = SetAll(d, map[string]interface{}{
		"all_environment": allEnvironment,
		"environment":     environment,
	})
	if err != nil {
		return diag.Errorf("fail to store application environment: %v", err)
	}

	return nil
}

func resourceAppUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")

		app, err := client.AppsRename(ctx, oldName.(string), newName.(string))
		if err != nil {
			return diag.Errorf("fail to rename app: %v", err)
		}

		err = SetAll(d, map[string]interface{}{
			"name":    app.Name,
			"url":     app.URL,
			"git_url": app.GitURL,
		})
		if err != nil {
			return diag.Errorf("fail to store application metadata: %v", err)
		}
	}

	if d.HasChange("environment") {
		oldVariables, newVariables := d.GetChange("environment")
		diff := MapDiff(oldVariables.(map[string]interface{}), newVariables.(map[string]interface{}))
		variables, _ := newVariables.(map[string]interface{})

		err := deleteVariablesByName(ctx, client, d.Id(), diff.Deleted)
		if err != nil {
			return diag.Errorf("fail to delete variables: %v", err)
		}

		var variablesToSet scalingo.Variables

		for _, name := range diff.Added {
			variablesToSet = append(variablesToSet, &scalingo.Variable{
				Name:  name,
				Value: variables[name].(string),
			})
		}

		for _, name := range diff.Modified {
			variablesToSet = append(variablesToSet, &scalingo.Variable{
				Name:  name,
				Value: variables[name].(string),
			})
		}

		_, _, err = client.VariableMultipleSet(ctx, d.Id(), variablesToSet)
		if err != nil {
			return diag.Errorf("fail to set variables: %v", err)
		}

		allEnvironment, err := appEnvironment(ctx, client, d.Id())
		if err != nil {
			return diag.Errorf("fail to get application environment: %v", err)
		}

		err = d.Set("all_environment", allEnvironment)
		if err != nil {
			return diag.Errorf("fail to store application environment: %v", err)
		}

		err = restartApp(ctx, client, d.Id())
		if err != nil {
			return diag.Errorf("fail to restart app: %v", err)
		}
	}

	if d.HasChange("force_https") {
		_, ForceHTTPS := d.GetChange("force_https")
		_, err := client.AppsForceHTTPS(ctx, d.Id(), ForceHTTPS.(bool))
		if err != nil {
			return diag.Errorf("fail to set force HTTPS: %v", err)
		}
	}

	if d.HasChange("router_logs") {
		_, RouterLogs := d.GetChange("router_logs")
		_, err := client.AppsRouterLogs(ctx, d.Id(), RouterLogs.(bool))
		if err != nil {
			return diag.Errorf("fail to set Router Logs: %v", err)
		}
	}

	if d.HasChange("sticky_session") {
		_, StickySession := d.GetChange("sticky_session")
		_, err := client.AppsStickySession(ctx, d.Id(), StickySession.(bool))
		if err != nil {
			return diag.Errorf("fail to set Sticky Session: %v", err)
		}
	}

	if d.HasChange("stack_id") {
		_, stackID := d.GetChange("stack_id")
		_, err := client.AppsSetStack(ctx, d.Id(), stackID.(string))
		if err != nil {
			return diag.Errorf("fail to set application stack: %v", err)
		}
	}

	return nil
}

func resourceAppDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	id := d.Id()
	name, _ := d.Get("name").(string)

	err := client.AppsDestroy(ctx, id, name)
	if err != nil {
		return diag.Errorf("fail to destroy app: %v", err)
	}
	return nil
}

func restartApp(ctx context.Context, client *scalingo.Client, id string) error {
	// Ignore the restart error, here the error is probably linked to the
	// application status, which means that the environment will be applied
	// later.
	// If the restart occurred, wait synchronously until the end of the restart
	// to validate that everything went fine
	res, err := client.AppsRestart(ctx, id, nil)
	if err == nil && res.StatusCode == 202 {
		defer res.Body.Close()
		location := res.Header.Get("Location")
		err = waitOperation(ctx, client, location)
		if err != nil {
			return fmt.Errorf("fail to get wait for the operation to finish: %v", err)
		}
	}
	return nil
}
