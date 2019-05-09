package scalingo

import (
	"errors"
	"strings"

	scalingo "github.com/Scalingo/go-scalingo"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceScalingoCollaborator() *schema.Resource {
	return &schema.Resource{
		Create: resourceCollaboratorCreate,
		Read:   resourceCollaboratorRead,
		Delete: resourceCollaboratorDelete,

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
			State: resourceCollaboratorImport,
		},
	}
}

func resourceCollaboratorCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)

	collaborator, err := client.CollaboratorAdd(d.Get("app").(string), d.Get("email").(string))
	if err != nil {
		return err
	}

	d.Set("username", collaborator.Username)
	d.Set("status", collaborator.Status)

	d.SetId(collaborator.ID)

	return nil
}

func resourceCollaboratorRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)

	collaborators, err := client.CollaboratorsList(d.Get("app").(string))
	if err != nil {
		return err
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

	d.Set("username", collaborator.Username)
	d.Set("email", collaborator.Email)
	d.Set("status", collaborator.Status)

	return nil
}

func resourceCollaboratorDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)

	err := client.CollaboratorRemove(d.Get("app").(string), d.Id())
	if err != nil {
		return err
	}

	return nil
}

func resourceCollaboratorImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*scalingo.Client)

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
			d.Set("app", appID)
			return []*schema.ResourceData{d}, nil
		}
	}

	return nil, errors.New("not found")
}
