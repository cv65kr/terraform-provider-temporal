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
				Default:     "127.0.0.1:7233",
				Description: "Host and port for the Temporal Frontend Service",
			},
			"tls_ca_path": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Path to a server Certificate Authority (CA) certificate file",
			},
			"tls_cert_path": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Path to a public X.509 certificate file for mutual TLS authentication",
			},
			"tls_key_path": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Path to a private key file for mutual TLS authentication",
			},
			"tls_server_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the target server that is used for TLS host verification",
			},
			"tls_disable_host_verification": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Disable verification of the server certificate (and thus host verification)",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"temporal_namespace": resourceNamespace(),
		},
		DataSourcesMap:       map[string]*schema.Resource{},
		ConfigureContextFunc: ClientConfigurer,
	}
}
