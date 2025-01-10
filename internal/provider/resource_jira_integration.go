package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var jiraIntegrationSchema = map[string]*schema.Schema{
	"id": {
		Description: "The ID of the Jira Integration.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"better_stack_id": {
		Description: "Due to required authentication in Jira, the integration has to be created and removed in Better Stack web UI. You can set the ID of the Jira Integration to control in Better Stack, and it will be auto-imported during resource creation.",
		Type:        schema.TypeString,
		Optional:    true,
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
	"jira_fields_json": {
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
	Name                   *string                 `json:"name,omitempty"`
	AutomaticIssueCreation *bool                   `json:"automatic_issue_creation,omitempty"`
	JiraProjectKey         *string                 `json:"jira_project_key,omitempty"`
	JiraIssueTypeID        *string                 `json:"jira_issue_type_id,omitempty"`
	JiraFieldsJSON         *string                 `json:"-"`
	JiraFields             *map[string]interface{} `json:"jira_fields,omitempty"`
}

type jiraIntegrationHTTPResponse struct {
	Data struct {
		ID         string          `json:"id"`
		Attributes jiraIntegration `json:"attributes"`
	} `json:"data"`
}

func jiraIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if id, ok := d.GetOk("better_stack_id"); ok {
		// creation is not supported for this resource, import better_stack_id and do a complete update instead
		d.SetId(id.(string))
		var current jiraIntegrationHTTPResponse
		if err, _ := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/jira-integrations/%s", url.PathEscape(d.Id())), &current); err != nil {
			return err
		}
		if err := jiraIntegrationUpdate(ctx, d, meta); err != nil {
			return err
		}
		return nil
	}

	return diag.Errorf("Due to required authentication in Jira, the integration has to be created and removed in Better Stack web UI. You can either import the resource, or set the ID of the Jira Integration in better_stack_id field and it will be auto-imported during resource creation.")
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

	// Handle jira_fields separately
	if d.HasChange("jira_fields_json") {
		if newJson, ok := d.GetOk("jira_fields_json"); ok {
			stringJson := newJson.(string)
			in.JiraFieldsJSON = &stringJson
			fields := make(map[string]interface{})
			if err := json.Unmarshal([]byte(stringJson), &fields); err != nil {
				return diag.FromErr(err)
			}
			in.JiraFields = &fields
		}
	}

	if err := resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/jira-integrations/%s", url.PathEscape(d.Id())), &in, &out); err != nil {
		return err
	}
	return jiraIntegrationCopyAttrs(d, &out.Data.Attributes)
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

	// Handle jira_fields separately
	if in.JiraFields != nil {
		newJson, err := json.Marshal(&in.JiraFields)
		if err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
		stringJson := string(newJson)
		in.JiraFieldsJSON = &stringJson
		if err := d.Set("jira_fields_json", stringJson); err != nil {
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
	}
}
