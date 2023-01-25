package temporal

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"address": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "localhost:7233",
				Description: "Host and port for the Temporal Frontend Service",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"temporal_namespace": resourceNamespace(),
		},
		DataSourcesMap:       map[string]*schema.Resource{},
		ConfigureContextFunc: ClientConfigurer,
	}
}
