package scalingo

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	scalingo "github.com/Scalingo/go-scalingo"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceScalingoRun() *schema.Resource {
	return &schema.Resource{
		Create: resourceRunCreate,
		Read:   schema.Noop,
		Delete: schema.Noop,
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

func resourceRunCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)

	command := stringListToStringSlice(d.Get("command").([]interface{}))
	detached := d.Get("detached").(bool)

	res, err := client.Run(scalingo.RunOpts{
		Command:  command,
		App:      d.Get("app").(string),
		Detached: detached,
	})

	if err != nil {
		return err
	}

	if detached {
		d.Set("output", "")
		return nil
	}

	// If the container is attached, open a WS connection to get the command output
	token, err := client.GetAccessToken()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("CONNECT", res.AttachURL, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth("", token)

	url, err := url.Parse(res.AttachURL)
	if err != nil {
		return err
	}

	dial, err := net.Dial("tcp", url.Host)
	if err != nil {
		return err
	}

	var conn *httputil.ClientConn
	if url.Scheme == "https" {
		host := strings.Split(url.Host, ":")[0]
		config := tls.Config{}
		config.ServerName = host
		tls_conn := tls.Client(dial, &config)
		conn = httputil.NewClientConn(tls_conn, nil)
	} else if url.Scheme == "http" {
		conn = httputil.NewClientConn(dial, nil)
	} else {
		return fmt.Errorf("Invalid scheme format %s", url.Scheme)
	}

	_, err = conn.Do(req)
	if err != nil && err != httputil.ErrPersistEOF {
		return err
	}

	connection, _ := conn.Hijack()

	output, err := ioutil.ReadAll(connection)
	if err != nil {
		return err
	}

	d.Set("output", string(output))

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
