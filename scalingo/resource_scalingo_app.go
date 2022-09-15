package scalingo

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	scalingo "github.com/Scalingo/go-scalingo/v5"
)

func resourceScalingoApp() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAppCreate,
		ReadContext:   resourceAppRead,
		UpdateContext: resourceAppUpdate,
		DeleteContext: resourceAppDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"environment": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"all_environment": {
				Type:     schema.TypeMap,
				Computed: true,
			},
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"git_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"force_https": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"stack_id": {
				Type:     schema.TypeString,
				Optional: true,
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					if newValue == "" {
						return true
					}
					return true
				},
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

	app, err := client.AppsCreate(ctx, createOpts)
	if err != nil {
		return diag.Errorf("fail to create app: %v", err)
	}

	d.SetId(app.ID)
	err = SetAll(d, map[string]interface{}{
		"url":      app.URL,
		"git_url":  app.GitUrl,
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
			return diag.Errorf("fail to set environement variables: %v", err)
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
		"name":        app.Name,
		"url":         app.URL,
		"git_url":     app.GitUrl,
		"force_https": app.ForceHTTPS,
		"stack_id":    app.StackID,
	})
	if err != nil {
		return diag.Errorf("fail to store application informations: %v", err)
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
			"git_url": app.GitUrl,
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
