package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func newOnCallCalendarDataSource() *schema.Resource {
	s := make(map[string]*schema.Schema)
	for k, v := range onCallCalendarSchema {
		cp := *v
		switch k {
		case "name":
			cp.Required = false
			cp.Optional = true
			cp.Computed = false
		default:
			cp.Computed = true
			cp.Optional = false
			cp.Required = false
			cp.ValidateDiagFunc = nil
			cp.Default = nil
			cp.DefaultFunc = nil
			cp.DiffSuppressFunc = nil
		}
		s[k] = &cp
	}
	return &schema.Resource{
		ReadContext: onCallCalendarDefaultOrLookup,
		Description: "On-call calendar lookup.",
		Schema:      s,
	}
}

func onCallCalendarDefaultOrLookup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    if name, ok := d.Get("name").(string); ok && name != "" {
        return onCallCalendarLookup(ctx, d, meta)
    }
    return onCallDefaultCalendar(ctx, d, meta)
}

type onCallDefaultCalendarHTTPResponse struct {
	Data struct {
		ID            string              `json:"id"`
		Attributes    onCallCalendar      `json:"attributes"`
		Relationships onCallRelationships `json:"relationships"`
	} `json:"data"`
	Included []onCallIncluded `json:"included"`
}

func onCallDefaultCalendar(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fetch := func() (*onCallDefaultCalendarHTTPResponse, error) {
		res, err := meta.(*client).Get(ctx, "/api/v2/on-calls/default")
		if err != nil {
			return nil, err
		}
		defer func() {
			// Keep-Alive.
			_, _ = io.Copy(io.Discard, res.Body)
			_ = res.Body.Close()
		}()
		body, err := io.ReadAll(res.Body)
		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("GET %s returned %d: %s", res.Request.URL.String(), res.StatusCode, string(body))
		}
		if err != nil {
			return nil, err
		}
		var tr onCallDefaultCalendarHTTPResponse
		return &tr, json.Unmarshal(body, &tr)
	}
	res, err := fetch()
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(res.Data.ID)
	if derr := onCallCalendarCopyAttrs(d, &res.Data.Attributes, res.Data.Relationships, res.Included); derr != nil {
		return derr
	}
	return nil
}

type onCallCalendarsPageHTTPResponse struct {
	Data []struct {
		ID            string              `json:"id"`
		Attributes    onCallCalendar      `json:"attributes"`
		Relationships onCallRelationships `json:"relationships"`
	} `json:"data"`
	Included []onCallIncluded `json:"included"`
	Pagination struct {
		First string `json:"first"`
		Last  string `json:"last"`
		Prev  string `json:"prev"`
		Next  string `json:"next"`
	} `json:"pagination"`
}

func onCallCalendarLookup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fetch := func(page int) (*onCallCalendarsPageHTTPResponse, error) {
		res, err := meta.(*client).Get(ctx, fmt.Sprintf("/api/v2/on-calls?page=%d", page))
		if err != nil {
			return nil, err
		}
		defer func() {
			// Keep-Alive.
			_, _ = io.Copy(io.Discard, res.Body)
			_ = res.Body.Close()
		}()
		body, err := io.ReadAll(res.Body)
		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("GET %s returned %d: %s", res.Request.URL.String(), res.StatusCode, string(body))
		}
		if err != nil {
			return nil, err
		}
		var tr onCallCalendarsPageHTTPResponse
		return &tr, json.Unmarshal(body, &tr)
	}
	calendarName := d.Get("name").(string)
	page := 1
	for {
		res, err := fetch(page)
		if err != nil {
			return diag.FromErr(err)
		}
		for _, e := range res.Data {
			if *e.Attributes.Name == calendarName {
				if d.Id() != "" {
					return diag.Errorf("There are multiple on-call calendars with the same name: %s", calendarName)
				}
				d.SetId(e.ID)
				if derr := onCallCalendarCopyAttrs(d, &e.Attributes, e.Relationships, res.Included); derr != nil {
					return derr
				}
			}
		}
		page++
		if res.Pagination.Next == "" {
			return nil
		}
	}
}
