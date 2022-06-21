package scalingo

import (
	"context"
	"errors"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	scalingo "github.com/Scalingo/go-scalingo/v4"
)

func resourceScalingoCollaborator() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCollaboratorCreate,
		ReadContext:   resourceCollaboratorRead,
		DeleteContext: resourceCollaboratorDelete,

		Schema: map[string]*schema.Schema{
			"app": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"username": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceCollaboratorImport,
		},
	}
}

func resourceCollaboratorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	collaborator, err := client.CollaboratorAdd(d.Get("app").(string), d.Get("email").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(collaborator.ID)

	err = SetAll(d, map[string]interface{}{
		"username": collaborator.Username,
		"status":   collaborator.Status,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceCollaboratorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	collaborators, err := client.CollaboratorsList(d.Get("app").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	var collaborator scalingo.Collaborator
	found := false

	for _, c := range collaborators {
		if c.ID == d.Id() {
			collaborator = c
			found = true
			break
		}
	}

	if !found {
		d.MarkNewResource()
		return nil
	}

	err = SetAll(d, map[string]interface{}{
		"username": collaborator.Username,
		"email":    collaborator.Email,
		"status":   collaborator.Status,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceCollaboratorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	err := client.CollaboratorRemove(d.Get("app").(string), d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceCollaboratorImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client, _ := meta.(*scalingo.Client)

	ids := strings.Split(d.Id(), ":")
	if len(ids) != 2 {
		return nil, errors.New("address should have the following format: <appid>:<collaborator ID>")
	}
	appID := ids[0]
	collaboratorID := ids[1] // can be either the email address or the ID

	collaborators, err := client.CollaboratorsList(appID)
	if err != nil {
		return nil, err
	}

	for _, collaborator := range collaborators {
		if collaborator.Email == collaboratorID || collaborator.ID == collaboratorID {
			d.SetId(collaborator.ID)
			err = d.Set("app", appID)
			if err != nil {
				return nil, err
			}
			return []*schema.ResourceData{d}, nil
		}
	}

	return nil, errors.New("not found")
}
