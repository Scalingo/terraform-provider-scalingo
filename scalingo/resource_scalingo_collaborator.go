package scalingo

import (
	"errors"

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
		return errors.New("not found")
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
