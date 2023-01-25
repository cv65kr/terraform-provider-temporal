package temporal

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	enumspb "go.temporal.io/api/enums/v1"
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
				Default:     3 * 24 * time.Hour,
				Description: "The Retention Period applies to Closed Workflow Executions",
			},
		},
	}
}

func resourceNamespaceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(Client)
	namespaceClient, err := client.NamespaceClient()
	if err != nil {
		return diag.Errorf("Failed to create SDK client: %s", err.Error())
	}

	ns := d.Get("name").(string)
	description := d.Get("description").(string)
	ownerEmail := d.Get("owner_email").(string)
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
		Description:                      description,
		OwnerEmail:                       ownerEmail,
		WorkflowExecutionRetentionPeriod: &retention,
		HistoryArchivalState:             archState,
	}

	if err = namespaceClient.Register(ctx, request); err != nil {
		if _, ok := err.(*serviceerror.NamespaceAlreadyExists); !ok {
			return diag.Errorf("namespace registration failed: %s", err.Error())
		}
	}

	d.SetId(ns)

	return nil
}

func resourceNamespaceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// client := m.(*sdkclient.Client)

	return nil
}

func resourceNamespaceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// client := m.(*sdkclient.Client)

	return nil
}

func resourceNamespaceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// client := m.(*sdkclient.Client)

	return nil
}
