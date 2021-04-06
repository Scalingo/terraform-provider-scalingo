package scalingo

import (
	"errors"

	scalingo "github.com/Scalingo/go-scalingo"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"gopkg.in/errgo.v1"
)

func resourceScalingoGithubLink() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubLinkCreate,
		Read:   resourceGithubLinkRead,
		Update: resourceGithubLinkUpdate,
		Delete: resourceGithubLinkDelete,

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
			State: resourceGithubLinkImport,
		},
	}
}

func resourceGithubLinkCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)

	app := d.Get("app").(string)
	source := d.Get("source").(string)
	autoDeploy := d.Get("auto_deploy").(bool)
	deployOnBranchChange := d.Get("deploy_on_branch_change").(bool)
	branch := d.Get("branch").(string)

	if len(branch) == 0 && (deployOnBranchChange || autoDeploy) {
		return errors.New("Branch must be set when deploy_on_branch_change or auto_deploy is enabled")
	}

	reviewApps := d.Get("review_apps").(bool)
	destroyReviewAppOnClose := d.Get("destroy_review_app_on_close").(bool)
	destroyStaledReviewApp := d.Get("destroy_stale_review_app").(bool)
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
		return err
	}

	if deployOnBranchChange {
		err := client.GithubLinkManualDeploy(app, link.ID, branch)
		if err != nil {
			d.Set("deploy_on_branch_change", false)
			return err
		}
	}

	d.SetId(link.ID)

	return nil
}
func resourceGithubLinkUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)

	app := d.Get("app").(string)

	changed := false
	params := scalingo.GithubLinkParams{}
	branch := d.Get("branch").(string)
	autoDeploy := d.Get("auto_deploy").(bool)
	deployOnBranchChange := d.Get("deploy_on_branch_change").(bool)

	if len(branch) == 0 && (deployOnBranchChange || autoDeploy) {
		return errors.New("Branch must be set when deploy_on_branch_change or auto_deploy is enabled")
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

	d.Partial(true)

	if (d.HasChange("branch") || d.HasChange("deploy_on_branch_change")) && deployOnBranchChange {
		err := client.GithubLinkManualDeploy(app, d.Id(), branch)
		if err != nil {
			d.Set("deploy_on_branch_change", false)
			return err
		}
		d.Set("branch", branch)
		d.SetPartial("branch")
	}

	if changed {
		link, err := client.GithubLinkUpdate(app, d.Id(), params)
		if err != nil {
			return err
		}

		d.Set("branch", link.GithubBranch)
		d.Set("auto_deploy", link.AutoDeployEnabled)
		d.Set("review_apps", link.DeployReviewAppsEnabled)
		d.Set("destroy_review_app_on_close", link.DestroyOnCloseEnabled)
		d.Set("destroy_stale_review_app", link.DestroyOnStaleEnabled)
		d.Set("destroy_closed_review_app_after", int(link.HoursBeforeDeleteOnClose))
		d.Set("destroy_stale_review_app_after", int(link.HoursBeforeDeleteStale))
		d.SetPartial("branch")
		d.SetPartial("auto_deploy")
		d.SetPartial("review_apps")
		d.SetPartial("destroy_review_app_on_close")
		d.SetPartial("destroy_stale_review_app")
		d.SetPartial("destroy_closed_review_app_after")
		d.SetPartial("destroy_stale_review_app_after")
	}
	d.Partial(false)

	return nil
}
func resourceGithubLinkRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)
	app := d.Get("app").(string)

	link, err := client.GithubLinkShow(app)
	if err != nil {
		return errgo.Notef(err, "error when fetching github repo link for app %v", app)
	}
	d.SetId(link.ID)
	d.Set("auto_deploy", link.AutoDeployEnabled)
	d.Set("review_apps", link.DeployReviewAppsEnabled)
	d.Set("destroy_review_app_on_close", link.DestroyOnCloseEnabled)
	d.Set("destroy_stale_review_app", link.DestroyOnStaleEnabled)
	d.Set("destroy_closed_review_app_after", int(link.HoursBeforeDeleteOnClose))
	d.Set("destroy_stale_review_app_after", int(link.HoursBeforeDeleteStale))
	d.Set("branch", link.GithubBranch)
	d.Set("source", link.GithubSource)

	return nil
}
func resourceGithubLinkDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)
	app := d.Get("app").(string)

	err := client.GithubLinkDelete(app, d.Id())
	if err != nil {
		return err
	}

	return nil
}

func resourceGithubLinkImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set("app", d.Id())

	return []*schema.ResourceData{d}, nil
}
