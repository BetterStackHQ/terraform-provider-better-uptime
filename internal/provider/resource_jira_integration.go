package provider

import (
	"context"
	"fmt"
	"net/url"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var jiraIntegrationSchema = map[string]*schema.Schema{
	"id": {
		Description: "The internal ID of the resource, can be ignored.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"better_stack_id": {
		Description: "The ID of the Jira Integration to control in Better Stack. Due to required authentication in Jira, the integration has to be created and removed in Better Stack web UI.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"name": {
		Description: "The name of the Jira Integration.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"automatic_issue_creation": {
		Description: "Whether to automatically create issues in Jira on incident start.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"jira_project_key": {
		Description: "The Jira project key.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"jira_issue_type_id": {
		Description: "The Jira issue type ID.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"jira_fields": {
		Description: "JSON object representing Jira fields.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
}

func newJiraIntegrationResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: jiraIntegrationCreate,
		ReadContext:   jiraIntegrationRead,
		UpdateContext: jiraIntegrationUpdate,
		DeleteContext: jiraIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "https://betterstack.com/docs/uptime/api/jira-integrations/",
		Schema:      jiraIntegrationSchema,
	}
}

type jiraIntegration struct {
	Name                   *string `json:"name,omitempty"`
	AutomaticIssueCreation *bool   `json:"automatic_issue_creation,omitempty"`
	JiraProjectKey         *string `json:"jira_project_key,omitempty"`
	JiraIssueTypeID        *string `json:"jira_issue_type_id,omitempty"`
	JiraFields             *string `json:"jira_fields,omitempty"`
}

type jiraIntegrationHTTPResponse struct {
	Data struct {
		ID         string          `json:"id"`
		Attributes jiraIntegration `json:"attributes"`
	} `json:"data"`
}

func jiraIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// creation is not supported for this resource, import better_stack_id and do a complete update instead
	d.SetId(d.Get("better_stack_id").(string))
	var current jiraIntegrationHTTPResponse
	if err, _ := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/jira-integrations/%s", url.PathEscape(d.Id())), &current); err != nil {
		return diag.Errorf("Jira integration %s not found.", d.Id())
	}
	if err := jiraIntegrationUpdate(ctx, d, meta); err != nil {
		return err
	}
	return nil
}

func jiraIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var out jiraIntegrationHTTPResponse
	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/jira-integrations/%s", url.PathEscape(d.Id())), &out); err != nil {
		return err
	} else if !ok {
		return diag.Errorf("Jira integration %s not found.", d.Id())
	}
	return jiraIntegrationCopyAttrs(d, &out.Data.Attributes)
}

func jiraIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in jiraIntegration
	var out jiraIntegrationHTTPResponse
	for _, e := range jiraIntegrationRef(&in) {
		if d.HasChange(e.k) {
			load(d, e.k, e.v)
		}
	}
	return resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/jira-integrations/%s", url.PathEscape(d.Id())), &in, &out)
}

func jiraIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// deletion is not supported for this resource, silently ignore
	d.SetId("")
	return nil
}

func jiraIntegrationCopyAttrs(d *schema.ResourceData, in *jiraIntegration) diag.Diagnostics {
	var derr diag.Diagnostics
	for _, e := range jiraIntegrationRef(in) {
		if err := d.Set(e.k, reflect.Indirect(reflect.ValueOf(e.v)).Interface()); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}
	return derr
}

func jiraIntegrationRef(in *jiraIntegration) []struct {
	k string
	v interface{}
} {
	return []struct {
		k string
		v interface{}
	}{
		{k: "name", v: &in.Name},
		{k: "automatic_issue_creation", v: &in.AutomaticIssueCreation},
		{k: "jira_project_key", v: &in.JiraProjectKey},
		{k: "jira_issue_type_id", v: &in.JiraIssueTypeID},
		{k: "jira_fields", v: &in.JiraFields},
	}
}
