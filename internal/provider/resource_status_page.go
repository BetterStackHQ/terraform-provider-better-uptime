package provider

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var statusPageSchema = map[string]*schema.Schema{
	"id": {
		Description: "The ID of this Status Page.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"history": {
		Type:        schema.TypeInt,
		Optional:    true,
		Description: "Number of days to display on the status page. Minimum 30 days.",
		ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
			history := val.(int)
			if history < 30 {
				errs = append(errs, fmt.Errorf("%q must be at least 30, got: %d", key, history))
			}
			return
		},
		Computed: true,
	},
	"company_name": {
		Description: "Name of your company.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"company_url": {
		Description: "URL of your company's website.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"contact_url": {
		Description: "URL that should be used for contacting you in case of an emergency.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"logo_url": {
		Description: "A direct link to your company's logo. The image should be under 20MB in size.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"timezone": {
		Description: "What timezone should we display your status page in? The accepted values can be found in the Rails TimeZone documentation. https://api.rubyonrails.org/classes/ActiveSupport/TimeZone.html",
		Type:        schema.TypeString,
		Required:    true,
	},
	"subdomain": {
		Description: "What subdomain should we use for your status page? This needs to be unique across our entire application, so choose carefully",
		Type:        schema.TypeString,
		Required:    true,
	},
	"custom_domain": {
		Description: "Do you want a custom domain on your status page? Add a CNAME record that points your domain to status.betteruptime.com. Example: `CNAME status.walmine.com statuspage.betteruptime.com`",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"min_incident_length": {
		Type:        schema.TypeInt,
		Optional:    true,
		Description: "If you don't want to display short incidents on your status page, this attribute is for you.",
	},
	"subscribable": {
		Type:        schema.TypeBool,
		Optional:    true,
		Description: "Do you want to allow users to subscribe to your status page changes?",
	},
	"hide_from_search_engines": {
		Type:        schema.TypeBool,
		Optional:    true,
		Description: "Hide your status page from search engines.",
	},
	"custom_css": {
		Description: "Unleash your inner designer and tweak our status page design to fit your branding.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"custom_javascript": {
		Description: "Add custom behavior to your status page. It is only allowed for status pages with a custom domain name.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"google_analytics_id": {
		Description: "Specify your own Google Analytics ID if you want to receive hits on your status page.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"announcement": {
		Description: "Add an announcement to your status page.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"announcement_embed_visible": {
		Type:        schema.TypeBool,
		Optional:    true,
		Description: strings.ReplaceAll(`Toggle this field if you want to show an announcement in your embed. You can embed the announcement using this snippet: **<script src="https://betteruptime.com/widgets/announcement.js" data-id="<SET STATUS_PAGE_ID>" async="async" type="text/javascript"></script>**`, "**", "`"),
	},
	"announcement_embed_link": {
		Description: "Point your embedded announcement to a specified URL.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"announcement_embed_css": {
		Description: "Modify the design of the announcement embed.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"password_enabled": {
		Description: "Do you want to enable password protection on your status page?",
		Type:        schema.TypeBool,
		Optional:    true,
	},
	"password": {
		Description: "Set a password of your status page (we won't store it as plaintext, promise). Required when password_enabled: true. We will set password_enabled: false automatically when you send us an empty password.",
		Type:        schema.TypeString,
		Optional:    true,
		Sensitive:   true,
	},
	"aggregate_state": {
		Description: "The overall status of this status page.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"created_at": {
		Description: "The time when this status page was created.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"updated_at": {
		Description: "The time when this status page was updated.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"design": {
		Description: "Choose between classic and modern status page design. Possible values: 'v1', 'v2'.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"theme": {
		Description: "Choose theme of your status page. Only applicable when design: v2. Possible values: 'light', 'dark'.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"layout": {
		Description: "Choose usual vertical layout or space-saving horizontal layout. Only applicable when design: v2. Possible values: 'vertical', 'horizontal'.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"automatic_reports": {
		Description: "Generate automatic reports when your services go down",
		Type:        schema.TypeBool,
		Optional:    true,
	},
}

func newStatusPageResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: statusPageCreate,
		ReadContext:   statusPageRead,
		UpdateContext: statusPageUpdate,
		DeleteContext: statusPageDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "https://betterstack.com/docs/uptime/api/status-pages/",
		Schema:      statusPageSchema,
	}
}

