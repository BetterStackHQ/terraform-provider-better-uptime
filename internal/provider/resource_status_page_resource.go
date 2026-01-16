package provider

import (
	"context"
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
		Description: "The ID of the resource you are adding.",
		Type:        schema.TypeInt,
		Required:    true,
	},
	"resource_type": {
		Description:  "The type of the resource you are adding. Available values: Monitor, MonitorGroup, Heartbeat, HeartbeatGroup, WebhookIntegration, EmailIntegration, IncomingWebhook, ResourceGroup, LogsChart, CatalogReference.",
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringInSlice([]string{"Monitor", "MonitorGroup", "Heartbeat", "HeartbeatGroup", "WebhookIntegration", "EmailIntegration", "IncomingWebhook", "ResourceGroup", "LogsChart", "CatalogReference"}, false),
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
		Description: "What widget to display for this resource. Expects one of three values: plain - only display status, history - display detailed historical status, response_times - add a response times chart (only for Monitor resource type). This takes preference over history when both parameters are present.",
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

func validateMetadataRule(v interface{}, k string) (ws []string, errors []error) {
	ruleList := v.([]interface{})
	if len(ruleList) == 0 {
		return
	}

	rule := ruleList[0].(map[string]interface{})

	// Check if key exists and is not empty
	if key, ok := rule["key"]; !ok || key.(string) == "" {
		errors = append(errors, fmt.Errorf("%s: key is required and cannot be empty", k))
	}

	// Check if metadata_value exists and has at least one item
	if metadataValues, ok := rule["metadata_value"]; !ok {
		errors = append(errors, fmt.Errorf("%s: metadata_value is required", k))
	} else {
		metadataValueList := metadataValues.([]interface{})
		if len(metadataValueList) == 0 {
			errors = append(errors, fmt.Errorf("%s: at least one metadata_value is required", k))
		}

		// Validate each metadata value
		for i, value := range metadataValueList {
			valueMap := value.(map[string]interface{})
			if itemID, ok := valueMap["item_id"]; !ok || itemID.(string) == "" {
				errors = append(errors, fmt.Errorf("%s: metadata_value[%d]: item_id is required and cannot be empty", k, i))
			}
			// The type field from metadataValueSchema should be present
			if valueType, ok := valueMap["type"]; !ok || valueType.(string) == "" {
				errors = append(errors, fmt.Errorf("%s: metadata_value[%d]: type is required and cannot be empty", k, i))
			}
		}
	}

	return
}

// Helper functions to convert between Terraform schema format and API format for metadata rules

func convertMetadataRuleToAPI(tfRule []interface{}) map[string]interface{} {
	if len(tfRule) == 0 {
		return nil
	}

	rule := tfRule[0].(map[string]interface{})
	key := rule["key"].(string)
	metadataValues := rule["metadata_value"].([]interface{})

	apiValues := make([]interface{}, len(metadataValues))
	for i, v := range metadataValues {
		valueMap := v.(map[string]interface{})
		valueType := valueMap["type"].(string)

		if valueType == "String" {
			// For string values, only send the value field
			if value, ok := valueMap["value"].(string); ok {
				apiValues[i] = map[string]interface{}{
					"value": value,
				}
			}
		} else {
			// For reference values, send type and one of item_id/name/email
			apiValue := map[string]interface{}{
				"type": valueType,
			}

			if itemID, ok := valueMap["item_id"].(string); ok && itemID != "" {
				apiValue["item_id"] = itemID
			}
			if name, ok := valueMap["name"].(string); ok && name != "" {
				apiValue["name"] = name
			}
			if email, ok := valueMap["email"].(string); ok && email != "" {
				apiValue["email"] = email
			}

			apiValues[i] = apiValue
		}
	}

	return map[string]interface{}{
		"key":    key,
		"values": apiValues,
	}
}

func convertMetadataRuleFromAPI(apiRule map[string]interface{}) []interface{} {
	if apiRule == nil {
		return nil
	}

	key, ok := apiRule["key"]
	if !ok {
		return nil
	}
	keyStr, ok := key.(string)
	if !ok {
		return nil
	}

	values, ok := apiRule["values"]
	if !ok {
		return nil
	}
	valuesArr, ok := values.([]interface{})
	if !ok {
		return nil
	}

	tfMetadataValues := make([]interface{}, len(valuesArr))

	for i, v := range valuesArr {
		valueMap, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		tfValue := map[string]interface{}{}

		// Handle type (always present in API response)
		if itemType, ok := valueMap["type"]; ok {
			if itemTypeStr, ok := itemType.(string); ok {
				tfValue["type"] = itemTypeStr
			}
		}

		// For string values, handle the value field
		if value, ok := valueMap["value"]; ok {
			if valueStr, ok := value.(string); ok {
				tfValue["value"] = valueStr
			}
		}

		// For reference values, handle item_id, name, email
		if itemID, ok := valueMap["item_id"]; ok {
			if itemIDStr, ok := itemID.(string); ok {
				tfValue["item_id"] = itemIDStr
			}
		}

		if name, ok := valueMap["name"]; ok {
			if nameStr, ok := name.(string); ok {
				tfValue["name"] = nameStr
			}
		}

		if email, ok := valueMap["email"]; ok {
			if emailStr, ok := email.(string); ok {
				tfValue["email"] = emailStr
			}
		}

		tfMetadataValues[i] = tfValue
	}

	return []interface{}{
		map[string]interface{}{
			"key":            keyStr,
			"metadata_value": tfMetadataValues,
		},
	}
}

func statusPageResourceCustomizeDiff(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	// Validate mark_as_down_for and mark_as_down_metadata_rule relationship
	markAsDownFor, hasMarkAsDownFor := d.GetOk("mark_as_down_for")
	markAsDownMetadataRule := d.Get("mark_as_down_metadata_rule").([]interface{})

	if hasMarkAsDownFor && markAsDownFor.(string) == "incident_matching_metadata" {
		if len(markAsDownMetadataRule) == 0 {
			return fmt.Errorf("mark_as_down_metadata_rule is required when mark_as_down_for is 'incident_matching_metadata'")
		}
	} else if hasMarkAsDownFor && markAsDownFor.(string) != "incident_matching_metadata" {
		if len(markAsDownMetadataRule) > 0 {
			return fmt.Errorf("mark_as_down_metadata_rule can only be used when mark_as_down_for is 'incident_matching_metadata'")
		}
	}

	// Validate mark_as_degraded_for and mark_as_degraded_metadata_rule relationship
	markAsDegradedFor, hasMarkAsDegradedFor := d.GetOk("mark_as_degraded_for")
	markAsDegradedMetadataRule := d.Get("mark_as_degraded_metadata_rule").([]interface{})

	if hasMarkAsDegradedFor && markAsDegradedFor.(string) == "incident_matching_metadata" {
		if len(markAsDegradedMetadataRule) == 0 {
			return fmt.Errorf("mark_as_degraded_metadata_rule is required when mark_as_degraded_for is 'incident_matching_metadata'")
		}
	} else if hasMarkAsDegradedFor && markAsDegradedFor.(string) != "incident_matching_metadata" {
		if len(markAsDegradedMetadataRule) > 0 {
			return fmt.Errorf("mark_as_degraded_metadata_rule can only be used when mark_as_degraded_for is 'incident_matching_metadata'")
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
		CustomizeDiff: statusPageResourceCustomizeDiff,
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
	// TODO:  if reflect.TypeOf(in).NumField() != len([]struct)
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
		// Skip loading status history when preparing to send the request
		if e.k != "status_history" {
			if e.k == "mark_as_down_metadata_rule" {
				if tfRule := d.Get(e.k).([]interface{}); len(tfRule) > 0 {
					apiRule := convertMetadataRuleToAPI(tfRule)
					in.MarkAsDownMetadataRule = &apiRule
				}
			} else if e.k == "mark_as_degraded_metadata_rule" {
				if tfRule := d.Get(e.k).([]interface{}); len(tfRule) > 0 {
					apiRule := convertMetadataRuleToAPI(tfRule)
					in.MarkAsDegradedMetadataRule = &apiRule
				}
			} else {
				load(d, e.k, e.v)
			}
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
			if in.MarkAsDownMetadataRule != nil && *in.MarkAsDownMetadataRule != nil {
				tfRule := convertMetadataRuleFromAPI(*in.MarkAsDownMetadataRule)
				if tfRule != nil {
					if err := d.Set(e.k, tfRule); err != nil {
						derr = append(derr, diag.FromErr(err)[0])
					}
				}
			}
		} else if e.k == "mark_as_degraded_metadata_rule" {
			if in.MarkAsDegradedMetadataRule != nil && *in.MarkAsDegradedMetadataRule != nil {
				tfRule := convertMetadataRuleFromAPI(*in.MarkAsDegradedMetadataRule)
				if tfRule != nil {
					if err := d.Set(e.k, tfRule); err != nil {
						derr = append(derr, diag.FromErr(err)[0])
					}
				}
			}
		} else {
			if err := d.Set(e.k, reflect.Indirect(reflect.ValueOf(e.v)).Interface()); err != nil {
				derr = append(derr, diag.FromErr(err)[0])
			}
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
				if tfRule := d.Get(e.k).([]interface{}); len(tfRule) > 0 {
					apiRule := convertMetadataRuleToAPI(tfRule)
					in.MarkAsDownMetadataRule = &apiRule
				}
			} else if e.k == "mark_as_degraded_metadata_rule" {
				if tfRule := d.Get(e.k).([]interface{}); len(tfRule) > 0 {
					apiRule := convertMetadataRuleToAPI(tfRule)
					in.MarkAsDegradedMetadataRule = &apiRule
				}
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
