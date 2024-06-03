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

func newSlackIntegrationDataSource() *schema.Resource {
	s := make(map[string]*schema.Schema)
	for k, v := range slackIntegrationSchema {
		cp := *v
		switch k {
		case "slack_channel_name":
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
		}
		s[k] = &cp
	}
	delete(s, "team_name")
	return &schema.Resource{
		ReadContext: slackIntegrationLookup,
		Description: "Slack integration lookup.",
		Schema:      s,
	}
}

type slackIntegrationsPageHTTPResponse struct {
	Data []struct {
		ID         string           `json:"id"`
		Attributes slackIntegration `json:"attributes"`
	} `json:"data"`
	Pagination struct {
		First string `json:"first"`
		Last  string `json:"last"`
		Prev  string `json:"prev"`
		Next  string `json:"next"`
	} `json:"pagination"`
}

func slackIntegrationLookup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fetch := func(page int) (*slackIntegrationsPageHTTPResponse, error) {
		res, err := meta.(*client).Get(ctx, fmt.Sprintf("/api/v2/slack-integrations?page=%d", page))
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
		var tr slackIntegrationsPageHTTPResponse
		return &tr, json.Unmarshal(body, &tr)
	}
	slackChannelName := d.Get("slack_channel_name").(string)
	page := 1
	for {
		res, err := fetch(page)
		if err != nil {
			return diag.FromErr(err)
		}
		for _, e := range res.Data {
			if *e.Attributes.SlackChannelName == slackChannelName {
				if d.Id() != "" {
					return diag.Errorf("There are multiple Slack integrations with the same slack_channel_name: %s", slackChannelName)
				}
				d.SetId(e.ID)
				if derr := slackIntegrationCopyAttrs(d, &e.Attributes); derr != nil {
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
