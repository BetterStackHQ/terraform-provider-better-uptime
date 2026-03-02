package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var statusPageResourceSchema = map[string]*schema.Schema{
	"id": {
		Description: "The ID of this Status Page Resource.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"status_page_id": {
		Description: "The ID of the Status Page.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"status_page_section_id": {
		Description: "The ID of the Status Page Section. If you don't specify a status_page_section_id, we add the resource to the first section. If there are no sections in the status page yet, one will be automatically created for you.",
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
	},
	"resource_id": {
		Description: "The ID of the resource you are adding. Omit when resource_type is ManuallyTrackedItem.",
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
	},
	"resource_type": {
		Description:  "The type of the resource you are adding. Available values: ManuallyTrackedItem, Monitor, MonitorGroup, Heartbeat, HeartbeatGroup, WebhookIntegration, EmailIntegration, IncomingWebhook, ResourceGroup, LogsChart, CatalogReference.",
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringInSlice([]string{"ManuallyTrackedItem", "Monitor", "MonitorGroup", "Heartbeat", "HeartbeatGroup", "WebhookIntegration", "EmailIntegration", "IncomingWebhook", "ResourceGroup", "LogsChart", "CatalogReference"}, false),
	},
	"public_name": {
		Description: "The resource name displayed publicly on your status page.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"explanation": {
		Description: "A detailed text displayed as a help icon.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"history": {
		Description: "Do you want to display detailed historical status for this item? This field is deprecated, use widget_type instead.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
		Deprecated:  "Use widget_type instead.",
	},
	// TODO: add 'effective_position' computed property?
	"position": {
		Description: "The position of this resource on your status page, indexed from zero. If you don't specify a position, we add the resource to the end of the status page. When you specify a position of an existing resource, we add the resource to this position and shift resources below to accommodate.",
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return !d.HasChange(k)
		},
	},
	"widget_type": {
		Description: "What widget to display for this resource. Available values: plain - only display status, history - display historical status, intraday_history - display detailed historical status, response_times - add a response times chart (only for Monitor resource type). This takes preference over history when both parameters are present.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"availability": {
		Description: "The availability of this resource (from 0.0 to 1.0).",
		Type:        schema.TypeFloat,
		Optional:    false,
		Computed:    true,
	},
	"status_history": {
		Description: "History of a single status page resource history",
		Type:        schema.TypeList,
		Optional:    false,
		Computed:    true,
		Elem: &schema.Resource{
			Schema: statusPageStatusHistorySchema,
		},
	},
	"status": {
		Description: "The current status of the resource. Can be one of `not_monitored` (when the underlying monitor is paused), `operational`, `maintenance`, `degraded`, or `downtime`",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"mark_as_down_for": {
		Description:  "How to mark this resource as down. Can be one of `no_incident`, `any_incident`, or `incident_matching_metadata`.",
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		ValidateFunc: validation.StringInSlice([]string{"no_incident", "any_incident", "incident_matching_metadata"}, false),
	},
	"mark_as_down_metadata_rule": {
		Description: "Metadata rule for marking resource as down. Only applicable when mark_as_down_for is 'incident_matching_metadata'.",
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key": {
					Description: "The metadata key to match against.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"metadata_value": {
					Description: "List of metadata values that should trigger the down status.",
					Type:        schema.TypeList,
					Required:    true,
					MinItems:    1,
					Elem:        &schema.Resource{Schema: metadataValueSchema},
				},
			},
		},
	},
	"mark_as_degraded_for": {
		Description:  "How to mark this resource as degraded. Can be one of `no_incident`, `any_incident`, or `incident_matching_metadata`.",
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		ValidateFunc: validation.StringInSlice([]string{"no_incident", "any_incident", "incident_matching_metadata"}, false),
	},
	"mark_as_degraded_metadata_rule": {
		Description: "Metadata rule for marking resource as degraded. Only applicable when mark_as_degraded_for is 'incident_matching_metadata'.",
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key": {
					Description: "The metadata key to match against.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"metadata_value": {
					Description: "List of metadata values that should trigger the degraded status.",
					Type:        schema.TypeList,
					Required:    true,
					MinItems:    1,
					Elem:        &schema.Resource{Schema: metadataValueSchema},
				},
			},
		},
	},
}

