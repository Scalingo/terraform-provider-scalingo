package scalingo

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v5"
)

func resourceScalingoScmRepoLink() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScmRepoLinkCreate,
		ReadContext:   resourceScmRepoLinkRead,
		UpdateContext: resourceScmRepoLinkUpdate,
		DeleteContext: resourceScmRepoLinkDelete,

		Schema: map[string]*schema.Schema{
			"app": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"auth_integration_uuid": {
				Type:     schema.TypeString,
				Required: true,
			},
			"source": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"branch": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"auto_deploy_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"deploy_on_branch_change": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"deploy_review_apps_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"delete_on_close_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"hours_before_delete_on_close": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"delete_stale_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"hours_before_delete_stale": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceScmRepoLinkImport,
		},
	}
}

func resourceScmRepoLinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	app, _ := d.Get("app").(string)
	source, _ := d.Get("source").(string)
	branch, _ := d.Get("branch").(string)
	autoDeployEnabled, _ := d.Get("auto_deploy_enabled").(bool)
	deployOnBranchChange, _ := d.Get("deploy_on_branch_change").(bool)
	authIntegrationUUID, _ := d.Get("auth_integration_uuid").(string)

	if branch == "" && (deployOnBranchChange || autoDeployEnabled) {
		return diag.Errorf("branch must be set when deploy_on_branch_change or auto_deploy_enabled is enabled")
	}

	deployReviewAppsEnabled, _ := d.Get("deploy_review_apps_enabled").(bool)
	deleteOnCloseEnabled, _ := d.Get("delete_on_close_enabled").(bool)
	hoursBeforeDeleteOnClose, _ := d.Get("hours_before_delete_on_close").(int)
	deleteStaleEnabled, _ := d.Get("delete_stale_enabled").(bool)
	hoursBeforeDeleteStale, _ := d.Get("hours_before_delete_stale").(int)

	if hoursBeforeDeleteOnClose < 0 || hoursBeforeDeleteStale < 0 {
		return diag.Errorf("Hours must be an unsigned int")
	}
	hoursBeforeDeleteOnCloseUint := uint(hoursBeforeDeleteOnClose)
	hoursBeforeDeleteStaleUint := uint(hoursBeforeDeleteStale)

	params := scalingo.SCMRepoLinkCreateParams{
		Source:                  &source,
		AutoDeployEnabled:       &autoDeployEnabled,
		DeployReviewAppsEnabled: &deployReviewAppsEnabled,
		AuthIntegrationUUID:     &authIntegrationUUID,
		Branch:                  &branch,
	}

	if deployReviewAppsEnabled {
		if deleteOnCloseEnabled {
			params.DestroyOnCloseEnabled = &deleteOnCloseEnabled
			params.HoursBeforeDeleteOnClose = &hoursBeforeDeleteOnCloseUint
		}
		if deleteStaleEnabled {
			params.DestroyStaleEnabled = &deleteStaleEnabled
			params.HoursBeforeDeleteStale = &hoursBeforeDeleteStaleUint
		}
	}

	link, err := client.SCMRepoLinkCreate(ctx, app, params)
	if err != nil {
		return diag.Errorf("fail to add SCM repo link: %v", err)
	}

	if deployOnBranchChange {
		err := client.SCMRepoLinkManualDeploy(ctx, app, branch)
		if err != nil {
			return diag.Errorf("fail to trigger manual deploy: %v", err)
		}
	}

	d.SetId(link.ID)

	return nil
}

func resourceScmRepoLinkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	app, _ := d.Get("app").(string)

	changed := false
	branch, _ := d.Get("branch").(string)
	autoDeploy, _ := d.Get("auto_deploy_enabled").(bool)
	deployOnBranchChange, _ := d.Get("deploy_on_branch_change").(bool)

	if branch == "" && (deployOnBranchChange || autoDeploy) {
		return diag.Errorf("branch must be set when deploy_on_branch_change or auto_deploy_enabled is enabled")
	}

	params := scalingo.SCMRepoLinkUpdateParams{}
	if d.HasChange("branch") {
		params.Branch = &branch
		changed = true
	}

	if d.HasChange("auto_deploy_enabled") {
		params.AutoDeployEnabled = &autoDeploy
		changed = true
	}

	if d.HasChange("deploy_review_apps_enabled") {
		params.DeployReviewAppsEnabled = boolAddr(d.Get("deploy_review_apps_enabled").(bool))
		changed = true
	}

	if d.HasChange("delete_on_close_enabled") {
		params.DestroyOnCloseEnabled = boolAddr(d.Get("delete_on_close_enabled").(bool))
		changed = true
	}

	if d.HasChange("delete_stale_enabled") {
		params.DestroyStaleEnabled = boolAddr(d.Get("delete_stale_enabled").(bool))
		changed = true
	}

	if d.HasChange("hours_before_delete_on_close") {
		params.HoursBeforeDeleteOnClose = uintAddr(uint(d.Get("hours_before_delete_on_close").(int)))
		changed = true
	}

	if d.HasChange("hours_before_delete_stale") {
		params.HoursBeforeDeleteStale = uintAddr(uint(d.Get("hours_before_delete_stale").(int)))
		changed = true
	}

	if (d.HasChange("branch") || d.HasChange("deploy_on_branch_change")) && deployOnBranchChange {
		err := client.SCMRepoLinkManualDeploy(ctx, app, branch)
		if err != nil {
			return diag.Errorf("fail to trigger manual deploy: %v", err)
		}
		err = d.Set("branch", branch)
		if err != nil {
			return diag.Errorf("fail to store new branch name: %v", err)
		}
	}

	if changed {
		link, err := client.SCMRepoLinkUpdate(ctx, app, params)
		if err != nil {
			return diag.Errorf("fail to update github repo link: %v", err)
		}
		err = SetAll(d, map[string]interface{}{
			"branch":                       link.Branch,
			"auto_deploy_enabled":          link.AutoDeployEnabled,
			"deploy_review_apps_enabled":   link.DeployReviewAppsEnabled,
			"delete_on_close_enabled":      link.DeleteOnCloseEnabled,
			"delete_stale_enabled":         link.DeleteStaleEnabled,
			"hours_before_delete_on_close": int(link.HoursBeforeDeleteOnClose),
			"hours_before_delete_stale":    int(link.HoursBeforeDeleteStale),
		})
		if err != nil {
			return diag.Errorf("fail to store github repo link information: %v", err)
		}
	}
	return nil
}

func resourceScmRepoLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)
	app, _ := d.Get("app").(string)

	link, err := client.SCMRepoLinkShow(ctx, app)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(link.ID)

	err = SetAll(d, map[string]interface{}{
		"auto_deploy_enabled":          link.AutoDeployEnabled,
		"deploy_review_apps_enabled":   link.DeployReviewAppsEnabled,
		"delete_on_close_enabled":      link.DeleteOnCloseEnabled,
		"delete_stale_enabled":         link.DeleteStaleEnabled,
		"hours_before_delete_on_close": int(link.HoursBeforeDeleteOnClose),
		"hours_before_delete_stale":    int(link.HoursBeforeDeleteStale),
		"branch":                       link.Branch,
		"auth_integration_uuid":        link.AuthIntegrationUUID,
	})
	if err != nil {
		return diag.Errorf("fail to store scm repo link information: %v", err)
	}

	return nil
}

func resourceScmRepoLinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)
	app, _ := d.Get("app").(string)

	err := client.SCMRepoLinkDelete(ctx, app)
	if err != nil {
		return diag.Errorf("fail to delete scm repo link: %v", err)
	}

	return nil
}

func resourceScmRepoLinkImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	err := d.Set("app", d.Id())
	if err != nil {
		return nil, fmt.Errorf("fail to store app id: %v", err)
	}

	diags := resourceScmRepoLinkRead(ctx, d, meta)
	err = DiagnosticError(diags)
	if err != nil {
		return nil, fmt.Errorf("fail to read scm repo link: %v", err)
	}

	return []*schema.ResourceData{d}, nil
}
