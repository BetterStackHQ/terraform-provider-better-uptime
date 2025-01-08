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
	"team_name": {
		Description: "Used to specify the team the resource should be created in when using global tokens.",
		Type:        schema.TypeString,
		Optional:    true,
		Default:     nil,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return d.Id() != ""
		},
	},
	"id": {
		Description: "The ID of the Jira Integration.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
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
	},
	"jira_project_key": {
		Description: "The Jira project key.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"jira_issue_type_id": {
		Description: "The Jira issue type ID.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"jira_fields": {
		Description: "JSON object representing Jira fields.",
		Type:        schema.TypeString,
		Optional:    true,
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
	ID                     *string `json:"id,omitempty"`
	Name                   *string `json:"name,omitempty"`
	AutomaticIssueCreation *bool   `json:"automatic_issue_creation,omitempty"`
	JiraProjectKey         *string `json:"jira_project_key,omitempty"`
	JiraIssueTypeID        *string `json:"jira_issue_type_id,omitempty"`
	JiraFields             *string `json:"jira_fields,omitempty"`
}

func jiraIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// If ID is provided, treat this as an import
	if id, ok := d.GetOk("id"); ok {
		d.SetId(id.(string))
		// Read the existing resource
		return jiraIntegrationRead(ctx, d, meta)
	}

	return diag.Errorf("resource ID is required, please import the resource first")
}

func jiraIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var out jiraIntegration
	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/jira-integrations/%s", url.PathEscape(d.Id())), &out); err != nil {
		return err
	} else if !ok {
		return diag.Errorf("Jira integration %s not found.", d.Id())
	}
	return jiraIntegrationCopyAttrs(d, &out)
}

func jiraIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in jiraIntegration
	var out jiraIntegration
	for _, e := range jiraIntegrationRef(&in) {
		if d.HasChange(e.k) {
			load(d, e.k, e.v)
		}
	}
	return resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/jira-integrations/%s", url.PathEscape(d.Id())), &in, &out)
}

func jiraIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.Errorf("deletion is not supported for this resource")
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
		{k: "id", v: &in.ID},
		{k: "name", v: &in.Name},
		{k: "automatic_issue_creation", v: &in.AutomaticIssueCreation},
		{k: "jira_project_key", v: &in.JiraProjectKey},
		{k: "jira_issue_type_id", v: &in.JiraIssueTypeID},
		{k: "jira_fields", v: &in.JiraFields},
	}
}
