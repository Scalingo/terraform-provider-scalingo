package scalingo

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	scalingo "github.com/Scalingo/go-scalingo/v4"
)

func resourceScalingoGithubLink() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGithubLinkCreate,
		ReadContext:   resourceGithubLinkRead,
		UpdateContext: resourceGithubLinkUpdate,
		DeleteContext: resourceGithubLinkDelete,

		Schema: map[string]*schema.Schema{
			"app": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"source": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"deploy_on_branch_change": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"auto_deploy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"branch": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"review_apps": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"destroy_review_app_on_close": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"destroy_stale_review_app": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"destroy_closed_review_app_after": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"destroy_stale_review_app_after": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceGithubLinkImport,
		},
	}
}

func resourceGithubLinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	app, _ := d.Get("app").(string)
	source, _ := d.Get("source").(string)
	autoDeploy, _ := d.Get("auto_deploy").(bool)
	deployOnBranchChange, _ := d.Get("deploy_on_branch_change").(bool)
	branch, _ := d.Get("branch").(string)

	if branch == "" && (deployOnBranchChange || autoDeploy) {
		return diag.Errorf("Branch must be set when deploy_on_branch_change or auto_deploy is enabled")
	}

	reviewApps, _ := d.Get("review_apps").(bool)
	destroyReviewAppOnClose, _ := d.Get("destroy_review_app_on_close").(bool)
	destroyStaledReviewApp, _ := d.Get("destroy_stale_review_app").(bool)
	destroyClosedReviewAppAfter := uint(d.Get("destroy_closed_review_app_after").(int))
	destroyStaleReviewAppAfter := uint(d.Get("destroy_stale_review_app_after").(int))

	params := scalingo.GithubLinkParams{
		GithubSource:            &source,
		AutoDeployEnabled:       &autoDeploy,
		DeployReviewAppsEnabled: &reviewApps,
	}

	if autoDeploy {
		params.GithubBranch = &branch
	}

	if reviewApps {
		if destroyReviewAppOnClose {
			params.DestroyOnCloseEnabled = &destroyReviewAppOnClose
			params.HoursBeforeDeleteOnClose = &destroyClosedReviewAppAfter
		}
		if destroyStaledReviewApp {
			params.DestroyStaleEnabled = &destroyStaledReviewApp
			params.HoursBeforeDeleteStale = &destroyStaleReviewAppAfter
		}
	}

	link, err := client.GithubLinkAdd(app, params)
	if err != nil {
		return diag.FromErr(err)
	}

	if deployOnBranchChange {
		err := client.GithubLinkManualDeploy(app, link.ID, branch)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(link.ID)

	return nil
}
func resourceGithubLinkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	app, _ := d.Get("app").(string)

	changed := false
	params := scalingo.GithubLinkParams{}
	branch, _ := d.Get("branch").(string)
	autoDeploy, _ := d.Get("auto_deploy").(bool)
	deployOnBranchChange, _ := d.Get("deploy_on_branch_change").(bool)

	if branch == "" && (deployOnBranchChange || autoDeploy) {
		return diag.Errorf("Branch must be set when deploy_on_branch_change or auto_deploy is enabled")
	}

	if d.HasChange("branch") {
		params.GithubBranch = &branch
		changed = true
	}

	if d.HasChange("auto_deploy") {
		params.AutoDeployEnabled = &autoDeploy
		changed = true
	}

	if d.HasChange("review_apps") {
		params.DeployReviewAppsEnabled = boolAddr(d.Get("review_apps").(bool))
		changed = true
	}

	if d.HasChange("destroy_review_app_on_close") {
		params.DestroyOnCloseEnabled = boolAddr(d.Get("destroy_review_app_on_close").(bool))
		changed = true
	}

	if d.HasChange("destroy_stale_review_app") {
		params.DestroyStaleEnabled = boolAddr(d.Get("destroy_stale_review_app").(bool))
		changed = true
	}

	if d.HasChange("destroy_closed_review_app_after") {
		params.HoursBeforeDeleteOnClose = uintAddr(uint(d.Get("destroy_closed_review_app_after").(int)))
		changed = true
	}

	if d.HasChange("destroy_stale_review_app_after") {
		params.HoursBeforeDeleteStale = uintAddr(uint(d.Get("destroy_stale_review_app_after").(int)))
		changed = true
	}

	if (d.HasChange("branch") || d.HasChange("deploy_on_branch_change")) && deployOnBranchChange {
		err := client.GithubLinkManualDeploy(app, d.Id(), branch)
		if err != nil {
			return diag.FromErr(err)
		}
		err = d.Set("branch", branch)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if changed {
		link, err := client.GithubLinkUpdate(app, d.Id(), params)
		if err != nil {
			return diag.FromErr(err)
		}

		err = SetAll(d, map[string]interface{}{
			"branch":                          link.GithubBranch,
			"auto_deploy":                     link.AutoDeployEnabled,
			"review_apps":                     link.DeployReviewAppsEnabled,
			"destroy_review_app_on_close":     link.DestroyOnCloseEnabled,
			"destroy_stale_review_app":        link.DestroyOnStaleEnabled,
			"destroy_closed_review_app_after": int(link.HoursBeforeDeleteOnClose),
			"destroy_stale_review_app_after":  int(link.HoursBeforeDeleteStale),
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
func resourceGithubLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)
	app, _ := d.Get("app").(string)

	link, err := client.GithubLinkShow(app)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(link.ID)

	err = SetAll(d, map[string]interface{}{
		"auto_deploy":                     link.AutoDeployEnabled,
		"review_apps":                     link.DeployReviewAppsEnabled,
		"destroy_review_app_on_close":     link.DestroyOnCloseEnabled,
		"destroy_stale_review_app":        link.DestroyOnStaleEnabled,
		"destroy_closed_review_app_after": int(link.HoursBeforeDeleteOnClose),
		"destroy_stale_review_app_after":  int(link.HoursBeforeDeleteStale),
		"branch":                          link.GithubBranch,
		"source":                          link.GithubSource,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
func resourceGithubLinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)
	app, _ := d.Get("app").(string)

	err := client.GithubLinkDelete(app, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceGithubLinkImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	err := d.Set("app", d.Id())
	if err != nil {
		return nil, err
	}

	diags := resourceGithubLinkRead(ctx, d, meta)
	err = DiagnosticError(diags)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
