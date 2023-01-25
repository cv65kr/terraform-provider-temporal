provider "temporal" {
    address = "localhost:7233"
    tls_ca_path = "./xyz.ca"
    tls_cert_path = "./xyz.cert"
    tls_key_path = "./xyz.key"
    tls_disable_host_verification = false
    tls_server_name = "name"
}