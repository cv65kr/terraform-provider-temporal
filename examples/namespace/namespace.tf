provider "temporal" {
}

resource "temporal_namespace" "ns_test1" {
    name = "test1"
	escription = "Test namespace 1"
	owner_email = "test@example.com"
	retention = "240"
}