package temporal

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkclient "go.temporal.io/sdk/client"
)

func ClientConfigurer(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	hostPort := d.Get("address").(string)

	// @TODO TLS support

	options := sdkclient.Options{
		HostPort: hostPort,
	}

	return NewClient(options), nil
}

type Client struct {
	options sdkclient.Options
}

func NewClient(options sdkclient.Options) Client {
	return Client{
		options: options,
	}
}

func (c Client) NamespaceClient() (sdkclient.NamespaceClient, error) {
	var namespaceClient sdkclient.NamespaceClient
	namespaceClient, err := sdkclient.NewNamespaceClient(c.options)
	if err != nil {
		return nil, err
	}

	return namespaceClient, nil
}
