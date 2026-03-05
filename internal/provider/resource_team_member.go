package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var teamMemberSchema = map[string]*schema.Schema{
	"team_name": {
		Description: "Used to specify the team the resource should be created in when using global tokens.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"email": {
		Description: "The email address of the team member to invite.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"role": {
		Description: "The role of the team member. Allowed values: responder, member, team_lead, billing_admin. Defaults to responder.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return d.Id() != ""
		},
	},
	"first_name": {
		Description: "The first name of the team member (available after invitation is accepted).",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"last_name": {
		Description: "The last name of the team member (available after invitation is accepted).",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"created_at": {
		Description: "The creation timestamp of the team member account.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"invited_at": {
		Description: "The timestamp when the invitation was sent.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"member_id": {
		Description: "The numeric ID of the team member. Empty for pending invitations.",
		Type:        schema.TypeString,
		Computed:    true,
	},
}

type teamMember struct {
	Email     *string `json:"email,omitempty"`
	Role      *string `json:"role,omitempty"`
	TeamName  *string `json:"team_name,omitempty"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	CreatedAt *string `json:"created_at,omitempty"`
	InvitedAt *string `json:"invited_at,omitempty"`
}

type teamMemberHTTPResponse struct {
	Message string `json:"message"`
	Data    struct {
		ID         string     `json:"id"`
		Type       string     `json:"type"`
		Attributes teamMember `json:"attributes"`
	} `json:"data"`
}

func newTeamMemberResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: teamMemberCreate,
		ReadContext:   teamMemberRead,
		DeleteContext: teamMemberDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				if err := d.Set("email", d.Id()); err != nil {
					return nil, err
				}
				return []*schema.ResourceData{d}, nil
			},
		},
		Description: "Allows managing **non-admin team members** using Terraform. Learn more about [inviting team members](https://betterstack.com/docs/uptime/inviting-team-members/).",
		Schema:      teamMemberSchema,
	}
}

func teamMemberCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var in teamMember
	load(d, "email", &in.Email)
	load(d, "role", &in.Role)
	load(d, "team_name", &in.TeamName)

	reqBody, err := json.Marshal(&in)
	if err != nil {
		return diag.FromErr(err)
	}
	baseURL := meta.(*client).BetterStackBaseURL()
	log.Printf("POST %s/api/v2/team-members: %s", baseURL, string(reqBody))
	res, err := meta.(*client).PostWithBaseURL(ctx, baseURL, "/api/v2/team-members", bytes.NewReader(reqBody))
	if err != nil {
		return diag.FromErr(err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, res.Body)
		_ = res.Body.Close()
	}()
	body, err := io.ReadAll(res.Body)
	// Accept both 201 (invitation created) and 200 (already a member)
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		return diag.Errorf("POST %s returned %d: %s", res.Request.URL.String(), res.StatusCode, string(body))
	}
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("POST /api/v2/team-members returned %d: %s", res.StatusCode, string(body))

	var out teamMemberHTTPResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(d.Get("email").(string))
	return teamMemberCopyAttrs(d, &out)
}

func teamMemberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	out, derr, ok := teamMemberReadByEmail(ctx, d, meta)
	if derr != nil {
		return derr
	}
	if !ok {
		d.SetId("") // Force "create" on 404.
		return nil
	}
	return teamMemberCopyAttrs(d, out)
}

func teamMemberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	email := d.Id()
	path := fmt.Sprintf("/api/v2/team-members?email=%s", url.QueryEscape(email))
	if v, ok := d.GetOk("team_name"); ok {
		path += fmt.Sprintf("&team_name=%s", url.QueryEscape(v.(string)))
	}

	return resourceDeleteWithBaseURL(ctx, meta, meta.(*client).BetterStackBaseURL(), path)
}

func teamMemberReadByEmail(ctx context.Context, d *schema.ResourceData, meta interface{}) (*teamMemberHTTPResponse, diag.Diagnostics, bool) {
	email := d.Id()
	path := fmt.Sprintf("/api/v2/team-members?email=%s", url.QueryEscape(email))
	if v, ok := d.GetOk("team_name"); ok {
		path += fmt.Sprintf("&team_name=%s", url.QueryEscape(v.(string)))
	}

	baseURL := meta.(*client).BetterStackBaseURL()
	log.Printf("GET %s%s", baseURL, path)
	res, err := meta.(*client).GetWithBaseURL(ctx, baseURL, path)
	if err != nil {
		return nil, diag.FromErr(err), false
	}
	defer func() {
		_, _ = io.Copy(io.Discard, res.Body)
		_ = res.Body.Close()
	}()

	if res.StatusCode == http.StatusNotFound {
		return nil, nil, false
	}
	body, err := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		return nil, diag.Errorf("GET %s returned %d: %s", res.Request.URL.String(), res.StatusCode, string(body)), false
	}
	if err != nil {
		return nil, diag.FromErr(err), false
	}
	log.Printf("GET %s returned %d: %s", res.Request.URL.String(), res.StatusCode, string(body))

	var out teamMemberHTTPResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, diag.FromErr(err), false
	}
	return &out, nil, true
}

func teamMemberCopyAttrs(d *schema.ResourceData, out *teamMemberHTTPResponse) diag.Diagnostics {
	var derr diag.Diagnostics
	set := func(k string, v interface{}) {
		if err := d.Set(k, v); err != nil {
			derr = append(derr, diag.FromErr(err)[0])
		}
	}
	if out.Data.Type == "team_member" {
		set("member_id", out.Data.ID)
	} else {
		set("member_id", nil)
	}
	a := out.Data.Attributes
	if a.Email != nil {
		set("email", *a.Email)
	}
	if a.Role != nil {
		set("role", *a.Role)
	}
	set("first_name", ptrToStr(a.FirstName))
	set("last_name", ptrToStr(a.LastName))
	set("created_at", ptrToStr(a.CreatedAt))
	set("invited_at", ptrToStr(a.InvitedAt))
	return derr
}

func ptrToStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
