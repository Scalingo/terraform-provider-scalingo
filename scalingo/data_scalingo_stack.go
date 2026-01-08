package scalingo

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v9"
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
			"deprecated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "When has been/will be deprecated the stack",
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

	stackFields := map[string]interface{}{
		"name":        selected.Name,
		"description": selected.Description,
		"base_image":  selected.BaseImage,
		"default":     selected.Default,
	}
	tflog.Info(ctx, fmt.Sprintf("Fetched stack '%v' with ID %v", name, selected.ID), stackFields)

	d.SetId(selected.ID)
	err = SetAll(d, stackFields)
	if err != nil {
		return diag.Errorf("fail to save stack information: %v", err)
	}

	return nil
}
