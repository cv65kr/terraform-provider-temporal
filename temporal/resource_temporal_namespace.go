package temporal

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/namespace/v1"
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
			"history_archival_state": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "State of archival",
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

	retention, err := timestamp.ParseDurationDefaultDays(d.Get("retention").(string))
	if err != nil {
		return diag.Errorf("Invalid format for rention option: %s", err.Error())
	}

	request := &workflowservice.RegisterNamespaceRequest{
		Namespace:                        ns,
		Description:                      d.Get("description").(string),
		OwnerEmail:                       d.Get("owner_email").(string),
		WorkflowExecutionRetentionPeriod: &retention,
		HistoryArchivalState:             archState,
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

	archState := false
	if resp.Config.HistoryArchivalState == enumspb.ARCHIVAL_STATE_ENABLED {
		archState = true
	}
	if err := d.Set("history_archival_state", archState); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("owner_email", resp.NamespaceInfo.OwnerEmail); err != nil {
		return diag.FromErr(err)
	}

	retentionDays := fmt.Sprintf("%v", resp.Config.WorkflowExecutionRetentionTtl.Hours()/24)
	if err := d.Set("retention", retentionDays); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceNamespaceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	if id != d.Get("name").(string) {
		return diag.Errorf("You cannot change the name of namespace")
	}

	if !d.HasChanges("description", "history_archival_state", "owner_email", "retention") {
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

	request := &workflowservice.UpdateNamespaceRequest{
		Namespace: id,
		UpdateInfo: &namespace.UpdateNamespaceInfo{
			Description: d.Get("description").(string),
			OwnerEmail:  d.Get("owner_email").(string),
		},
		Config: &namespace.NamespaceConfig{
			WorkflowExecutionRetentionTtl: &retention,
			HistoryArchivalState:          archState,
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
