package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func newTeamMemberDataSource() *schema.Resource {
	s := make(map[string]*schema.Schema)
	for k, v := range teamMemberSchema {
		cp := *v
		switch k {
		case "email":
			cp.Required = true
			cp.Computed = false
			cp.ForceNew = false
		case "team_name":
			cp.ForceNew = false
			cp.DiffSuppressFunc = nil
		default:
			cp.Computed = true
			cp.Optional = false
			cp.Required = false
			cp.ForceNew = false
			cp.ValidateDiagFunc = nil
			cp.Default = nil
			cp.DefaultFunc = nil
			cp.DiffSuppressFunc = nil
		}
		s[k] = &cp
	}
	return &schema.Resource{
		ReadContext: teamMemberLookup,
		Description: "Team member lookup by email.",
		Schema:      s,
	}
}

func teamMemberLookup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	email := d.Get("email").(string)
	d.SetId(email)

	out, derr, ok := teamMemberReadByEmail(ctx, d, meta)
	if derr != nil {
		return derr
	}
	if !ok {
		return diag.Errorf("team member with email %q not found", email)
	}
	return teamMemberCopyAttrs(d, out)
}
