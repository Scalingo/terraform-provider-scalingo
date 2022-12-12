package scalingo

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/Scalingo/go-scalingo/v6"
)

func dataSourceScStack() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScStackRead,
		Description: "List of available stacks to base applications on",

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Slug name of the stack (scalingo-18, scalingo-20, …)",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Textual description of the stack",
			},
			"base_image": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Base docker image on which is based the stack",
			},
			"default": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Is it the current default stack?",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the stack",
			},
			"deprecated_at": {
				Type:         schema.TypeString,
				Computed:     true,
				ValidateFunc: validation.IsRFC3339Time,
				Description:  "When has been/will be deprecated the stack",
			},
		},
	}
}

func dataSourceScStackRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	name, ok := d.Get("name").(string)
	if !ok || name == "" {
		return diag.Errorf("name attribute is mandatory")
	}

	stacks, err := client.StacksList(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	var selected *scalingo.Stack

	for _, s := range stacks {
		if s.Name == name {
			selected = &s
			break
		}
	}

	if selected == nil {
		return diag.Errorf("fail to find stack with name '%s'", name)
	}

	d.SetId(selected.ID)
	err = SetAll(d, map[string]interface{}{
		"name":        selected.Name,
		"description": selected.Description,
		"base_image":  selected.BaseImage,
		"default":     selected.Default,
	})
	if err != nil {
		return diag.Errorf("fail to save stack information: %v", err)
	}

	return nil
}
