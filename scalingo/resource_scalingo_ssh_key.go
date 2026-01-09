package scalingo

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v9"
)

func resourceScalingoSSHKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSSHKeyCreate,
		ReadContext:   resourceSSHKeyRead,
		DeleteContext: resourceSSHKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceSSHKeyImporter,
		},
		Description: "Resource representing a SSH Key used for git operations authentication",

		Schema: map[string]*schema.Schema{
			"key_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the SSH Key",
			},
			"public_key": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Sensitive:   true,
				Description: "Content of the public SSH Key",
			},
		},
	}
}

func resourceSSHKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	keyName, _ := d.Get("key_name").(string)
	keyContent, _ := d.Get("public_key").(string)

	sshKey, err := client.KeysAdd(ctx, keyName, keyContent)
	if err != nil {
		return diag.Errorf("fail to add ssh key: %v", err)
	}
	d.SetId(sshKey.ID)

	return nil
}

func resourceSSHKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	keyID := d.Id()

	keysList, err := client.KeysList(ctx)
	if err != nil {
		return diag.Errorf("fail to get list of ssh keys: %v", err)
	}
	filteredKeys := keepIf(keysList, func(k scalingo.Key) bool {
		return k.ID == keyID
	})

	if len(filteredKeys) != 1 {
		return diag.Errorf("fail to find the selected ssh key")
	}

	sshKey := filteredKeys[0]
	d.SetId(sshKey.ID)
	err = SetAll(d, map[string]interface{}{
		"key_name":   sshKey.Name,
		"public_key": sshKey.Content,
	})
	if err != nil {
		return diag.Errorf("fail to store ssh key: %v", err)
	}

	return nil
}

func resourceSSHKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	keyID := d.Id()

	err := client.KeysDelete(ctx, keyID)
	if err != nil {
		return diag.Errorf("fail to remove ssh key: %v", err)
	}

	return nil
}

func resourceSSHKeyImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	diags := resourceSSHKeyRead(ctx, d, meta)
	err := DiagnosticError(diags)
	if err != nil {
		return nil, fmt.Errorf("fail to read ssh key: %v", err)
	}

	return []*schema.ResourceData{d}, nil
}
