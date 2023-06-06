package scalingo

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v6"
)

func resourceScalingoScmRepoLink() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScmRepoLinkCreate,
		ReadContext:   resourceScmRepoLinkRead,
		UpdateContext: resourceScmRepoLinkUpdate,
		DeleteContext: resourceScmRepoLinkDelete,
		Description:   "Resource SCM Repo Link representing a link between an application and a repository of a SCM",

		Schema: map[string]*schema.Schema{
			"app": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID or slug name of the targeted application",
			},
			"auth_integration_uuid": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "SCM Integration UUID to base the link on",
			},
			"source": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "URL to the SCM repository, example: https://github.com/user/repository",
			},
			"branch": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Branch to use for autodeploy",
			},
			"auto_deploy_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Is auto-deploy enabled?",
			},
			"deploy_review_apps_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Is automatic review apps creation enabled?",
			},
			"delete_on_close_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Should review apps be deleted when Pull/Merge Requests are closed?",
			},
			"hours_before_delete_on_close": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Additional delay before deleting a review app once a Pull/Merge Request has been closed",
			},
			"delete_stale_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Should a review app be deleted if attached Pull/Merge Request is considered as stale?",
			},
			"hours_before_delete_stale": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "After how many hours should a Pull/Merge Request be considered as stale",
			},
			"automatic_creation_from_forks_allowed": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Should review apps be created automatically if a Pull/Merge Request is based on the branch of a fork",
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
	authIntegrationUUID, _ := d.Get("auth_integration_uuid").(string)

	if branch == "" && autoDeployEnabled {
		return diag.Errorf("branch must be set when auto_deploy_enabled is enabled")
	}

	deployReviewAppsEnabled, _ := d.Get("deploy_review_apps_enabled").(bool)
	deleteOnCloseEnabled, _ := d.Get("delete_on_close_enabled").(bool)
	hoursBeforeDeleteOnClose, _ := d.Get("hours_before_delete_on_close").(int)
	deleteStaleEnabled, _ := d.Get("delete_stale_enabled").(bool)
	hoursBeforeDeleteStale, _ := d.Get("hours_before_delete_stale").(int)
	automaticCreationFromForksAllowed, _ := d.Get("automatic_creation_from_forks_allowed").(bool)

	if hoursBeforeDeleteOnClose < 0 || hoursBeforeDeleteStale < 0 {
		return diag.Errorf("hours must be an unsigned int")
	}
	hoursBeforeDeleteOnCloseUint := uint(hoursBeforeDeleteOnClose)
	hoursBeforeDeleteStaleUint := uint(hoursBeforeDeleteStale)

	params := scalingo.SCMRepoLinkCreateParams{
		Source:                            &source,
		AutoDeployEnabled:                 &autoDeployEnabled,
		DeployReviewAppsEnabled:           &deployReviewAppsEnabled,
		AuthIntegrationUUID:               &authIntegrationUUID,
		Branch:                            &branch,
		AutomaticCreationFromForksAllowed: &automaticCreationFromForksAllowed,
		DestroyOnCloseEnabled:             &deleteOnCloseEnabled,
		HoursBeforeDeleteOnClose:          &hoursBeforeDeleteOnCloseUint,
		DestroyStaleEnabled:               &deleteStaleEnabled,
		HoursBeforeDeleteStale:            &hoursBeforeDeleteStaleUint,
	}

	link, err := client.SCMRepoLinkCreate(ctx, app, params)
	if err != nil {
		return diag.Errorf("fail to add SCM repo link: %v", err)
	}

	d.SetId(link.AppID)

	return nil
}

func resourceScmRepoLinkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	app, _ := d.Get("app").(string)

	changed := false
	branch, _ := d.Get("branch").(string)
	autoDeployEnabled, _ := d.Get("auto_deploy_enabled").(bool)

	if branch == "" && autoDeployEnabled {
		return diag.Errorf("branch must be set when auto_deploy_enabled is enabled")
	}

	params := scalingo.SCMRepoLinkUpdateParams{}
	if d.HasChange("branch") {
		params.Branch = &branch
		changed = true
	}

	if d.HasChange("auto_deploy_enabled") {
		params.AutoDeployEnabled = &autoDeployEnabled
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

	if d.HasChange("automatic_creation_from_forks_allowed") {
		params.AutomaticCreationFromForksAllowed = boolAddr(d.Get("automatic_creation_from_forks_allowed").(bool))
		changed = true
	}

	if changed {
		link, err := client.SCMRepoLinkUpdate(ctx, app, params)
		if err != nil {
			return diag.Errorf("fail to update github repo link: %v", err)
		}
		err = SetAll(d, map[string]interface{}{
			"branch":                                link.Branch,
			"auto_deploy_enabled":                   link.AutoDeployEnabled,
			"deploy_review_apps_enabled":            link.DeployReviewAppsEnabled,
			"delete_on_close_enabled":               link.DeleteOnCloseEnabled,
			"delete_stale_enabled":                  link.DeleteStaleEnabled,
			"automatic_creation_from_forks_allowed": link.AutomaticCreationFromForksAllowed,
			"hours_before_delete_on_close":          int(link.HoursBeforeDeleteOnClose),
			"hours_before_delete_stale":             int(link.HoursBeforeDeleteStale),
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
	d.SetId(link.AppID)

	err = SetAll(d, map[string]interface{}{
		"app":                                   link.AppID,
		"auto_deploy_enabled":                   link.AutoDeployEnabled,
		"deploy_review_apps_enabled":            link.DeployReviewAppsEnabled,
		"delete_on_close_enabled":               link.DeleteOnCloseEnabled,
		"delete_stale_enabled":                  link.DeleteStaleEnabled,
		"automatic_creation_from_forks_allowed": link.AutomaticCreationFromForksAllowed,
		"hours_before_delete_on_close":          int(link.HoursBeforeDeleteOnClose),
		"hours_before_delete_stale":             int(link.HoursBeforeDeleteStale),
		"branch":                                link.Branch,
		"auth_integration_uuid":                 link.AuthIntegrationUUID,
		"source":                                fmt.Sprintf("%s/%s/%s", link.URL, link.Owner, link.Repo),
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
