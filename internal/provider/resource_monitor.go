package provider

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TODO: change to map<name, description> and then use to gen monitor_type description
var monitorTypes = []string{"status", "expected_status_code", "keyword", "keyword_absence", "ping", "tcp", "udp", "smtp", "pop", "imap"}
var monitorSchema = map[string]*schema.Schema{
	"id": {
		Description: "The ID of this Monitor.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"ssl_expiration": {
		Description: "How many days before the SSL certificate expires do you want to be alerted?" +
			" Valid values are 1, 2, 3, 7, 14, 30, and 60.",
		Type:     schema.TypeInt,
		Optional: true,
		// TODO: ValidateDiagFunc: validation.IntInSlice
	},
	"domain_expiration": {
		Description: "How many days before the domain expires do you want to be alerted?" +
			" Valid values are 1, 2, 3, 7, 14, 30, and 60.",
		Type:     schema.TypeInt,
		Optional: true,
		// TODO: ValidateDiagFunc: validation.IntInSlice
	},
	"policy_id": {
		Description: "Set the escalation policy for the monitor.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"url": {
		Description: "URL of your website or the host you want to ping (see monitor_type below).",
		Type:        schema.TypeString,
		Required:    true,
	},
	"monitor_type": {
		Description: strings.ReplaceAll(`Valid values:

    **status** We will check your website for 2XX HTTP status code.

	**expected_status_code** We will check if your website returned one of the values in expected_status_codes.

    **keyword** We will check if your website contains the required_keyword.

    **keyword_absence** We will check if your website doesn't contain the required_keyword.

    **ping** We will ping your host specified in the url parameter.

    **tcp** We will test a TCP port at your host specified in the url parameter
(port is required).

    **udp** We will test a UDP port at your host specified in the url parameter
(port and required_keyword are required).

    **smtp** We will check for a SMTP server at the host specified in the url parameter
(port is required, and can be one of 25, 465, 587, or a combination of those ports separated by comma).

    **pop** We will check for a POP3 server at the host specified in the url parameter
(port is required, and can be 110, 995, or both).

    **imap** We will check for an IMAP server at the host specified in the url parameter
(port is required, and can be 143, 993, or both).`, "**", "`"),
		Type:     schema.TypeString,
		Required: true,
		ValidateDiagFunc: func(v interface{}, path cty.Path) diag.Diagnostics {
			s := v.(string)
			for _, monitorType := range monitorTypes {
				if s == monitorType {
					return nil
				}
			}
			return diag.Diagnostics{
				diag.Diagnostic{
					AttributePath: path,
					Severity:      diag.Error,
					Summary:       `Invalid "monitor_type"`,
					Detail:        fmt.Sprintf("Expected one of %v", monitorTypes),
				},
			}
		},
	},
	"required_keyword": {
		Description: "Required if monitor_type is set to keyword  or udp. We will create a new incident if this keyword is missing on your page.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"expected_status_codes": {
		Description: "Required if monitor_type is set to expected_status_code. We will create a new incident if the status code returned from the server is not in the list of expected status codes.",
		Type:        schema.TypeList,
		Elem: &schema.Schema{
			Type: schema.TypeInt,
		},
		Optional: true,
	},
	"call": {
		Description: "Should we call the on-call person?",
		Type:        schema.TypeBool,
		Optional:    true,
	},
	"sms": {
		Description: "Should we send an SMS to the on-call person?",
		Type:        schema.TypeBool,
		Optional:    true,
	},
	"email": {
		Description: "Should we send an email to the on-call person?",
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
	},
	"push": {
		Description: "Should we send a push notification to the on-call person?",
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
	},
	"team_wait": {
		Description: "How long to wait before escalating the incident alert to the team. Leave blank to disable escalating to the entire team.",
		Type:        schema.TypeInt,
		Optional:    true,
	},
	"paused": {
		Description: "Set to true to pause monitoring - we won't notify you about downtime. Set to false to resume monitoring.",
		Type:        schema.TypeBool,
		Optional:    true,
	},
	"paused_at": {
		Description: "The time when this monitor was paused.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"follow_redirects": {
		Description: "Set to true for the monitor to follow redirects.",
		Type:        schema.TypeBool,
		Default:     true,
		Optional:    true,
	},
	"port": {
		Description: "Required if monitor_type is set to tcp, udp, smtp, pop, or imap." +
			" tcp and udp monitors accept any ports, while smtp, pop, and imap accept only the specified ports corresponding with their servers (e.g. \"25,465,587\" for smtp).",
		Type:     schema.TypeString,
		Optional: true,
	},
	"regions": {
		Description: "An array of regions to set. Allowed values are [\"us\", \"eu\", \"as\", \"au\"] or any subset of these regions.",
		Type:        schema.TypeList,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Optional: true,
		// TODO: ValidateDiagFunc
	},
	"monitor_group_id": {
		Description: "Set this attribute if you want to add this monitor to a monitor group.",
		Type:        schema.TypeInt,
		Optional:    true,
	},
	"pronounceable_name": {
		Description: "Pronounceable name of the monitor. We will use this when we call you. Try to make it tongue-friendly, please?",
		Type:        schema.TypeString,
		Optional:    true,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return new == "" || old == new
		},
	},
	"recovery_period": {
		Description: "How long the monitor must be up to automatically mark an incident as resolved after being down.",
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     180,
	},
	"verify_ssl": {
		Description: "Should we verify SSL certificate validity?",
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
	},
	"check_frequency": {
		Description: "How often should we check your website? In seconds.",
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     180,
	},
	"confirmation_period": {
		Description: "How long should we wait after observing a failure before we start a new incident?",
		Type:        schema.TypeInt,
		Optional:    true,
	},
	"http_method": {
		Description: "HTTP Method used to make a request. Valid options: GET, HEAD, POST, PUT, PATCH",
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "GET",
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return strings.EqualFold(old, new)
		},
		// TODO: ValidateDiagFunc: validation.StringInSlice
	},
	"request_timeout": {
		Description: "How long to wait before timing out the request? In seconds.",
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     30,
	},
	"request_body": {
		Description: "Request body for POST, PUT, PATCH requests.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"request_headers": {
		Description: "An array of request headers, consisting of name and value pairs",
		Type:        schema.TypeList,
		Elem: &schema.Schema{
			Type: schema.TypeMap,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		Optional: true,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			// Ignore ID changes
			attribute := strings.Split(k, ".")
			if len(attribute) > 2 && attribute[2] == "id" {
				return true
			} else {
				return false
			}
		},
	},
	"auth_username": {
		Description: "Basic HTTP authentication username to include with the request.",
		Type:        schema.TypeString,
		Optional:    true,
		Sensitive:   true,
	},
	"auth_password": {
		Description: "Basic HTTP authentication password to include with the request.",
		Type:        schema.TypeString,
		Optional:    true,
		Sensitive:   true,
	},
	"maintenance_from": {
		Description: "Start of the maintenance window each day. We won't check your website during this window. Example: \"01:00:00\"",
		Type:        schema.TypeString,
		Optional:    true,
		// TODO: ValidateDiagFunc
	},
	"maintenance_to": {
		Description: "End of the maintenance window each day. Example: \"03:00:00\"",
		Type:        schema.TypeString,
		Optional:    true,
		// TODO: ValidateDiagFunc
	},
	"maintenance_timezone": {
		Description: "The timezone to use for the maintenance window each day. Defaults to UTC. The accepted values can be found in the Rails TimeZone documentation. https://api.rubyonrails.org/classes/ActiveSupport/TimeZone.html",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"remember_cookies": {
		Description: "Set to true to keep cookies when redirecting.",
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
	},
	"last_checked_at": {
		Description: "When the website was last checked.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"status": {
		Description: "The status of this website check.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"created_at": {
		Description: "The time when this monitor was created.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"updated_at": {
		Description: "The time when this monitor was updated.",
		Type:        schema.TypeString,
		Computed:    true,
	},
}

func newMonitorResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: monitorCreate,
		ReadContext:   monitorRead,
		UpdateContext: monitorUpdate,
		DeleteContext: monitorDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "https://docs.betteruptime.com/api/monitors-api",
		Schema:      monitorSchema,
	}
}

type monitor struct {
	SSLExpiration       *int                      `json:"ssl_expiration,omitempty"`
	DomainExpiration    *int                      `json:"domain_expiration,omitempty"`
	PolicyID            *string                   `json:"policy_id,omitempty"`
	URL                 *string                   `json:"url,omitempty"`
	MonitorType         *string                   `json:"monitor_type,omitempty"`
	RequiredKeyword     *string                   `json:"required_keyword,omitempty"`
	ExpectedStatusCodes *[]int                    `json:"expected_status_codes,omitempty"`
	Call                *bool                     `json:"call,omitempty"`
	SMS                 *bool                     `json:"sms,omitempty"`
	Email               *bool                     `json:"email,omitempty"`
	Push                *bool                     `json:"push,omitempty"`
	TeamWait            *int                      `json:"team_wait,omitempty"`
	Paused              *bool                     `json:"paused,omitempty"`
	PausedAt            *string                   `json:"paused_at,omitempty"`
	FollowRedirects     *bool                     `json:"follow_redirects,omitempty"`
	Port                *string                   `json:"port,omitempty"`
	Regions             *[]string                 `json:"regions,omitempty"`
	MonitorGroupID      *int                      `json:"monitor_group_id,omitempty"`
	PronounceableName   *string                   `json:"pronounceable_name,omitempty"`
	RecoveryPeriod      *int                      `json:"recovery_period,omitempty"`
	VerifySSL           *bool                     `json:"verify_ssl,omitempty"`
	CheckFrequency      *int                      `json:"check_frequency,omitempty"`
	ConfirmationPeriod  *int                      `json:"confirmation_period,omitempty"`
	HTTPMethod          *string                   `json:"http_method,omitempty"`
	RequestTimeout      *int                      `json:"request_timeout,omitempty"`
	RequestBody         *string                   `json:"request_body,omitempty"`
	RequestHeaders      *[]map[string]interface{} `json:"request_headers,omitempty"`
	AuthUsername        *string                   `json:"auth_username,omitempty"`
	AuthPassword        *string                   `json:"auth_password,omitempty"`
	MaintenanceFrom     *string                   `json:"maintenance_from,omitempty"`
	MaintenanceTo       *string                   `json:"maintenance_to,omitempty"`
	MaintenanceTimezone *string                   `json:"maintenance_timezone,omitempty"`
	RememberCookies     *bool                     `json:"remember_cookies,omitempty"`
	LastCheckedAt       *string                   `json:"last_checked_at,omitempty"`
	Status              *string                   `json:"status,omitempty"`
	CreatedAt           *string                   `json:"created_at,omitempty"`
	UpdatedAt           *string                   `json:"updated_at,omitempty"`
}

