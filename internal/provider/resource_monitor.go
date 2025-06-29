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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// TODO: change to map<name, description> and then use to gen monitor_type description
var monitorTypes = []string{"status", "expected_status_code", "keyword", "keyword_absence", "ping", "tcp", "udp", "smtp", "pop", "imap", "dns", "playwright"}
var ipVersions = []string{"ipv4", "ipv6"}
var monitorSchema = map[string]*schema.Schema{
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
		Description: "The ID of this Monitor.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"ssl_expiration": {
		Description: "How many days before the SSL certificate expires do you want to be alerted?" +
			" Valid values are 1, 2, 3, 7, 14, 30, and 60. Set to -1 to disable SSL expiration check.",
		Type:         schema.TypeInt,
		Optional:     true,
		Computed:     true,
		ValidateFunc: validation.IntInSlice([]int{1, 2, 3, 7, 14, 30, 60, -1}),
	},
	"domain_expiration": {
		Description: "How many days before the domain expires do you want to be alerted?" +
			" Valid values are 1, 2, 3, 7, 14, 30, and 60. Set to -1 to disable domain expiration check.",
		Type:         schema.TypeInt,
		Optional:     true,
		Computed:     true,
		ValidateFunc: validation.IntInSlice([]int{1, 2, 3, 7, 14, 30, 60, -1}),
	},
	"policy_id": {
		Description: "Set the escalation policy for the monitor.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"expiration_policy_id": {
		Description: "Set the expiration escalation policy for the monitor. It is used for SSL certificate and domain expiration checks. When set to null, an e-mail is sent to the entire team.",
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     nil,
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
(port is required, and can be 143, 993, or both).

    **dns** We will check for a DNS server at the host specified in the url parameter
(request_body is required, and should contain the domain to query the DNS server with).

    **playwright** We will run the scenario defined by playwright_script, identified in the UI by scenario_name`, "**", "`"),
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
		Computed:    true,
	},
	"expected_status_codes": {
		Description: "Required if monitor_type is set to expected_status_code. We will create a new incident if the status code returned from the server is not in the list of expected status codes.",
		Type:        schema.TypeList,
		Elem: &schema.Schema{
			Type: schema.TypeInt,
		},
		Optional: true,
		Computed: true,
	},
	"call": {
		Description: "Whether to call when a new incident is created.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"sms": {
		Description: "Whether to send an SMS when a new incident is created.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"email": {
		Description: "Whether to send an email when a new incident is created.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"push": {
		Description: "Whether to send a push notification when a new incident is created.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"critical_alert": {
		Description: "Whether to send a critical push notification that ignores the mute switch and Do not Disturb mode when a new incident is created.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"team_wait": {
		Description: "How long to wait before escalating the incident alert to the team. Leave blank to disable escalating to the entire team. In seconds.",
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
	},
	"paused": {
		Description: "Set to true to pause monitoring - we won't notify you about downtime. Set to false to resume monitoring.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"paused_at": {
		Description: "The time when this monitor was paused.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"follow_redirects": {
		Description: "Set to true for the monitor to follow redirects.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"port": {
		Description: "Required if monitor_type is set to tcp, udp, smtp, pop, or imap." +
			" tcp and udp monitors accept any ports, while smtp, pop, and imap accept only the specified ports corresponding with their servers (e.g. \"25,465,587\" for smtp).",
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"regions": {
		Description: "An array of regions to set. Allowed values are [\"us\", \"eu\", \"as\", \"au\"] or any subset of these regions.",
		Type:        schema.TypeList,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Optional: true,
		Computed: true,
		// TODO: ValidateDiagFunc
	},
	"monitor_group_id": {
		Description: "Set this attribute if you want to add this monitor to a monitor group.",
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
	},
	"pronounceable_name": {
		Description: "Pronounceable name of the monitor. We will use this when we call you. Try to make it tongue-friendly, please?",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return new == "" || old == new
		},
	},
	"recovery_period": {
		Description: "How long the monitor must be up to automatically mark an incident as resolved after being down. In seconds.",
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
	},
	"verify_ssl": {
		Description: "Should we verify SSL certificate validity?",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"check_frequency": {
		Description: "How often should we check your website? In seconds.",
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
	},
	"confirmation_period": {
		Description: "How long should we wait after observing a failure before we start a new incident? In seconds.",
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
	},
	"http_method": {
		Description: "HTTP Method used to make a request. Valid options: GET, HEAD, POST, PUT, PATCH",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return strings.EqualFold(old, new)
		},
		// TODO: ValidateDiagFunc: validation.StringInSlice
	},
	"request_timeout": {
		Description: "How long to wait before timing out the request?\n" +
			"  - For Server and Port monitors (types `ping`, `tcp`, `udp`, `smtp`, `pop`, `imap` and `dns`) the timeout is specified in *milliseconds*. Valid options: 500, 1000, 2000, 3000, 5000.\n" +
			"  - For Playwright monitors (type `playwright`), this determines the Playwright scenario timeout instead in *seconds*. Valid options: 15, 30, 45, 60.\n" +
			"  - For all other monitors, the timeout is specified in *seconds*. Valid options: 2, 3, 5, 10, 15, 30, 45, 60.\n",
		Type:     schema.TypeInt,
		Optional: true,
		Computed: true,
	},
	"request_body": {
		Description: "Request body for POST, PUT, PATCH requests. Required if monitor_type is set to dns (domain to query the DNS server with).",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
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
		Computed: true,
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
		Computed:    true,
		Sensitive:   true,
	},
	"auth_password": {
		Description: "Basic HTTP authentication password to include with the request.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		Sensitive:   true,
	},

	"ip_version": {
		Description: strings.ReplaceAll(`Valid values:

    **ipv4** Use IPv4 only,

    **ipv6** Use IPv6 only.`, "**", "`"),
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
		ValidateDiagFunc: func(v interface{}, path cty.Path) diag.Diagnostics {
			if v == nil {
				return nil
			}
			s := v.(string)
			for _, ipVersion := range ipVersions {
				if s == ipVersion {
					return nil
				}
			}
			return diag.Diagnostics{
				diag.Diagnostic{
					AttributePath: path,
					Severity:      diag.Error,
					Summary:       `Invalid "ip_version"`,
					Detail:        fmt.Sprintf("Expected one of %v or nil", ipVersions),
				},
			}
		},
	},
	"maintenance_from": {
		Description: "Start of the maintenance window each day. We won't check your website during this window. Example: \"01:00:00\"",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		// TODO: ValidateDiagFunc
	},
	"maintenance_to": {
		Description: "End of the maintenance window each day. Example: \"03:00:00\"",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		// TODO: ValidateDiagFunc
	},
	"maintenance_timezone": {
		Description: "The timezone to use for the maintenance window each day. Defaults to UTC. The accepted values can be found in the Rails TimeZone documentation. https://api.rubyonrails.org/classes/ActiveSupport/TimeZone.html",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"maintenance_days": {
		Description: "An array of maintenance days to set. If a maintenance window is overnight both affected days should be set. Allowed values are [\"mon\", \"tue\", \"wed\", \"thu\", \"fri\", \"sat\", \"sun\"] or any subset of these days.",
		Type:        schema.TypeList,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Optional: true,
		Computed: true,
	},
	"remember_cookies": {
		Description: "Set to true to keep cookies when redirecting.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"last_checked_at": {
		Description: "When the website was last checked.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"status": {
		Description: "The status of this website check.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"created_at": {
		Description: "The time when this monitor was created.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"updated_at": {
		Description: "The time when this monitor was updated.",
		Type:        schema.TypeString,
		Optional:    false,
		Computed:    true,
	},
	"playwright_script": {
		Description: "For Playwright monitors, the JavaScript source code of the scenario.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"scenario_name": {
		Description: "For Playwright monitors, the scenario name identifying the monitor in the UI.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"environment_variables": {
		Description: "For Playwright monitors, the environment variables that can be used in the scenario. Example: `{ \"PASSWORD\" = \"passw0rd\" }`.",
		Type:        schema.TypeMap,
		Elem:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		Sensitive:   true,
		ValidateDiagFunc: func(v interface{}, path cty.Path) diag.Diagnostics {
			m := v.(map[string]interface{})
			for k, v := range m {
				if k == "" {
					return diag.Diagnostics{
						diag.Diagnostic{
							AttributePath: path,
							Severity:      diag.Error,
							Summary:       `Invalid "environment_variables"`,
							Detail:        "Environment variable name cannot be empty",
						},
					}
				}
				if v == "" {
					return diag.Diagnostics{
						diag.Diagnostic{
							AttributePath: path,
							Severity:      diag.Error,
							Summary:       `Invalid "environment_variables"`,
							Detail:        fmt.Sprintf("Environment variable value for key %q cannot be empty", k),
						},
					}
				}
			}
			return nil
		},
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
		CustomizeDiff: validateRequestHeaders,
		Description:   "https://betterstack.com/docs/uptime/api/monitors/",
		Schema:        monitorSchema,
	}
}

