package scalingo

import (
	scalingo "github.com/Scalingo/go-scalingo"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceScalingoDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceDomainCreate,
		Read:   resourceDomainRead,
		Delete: resourceDomainDelete,

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
		return err
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
