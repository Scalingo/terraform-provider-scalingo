package scalingo

import (
	scalingo "github.com/Scalingo/go-scalingo"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceScalingoApp() *schema.Resource {
	return &schema.Resource{
		Create: resourceAppCreate,
		Read:   resourceAppRead,
		Update: resourceAppUpdate,
		Delete: resourceAppDelete,

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
		},

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceAppCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)

	appName := d.Get("name").(string)

	app, err := client.AppsCreate(scalingo.AppsCreateOpts{
		Name: appName,
	})

	if err != nil {
		return err
	}

	d.SetId(app.Id)
	d.Set("url", app.Url)
	d.Set("git_url", app.GitUrl)

	if d.Get("environment") != nil {
		environment := d.Get("environment").(map[string]interface{})
		var variables scalingo.Variables
		for name, value := range environment {
			variables = append(variables, &scalingo.Variable{
				Name:  name,
				Value: value.(string),
			})
		}

		_, _, err := client.VariableMultipleSet(d.Id(), variables)
		if err != nil {
			return err
		}

		allEnvironment, err := appEnvironment(client, app.Id)
		if err != nil {
			return err
		}
		d.Set("all_environment", allEnvironment)
	}

	return nil
}

func resourceAppRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)

	id := d.Id()

	app, err := client.AppsShow(id)
	if err != nil {
		return err
	}

	d.SetId(app.Id)
	d.Set("name", app.Name)
	d.Set("url", app.Url)
	d.Set("git_url", app.GitUrl)

	variables, err := client.VariablesList(d.Id())
	if err != nil {
		return err
	}

	currentEnvironment := d.Get("environment").(map[string]interface{})

	environment := make(map[string]interface{})
	allEnvironment := make(map[string]interface{})

	for _, variable := range variables {
		if _, ok := currentEnvironment[variable.Name]; ok {
			environment[variable.Name] = variable.Value
		}
		allEnvironment[variable.Name] = variable.Value
	}

	d.Set("all_environment", allEnvironment)
	d.Set("environment", environment)

	return nil
}

func resourceAppUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)
	d.Partial(true)

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")

		app, err := client.AppsRename(oldName.(string), newName.(string))
		if err != nil {
			return err
		}
		d.Set("url", app.Url)
		d.Set("git_url", app.GitUrl)

		d.SetPartial("name")
		d.SetPartial("url")
		d.SetPartial("git_url")
	}

	if d.HasChange("environment") {
		oldVariables, newVariables := d.GetChange("environment")
		diff := MapDiff(oldVariables.(map[string]interface{}), newVariables.(map[string]interface{}))
		variables := newVariables.(map[string]interface{})

		err := deleteVariablesByName(client, d.Id(), diff.Deleted)
		if err != nil {
			return err
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

		_, _, err = client.VariableMultipleSet(d.Id(), variablesToSet)
		if err != nil {
			return err
		}

		d.SetPartial("environment")

		allEnvironment, err := appEnvironment(client, d.Id())
		if err != nil {
			return err
		}

		d.Set("all_environment", allEnvironment)
		d.SetPartial("all_environment")

		// Ignore the restart error, here the error is probably linked to the
		// application status, which means that the environment will be applied
		// later.
		// If the restart occured, wait synchronously until the end of the restart
		// to validate that everything went fine
		res, err := client.AppsRestart(d.Id(), nil)
		if err == nil && res.StatusCode == 202 {
			defer res.Body.Close()
			location := res.Header.Get("Location")
			err = waitOperation(client, location)
		}
	}

	d.Partial(false)
	return nil
}

func resourceAppDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)

	id := d.Id()
	name := d.Get("name").(string)

	err := client.AppsDestroy(id, name)
	if err != nil {
		return err
	}
	return nil
}
