package provider

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"strconv"

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
		Description: "Map representing Jira fields. Values can be strings or numbers.",
		Type:        schema.TypeMap,
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
	JiraFields             *map[string]interface{} `json:"jira_fields,omitempty"`
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
		return err
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

	// Handle jira_fields separately
	if d.HasChange("jira_fields") {
		if v, ok := d.GetOk("jira_fields"); ok {
			fields := make(map[string]interface{})
			for k, v := range v.(map[string]interface{}) {
				// Dynamically determine type, as internally it will always be a string
				if intVal, err := strconv.ParseInt(v.(string), 10, 64); err == nil {
					fields[k] = intVal
				} else if floatVal, err := strconv.ParseFloat(v.(string), 64); err == nil {
					fields[k] = floatVal
				} else {
					fields[k] = v // Assume string
				}
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
		fields := make(map[string]string)
		for k, v := range *in.JiraFields {
			// Internally, it should always be a string in Terraform state
			if floatVal, isFloat := v.(float64); isFloat {
				fields[k] = strconv.FormatFloat(floatVal, 'f', -1, 64)
			} else if intVal, isInt := v.(int64); isInt {
				fields[k] = strconv.FormatInt(intVal, 10)
			} else {
				fields[k] = v.(string) // Assume string
			}
		}
		if err := d.Set("jira_fields", fields); err != nil {
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
