package temporal

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/namespace/v1"
	replicationpb "go.temporal.io/api/replication/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/server/common/primitives/timestamp"
)

func resourceNamespace() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNamespaceCreate,
		ReadContext:   resourceNamespaceRead,
		UpdateContext: resourceNamespaceUpdate,
		DeleteContext: resourceNamespaceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Description: "Resource to manage a Temporal namespace.",

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of namespace",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of namespace",
			},
			"owner_email": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Specify the email address of the Namespace owner",
			},
			"retention": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     30,
				Description: "The Retention Period applies to Closed Workflow Executions (in days)",
			},
			"global_namespace": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether a Namespace i a Global Namespace. This is a read-only setting and cannot be changed",
			},
			"active_cluster": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "active",
				Description: "Specify the name of the active Temporal Cluster",
			},
			"clusters": {
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The list contains the names of Clusters to which the Namespace can fail over",
			},
			"history_archival_state": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "The State of archival",
			},
			"history_uri": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The URI for Archival. The URI cannot be changed after Archival is first enabled",
			},
			"namespace_data": {
				Type:        schema.TypeMap,
				Optional:    true,
				Default:     30,
				Description: "Data for a Namespace",
			},
			"visibility_archival_state": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "The visibility state for archival",
			},
			"visibility_uri": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The URI for visibility Archival. The URI cannot be changed after Archival is first enabled",
			},
		},
	}
}

func resourceNamespaceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	namespaceClient, err := m.(Client).NamespaceClient()
	if err != nil {
		return diag.Errorf("Failed to create SDK client: %s", err.Error())
	}

	ns := d.Get("name").(string)
	archState := enumspb.ARCHIVAL_STATE_DISABLED
	if d.Get("history_archival_state").(bool) {
		archState = enumspb.ARCHIVAL_STATE_ENABLED
	}

	archVisState := enumspb.ARCHIVAL_STATE_DISABLED
	if d.Get("visibility_archival_state").(bool) {
		archVisState = enumspb.ARCHIVAL_STATE_ENABLED
	}

	retention, err := timestamp.ParseDurationDefaultDays(d.Get("retention").(string))
	if err != nil {
		return diag.Errorf("Invalid format for rention option: %s", err.Error())
	}

	activeCluster := d.Get("active_cluster").(string)
	var clusters []*replicationpb.ClusterReplicationConfig
	if activeCluster != "" {
		clusterNames := d.Get("clusters").([]interface{})
		for _, clusterName := range clusterNames {
			clusters = append(clusters, &replicationpb.ClusterReplicationConfig{
				ClusterName: clusterName.(string),
			})
		}
	}

	request := &workflowservice.RegisterNamespaceRequest{
		Namespace:                        ns,
		Description:                      d.Get("description").(string),
		OwnerEmail:                       d.Get("owner_email").(string),
		WorkflowExecutionRetentionPeriod: &retention,
		IsGlobalNamespace:                d.Get("global_namespace").(bool),
		Data:                             strInterfaceToStrMap(d.Get("namespace_data").(map[string]interface{})),
		ActiveClusterName:                activeCluster,
		Clusters:                         clusters,
		HistoryArchivalState:             archState,
		HistoryArchivalUri:               d.Get("history_uri").(string),
		VisibilityArchivalState:          archVisState,
		VisibilityArchivalUri:            d.Get("visibility_uri").(string),
	}

	if err = namespaceClient.Register(ctx, request); err != nil {
		if _, ok := err.(*serviceerror.NamespaceAlreadyExists); !ok {
			return diag.Errorf("Namespace registration failed: %s", err.Error())
		}
	}

	d.SetId(ns)

	return nil
}

func resourceNamespaceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	namespaceClient, err := m.(Client).NamespaceClient()
	if err != nil {
		return diag.Errorf("Failed to create SDK client: %s", err.Error())
	}

	resp, err := namespaceClient.Describe(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("name", resp.NamespaceInfo.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("description", resp.NamespaceInfo.Description); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("owner_email", resp.NamespaceInfo.OwnerEmail); err != nil {
		return diag.FromErr(err)
	}

	retentionDays := fmt.Sprintf("%v", resp.Config.WorkflowExecutionRetentionTtl.Hours()/24)
	if err := d.Set("retention", retentionDays); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("global_namespace", resp.IsGlobalNamespace); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("namespace_data", resp.NamespaceInfo.Data); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("active_cluster", resp.ReplicationConfig.ActiveClusterName); err != nil {
		return diag.FromErr(err)
	}

	var clusterList []interface{}
	for _, cluster := range resp.ReplicationConfig.Clusters {
		clusterList = append(clusterList, cluster.ClusterName)
	}
	if err := d.Set("clusters", clusterList); err != nil {
		return diag.FromErr(err)
	}

	archState := false
	if resp.Config.HistoryArchivalState == enumspb.ARCHIVAL_STATE_ENABLED {
		archState = true
	}
	if err := d.Set("history_archival_state", archState); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("history_uri", resp.Config.HistoryArchivalUri); err != nil {
		return diag.FromErr(err)
	}

	archVisState := false
	if resp.Config.VisibilityArchivalState == enumspb.ARCHIVAL_STATE_ENABLED {
		archVisState = true
	}
	if err := d.Set("visibility_archival_state", archVisState); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("visibility_uri", resp.Config.VisibilityArchivalUri); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceNamespaceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	if id != d.Get("name").(string) {
		return diag.Errorf("You cannot change the name of namespace")
	}

	if !d.HasChanges("description", "owner_email", "retention", "global_namespace", "active_cluster", "clusters", "history_archival_state", "history_uri", "namespace_data", "visibility_archival_state", "visibility_uri") {
		return nil
	}

	namespaceClient, err := m.(Client).NamespaceClient()
	if err != nil {
		return diag.Errorf("Failed to create SDK client: %s", err.Error())
	}

	retention, err := timestamp.ParseDurationDefaultDays(d.Get("retention").(string))
	if err != nil {
		return diag.Errorf("Invalid format for rention option: %s", err.Error())
	}

	archState := enumspb.ARCHIVAL_STATE_DISABLED
	if d.Get("history_archival_state").(bool) {
		archState = enumspb.ARCHIVAL_STATE_ENABLED
	}

	archVisState := enumspb.ARCHIVAL_STATE_DISABLED
	if d.Get("visibility_archival_state").(bool) {
		archVisState = enumspb.ARCHIVAL_STATE_ENABLED
	}

	activeCluster := d.Get("active_cluster").(string)
	var clusters []*replicationpb.ClusterReplicationConfig
	if activeCluster != "" {
		clusterNames := d.Get("clusters").([]interface{})
		for _, clusterName := range clusterNames {
			clusters = append(clusters, &replicationpb.ClusterReplicationConfig{
				ClusterName: clusterName.(string),
			})
		}
	}

	request := &workflowservice.UpdateNamespaceRequest{
		Namespace: id,
		UpdateInfo: &namespace.UpdateNamespaceInfo{
			Description: d.Get("description").(string),
			OwnerEmail:  d.Get("owner_email").(string),
			Data:        strInterfaceToStrMap(d.Get("namespace_data").(map[string]interface{})),
		},
		Config: &namespace.NamespaceConfig{
			WorkflowExecutionRetentionTtl: &retention,
			HistoryArchivalState:          archState,
			VisibilityArchivalState:       archVisState,
		},
		ReplicationConfig: &replicationpb.NamespaceReplicationConfig{
			ActiveClusterName: activeCluster,
			Clusters:          clusters,
		},
	}

	if err := namespaceClient.Update(ctx, request); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceNamespaceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Not supported by Temporal
	return nil
}

func strInterfaceToStrMap(c map[string]interface{}) map[string]string {
	foo := map[string]string{}
	for k, v := range c {
		foo[k] = v.(string)
	}
	return foo
}
