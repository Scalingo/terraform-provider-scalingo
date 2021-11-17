package scalingo

import (
	"errors"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"gopkg.in/errgo.v1"

	scalingo "github.com/Scalingo/go-scalingo"
)

func resourceScalingoDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceDomainCreate,
		Read:   resourceDomainRead,
		Delete: resourceDomainDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDomainImporter,
		},

		Schema: map[string]*schema.Schema{
			"common_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"app": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceDomainCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)

	domainName := d.Get("common_name").(string)
	appId := d.Get("app").(string)

	domain, err := client.DomainsAdd(appId, scalingo.Domain{
		Name: domainName,
	})
	if err != nil {
		return err
	}
	d.SetId(domain.ID)

	return nil
}

func resourceDomainRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)

	appId := d.Get("app").(string)

	domain, err := client.DomainsShow(appId, d.Id())
	if err != nil {
		return errgo.Notef(err, "fail to get domain %v of app %v", d.Id(), appId)
	}

	d.SetId(domain.ID)
	d.Set("common_name", domain.Name)

	return nil
}

func resourceDomainDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)

	appId := d.Get("app").(string)

	err := client.DomainsRemove(appId, d.Id())
	if err != nil {
		return err
	}

	return nil
}

func resourceDomainImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if !strings.Contains(d.Id(), ":") {
		return nil, errors.New("schema must be app_id:domain_id")
	}
	split := strings.Split(d.Id(), ":")
	d.Set("app", split[0])
	d.SetId(split[1])

	return []*schema.ResourceData{d}, nil
}
