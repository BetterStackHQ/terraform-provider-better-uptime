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

var roleSchema = map[string]*schema.Schema{
	"name": {
		Description: "The role to look up: a system role value (responder, member, team_lead, billing_admin, admin) or a custom role's name.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"role": {
		Description: "The role identifier: a system role (admin, billing_admin, team_lead, responder, member) or 'custom' for a custom role.",
		Type:        schema.TypeString,
		Computed:    true,
	},
}

type role struct {
	Name *string `json:"name,omitempty"`
	Role *string `json:"role,omitempty"`
}

type rolesPageHTTPResponse struct {
	Data []struct {
		ID         string `json:"id"`
		Attributes role   `json:"attributes"`
	} `json:"data"`
	Pagination struct {
		Next string `json:"next"`
	} `json:"pagination"`
}

func newRoleDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: roleLookup,
		Description: "Look up a team role by name to obtain its ID (for use with `betteruptime_team_member` or the change-role API).",
		Schema:      roleSchema,
	}
}

func roleLookup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	id, attrs, err := findRoleByName(ctx, meta, name)
	if err != nil {
		return diag.FromErr(err)
	}
	if id == "" {
		return diag.Errorf("No role found with name: %s", name)
	}
	d.SetId(id)
	roleValue := ""
	if attrs.Role != nil {
		roleValue = *attrs.Role
	}
	if err := d.Set("role", roleValue); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

// findRoleByName paginates GET /api/v2/roles (served by betterstack) and returns
// the id + attributes of the role whose `role` value (the system-role identifier,
// or "custom") or display name matches the argument.
func findRoleByName(ctx context.Context, meta interface{}, name string) (string, *role, error) {
	baseURL := meta.(*client).BetterStackBaseURL()
	page := 1
	for {
		res, err := meta.(*client).GetWithBaseURL(ctx, baseURL, fmt.Sprintf("/api/v2/roles?page=%d", page))
		if err != nil {
			return "", nil, err
		}
		body, readErr := io.ReadAll(res.Body)
		_ = res.Body.Close()
		if res.StatusCode != http.StatusOK {
			return "", nil, fmt.Errorf("GET %s returned %d: %s", res.Request.URL.String(), res.StatusCode, string(body))
		}
		if readErr != nil {
			return "", nil, readErr
		}
		var pageResp rolesPageHTTPResponse
		if err := json.Unmarshal(body, &pageResp); err != nil {
			return "", nil, err
		}
		for _, e := range pageResp.Data {
			if (e.Attributes.Role != nil && *e.Attributes.Role == name) ||
				(e.Attributes.Name != nil && *e.Attributes.Name == name) {
				attrs := e.Attributes
				return e.ID, &attrs, nil
			}
		}
		if pageResp.Pagination.Next == "" {
			return "", nil, nil
		}
		page++
	}
}
