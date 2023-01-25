package temporal

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkclient "go.temporal.io/sdk/client"
	"go.temporal.io/server/common/auth"
)

func ClientConfigurer(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	hostPort := d.Get("address").(string)

	tls, err := createTLSConfig(d)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	clientOptions := sdkclient.Options{
		HostPort: hostPort,
		ConnectionOptions: sdkclient.ConnectionOptions{
			TLS: tls,
		},
	}

	return NewClient(clientOptions), nil
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

// Ported from https://github.com/temporalio/tctl/blob/main/cli_curr/factory.go
func createTLSConfig(d *schema.ResourceData) (*tls.Config, error) {
	hostPort := d.Get("address").(string)
	caPath := d.Get("tls_ca_path").(string)
	certPath := d.Get("tls_cert_path").(string)
	keyPath := d.Get("tls_key_path").(string)
	disableHostNameVerification := d.Get("tls_disable_host_verification").(bool)
	serverName := d.Get("tls_server_name").(string)

	var host string
	var cert *tls.Certificate
	var caPool *x509.CertPool

	if caPath != "" {
		caCertPool, err := fetchCACert(caPath)
		if err != nil {
			return nil, err
		}
		caPool = caCertPool
	}

	if certPath != "" {
		myCert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return nil, err
		}
		cert = &myCert
	}

	if caPool != nil || cert != nil {
		if serverName != "" {
			host = serverName
		} else {
			host, _, _ = net.SplitHostPort(hostPort)
		}

		tlsConfig := auth.NewTLSConfigForServer(host, !disableHostNameVerification)
		if caPool != nil {
			tlsConfig.RootCAs = caPool
		}

		if cert != nil {
			tlsConfig.Certificates = []tls.Certificate{*cert}
		}

		return tlsConfig, nil
	}

	if serverName != "" {
		host = serverName
		tlsConfig := auth.NewTLSConfigForServer(host, !disableHostNameVerification)
		return tlsConfig, nil
	}

	return nil, nil
}

// Ported from https://github.com/temporalio/tctl/blob/main/cli_curr/factory.go
func fetchCACert(pathOrUrl string) (caPool *x509.CertPool, err error) {
	caPool = x509.NewCertPool()
	var caBytes []byte

	if strings.HasPrefix(pathOrUrl, "http://") {
		return nil, errors.New("HTTP is not supported for CA cert URLs. Provide HTTPS URL")
	}

	if strings.HasPrefix(pathOrUrl, "https://") {
		resp, err := http.Get(pathOrUrl)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		caBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
	} else {
		caBytes, err = os.ReadFile(pathOrUrl)
		if err != nil {
			return nil, err
		}
	}

	if !caPool.AppendCertsFromPEM(caBytes) {
		return nil, errors.New("unknown failure constructing cert pool for ca")
	}

	return caPool, nil
}
