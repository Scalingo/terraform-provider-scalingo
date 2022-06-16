package scalingo

import (
	"time"

	"gopkg.in/errgo.v1"

	httpclient "github.com/Scalingo/go-scalingo/v4/http"
)

type Stack struct {
	ID          string    `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	BaseImage   string    `json:"base_image"`
	Default     bool      `json:"default"`
}

type StacksService interface {
	StacksList() ([]Stack, error)
}

var _ StacksService = (*Client)(nil)

func (c *Client) StacksList() ([]Stack, error) {
	req := &httpclient.APIRequest{
		Endpoint: "/features/stacks",
	}

	resmap := map[string][]Stack{}
	err := c.ScalingoAPI().DoRequest(req, &resmap)
	if err != nil {
		return nil, errgo.Notef(err, "fail to request Scalingo API")
	}
	return resmap["stacks"], nil
}
