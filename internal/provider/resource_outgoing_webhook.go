package provider

import (
	"context"
	"fmt"
	"net/url"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var outgoingWebhookSchema = map[string]*schema.Schema{
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
		Description: "The ID of the outgoing webhook.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"name": {
		Description: "The name of the outgoing webhook.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"url": {
		Description: "The URL to send webhooks to.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"trigger_type": {
		Description: "The type of trigger for the webhook. Only settable during creation.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		ValidateFunc: validation.StringInSlice([]string{
			"incident_change",
			"on_call_change",
			"monitor_change",
		}, false),
	},
	"on_incident_started": {
		Description: "Whether to trigger webhook when incident starts.",
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
	},
	"on_incident_acknowledged": {
		Description: "Whether to trigger webhook when incident is acknowledged.",
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
	},
	"on_incident_resolved": {
		Description: "Whether to trigger webhook when incident is resolved.",
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
	},
	"custom_webhook_template_attributes": {
		Description: "Custom webhook template configuration.",
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"http_method": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "POST",
				},
				"auth_user": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"auth_password": {
					Type:      schema.TypeString,
					Optional:  true,
					Sensitive: true,
				},
				"headers_template": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Type:     schema.TypeString,
								Required: true,
							},
							"value": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
				"body_template": {
					Type:             schema.TypeString,
					Optional:         true,
					DiffSuppressFunc: suppressEquivalentJSONDiffs,
				},
			},
		},
	},
}

type headerTemplate struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type customWebhookTemplateAttributes struct {
	ID             *string          `json:"id,omitempty"`
	HTTPMethod     *string          `json:"http_method,omitempty"`
	AuthUser       *string          `json:"auth_user,omitempty"`
	AuthPassword   *string          `json:"auth_password,omitempty"`
	HeaderTemplate []headerTemplate `json:"headers_template,omitempty"`
	BodyTemplate   interface{}      `json:"body_template,omitempty"`
}

type outgoingWebhook struct {
	ID                              *string                          `json:"id,omitempty"`
	Name                            *string                          `json:"name,omitempty"`
	URL                             *string                          `json:"url,omitempty"`
	TriggerType                     *string                          `json:"trigger_type,omitempty"`
	OnIncidentStarted               *bool                            `json:"on_incident_started,omitempty"`
	OnIncidentAcknowledged          *bool                            `json:"on_incident_acknowledged,omitempty"`
	OnIncidentResolved              *bool                            `json:"on_incident_resolved,omitempty"`
	CustomWebhookTemplateAttributes *customWebhookTemplateAttributes `json:"custom_webhook_template_attributes,omitempty"`
	TeamName                        *string                          `json:"team_name,omitempty"`
}

type outgoingWebhookHTTPResponse struct {
	Data struct {
		ID         string          `json:"id"`
		Attributes outgoingWebhook `json:"attributes"`
	} `json:"data"`
}

func validateOutgoingWebhook(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
	triggerType := d.Get("trigger_type").(string)

	// Validate incident_change specific fields
	incidentFields := []string{"on_incident_started", "on_incident_acknowledged", "on_incident_resolved"}

	for _, field := range incidentFields {
		if value, ok := d.GetOk(field); ok && value.(bool) {
			if triggerType != "incident_change" {
				return fmt.Errorf("%s can only be set when trigger_type is 'incident_change'", field)
			}
		}
	}

	return nil
}

func newOutgoingWebhookResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: outgoingWebhookCreate,
		ReadContext:   outgoingWebhookRead,
		UpdateContext: outgoingWebhookUpdate,
		DeleteContext: outgoingWebhookDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: validateOutgoingWebhook,
		Schema:        outgoingWebhookSchema,
	}
}

func outgoingWebhookRef(in *outgoingWebhook) []struct {
	k string
	v interface{}
} {
	return []struct {
		k string
		v interface{}
	}{
		{k: "id", v: &in.ID},
		{k: "name", v: &in.Name},
		{k: "url", v: &in.URL},
		{k: "trigger_type", v: &in.TriggerType},
		{k: "on_incident_started", v: &in.OnIncidentStarted},
		{k: "on_incident_acknowledged", v: &in.OnIncidentAcknowledged},
		{k: "on_incident_resolved", v: &in.OnIncidentResolved},
	}
}

func outgoingWebhookCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in outgoingWebhook

	// Load basic fields
	for _, e := range outgoingWebhookRef(&in) {
		load(d, e.k, e.v)
	}

	// Load team name
	load(d, "team_name", &in.TeamName)

	// Handle custom webhook template attributes
	if v, ok := d.GetOk("custom_webhook_template_attributes"); ok && len(v.([]interface{})) > 0 {
		attrs := v.([]interface{})[0].(map[string]interface{})
		template := &customWebhookTemplateAttributes{}

		if method, ok := attrs["http_method"].(string); ok {
			template.HTTPMethod = &method
		}
		if user, ok := attrs["auth_user"].(string); ok {
			template.AuthUser = &user
		}
		if pass, ok := attrs["auth_password"].(string); ok {
			template.AuthPassword = &pass
		}

		// Handle headers template
		if headers, ok := attrs["headers_template"].([]interface{}); ok {
			template.HeaderTemplate = make([]headerTemplate, len(headers))
			for i, h := range headers {
				header := h.(map[string]interface{})
				template.HeaderTemplate[i] = headerTemplate{
					Name:  header["name"].(string),
					Value: header["value"].(string),
				}
			}
		}

		// Handle body template
		if body, ok := attrs["body_template"].(string); ok {
			template.BodyTemplate = body
		}

		in.CustomWebhookTemplateAttributes = template
	}

	var out outgoingWebhookHTTPResponse
	if err := resourceCreate(ctx, meta, "/api/v2/outgoing-webhooks", &in, &out); err != nil {
		return err
	}

	d.SetId(out.Data.ID)
	return outgoingWebhookCopyAttrs(d, &out.Data.Attributes)
}

func outgoingWebhookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var out outgoingWebhookHTTPResponse
	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/outgoing-webhooks/%s", url.PathEscape(d.Id())), &out); err != nil {
		return err
	} else if !ok {
		d.SetId("")
		return nil
	}
	return outgoingWebhookCopyAttrs(d, &out.Data.Attributes)
}

func outgoingWebhookCopyAttrs(d *schema.ResourceData, in *outgoingWebhook) diag.Diagnostics {
	var derr diag.Diagnostics

	// Copy basic fields
	for _, e := range outgoingWebhookRef(in) {
		if err := d.Set(e.k, reflect.Indirect(reflect.ValueOf(e.v)).Interface()); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}

	// Handle custom webhook template attributes
	if in.CustomWebhookTemplateAttributes != nil {
		template := map[string]interface{}{
			"id":            in.CustomWebhookTemplateAttributes.ID,
			"http_method":   in.CustomWebhookTemplateAttributes.HTTPMethod,
			"auth_user":     in.CustomWebhookTemplateAttributes.AuthUser,
			"auth_password": in.CustomWebhookTemplateAttributes.AuthPassword,
			"body_template": in.CustomWebhookTemplateAttributes.BodyTemplate,
		}

		if len(in.CustomWebhookTemplateAttributes.HeaderTemplate) > 0 {
			headers := make([]map[string]interface{}, len(in.CustomWebhookTemplateAttributes.HeaderTemplate))
			for i, h := range in.CustomWebhookTemplateAttributes.HeaderTemplate {
				headers[i] = map[string]interface{}{
					"name":  h.Name,
					"value": h.Value,
				}
			}
			template["headers_template"] = headers
		}

		if err := d.Set("custom_webhook_template_attributes", []interface{}{template}); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}

	return derr
}

func outgoingWebhookUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in outgoingWebhook

	// Only include changed fields in the update
	for _, e := range outgoingWebhookRef(&in) {
		if d.HasChange(e.k) {
			load(d, e.k, e.v)
		}
	}

	// Handle custom webhook template attributes if changed
	if d.HasChange("custom_webhook_template_attributes") {
		if v, ok := d.GetOk("custom_webhook_template_attributes"); ok && len(v.([]interface{})) > 0 {
			attrs := v.([]interface{})[0].(map[string]interface{})
			template := &customWebhookTemplateAttributes{}

			if method, ok := attrs["http_method"].(string); ok {
				template.HTTPMethod = &method
			}
			if user, ok := attrs["auth_user"].(string); ok {
				template.AuthUser = &user
			}
			if pass, ok := attrs["auth_password"].(string); ok {
				template.AuthPassword = &pass
			}

			if headers, ok := attrs["headers_template"].([]interface{}); ok {
				template.HeaderTemplate = make([]headerTemplate, len(headers))
				for i, h := range headers {
					header := h.(map[string]interface{})
					template.HeaderTemplate[i] = headerTemplate{
						Name:  header["name"].(string),
						Value: header["value"].(string),
					}
				}
			}

			if body, ok := attrs["body_template"].(string); ok {
				template.BodyTemplate = body
			}

			in.CustomWebhookTemplateAttributes = template
		}
	}

	var out outgoingWebhookHTTPResponse
	return resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/outgoing-webhooks/%s", url.PathEscape(d.Id())), &in, &out)
}

func outgoingWebhookDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDelete(ctx, meta, fmt.Sprintf("/api/v2/outgoing-webhooks/%s", url.PathEscape(d.Id())))
}
