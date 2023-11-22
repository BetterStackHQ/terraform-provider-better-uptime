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
		url: "https://betteruptime.com",
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
				Description: "Better Uptime API Token. The value can be omitted if `BETTERUPTIME_API_TOKEN` environment variable is set. See https://docs.betteruptime.com/api/getting-started#obtaining-an-api-token on how to obtain the API token for your team.",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"betteruptime_monitor": newMonitorDataSource(),
			"betteruptime_policy":  newPolicyDataSource(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"betteruptime_email_integration":    newEmailIntegrationResource(),
			"betteruptime_heartbeat":            newHeartbeatResource(),
			"betteruptime_heartbeat_group":      newHeartbeatGroupResource(),
			"betteruptime_incoming_webhook":     newIncomingWebhookResource(),
			"betteruptime_metadata":             newMetadataResource(),
			"betteruptime_monitor":              newMonitorResource(),
			"betteruptime_monitor_group":        newMonitorGroupResource(),
			"betteruptime_policy":               newPolicyResource(),
			"betteruptime_status_page":          newStatusPageResource(),
			"betteruptime_status_page_section":  newStatusPageSectionResource(),
			"betteruptime_status_page_resource": newStatusPageResourceResource(),
		},
		ConfigureContextFunc: func(ctx context.Context, r *schema.ResourceData) (interface{}, diag.Diagnostics) {
			var userAgent string
			if spec.version != "" {
				userAgent = "terraform-provider-better-uptime/" + spec.version
			}
			c, err := newClient(spec.url, r.Get("api_token").(string),
				withHTTPClient(&http.Client{
					Timeout: time.Second * 60,
				}),
				withUserAgent(userAgent))
			return c, diag.FromErr(err)
		},
	}
}