var statusPageStatusHistorySchema = map[string]*schema.Schema{
	"day": {
		Description: "Status date",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"status": {
		Description: "Status",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"downtime_duration": {
		Description: "Status duration",
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
	},
	"maintenance_duration": {
		Description: "Status duration",
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
	},
}

func loadStatusPageResourceMetadataRule(d *schema.ResourceData, ruleKey string) *map[string]interface{} {
	metadataRules := d.Get(ruleKey).([]interface{})
	if len(metadataRules) == 0 {
		return nil
	}

	rule := metadataRules[0].(map[string]interface{})
	metadataRule := make(map[string]interface{})

	if key, ok := rule["key"].(string); ok && key != "" {
		metadataRule["key"] = key
	}

	if metadataValues, ok := rule["metadata_value"].([]interface{}); ok && len(metadataValues) > 0 {
		values := make([]metadataValue, 0, len(metadataValues))
		for _, v := range metadataValues {
			valueMap := v.(map[string]interface{})
			value := metadataValue{}

			if v, ok := valueMap["type"].(string); ok && v != "" {
				value.Type = v
			}
			if v, ok := valueMap["value"].(string); ok && v != "" {
				value.Value = &v
			}
			if v, ok := valueMap["item_id"].(string); ok && v != "" {
				value.ItemID = json.Number(v)
			}
			if v, ok := valueMap["email"].(string); ok && v != "" {
				value.Email = &v
			}
			if v, ok := valueMap["name"].(string); ok && v != "" {
				value.Name = &v
			}

			values = append(values, value)
		}
		metadataRule["values"] = values
	}

	return &metadataRule
}

func statusPageResourceMetadataRuleCopyAttrs(d *schema.ResourceData, rule *map[string]interface{}, ruleKey string) error {
	if rule == nil {
		return d.Set(ruleKey, []interface{}{})
	}

	metadataRule := make(map[string]interface{})

	if key, ok := (*rule)["key"].(string); ok {
		metadataRule["key"] = key
	}

	if values, ok := (*rule)["values"].([]interface{}); ok && len(values) > 0 {
		metadataValues := make([]interface{}, 0, len(values))
		for i, v := range values {
			valueMap, ok := v.(map[string]interface{})
			if !ok {
				continue
			}
			metadataValue := make(map[string]interface{})

			if valueType, ok := valueMap["type"].(string); ok {
				metadataValue["type"] = valueType
			}
			if value, ok := valueMap["value"].(string); ok {
				metadataValue["value"] = value
			}

			// Only include item_id, email, name if they were configured in Terraform
			configPrefix := fmt.Sprintf("%s.0.metadata_value.%d", ruleKey, i)
			if _, ok := d.GetOk(configPrefix + ".item_id"); ok {
				if id, ok := valueMap["item_id"]; ok && id != nil {
					switch typedId := id.(type) {
					case json.Number:
						metadataValue["item_id"] = typedId.String()
					case float64:
						metadataValue["item_id"] = fmt.Sprintf("%.0f", typedId)
					}
				}
			}
			if _, ok := d.GetOk(configPrefix + ".email"); ok {
				if email, ok := valueMap["email"].(string); ok {
					metadataValue["email"] = email
				}
			}
			if _, ok := d.GetOk(configPrefix + ".name"); ok {
				if name, ok := valueMap["name"].(string); ok {
					metadataValue["name"] = name
				}
			}

			metadataValues = append(metadataValues, metadataValue)
		}
		metadataRule["metadata_value"] = metadataValues
	}

	if len(metadataRule) == 0 {
		return d.Set(ruleKey, []interface{}{})
	}

	return d.Set(ruleKey, []interface{}{metadataRule})
}

func validateStatusPageResource(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	degradedRuleError := validateMetadataRule(d, "mark_as_degraded_for", "mark_as_degraded_metadata_rule")
	downRuleError := validateMetadataRule(d, "mark_as_down_for", "mark_as_down_metadata_rule")
	resourceIDError := validateResourceTypeID(d)

	return errors.Join(degradedRuleError, downRuleError, resourceIDError)
}

func validateResourceTypeID(d *schema.ResourceDiff) error {
	resourceType := d.Get("resource_type").(string)
	_, hasResourceID := d.GetOk("resource_id")
	if resourceType != "ManuallyTrackedItem" && !hasResourceID && d.NewValueKnown("resource_id") {
		return fmt.Errorf("resource_id is required when resource_type is %s", resourceType)
	}
	return nil
}

func validateMetadataRule(d *schema.ResourceDiff, keyFor string, keyMetadataRule string) error {
	// Validate mark_as_degraded_for and mark_as_degraded_metadata_rule relationship
	markFor, hasMarkFor := d.GetOk(keyFor)
	metadataRule := d.Get(keyMetadataRule).([]interface{})

	if hasMarkFor && markFor.(string) == "incident_matching_metadata" {
		if len(metadataRule) == 0 {
			return fmt.Errorf("%s is required when %s is 'incident_matching_metadata'", keyMetadataRule, keyFor)
		}
		// Validate metadata rule structure
		if len(metadataRule) > 0 {
			rule := metadataRule[0].(map[string]interface{})
			metadataValues := rule["metadata_value"].([]interface{})
			for i, v := range metadataValues {
				value := v.(map[string]interface{})
				if err := validateMetadataValue(value, fmt.Sprintf("%s.metadata_value.%d", keyMetadataRule, i)); err != nil {
					return err
				}
			}
		}
	} else if hasMarkFor && markFor.(string) != "incident_matching_metadata" {
		if len(metadataRule) > 0 {
			return fmt.Errorf("%s can only be used when %s is 'incident_matching_metadata'", keyMetadataRule, keyFor)
		}
	}

	return nil
}

func newStatusPageResourceResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: statusPageResourceCreate,
		ReadContext:   statusPageResourceRead,
		UpdateContext: statusPageResourceUpdate,
		DeleteContext: statusPageResourceDelete,
		CustomizeDiff: validateStatusPageResource,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				split := strings.SplitN(d.Id(), "/", 2)
				if len(split) != 2 {
					return nil, errors.New("betteruptime_status_page_resource can be imported via \"status_page_id/id\" only (e.g. \"0/1\")")
				}
				if err := d.Set("status_page_id", split[0]); err != nil {
					return nil, err
				}
				d.SetId(split[1])
				return []*schema.ResourceData{d}, nil
			},
		},
		Description: "https://betterstack.com/docs/uptime/api/status-page-resources/",
		Schema:      statusPageResourceSchema,
	}
}