type monitorHTTPResponse struct {
	Data struct {
		ID         string  `json:"id"`
		Attributes monitor `json:"attributes"`
	} `json:"data"`
}

func monitorRef(in *monitor) []struct {
	k string
	v interface{}
} {
	// TODO:  if reflect.TypeOf(in).NumField() != len([]struct)
	return []struct {
		k string
		v interface{}
	}{
		{k: "ssl_expiration", v: &in.SSLExpiration},
		{k: "domain_expiration", v: &in.DomainExpiration},
		{k: "policy_id", v: &in.PolicyID},
		{k: "url", v: &in.URL},
		{k: "monitor_type", v: &in.MonitorType},
		{k: "required_keyword", v: &in.RequiredKeyword},
		{k: "expected_status_codes", v: &in.ExpectedStatusCodes},
		{k: "call", v: &in.Call},
		{k: "sms", v: &in.SMS},
		{k: "email", v: &in.Email},
		{k: "push", v: &in.Push},
		{k: "team_wait", v: &in.TeamWait},
		{k: "paused", v: &in.Paused},
		{k: "paused_at", v: &in.PausedAt},
		{k: "follow_redirects", v: &in.FollowRedirects},
		{k: "port", v: &in.Port},
		{k: "regions", v: &in.Regions},
		{k: "monitor_group_id", v: &in.MonitorGroupID},
		{k: "pronounceable_name", v: &in.PronounceableName},
		{k: "recovery_period", v: &in.RecoveryPeriod},
		{k: "verify_ssl", v: &in.VerifySSL},
		{k: "check_frequency", v: &in.CheckFrequency},
		{k: "confirmation_period", v: &in.ConfirmationPeriod},
		{k: "http_method", v: &in.HTTPMethod},
		{k: "request_timeout", v: &in.RequestTimeout},
		{k: "request_body", v: &in.RequestBody},
		{k: "request_headers", v: &in.RequestHeaders},
		{k: "auth_username", v: &in.AuthUsername},
		{k: "auth_password", v: &in.AuthPassword},
		{k: "maintenance_from", v: &in.MaintenanceFrom},
		{k: "maintenance_to", v: &in.MaintenanceTo},
		{k: "maintenance_timezone", v: &in.MaintenanceTimezone},
		{k: "remember_cookies", v: &in.RememberCookies},
		{k: "last_checked_at", v: &in.LastCheckedAt},
		{k: "status", v: &in.Status},
		{k: "created_at", v: &in.CreatedAt},
		{k: "updated_at", v: &in.UpdatedAt},
	}
}

func monitorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in monitor
	for _, e := range monitorRef(&in) {
		if e.k == "request_headers" {
			loadRequestHeaders(d, e.v.(**[]map[string]interface{}))
		} else {
			load(d, e.k, e.v)
		}
	}
	var out monitorHTTPResponse
	if err := resourceCreate(ctx, meta, "/api/v2/monitors", &in, &out); err != nil {
		return err
	}
	d.SetId(out.Data.ID)
	return monitorCopyAttrs(d, &out.Data.Attributes)
}

func monitorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var out monitorHTTPResponse
	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/monitors/%s", url.PathEscape(d.Id())), &out); err != nil {
		return err
	} else if !ok {
		d.SetId("") // Force "create" on 404.
		return nil
	}
	return monitorCopyAttrs(d, &out.Data.Attributes)
}

func monitorCopyAttrs(d *schema.ResourceData, in *monitor) diag.Diagnostics {
	var derr diag.Diagnostics
	for _, e := range monitorRef(in) {
		if err := d.Set(e.k, reflect.Indirect(reflect.ValueOf(e.v)).Interface()); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}
	return derr
}

func monitorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in monitor
	var out policyHTTPResponse
	for _, e := range monitorRef(&in) {
		if d.HasChange(e.k) {
			if e.k == "request_headers" {
				loadRequestHeaders(d, e.v.(**[]map[string]interface{}))
			} else {
				load(d, e.k, e.v)
			}
		}
	}

	return resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/monitors/%s", url.PathEscape(d.Id())), &in, &out)
}

func monitorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDelete(ctx, meta, fmt.Sprintf("/api/v2/monitors/%s", url.PathEscape(d.Id())))
}

func loadRequestHeaders(d *schema.ResourceData, receiver **[]map[string]interface{}) {
	x := receiver
	v := d.Get("request_headers")
	var t []map[string]interface{}
	for _, v := range v.([]interface{}) {
		header := v.(map[string]interface{})
		newHeader := map[string]interface{}{"name": header["name"], "value": header["value"]}
		t = append(t, newHeader)
	}

	// Retrieve old requests and construct "_destroy" attribute for the ones we no longer want
	old, _ := d.GetChange("request_headers")
	for _, v := range old.([]interface{}) {
		header := v.(map[string]interface{})
		foundHeader := findRequestHeader(&t, &header)

		var hasId bool // Check if the found header already has an ID
		if foundHeader == nil {
			hasId = false
		} else {
			_, hasId = (*foundHeader)["id"]
		}

		if foundHeader == nil || hasId {
			headerToDestroy := map[string]interface{}{"id": header["id"].(string), "_destroy": "true"}
			t = append(t, headerToDestroy)
		} else if header["id"] != nil {
			(*foundHeader)["id"] = header["id"].(string)
		}
	}

	*x = &t
}

func findRequestHeader(headers *[]map[string]interface{}, header *map[string]interface{}) *map[string]interface{} {
	for _, h := range *headers {
		if h["name"] == (*header)["name"] && h["value"] == (*header)["value"] {
			return &h
		}
	}
	return nil
}
