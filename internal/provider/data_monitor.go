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

func newMonitorDataSource() *schema.Resource {
	s := make(map[string]*schema.Schema)
	for k, v := range monitorSchema {
		cp := *v
		switch k {
		case "url":
			// keep required
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
		ReadContext: monitorLookup,
		Description: "Monitor lookup.",
		Schema:      s,
	}
}

type monitorPageHTTPResponse struct {
	Data []struct {
		ID         string  `json:"id"`
		Attributes monitor `json:"attributes"`
	} `json:"data"`
	Pagination struct {
		First string `json:"first"`
		Last  string `json:"last"`
		Prev  string `json:"prev"`
		Next  string `json:"next"`
	} `json:"pagination"`
}

func monitorLookup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fetch := func(page int) (*monitorPageHTTPResponse, error) {
		res, err := meta.(*client).Get(ctx, fmt.Sprintf("/api/v2/monitors?page=%d", page))
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
		var tr monitorPageHTTPResponse
		return &tr, json.Unmarshal(body, &tr)
	}
	url := d.Get("url").(string)
	page := 1
	for {
		res, err := fetch(page)
		if err != nil {
			return diag.FromErr(err)
		}
		for _, e := range res.Data {
			if *e.Attributes.URL == url {
				if d.Id() != "" {
					return diag.Errorf("duplicate")
				}
				d.SetId(e.ID)
				if derr := monitorCopyAttrs(d, &e.Attributes); derr != nil {
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
