package provider

import (
	"context"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type provider struct {
	url     string
	version string
}

type Option func(*provider)

func WithURL(v string) Option {
	return func(p *provider) {
		p.url = v
	}
}

func WithVersion(v string) Option {
	return func(p *provider) {
		p.version = v
	}
}

func New(opts ...Option) *schema.Provider {
	spec := provider{
		url: "https://uptime.betterstack.com",
	}
	for _, opt := range opts {
		opt(&spec)
	}
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_token": {
				Type:        schema.TypeString,
				Sensitive:   true,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("BETTERUPTIME_API_TOKEN", nil),
				Description: "Better Stack Uptime API token. The value can be omitted if `BETTERUPTIME_API_TOKEN` environment variable is set. See https://betterstack.com/docs/uptime/api/getting-started-with-uptime-api/#obtaining-an-uptime-api-token on how to obtain the API token for your team.",
			},
			"api_retry_max": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     4,
				Description: "Maximum number of retries for API requests.",
			},
			"api_retry_wait_min": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     10,
				Description: "Minimum time to wait between retries in seconds.",
			},
			"api_retry_wait_max": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     300,
				Description: "Maximum time to wait between retries in seconds.",
			},
			"api_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     60,
				Description: "Timeout for individual HTTP requests in seconds.",
			},
			"api_rate_limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     8,
				Description: "Maximum number of API requests per second. 0 means no limit.",
			},
			"api_rate_burst": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Burst size for rate limiter, allows temporary bursts above the rate limit. 0 means use automatic default (2x rate limit, minimum 10).",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"betteruptime_monitor":           newMonitorDataSource(),
			"betteruptime_on_call_calendar":  newOnCallCalendarDataSource(),
			"betteruptime_policy":            newPolicyDataSource(),
			"betteruptime_severity":          newSeverityDataSource(),
			"betteruptime_slack_integration": newSlackIntegrationDataSource(),
			"betteruptime_incoming_webhook":  newIncomingWebhookDataSource(),
			"betteruptime_ip_list":           newIpListDataSource(),
			"betteruptime_team_member":       newTeamMemberDataSource(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"betteruptime_email_integration":             newEmailIntegrationResource(),
			"betteruptime_heartbeat":                     newHeartbeatResource(),
			"betteruptime_heartbeat_group":               newHeartbeatGroupResource(),
			"betteruptime_incoming_webhook":              newIncomingWebhookResource(),
			"betteruptime_metadata":                      newMetadataResource(),
			"betteruptime_monitor":                       newMonitorResource(),
			"betteruptime_monitor_group":                 newMonitorGroupResource(),
			"betteruptime_on_call_calendar":              newOnCallCalendarResource(),
			"betteruptime_policy":                        newPolicyResource(),
			"betteruptime_policy_group":                  newPolicyGroupResource(),
			"betteruptime_severity":                      newSeverityResource(),
			"betteruptime_severity_group":                newSeverityGroupResource(),
			"betteruptime_status_page":                   newStatusPageResource(),
			"betteruptime_status_page_group":             newStatusPageGroupResource(),
			"betteruptime_status_page_section":           newStatusPageSectionResource(),
			"betteruptime_status_page_resource":          newStatusPageResourceResource(),
			"betteruptime_pagerduty_integration":         newPagerdutyIntegrationResource(),
			"betteruptime_splunk_oncall_integration":     newSplunkOnCallIntegrationResource(),
			"betteruptime_aws_cloudwatch_integration":    newAwsCloudWatchIntegrationResource(),
			"betteruptime_azure_integration":             newAzureIntegrationResource(),
			"betteruptime_datadog_integration":           newDatadogIntegrationResource(),
			"betteruptime_google_monitoring_integration": newGoogleMonitoringIntegrationResource(),
			"betteruptime_new_relic_integration":         newNewRelicIntegrationResource(),
			"betteruptime_grafana_integration":           newGrafanaIntegrationResource(),
			"betteruptime_elastic_integration":           newElasticIntegrationResource(),
			"betteruptime_prometheus_integration":        newPrometheusIntegrationResource(),
			"betteruptime_outgoing_webhook":              newOutgoingWebhookResource(),
			"betteruptime_jira_integration":              newJiraIntegrationResource(),
			"betteruptime_catalog_relation":              newCatalogRelationResource(),
			"betteruptime_catalog_attribute":             newCatalogAttributeResource(),
			"betteruptime_catalog_record":                newCatalogRecordResource(),
			"betteruptime_team_member":                   newTeamMemberResource(),
		},
		ConfigureContextFunc: func(ctx context.Context, r *schema.ResourceData) (interface{}, diag.Diagnostics) {
			var userAgent string
			if spec.version != "" {
				userAgent = "terraform-provider-better-uptime/" + spec.version
			}

			timeout := time.Duration(r.Get("api_timeout").(int)) * time.Second

			c, err := newClient(ClientConfig{
				BaseURL:      spec.url,
				Token:        r.Get("api_token").(string),
				UserAgent:    userAgent,
				HTTPClient:   &http.Client{Timeout: timeout},
				RetryMax:     r.Get("api_retry_max").(int),
				RetryWaitMin: time.Duration(r.Get("api_retry_wait_min").(int)) * time.Second,
				RetryWaitMax: time.Duration(r.Get("api_retry_wait_max").(int)) * time.Second,
				RateLimit:    r.Get("api_rate_limit").(int),
				RateBurst:    r.Get("api_rate_burst").(int),
			})
			return c, diag.FromErr(err)
		},
	}
}
