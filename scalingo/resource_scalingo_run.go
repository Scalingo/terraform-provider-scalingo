package scalingo

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	scalingo "github.com/Scalingo/go-scalingo/v4"
)

func resourceScalingoRun() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRunCreate,
		Read:          schema.Noop,
		Delete:        schema.Noop,
		Schema: map[string]*schema.Schema{
			"app": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"command": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
				ForceNew: true,
			},
			"detached": {
				Type:     schema.TypeBool,
				ForceNew: true,
				Optional: true,
				Default:  false,
			},
			"output": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceRunCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	command := stringListToStringSlice(d.Get("command").([]interface{}))
	detached, _ := d.Get("detached").(bool)

	res, err := client.Run(scalingo.RunOpts{
		Command:  command,
		App:      d.Get("app").(string),
		Detached: detached,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if detached {
		err = d.Set("output", "")
		if err != nil {
			return diag.Errorf("fail to reset run output: %v", err)
		}
		return nil
	}

	// If the container is attached, open a WS connection to get the command output
	token, err := client.GetAccessToken()
	if err != nil {
		return diag.FromErr(err)
	}

	req, err := http.NewRequest("CONNECT", res.AttachURL, nil)
	if err != nil {
		return diag.FromErr(err)
	}

	req.SetBasicAuth("", token)

	url, err := url.Parse(res.AttachURL)
	if err != nil {
		return diag.FromErr(err)
	}

	dial, err := net.Dial("tcp", url.Host)
	if err != nil {
		return diag.FromErr(err)
	}

	// This code should be refactored to use http.Transport since httputil.ClientConn was deprecated
	var conn *httputil.ClientConn //nolint
	if url.Scheme == "https" {
		host := strings.Split(url.Host, ":")[0]
		config := tls.Config{}
		config.ServerName = host
		tlsConn := tls.Client(dial, &config)
		conn = httputil.NewClientConn(tlsConn, nil) // nolint
	} else if url.Scheme == "http" {
		conn = httputil.NewClientConn(dial, nil) // nolint
	} else {
		return diag.Errorf("Invalid scheme format %s", url.Scheme)
	}

	resp, err := conn.Do(req)
	if err != nil && err != httputil.ErrPersistEOF { //nolint
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	connection, _ := conn.Hijack()

	output, err := io.ReadAll(connection)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("output", string(output))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(d.Get("app").(string) + strings.Join(command, ","))

	return nil
}

func stringListToStringSlice(stringList []interface{}) []string {
	ret := []string{}
	for _, v := range stringList {
		if v == nil {
			ret = append(ret, "")
			continue
		}
		ret = append(ret, v.(string))
	}
	return ret
}
