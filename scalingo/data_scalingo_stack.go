package scalingo

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v4"
)

func dataSourceScStack() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScStackRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"base_image": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceScStackRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*scalingo.Client)

	name, ok := d.Get("name").(string)
	if !ok || name == "" {
		return diag.Errorf("name attribute is mendatory")
	}

	stacks, err := client.StacksList()
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
	d.Set("name", selected.Name)
	d.Set("description", selected.Description)
	d.Set("base_image", selected.BaseImage)
	d.Set("default", selected.Default)

	return nil
}
