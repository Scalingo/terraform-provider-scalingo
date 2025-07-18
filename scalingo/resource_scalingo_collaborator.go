package scalingo

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v8"
)

func resourceScalingoCollaborator() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCollaboratorCreate,
		ReadContext:   resourceCollaboratorRead,
		DeleteContext: resourceCollaboratorDelete,
		UpdateContext: resourceCollaboratorUpdate,
		Description:   "Resource representing a collaboration between a user and an application",

		Schema: map[string]*schema.Schema{
			"app": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the targeted application",
			},
			"email": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Email of the collaborator to invite",
			},
			"username": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Username of the attached account once the collaboration has been accepted",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the collaboration (pending/accepted)",
			},
			"limited": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether the collaborator is a limited collaborator for the application",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceCollaboratorImport,
		},
	}
}

func resourceCollaboratorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	collaborator, err := client.CollaboratorAdd(ctx, d.Get("app").(string), scalingo.CollaboratorAddParams{d.Get("email").(string), false})
	if err != nil {
		return diag.Errorf("fail to add collaborator: %v", err)
	}

	d.SetId(collaborator.ID)

	err = SetAll(d, map[string]interface{}{
		"username": collaborator.Username,
		"status":   collaborator.Status,
		"limited":  collaborator.IsLimited,
	})
	if err != nil {
		return diag.Errorf("fail to store collaborator information: %v", err)
	}

	return nil
}

func resourceCollaboratorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	collaborators, err := client.CollaboratorsList(ctx, d.Get("app").(string))
	if err != nil {
		return diag.Errorf("fail to list collaborators: %v", err)
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
		// Flag that it doesn't exist anymore
		d.SetId("")
		return nil
	}

	err = SetAll(d, map[string]interface{}{
		"username": collaborator.Username,
		"email":    collaborator.Email,
		"status":   collaborator.Status,
		"limited":  collaborator.IsLimited,
	})
	if err != nil {
		return diag.Errorf("fail to store collaborator information: %v", err)
	}

	return nil
}

func resourceCollaboratorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	err := client.CollaboratorRemove(ctx, d.Get("app").(string), d.Id())
	if err != nil {
		return diag.Errorf("fail to remove collaborator: %v", err)
	}

	return nil
}

func resourceCollaboratorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	collaborator, err := client.CollaboratorUpdate(ctx, d.Get("app").(string), d.Id(), scalingo.CollaboratorUpdateParams{d.Get("limited").(bool)})
	if err != nil {
		return diag.Errorf("fail to update collaborator: %v", err)
	}

	d.SetId(collaborator.ID)

	err = SetAll(d, map[string]interface{}{
		"username": collaborator.Username,
		"status":   collaborator.Status,
		"limited":  collaborator.IsLimited,
	})
	if err != nil {
		return diag.Errorf("fail to store collaborator information: %v", err)
	}

	return nil
}

func resourceCollaboratorImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client, _ := meta.(*scalingo.Client)

	ids := strings.Split(d.Id(), ":")
	if len(ids) != 2 {
		return nil, fmt.Errorf("address should have the following format: <appid>:<collaborator ID>")
	}
	appID := ids[0]
	collaboratorID := ids[1] // can be either the email address or the ID

	collaborators, err := client.CollaboratorsList(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("fail to list collaborators: %v", err)
	}

	for _, collaborator := range collaborators {
		if collaborator.Email == collaboratorID || collaborator.ID == collaboratorID {
			d.SetId(collaborator.ID)
			err = d.Set("app", appID)
			if err != nil {
				return nil, fmt.Errorf("fail to store app id: %v", err)
			}
			return []*schema.ResourceData{d}, nil
		}
	}

	return nil, fmt.Errorf("not found")
}
