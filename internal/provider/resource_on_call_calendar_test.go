package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceOnCallCalendar(t *testing.T) {
	id := "123"
	name := "Test Calendar"
	updatedName := "Updated Calendar"

	server := newResourceServer(t, "/api/v2/on-calls", id)
	defer server.Close()

	resource.UnitTest(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			// Step 1 - create
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_on_call_calendar" "test" {
					name = "%s"
				}
				`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "id", id),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "name", name),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "default_calendar", "false"),
				),
			},
			// Step 2 - update
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_on_call_calendar" "test" {
					name = "%s"
				}
				`, updatedName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betteruptime_on_call_calendar.test", "id"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "name", updatedName),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "default_calendar", "false"),
				),
			},
			// Step 3 - import
			{
				ResourceName:      "betteruptime_on_call_calendar.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