type monitor struct {
	SSLExpiration        *NullableInt              `json:"ssl_expiration,omitempty"`
	DomainExpiration     *NullableInt              `json:"domain_expiration,omitempty"`
	PolicyID             *string                   `json:"policy_id,omitempty"`
	ExpirationPolicyID   *int                      `json:"expiration_policy_id"`
	URL                  *string                   `json:"url,omitempty"`
	MonitorType          *string                   `json:"monitor_type,omitempty"`
	RequiredKeyword      *string                   `json:"required_keyword,omitempty"`
	ExpectedStatusCodes  *[]int                    `json:"expected_status_codes,omitempty"`
	Call                 *bool                     `json:"call,omitempty"`
	SMS                  *bool                     `json:"sms,omitempty"`
	Email                *bool                     `json:"email,omitempty"`
	Push                 *bool                     `json:"push,omitempty"`
	CriticalAlert        *bool                     `json:"critical_alert,omitempty"`
	TeamWait             *int                      `json:"team_wait,omitempty"`
	Paused               *bool                     `json:"paused,omitempty"`
	PausedAt             *string                   `json:"paused_at,omitempty"`
	FollowRedirects      *bool                     `json:"follow_redirects,omitempty"`
	Port                 *string                   `json:"port,omitempty"`
	Regions              *[]string                 `json:"regions,omitempty"`
	MonitorGroupID       *int                      `json:"monitor_group_id,omitempty"`
	PronounceableName    *string                   `json:"pronounceable_name,omitempty"`
	RecoveryPeriod       *int                      `json:"recovery_period,omitempty"`
	VerifySSL            *bool                     `json:"verify_ssl,omitempty"`
	CheckFrequency       *int                      `json:"check_frequency,omitempty"`
	ConfirmationPeriod   *int                      `json:"confirmation_period,omitempty"`
	HTTPMethod           *string                   `json:"http_method,omitempty"`
	RequestTimeout       *int                      `json:"request_timeout,omitempty"`
	RequestBody          *string                   `json:"request_body,omitempty"`
	RequestHeaders       *[]map[string]interface{} `json:"request_headers,omitempty"`
	AuthUsername         *string                   `json:"auth_username,omitempty"`
	AuthPassword         *string                   `json:"auth_password,omitempty"`
	IpVersion            *string                   `json:"ip_version,omitempty"`
	MaintenanceFrom      *string                   `json:"maintenance_from,omitempty"`
	MaintenanceTo        *string                   `json:"maintenance_to,omitempty"`
	MaintenanceTimezone  *string                   `json:"maintenance_timezone,omitempty"`
	MaintenanceDays      *[]string                 `json:"maintenance_days,omitempty"`
	RememberCookies      *bool                     `json:"remember_cookies,omitempty"`
	LastCheckedAt        *string                   `json:"last_checked_at,omitempty"`
	Status               *string                   `json:"status,omitempty"`
	CreatedAt            *string                   `json:"created_at,omitempty"`
	UpdatedAt            *string                   `json:"updated_at,omitempty"`
	PlaywrightScript     *string                   `json:"playwright_script,omitempty"`
	ScenarioName         *string                   `json:"scenario_name,omitempty"`
	EnvironmentVariables *map[string]string        `json:"environment_variables,omitempty"`
	TeamName             *string                   `json:"team_name,omitempty"`
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
		{k: "expiration_policy_id", v: &in.ExpirationPolicyID},
		{k: "url", v: &in.URL},
		{k: "monitor_type", v: &in.MonitorType},
		{k: "required_keyword", v: &in.RequiredKeyword},
		{k: "expected_status_codes", v: &in.ExpectedStatusCodes},
		{k: "call", v: &in.Call},
		{k: "sms", v: &in.SMS},
		{k: "email", v: &in.Email},
		{k: "push", v: &in.Push},
		{k: "critical_alert", v: &in.CriticalAlert},
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
		{k: "ip_version", v: &in.IpVersion},
		{k: "maintenance_from", v: &in.MaintenanceFrom},
		{k: "maintenance_to", v: &in.MaintenanceTo},
		{k: "maintenance_timezone", v: &in.MaintenanceTimezone},
		{k: "maintenance_days", v: &in.MaintenanceDays},
		{k: "remember_cookies", v: &in.RememberCookies},
		{k: "last_checked_at", v: &in.LastCheckedAt},
		{k: "status", v: &in.Status},
		{k: "created_at", v: &in.CreatedAt},
		{k: "updated_at", v: &in.UpdatedAt},
		{k: "playwright_script", v: &in.PlaywrightScript},
		{k: "scenario_name", v: &in.ScenarioName},
		{k: "environment_variables", v: &in.EnvironmentVariables},
	}
}

func monitorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in monitor
	for _, e := range monitorRef(&in) {
		if e.k == "request_headers" {
			if err := loadRequestHeaders(d, e.v.(**[]map[string]interface{})); err != nil {
				return diag.FromErr(err)
			}
		} else if e.k == "expiration_policy_id" {
			// Work around the fact that Terraform represents null value as 0
			loadExpirationPolicy(d, e.v.(**int))
		} else if e.k == "domain_expiration" {
			in.DomainExpiration = NullableIntFromResourceData(d, e.k, -1)
		} else if e.k == "ssl_expiration" {
			in.SSLExpiration = NullableIntFromResourceData(d, e.k, -1)
		} else {
			load(d, e.k, e.v)
		}
	}
	load(d, "team_name", &in.TeamName)
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
		if e.k == "ssl_expiration" {
			if err := SetNullableIntResourceData(d, "ssl_expiration", -1, in.SSLExpiration); err != nil {
				derr = append(derr, diag.FromErr(err)[0])
			}
		} else if e.k == "domain_expiration" {
			if err := SetNullableIntResourceData(d, "domain_expiration", -1, in.DomainExpiration); err != nil {
				derr = append(derr, diag.FromErr(err)[0])
			}
		} else if err := d.Set(e.k, reflect.Indirect(reflect.ValueOf(e.v)).Interface()); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}
	return derr
}

func monitorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in monitor
	var out policyHTTPResponse
	for _, e := range monitorRef(&in) {
		if e.k == "expiration_policy_id" {
			// Work around the fact that Terraform represents null value as 0
			loadExpirationPolicy(d, e.v.(**int))
		} else if d.HasChange(e.k) {
			if e.k == "request_headers" {
				if err := loadRequestHeaders(d, e.v.(**[]map[string]interface{})); err != nil {
					return diag.FromErr(err)
				}
			} else if e.k == "domain_expiration" {
				in.DomainExpiration = NullableIntFromResourceData(d, "domain_expiration", -1)
			} else if e.k == "ssl_expiration" {
				in.SSLExpiration = NullableIntFromResourceData(d, "ssl_expiration", -1)
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

func validateRequestHeaders(ctx context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if headers, ok := diff.GetOk("request_headers"); ok {
		for _, header := range headers.([]interface{}) {
			headerMap := header.(map[string]interface{})
			if err := validateRequestHeader(headerMap); err != nil {
				return fmt.Errorf("invalid request header %v: %v", headerMap, err)
			}
		}
	}
	return nil
}

func validateRequestHeader(header map[string]interface{}) error {
	if len(header) == 0 {
		// Headers with calculated fields that are not known at the time will be passed as empty maps, ignore them
		return nil
	}

	name, nameOk := header["name"].(string)
	value, valueOk := header["value"].(string)

	if !nameOk || name == "" {
		return fmt.Errorf("must contain 'name' key with a non-empty string value")
	}

	if !valueOk || value == "" {
		return fmt.Errorf("must contain 'value' key with a non-empty string value")
	}

	if len(header) != 2 {
		return fmt.Errorf("must only contain 'name' and 'value' keys")
	}

	return nil
}

func loadRequestHeaders(d *schema.ResourceData, receiver **[]map[string]interface{}) error {
	x := receiver
	v := d.Get("request_headers")
	var t []map[string]interface{}
	for _, v := range v.([]interface{}) {
		header := v.(map[string]interface{})

		// Validation at apply time, empty map is considered invalid (fields should be known at this point)
		if len(header) == 0 {
			return fmt.Errorf("invalid request header %v: map cannot be empty", header)
		}
		// Headers can have ID at apply time, temporarily remove it before validation and reattach it afterwards
		id, idPresent := header["id"]
		delete(header, "id")
		err := validateRequestHeader(header)
		if idPresent {
			header["id"] = id
		}
		if err != nil {
			return fmt.Errorf("invalid request header %v: %v", header, err)
		}

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

	return nil
}

func findRequestHeader(headers *[]map[string]interface{}, header *map[string]interface{}) *map[string]interface{} {
	for _, h := range *headers {
		if h["name"] == (*header)["name"] && h["value"] == (*header)["value"] {
			return &h
		}
	}
	return nil
}

func loadExpirationPolicy(d *schema.ResourceData, receiver **int) {
	if v, ok := d.GetOk("expiration_policy_id"); ok {
		t := v.(int)
		if t == 0 {
			*receiver = nil
		} else {
			*receiver = &t
		}
	}
}
