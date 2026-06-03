package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// teamNameSchema returns the schema for the team_name attribute shared by every resource that can
// be created in a specific team using a global API token. The value is only used when the resource
// is created; afterwards any change is suppressed (see DiffSuppressFunc) and changing it to a
// different non-empty value is rejected (see validateTeamNameNotChanged).
func teamNameSchema() *schema.Schema {
	return &schema.Schema{
		Description: "Used to specify the team the resource should be created in when using global tokens. You can't update this value later.",
		Type:        schema.TypeString,
		Optional:    true,
		Default:     nil,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return d.Id() != ""
		},
	}
}

// validateTeamNameNotChanged rejects changing team_name to a different, non-empty value after a
// resource has been created. team_name only selects the team when the resource is created with a
// global API token and is ignored afterwards, so silently dropping the change ("No changes") was
// confusing. Clearing the value remains a silent no-op (the diff is suppressed in teamNameSchema).
//
// The change is detected from the raw config and state rather than HasChange, because the
// DiffSuppressFunc removes team_name from the computed diff before CustomizeDiff runs.
func validateTeamNameNotChanged(ctx context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if diff.Id() == "" {
		// The resource is being created — team_name is honored.
		return nil
	}

	config := diff.GetRawConfig()
	if config.IsNull() || !config.IsKnown() {
		return nil
	}
	newAttr := config.GetAttr("team_name")
	if newAttr.IsNull() || !newAttr.IsKnown() {
		return nil
	}
	newName := newAttr.AsString()
	if newName == "" {
		// Clearing team_name has no effect, so allow it (the diff is suppressed).
		return nil
	}

	oldName := ""
	if state := diff.GetRawState(); !state.IsNull() && state.IsKnown() {
		if oldAttr := state.GetAttr("team_name"); !oldAttr.IsNull() && oldAttr.IsKnown() {
			oldName = oldAttr.AsString()
		}
	}
	if newName != oldName {
		return fmt.Errorf("team_name cannot be changed after resource is created - it's only used to specify the team when creating it using a global token, you can safely remove it from your configuration now")
	}
	return nil
}
