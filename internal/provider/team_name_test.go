package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TestTeamNameCannotBeChangedAfterCreate verifies team_name handling after a resource exists:
// clearing it is a silent no-op (no plan change), while changing it to a different, non-empty
// value fails with a helpful error. team_name is only used when creating a resource with a
// global token.
func TestTeamNameCannotBeChangedAfterCreate(t *testing.T) {
	server := newResourceServer(t, "/api/v2/urgency-groups", "1")
	defer server.Close()

	withTeamName := func(teamName string) string {
		return fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_severity_group" "this" {
					name       = "example"
					sort_index = 1
					team_name  = "%s"
				}
				`, teamName)
	}
	withoutTeamName := `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_severity_group" "this" {
					name       = "example"
					sort_index = 1
				}
				`

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			// Step 1 - create in a team.
			{
				Config: withTeamName("First team"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_severity_group.this", "team_name", "First team"),
				),
			},
			// Step 2 - clearing team_name is a no-op: no plan change, no error.
			{
				Config:   withoutTeamName,
				PlanOnly: true,
			},
			// Step 3 - changing team_name to a different team must fail.
			{
				Config:      withTeamName("Second team"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`team_name cannot be changed after resource is created`),
			},
		},
	})
}
