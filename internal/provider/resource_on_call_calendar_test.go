package provider

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccResourceOnCallCalendarWithRotation(t *testing.T) {
	id := "123"
	name := "Test Calendar"
	updatedName := "Updated Calendar"

	server := newResourceServer(t, "/api/v2/on-calls", id)

	server.ExpectRequest(
		"POST",
		"/api/v2/on-calls/123/rotation",
		`{"end_rotations_at":"2026-01-01T01:00:00+01:00","rotation_interval":"day","rotation_length":1,"start_rotations_at":"2025-01-01T01:00:00+01:00","users":["user1@example.com","user2@example.com"]}`,
		http.StatusCreated,
		`{"users":["user1@example.com","user2@example.com"],"rotation_length":1,"rotation_interval":"day","start_rotations_at":"2025-01-01T00:00:00Z","end_rotations_at":"2026-01-01T00:00:00Z","timezone":null,"effective_timezone":"UTC","working_hours":[]}`,
	)
	server.ExpectRequest(
		"GET",
		"/api/v2/on-calls/123/rotation",
		"",
		http.StatusOK,
		`{"users":["user1@example.com","user2@example.com"],"rotation_length":1,"rotation_interval":"day","start_rotations_at":"2025-01-01T00:00:00Z","end_rotations_at":"2026-01-01T00:00:00Z","timezone":null,"effective_timezone":"UTC","working_hours":[]}`,
	)

	defer server.Close()

	resource.UnitTest(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			// Step 1 - create with rotation
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_on_call_calendar" "test" {
					name = "%s"
					on_call_rotation {
						users = ["user1@example.com", "user2@example.com"]
						rotation_length = 1
						rotation_interval = "day"
						start_rotations_at = "2025-01-01T01:00:00+01:00"
						end_rotations_at = "2026-01-01T01:00:00+01:00"
					}
				}
				`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "id", id),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "name", name),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "default_calendar", "false"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.#", "1"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.users.#", "2"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.users.0", "user1@example.com"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.users.1", "user2@example.com"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.rotation_length", "1"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.rotation_interval", "day"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.start_rotations_at", "2025-01-01T00:00:00Z"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.end_rotations_at", "2026-01-01T00:00:00Z"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.timezone", ""),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.#", "0"),
				),
			},
			// Step 2 - test invalid rotation interval
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_on_call_calendar" "test" {
					name = "%s"
					on_call_rotation {
						users = ["user1@example.com"]
						rotation_length = 1
						rotation_interval = "invalid"
						start_rotations_at = "2025-01-01T00:00:00Z"
						end_rotations_at = "2026-01-01T00:00:00Z"
					}
				}
				`, name),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`expected on_call_rotation.0.rotation_interval to be one of \["hour" "day" "week"\], got invalid`),
			},
			// Step 3 - test invalid datetime format
			{
				Config: fmt.Sprintf(`
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_on_call_calendar" "test" {
					name = "%s"
					on_call_rotation {
						users = ["user1@example.com"]
						rotation_length = 1
						rotation_interval = "day"
						start_rotations_at = "2025-01-01"
						end_rotations_at = "2026-01-01T00:00:00Z"
					}
				}
				`, name),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`expected RFC 3339 datetime \(e\.g\. 2026-01-01T00:00:00Z\), got 2025-01-01`),
			},
			// Step 4 - update, remove rotation (kept managed in Better Stack)
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
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "id", id),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "name", updatedName),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "default_calendar", "false"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.#", "1"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.users.#", "2"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.users.0", "user1@example.com"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.users.1", "user2@example.com"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.rotation_length", "1"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.rotation_interval", "day"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.start_rotations_at", "2025-01-01T00:00:00Z"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.end_rotations_at", "2026-01-01T00:00:00Z"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.timezone", ""),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.#", "0"),
				),
			},
			// Step 5 - import
			{
				ResourceName:      "betteruptime_on_call_calendar.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceOnCallCalendarWithWorkingHours(t *testing.T) {
	id := "123"
	name := "Working Hours Calendar"
	rotationURL := "/api/v2/on-calls/123/rotation"
	rotationRequest := `{"end_rotations_at":"2027-01-05T00:00:00Z","rotation_interval":"day","rotation_length":1,"start_rotations_at":"2026-01-05T00:00:00Z","timezone":"Europe/Prague","users":["user1@example.com"],"working_hours":[{"day":"monday","end_time":"17:00","start_time":"09:00"},{"day":"friday","end_time":"06:00","start_time":"22:00:00"}]}`
	rotationRequestWithoutWorkingHours := `{"end_rotations_at":"2027-01-05T00:00:00Z","rotation_interval":"day","rotation_length":1,"start_rotations_at":"2026-01-05T00:00:00Z","timezone":"Europe/Prague","users":["user1@example.com"]}`
	rotationResponse := `{"users":["user1@example.com"],"rotation_length":1,"rotation_interval":"day","start_rotations_at":"2026-01-05T00:00:00Z","end_rotations_at":"2027-01-05T00:00:00Z","timezone":"Europe/Prague","effective_timezone":"Europe/Prague","working_hours":[{"day":"monday","start_time":"09:00","end_time":"17:00"},{"day":"friday","start_time":"22:00","end_time":"06:00"}]}`
	rotationResponseWithoutWorkingHours := `{"users":["user1@example.com"],"rotation_length":1,"rotation_interval":"day","start_rotations_at":"2026-01-05T00:00:00Z","end_rotations_at":"2027-01-05T00:00:00Z","timezone":"Europe/Prague","effective_timezone":"Europe/Prague","working_hours":[]}`

	server := newResourceServer(t, "/api/v2/on-calls", id)

	server.ExpectRequest("POST", rotationURL, rotationRequest, http.StatusCreated, rotationResponse)
	server.ExpectRequest("POST", rotationURL, rotationRequestWithoutWorkingHours, http.StatusCreated, rotationResponseWithoutWorkingHours)
	server.ExpectRequest("GET", rotationURL, "", http.StatusOK, rotationResponse)

	defer server.Close()

	resource.UnitTest(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testProviderFactories(server.URL),
		Steps: []resource.TestStep{
			// Step 1 - create with a time zone and working hours
			{
				Config: testProviderBlock + fmt.Sprintf(`
				resource "betteruptime_on_call_calendar" "test" {
				  name = "%s"
				  on_call_rotation {
				    users              = ["user1@example.com"]
				    rotation_length    = 1
				    rotation_interval  = "day"
				    start_rotations_at = "2026-01-05T00:00:00Z"
				    end_rotations_at   = "2027-01-05T00:00:00Z"
				    timezone           = "Europe/Prague"

				    working_hours {
				      day        = "monday"
				      start_time = "09:00"
				      end_time   = "17:00"
				    }
				    working_hours {
				      day        = "friday"
				      start_time = "22:00:00"
				      end_time   = "06:00"
				    }
				  }
				}
				`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.timezone", "Europe/Prague"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.#", "2"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.0.day", "monday"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.0.start_time", "09:00"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.0.end_time", "17:00"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.1.day", "friday"),
					// The config said 22:00:00, the API answers 22:00, and the step's follow-up plan is
					// only empty when the seconds are diffed away.
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.1.start_time", "22:00"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.1.end_time", "06:00"),
				),
			},
			// Step 2 - test invalid day
			{
				Config: testProviderBlock + fmt.Sprintf(`
				resource "betteruptime_on_call_calendar" "test" {
				  name = "%s"
				  on_call_rotation {
				    users              = ["user1@example.com"]
				    rotation_length    = 1
				    rotation_interval  = "day"
				    start_rotations_at = "2026-01-05T00:00:00Z"
				    end_rotations_at   = "2027-01-05T00:00:00Z"

				    working_hours {
				      day        = "funday"
				      start_time = "09:00"
				      end_time   = "17:00"
				    }
				  }
				}
				`, name),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`expected on_call_rotation.0.working_hours.0.day to be one of \["sunday" "monday" "tuesday" "wednesday" "thursday" "friday" "saturday"\], got funday`),
			},
			// Step 3 - test invalid time of day
			{
				Config: testProviderBlock + fmt.Sprintf(`
				resource "betteruptime_on_call_calendar" "test" {
				  name = "%s"
				  on_call_rotation {
				    users              = ["user1@example.com"]
				    rotation_length    = 1
				    rotation_interval  = "day"
				    start_rotations_at = "2026-01-05T00:00:00Z"
				    end_rotations_at   = "2027-01-05T00:00:00Z"

				    working_hours {
				      day        = "monday"
				      start_time = "9am"
				      end_time   = "17:00"
				    }
				  }
				}
				`, name),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`expected a time of day as HH:MM \(24-hour clock\), got 9am`),
			},
			// Step 4 - remove the working hours, keeping the rest of the rotation
			{
				Config: testProviderBlock + fmt.Sprintf(`
				resource "betteruptime_on_call_calendar" "test" {
				  name = "%s"
				  on_call_rotation {
				    users              = ["user1@example.com"]
				    rotation_length    = 1
				    rotation_interval  = "day"
				    start_rotations_at = "2026-01-05T00:00:00Z"
				    end_rotations_at   = "2027-01-05T00:00:00Z"
				    timezone           = "Europe/Prague"
				  }
				}
				`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.timezone", "Europe/Prague"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.#", "0"),
					server.TestCheckCalledRequest("POST", rotationURL, rotationRequestWithoutWorkingHours),
					server.ReplaceExpectedResponseAfterApply("GET", rotationURL, rotationResponseWithoutWorkingHours),
				),
			},
		},
	})
}