type statusPage struct {
	History                  *int    `json:"history,omitempty"`
	CompanyName              *string `json:"company_name,omitempty"`
	CompanyURL               *string `json:"company_url,omitempty"`
	ContactURL               *string `json:"contact_url,omitempty"`
	LogoURL                  *string `json:"logo_remote_url,omitempty"`
	Timezone                 *string `json:"timezone,omitempty"`
	Subdomain                *string `json:"subdomain,omitempty"`
	CustomDomain             *string `json:"custom_domain,omitempty"`
	MinIncidentLength        *int    `json:"min_incident_length,omitempty"`
	Subscribable             *bool   `json:"subscribable,omitempty"`
	HideFromSearchEngines    *bool   `json:"hide_from_search_engines,omitempty"`
	CustomCSS                *string `json:"custom_css,omitempty"`
	CustomJavaScript         *string `json:"custom_javascript,omitempty"`
	GoogleAnalyticsID        *string `json:"google_analytics_id,omitempty"`
	Announcement             *string `json:"announcement,omitempty"`
	AnnouncementEmbedVisible *bool   `json:"announcement_embed_visible,omitempty"`
	AnnouncementEmbedLink    *string `json:"announcement_embed_link,omitempty"`
	AnnouncementEmbedCSS     *string `json:"announcement_embed_css,omitempty"`
	PasswordEnabled          *bool   `json:"password_enabled,omitempty"`
	Password                 *string `json:"password,omitempty"`
	AggregateState           *string `json:"aggregate_state,omitempty"`
	CreatedAt                *string `json:"created_at,omitempty"`
	UpdatedAt                *string `json:"updated_at,omitempty"`
	Design                   *string `json:"design,omitempty"`
	Theme                    *string `json:"theme,omitempty"`
	Layout                   *string `json:"layout,omitempty"`
	AutomaticReports         *bool   `json:"automatic_reports,omitempty"`
}

type statusPageHTTPResponse struct {
	Data struct {
		ID         string     `json:"id"`
		Attributes statusPage `json:"attributes"`
	} `json:"data"`
}

func statusPageRef(in *statusPage) []struct {
	k string
	v interface{}
} {
	// TODO:  if reflect.TypeOf(in).NumField() != len([]struct)
	return []struct {
		k string
		v interface{}
	}{
		{k: "history", v: &in.History},
		{k: "company_name", v: &in.CompanyName},
		{k: "company_url", v: &in.CompanyURL},
		{k: "contact_url", v: &in.ContactURL},
		{k: "logo_url", v: &in.LogoURL},
		{k: "timezone", v: &in.Timezone},
		{k: "subdomain", v: &in.Subdomain},
		{k: "custom_domain", v: &in.CustomDomain},
		{k: "min_incident_length", v: &in.MinIncidentLength},
		{k: "subscribable", v: &in.Subscribable},
		{k: "hide_from_search_engines", v: &in.HideFromSearchEngines},
		{k: "custom_css", v: &in.CustomCSS},
		{k: "custom_javascript", v: &in.CustomJavaScript},
		{k: "google_analytics_id", v: &in.GoogleAnalyticsID},
		{k: "announcement", v: &in.Announcement},
		{k: "announcement_embed_visible", v: &in.AnnouncementEmbedVisible},
		{k: "announcement_embed_link", v: &in.AnnouncementEmbedLink},
		{k: "announcement_embed_css", v: &in.AnnouncementEmbedCSS},
		{k: "password_enabled", v: &in.PasswordEnabled},
		{k: "password", v: &in.Password},
		{k: "aggregate_state", v: &in.AggregateState},
		{k: "created_at", v: &in.CreatedAt},
		{k: "updated_at", v: &in.UpdatedAt},
		{k: "design", v: &in.Design},
		{k: "theme", v: &in.Theme},
		{k: "layout", v: &in.Layout},
		{k: "automatic_reports", v: &in.AutomaticReports},
	}
}
func statusPageCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in statusPage
	for _, e := range statusPageRef(&in) {
		load(d, e.k, e.v)
	}
	var out statusPageHTTPResponse
	if err := resourceCreate(ctx, meta, "/api/v2/status-pages", &in, &out); err != nil {
		return err
	}
	d.SetId(out.Data.ID)
	// Set password from user input since it's not included in the API response
	var derr diag.Diagnostics
	if err := d.Set("password", in.Password); err != nil {
		derr = append(derr, diag.FromErr(err)[0])
	}
	return statusPageCopyAttrs(d, &out.Data.Attributes, derr)
}

func statusPageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var out statusPageHTTPResponse
	if err, ok := resourceRead(ctx, meta, fmt.Sprintf("/api/v2/status-pages/%s", url.PathEscape(d.Id())), &out); err != nil {
		return err
	} else if !ok {
		d.SetId("") // Force "create" on 404.
		return nil
	}
	return statusPageCopyAttrs(d, &out.Data.Attributes, nil)
}

func statusPageCopyAttrs(d *schema.ResourceData, in *statusPage, derr diag.Diagnostics) diag.Diagnostics {
	for _, e := range statusPageRef(in) {
		if e.k == "password" {
			// Skip copying password as it's never returned from the API
			continue
		}
		if err := d.Set(e.k, reflect.Indirect(reflect.ValueOf(e.v)).Interface()); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}
	return derr
}

func statusPageUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in statusPage
	var out policyHTTPResponse
	for _, e := range statusPageRef(&in) {
		if d.HasChange(e.k) {
			load(d, e.k, e.v)
		}
	}
	return resourceUpdate(ctx, meta, fmt.Sprintf("/api/v2/status-pages/%s", url.PathEscape(d.Id())), &in, &out)
}

func statusPageDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDelete(ctx, meta, fmt.Sprintf("/api/v2/status-pages/%s", url.PathEscape(d.Id())))
}