type statusPageResource struct {
	StatusPageSectionID        *int                      `json:"status_page_section_id,omitempty"`
	ResourceID                 *int                      `json:"resource_id,omitempty"`
	ResourceType               *string                   `json:"resource_type,omitempty"`
	PublicName                 *string                   `json:"public_name,omitempty"`
	Explanation                *string                   `json:"explanation,omitempty"`
	History                    *bool                     `json:"history,omitempty"`
	Position                   *int                      `json:"position,omitempty"`
	FixedPosition              *bool                     `json:"fixed_position,omitempty"`
	WidgetType                 *string                   `json:"widget_type,omitempty"`
	Availability               *float32                  `json:"availability,omitempty"`
	Status                     *string                   `json:"status,omitempty"`
	StatusHistory              *[]map[string]interface{} `json:"status_history,omitempty"`
	MarkAsDownFor              *string                   `json:"mark_as_down_for,omitempty"`
	MarkAsDownMetadataRule     *map[string]interface{}   `json:"mark_as_down_metadata_rule,omitempty"`
	MarkAsDegradedFor          *string                   `json:"mark_as_degraded_for,omitempty"`
	MarkAsDegradedMetadataRule *map[string]interface{}   `json:"mark_as_degraded_metadata_rule,omitempty"`
}

type statusPageResourceHTTPResponse struct {
	Data struct {
		ID         string             `json:"id"`
		Attributes statusPageResource `json:"attributes"`
	} `json:"data"`
}

func statusPageResourceRef(in *statusPageResource) []struct {
	k string
	v interface{}
} {
	return []struct {
		k string
		v interface{}
	}{
		{k: "status_page_section_id", v: &in.StatusPageSectionID},
		{k: "resource_id", v: &in.ResourceID},
		{k: "resource_type", v: &in.ResourceType},
		{k: "public_name", v: &in.PublicName},
		{k: "explanation", v: &in.Explanation},
		{k: "history", v: &in.History},
		{k: "position", v: &in.Position},
		{k: "widget_type", v: &in.WidgetType},
		{k: "availability", v: &in.Availability},
		{k: "status_history", v: &in.StatusHistory},
		{k: "status", v: &in.Status},
		{k: "mark_as_down_for", v: &in.MarkAsDownFor},
		{k: "mark_as_down_metadata_rule", v: &in.MarkAsDownMetadataRule},
		{k: "mark_as_degraded_for", v: &in.MarkAsDegradedFor},
		{k: "mark_as_degraded_metadata_rule", v: &in.MarkAsDegradedMetadataRule},
	}
}

func statusPageResourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in statusPageResource
	for _, e := range statusPageResourceRef(&in) {
		if e.k == "status_history" {
			// Skip loading status history when preparing to send the request
		} else if e.k == "mark_as_down_metadata_rule" {
			in.MarkAsDownMetadataRule = loadStatusPageResourceMetadataRule(d, "mark_as_down_metadata_rule")
		} else if e.k == "mark_as_degraded_metadata_rule" {
			in.MarkAsDegradedMetadataRule = loadStatusPageResourceMetadataRule(d, "mark_as_degraded_metadata_rule")
		} else {
			load(d, e.k, e.v)
		}
	}
	in.FixedPosition = truePtr()
	statusPageID := d.Get("status_page_id").(string)
	var out statusPageResourceHTTPResponse
	if err := resourceCreate(ctx, meta, fmt.Sprintf("/api/v2/status-pages/%s/resources", url.PathEscape(statusPageID)), &in, &out); err != nil {
		return err
	}
	d.SetId(out.Data.ID)
	return statusPageResourceCopyAttrs(d, &out.Data.Attributes)
}

func statusPageResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	statusPageID := d.Get("status_page_id").(string)
	var out statusPageResourceHTTPResponse
	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/status-pages/%s/resources/%s", url.PathEscape(statusPageID), url.PathEscape(d.Id())), &out); err != nil {
		return err
	} else if !ok {
		d.SetId("") // Force "create" on 404.
		return nil
	}
	return statusPageResourceCopyAttrs(d, &out.Data.Attributes)
}

func statusPageResourceCopyAttrs(d *schema.ResourceData, in *statusPageResource) diag.Diagnostics {
	var derr diag.Diagnostics
	for _, e := range statusPageResourceRef(in) {
		if e.k == "mark_as_down_metadata_rule" {
			if err := statusPageResourceMetadataRuleCopyAttrs(d, in.MarkAsDownMetadataRule, "mark_as_down_metadata_rule"); err != nil {
				derr = append(derr, diag.FromErr(err)[0])
			}
		} else if e.k == "mark_as_degraded_metadata_rule" {
			if err := statusPageResourceMetadataRuleCopyAttrs(d, in.MarkAsDegradedMetadataRule, "mark_as_degraded_metadata_rule"); err != nil {
				derr = append(derr, diag.FromErr(err)[0])
			}
		} else {
			if err := d.Set(e.k, reflect.Indirect(reflect.ValueOf(e.v)).Interface()); err != nil {
				derr = append(derr, diag.FromErr(err)[0])
			}
		}
	}

	// Clear metadata rules if they are not active, they would be missing in the API request
	if d.Get("mark_as_down_for").(string) != "incident_matching_metadata" {
		if err := d.Set("mark_as_down_metadata_rule", []interface{}{}); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}
	if d.Get("mark_as_degraded_for").(string) != "incident_matching_metadata" {
		if err := d.Set("mark_as_degraded_metadata_rule", []interface{}{}); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}

	return derr
}

func statusPageResourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in statusPageResource
	var out policyHTTPResponse
	for _, e := range statusPageResourceRef(&in) {
		if d.HasChange(e.k) {
			if e.k == "mark_as_down_metadata_rule" {
				in.MarkAsDownMetadataRule = loadStatusPageResourceMetadataRule(d, "mark_as_down_metadata_rule")
			} else if e.k == "mark_as_degraded_metadata_rule" {
				in.MarkAsDegradedMetadataRule = loadStatusPageResourceMetadataRule(d, "mark_as_degraded_metadata_rule")
			} else {
				load(d, e.k, e.v)
				// When updating resource ID, we need to update resource type as well (and vice-versa)
				if e.k == "resource_id" {
					load(d, "resource_type", &in.ResourceType)
				}
				if e.k == "resource_type" {
					load(d, "resource_id", &in.ResourceID)
				}
			}
		}
	}
	in.FixedPosition = truePtr()
	statusPageID := d.Get("status_page_id").(string)
	return resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/status-pages/%s/resources/%s", url.PathEscape(statusPageID), url.PathEscape(d.Id())), &in, &out)
}

func statusPageResourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	statusPageID := d.Get("status_page_id").(string)
	return resourceDelete(ctx, meta, fmt.Sprintf("/api/v2/status-pages/%s/resources/%s", url.PathEscape(statusPageID), url.PathEscape(d.Id())))
}
