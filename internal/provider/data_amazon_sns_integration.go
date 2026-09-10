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

func newAmazonSnsIntegrationDataSource() *schema.Resource {
	s := make(map[string]*schema.Schema)
	for k, v := range amazonSnsIntegrationSchema {
		cp := *v
		switch k {
		case "name":
			cp.Required = true
			cp.Optional = false
			cp.Computed = false
		default:
			cp.Computed = true
			cp.Optional = false
			cp.Required = false
			cp.ValidateDiagFunc = nil
			cp.Default = nil
			cp.DefaultFunc = nil
			cp.DiffSuppressFunc = nil
			cp.MaxItems = 0
		}
		s[k] = &cp
	}
	delete(s, "team_name")
	// The resource's copy explains how to change this. Nothing here changes anything.
	if title := s["title_field"]; title != nil {
		title.Description = "The field describing how the incident title is extracted, if one is configured."
	}
	return &schema.Resource{
		ReadContext: amazonSnsIntegrationLookup,
		Description: "Amazon SNS integration lookup.",
		Schema:      s,
	}
}

type amazonSnsIntegrationsPageHTTPResponse struct {
	Data []struct {
		ID         string               `json:"id"`
		Attributes amazonSnsIntegration `json:"attributes"`
	} `json:"data"`
	Pagination struct {
		First string `json:"first"`
		Last  string `json:"last"`
		Prev  string `json:"prev"`
		Next  string `json:"next"`
	} `json:"pagination"`
}

func amazonSnsIntegrationLookup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fetch := func(page int) (*amazonSnsIntegrationsPageHTTPResponse, error) {
		res, err := meta.(*client).Get(ctx, fmt.Sprintf("/api/v2/amazon-sns?page=%d", page))
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
		var tr amazonSnsIntegrationsPageHTTPResponse
		return &tr, json.Unmarshal(body, &tr)
	}
	name := d.Get("name").(string)
	page := 1
	for {
		res, err := fetch(page)
		if err != nil {
			return diag.FromErr(err)
		}
		for _, e := range res.Data {
			if e.Attributes.Name != nil && *e.Attributes.Name == name {
				if d.Id() != "" {
					return diag.Errorf("There are multiple Amazon SNS integrations with the same name: %s", name)
				}
				d.SetId(e.ID)
				if derr := amazonSnsIntegrationCopyAttrs(d, &e.Attributes); derr != nil {
					return derr
				}
			}
		}
		page++
		if res.Pagination.Next == "" {
			// Erroring rather than returning an empty data source, the way the team member lookup
			// does: a silent no-match hands null attributes to whatever interpolated them, and the
			// failure surfaces somewhere else entirely.
			if d.Id() == "" {
				return diag.Errorf("Amazon SNS integration with name %q not found", name)
			}

			return nil
		}
	}
}
