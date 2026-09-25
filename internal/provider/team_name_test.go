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

// TestTeamNameAfterImport verifies an imported resource can keep team_name in its config: Read
// stores the team reported by the API, so the same team plans cleanly and a different one is
// still rejected.
func TestTeamNameAfterImport(t *testing.T) {
	server := newResourceServer(t, "/api/v2/urgency-groups", "1")
	defer server.Close()
	// An existing severity group, as returned by the API.
	server.Data.Store([]byte(`{"name":"example","sort_index":1,"team_name":"First team"}`))

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

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testProviderFactories(server.URL),
		Steps: []resource.TestStep{
			{
				Config:             withTeamName("First team"),
				ResourceName:       "betteruptime_severity_group.this",
				ImportState:        true,
				ImportStateId:      "1",
				ImportStatePersist: true,
			},
			{
				Config:   withTeamName("First team"),
				PlanOnly: true,
			},
			{
				Config:      withTeamName("Second team"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`team_name cannot be changed after resource is created`),
			},
		},
	})
}

// TestTeamNameAfterImportNotReportedByAPI verifies team_name is accepted after importing a resource
// whose API response doesn't include the team (status page groups), as there's nothing to compare.
func TestTeamNameAfterImportNotReportedByAPI(t *testing.T) {
	server := newResourceServer(t, "/api/v2/status-page-groups", "1")
	defer server.Close()
	server.Data.Store([]byte(`{"name":"example","sort_index":1}`))

	config := `
		provider "betteruptime" {
			api_token = "foo"
		}

		resource "betteruptime_status_page_group" "this" {
			name       = "example"
			sort_index = 1
			team_name  = "First team"
		}
		`

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testProviderFactories(server.URL),
		Steps: []resource.TestStep{
			{
				Config:             config,
				ResourceName:       "betteruptime_status_page_group.this",
				ImportState:        true,
				ImportStateId:      "1",
				ImportStatePersist: true,
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}

// TestTeamNameKeptAfterTeamRename verifies renaming a team doesn't break a config that still uses
// the old name: Read keeps the team_name already in state instead of taking the API's.
func TestTeamNameKeptAfterTeamRename(t *testing.T) {
	server := newResourceServer(t, "/api/v2/urgency-groups", "1")
	defer server.Close()

	config := `
		provider "betteruptime" {
			api_token = "foo"
		}

		resource "betteruptime_severity_group" "this" {
			name       = "example"
			sort_index = 1
			team_name  = "First team"
		}
		`

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testProviderFactories(server.URL),
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				PreConfig: func() {
					server.Data.Store([]byte(`{"name":"example","sort_index":1,"team_name":"Renamed team"}`))
				},
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}
