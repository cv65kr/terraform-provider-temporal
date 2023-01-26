provider "temporal" {
}

resource "temporal_namespace" "ns_test1" {
    name = "test1"
	escription = "Test namespace 1"
	owner_email = "test@example.com"
	retention = "240"

	active_cluster = "test"
	clusters = ["c1", "c2"]
	history_archival_state = false
	history_uri = "https://test.com"
	namespace_data = {
		"k1": "v1",
		"k2": "v2"
	}
	visibility_archival_state = false
	visibility_uri = "https://test2.com"
}