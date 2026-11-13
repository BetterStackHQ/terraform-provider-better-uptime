package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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

	tooManyWorkingHours := ""
	for i := 0; i < 51; i++ {
		tooManyWorkingHours += `
				    working_hours {
				      day = "monday"
				    }`
	}

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
			// Step 4 - test more working hours than the API accepts
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
%s
				  }
				}
				`, name, tooManyWorkingHours),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Too many working_hours blocks`),
			},
			// Step 5 - test a blank time zone
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
				    timezone           = " "
				  }
				}
				`, name),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?s)timezone.*to not be an empty string or whitespace`),
			},
			// Step 6 - remove the working hours, keeping the rest of the rotation
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

func TestAccResourceOnCallCalendarWithWholeDayWorkingHours(t *testing.T) {
	id := "123"
	name := "Whole Day Calendar"
	rotationURL := "/api/v2/on-calls/123/rotation"
	rotationRequest := `{"end_rotations_at":"2027-01-05T00:00:00Z","rotation_interval":"day","rotation_length":1,"start_rotations_at":"2026-01-05T00:00:00Z","users":["user1@example.com"],"working_hours":[{"day":"sunday","end_time":"00:00","start_time":"00:00"}]}`
	// A rotation created with working hours and no time zone is stored as UTC and answers with it.
	rotationResponse := `{"users":["user1@example.com"],"rotation_length":1,"rotation_interval":"day","start_rotations_at":"2026-01-05T00:00:00Z","end_rotations_at":"2027-01-05T00:00:00Z","timezone":"UTC","effective_timezone":"UTC","working_hours":[{"day":"sunday","start_time":"00:00","end_time":"00:00"}]}`

	server := newResourceServer(t, "/api/v2/on-calls", id)

	server.ExpectRequest("POST", rotationURL, rotationRequest, http.StatusCreated, rotationResponse)
	server.ExpectRequest("GET", rotationURL, "", http.StatusOK, rotationResponse)

	defer server.Close()

	resource.UnitTest(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testProviderFactories(server.URL),
		Steps: []resource.TestStep{
			// Step 1 - a block with only a day covers that day from midnight to midnight
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
				      day = "sunday"
				    }
				  }
				}
				`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.timezone", "UTC"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.#", "1"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.0.day", "sunday"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.0.start_time", "00:00"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.0.end_time", "00:00"),
					server.TestCheckCalledRequest("POST", rotationURL, rotationRequest),
				),
			},
		},
	})
}

func TestAccResourceOnCallCalendarWithSplitDayWorkingHours(t *testing.T) {
	id := "123"
	name := "Split Day Calendar"
	rotationURL := "/api/v2/on-calls/123/rotation"
	rotationRequest := `{"end_rotations_at":"2027-01-05T00:00:00Z","rotation_interval":"day","rotation_length":1,"start_rotations_at":"2026-01-05T00:00:00Z","users":["user1@example.com"],"working_hours":[{"day":"monday","end_time":"12:00","start_time":"09:00"},{"day":"monday","end_time":"17:00","start_time":"12:30"}]}`
	rotationResponse := `{"users":["user1@example.com"],"rotation_length":1,"rotation_interval":"day","start_rotations_at":"2026-01-05T00:00:00Z","end_rotations_at":"2027-01-05T00:00:00Z","timezone":null,"effective_timezone":"UTC","working_hours":[{"day":"monday","start_time":"09:00","end_time":"12:00"},{"day":"monday","start_time":"12:30","end_time":"17:00"}]}`

	server := newResourceServer(t, "/api/v2/on-calls", id)

	server.ExpectRequest("POST", rotationURL, rotationRequest, http.StatusCreated, rotationResponse)
	server.ExpectRequest("GET", rotationURL, "", http.StatusOK, rotationResponse)

	defer server.Close()

	resource.UnitTest(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testProviderFactories(server.URL),
		Steps: []resource.TestStep{
			// Step 1 - two blocks for one day, split around a lunch break
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
				      start_time = "09:00"
				      end_time   = "12:00"
				    }
				    working_hours {
				      day        = "monday"
				      start_time = "12:30"
				      end_time   = "17:00"
				    }
				  }
				}
				`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.#", "2"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.0.day", "monday"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.0.start_time", "09:00"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.0.end_time", "12:00"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.1.day", "monday"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.1.start_time", "12:30"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.1.end_time", "17:00"),
					server.TestCheckCalledRequest("POST", rotationURL, rotationRequest),
				),
			},
		},
	})
}

func TestAccResourceOnCallCalendarUnsortedWorkingHours(t *testing.T) {
	id := "123"
	name := "Unsorted Calendar"
	rotationURL := "/api/v2/on-calls/123/rotation"
	rotationRequest := `{"end_rotations_at":"2027-01-05T00:00:00Z","rotation_interval":"day","rotation_length":1,"start_rotations_at":"2026-01-05T00:00:00Z","users":["user1@example.com"],"working_hours":[{"day":"monday","end_time":"17:00","start_time":"09:00"},{"day":"saturday","end_time":"00:00","start_time":"00:00"},{"day":"friday","end_time":"12:00","start_time":"09:00"},{"day":"friday","end_time":"17:00","start_time":"12:30"}]}`
	// The API stores the windows as a set and answers in canonical order, Sunday first and then by start time.
	rotationResponse := `{"users":["user1@example.com"],"rotation_length":1,"rotation_interval":"day","start_rotations_at":"2026-01-05T00:00:00Z","end_rotations_at":"2027-01-05T00:00:00Z","timezone":null,"effective_timezone":"UTC","working_hours":[{"day":"monday","start_time":"09:00","end_time":"17:00"},{"day":"friday","start_time":"09:00","end_time":"12:00"},{"day":"friday","start_time":"12:30","end_time":"17:00"},{"day":"saturday","start_time":"00:00","end_time":"00:00"}]}`

	server := newResourceServer(t, "/api/v2/on-calls", id)

	server.ExpectRequest("POST", rotationURL, rotationRequest, http.StatusCreated, rotationResponse)
	server.ExpectRequest("GET", rotationURL, "", http.StatusOK, rotationResponse)

	defer server.Close()

	resource.UnitTest(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testProviderFactories(server.URL),
		Steps: []resource.TestStep{
			// Step 1 - blocks written out of canonical order keep the order of the config, so the
			// step's follow-up plan is empty instead of wanting to reshuffle them on every run
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
				      start_time = "09:00"
				      end_time   = "17:00"
				    }
				    working_hours {
				      day = "saturday"
				    }
				    working_hours {
				      day        = "friday"
				      start_time = "09:00"
				      end_time   = "12:00"
				    }
				    working_hours {
				      day        = "friday"
				      start_time = "12:30"
				      end_time   = "17:00"
				    }
				  }
				}
				`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.#", "4"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.0.day", "monday"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.0.start_time", "09:00"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.0.end_time", "17:00"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.1.day", "saturday"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.1.start_time", "00:00"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.1.end_time", "00:00"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.2.day", "friday"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.2.start_time", "09:00"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.2.end_time", "12:00"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.3.day", "friday"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.3.start_time", "12:30"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.3.end_time", "17:00"),
					server.TestCheckCalledRequest("POST", rotationURL, rotationRequest),
				),
			},
		},
	})
}

func TestAccResourceOnCallCalendarWorkingHoursDrift(t *testing.T) {
	id := "123"
	name := "Drifting Calendar"
	rotationURL := "/api/v2/on-calls/123/rotation"
	rotationRequest := `{"end_rotations_at":"2027-01-05T00:00:00Z","rotation_interval":"day","rotation_length":1,"start_rotations_at":"2026-01-05T00:00:00Z","users":["user1@example.com"],"working_hours":[{"day":"monday","end_time":"17:00","start_time":"09:00"},{"day":"friday","end_time":"17:00","start_time":"09:00"}]}`
	rotationResponse := `{"users":["user1@example.com"],"rotation_length":1,"rotation_interval":"day","start_rotations_at":"2026-01-05T00:00:00Z","end_rotations_at":"2027-01-05T00:00:00Z","timezone":null,"effective_timezone":"UTC","working_hours":[{"day":"monday","start_time":"09:00","end_time":"17:00"},{"day":"friday","start_time":"09:00","end_time":"17:00"}]}`
	rotationResponseWithWindowAddedInUI := `{"users":["user1@example.com"],"rotation_length":1,"rotation_interval":"day","start_rotations_at":"2026-01-05T00:00:00Z","end_rotations_at":"2027-01-05T00:00:00Z","timezone":null,"effective_timezone":"UTC","working_hours":[{"day":"monday","start_time":"09:00","end_time":"17:00"},{"day":"wednesday","start_time":"09:00","end_time":"17:00"},{"day":"friday","start_time":"09:00","end_time":"17:00"}]}`

	server := newResourceServer(t, "/api/v2/on-calls", id)

	server.ExpectRequest("POST", rotationURL, rotationRequest, http.StatusCreated, rotationResponse)
	server.ExpectRequest("GET", rotationURL, "", http.StatusOK, rotationResponse)

	defer server.Close()

	config := testProviderBlock + fmt.Sprintf(`
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
	      start_time = "09:00"
	      end_time   = "17:00"
	    }
	    working_hours {
	      day        = "friday"
	      start_time = "09:00"
	      end_time   = "17:00"
	    }
	  }
	}
	`, name)

	resource.UnitTest(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testProviderFactories(server.URL),
		Steps: []resource.TestStep{
			// Step 1 - create the rotation, then answer reads with a window someone added in Better Stack.
			// The drift lands before this step's own post-apply refresh, so its plan is the one that has to see it.
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.#", "2"),
					server.ReplaceExpectedResponseAfterApply("GET", rotationURL, rotationResponseWithWindowAddedInUI),
				),
				ExpectNonEmptyPlan: true,
			},
			// Step 2 - the unchanged config puts the rotation back to its two windows
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.#", "2"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.0.day", "monday"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.1.day", "friday"),
					server.TestCheckCalledRequest("POST", rotationURL, rotationRequest),
					server.TestCheckCalledRequestWithout("POST", rotationURL, "wednesday"),
					// The create in step 1 satisfies the checks above on its own, so this pins the step's own POST.
					server.TestCheckCalledRequestCount("POST", rotationURL, 2),
					server.ReplaceExpectedResponseAfterApply("GET", rotationURL, rotationResponse),
				),
			},
		},
	})
}

func TestAccResourceOnCallCalendarLegacyRotationResponse(t *testing.T) {
	id := "123"
	name := "Legacy Calendar"
	rotationURL := "/api/v2/on-calls/123/rotation"
	rotationRequest := `{"end_rotations_at":"2027-01-05T00:00:00Z","rotation_interval":"day","rotation_length":1,"start_rotations_at":"2026-01-05T00:00:00Z","users":["user1@example.com"]}`
	// The shape the API answered with before it learned about time zones and working hours.
	rotationResponse := `{"users":["user1@example.com"],"rotation_length":1,"rotation_interval":"day","start_rotations_at":"2026-01-05T00:00:00Z","end_rotations_at":"2027-01-05T00:00:00Z"}`

	server := newResourceServer(t, "/api/v2/on-calls", id)

	server.ExpectRequest("POST", rotationURL, rotationRequest, http.StatusCreated, rotationResponse)
	server.ExpectRequest("GET", rotationURL, "", http.StatusOK, rotationResponse)

	defer server.Close()

	resource.UnitTest(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testProviderFactories(server.URL),
		Steps: []resource.TestStep{
			// Step 1 - a response without the new keys decodes, and the step's follow-up plan is empty
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
				  }
				}
				`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.working_hours.#", "0"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "on_call_rotation.0.timezone", ""),
				),
			},
		},
	})
}

func TestResourceOnCallCalendarSlackGroupHandle(t *testing.T) {
	server := newResourceServer(t, "/api/v2/on-calls", "123")
	server.ExpectRequest("GET", "/api/v2/on-calls/123/rotation", "", http.StatusNotFound, `{"errors":"Not found"}`)
	defer server.Close()

	config := func(slackAttributes string) string {
		return fmt.Sprintf(`
		provider "betteruptime" {
			api_token = "foo"
		}

		resource "betteruptime_on_call_calendar" "test" {
			name = "Payments"
			%s
		}
		`, slackAttributes)
	}
	lastBody := func(method string) string {
		server.mu.Lock()
		defer server.mu.Unlock()
		for i := len(server.CalledRequests) - 1; i >= 0; i-- {
			if request := server.CalledRequests[i]; request.Method == method && strings.HasPrefix(request.URL, "/api/v2/on-calls") && !strings.HasSuffix(request.URL, "/rotation") {
				return request.Body
			}
		}
		return ""
	}
	sent := func(method, fragment string, want bool) resource.TestCheckFunc {
		return func(*terraform.State) error {
			if body := lastBody(method); strings.Contains(body, fragment) != want {
				return fmt.Errorf("expected the last %s body %q to contain %q: %v", method, body, fragment, want)
			}
			return nil
		}
	}

	resource.UnitTest(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			// Step 1 - create with a handle; the mode is left to the API
			{
				Config: config(`slack_group_handle = "payments-oncall"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "slack_group_handle", "payments-oncall"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "use_existing_slack_group", "false"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "slack_groups.#", "0"),
					sent("POST", `"slack_group_handle":"payments-oncall"`, true),
					sent("POST", `use_existing_slack_group`, false),
				),
			},
			// Step 2 - switch to a group the customer already has in Slack
			{
				Config: config(`
					slack_group_handle       = "ops.rotation"
					use_existing_slack_group = true
				`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "slack_group_handle", "ops.rotation"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "use_existing_slack_group", "true"),
					sent("PATCH", `"slack_group_handle":"ops.rotation"`, true),
					sent("PATCH", `"use_existing_slack_group":true`, true),
				),
			},
			// Step 3 - leaving both out keeps what Better Stack has, so a calendar tagged in the dashboard plans no change
			{
				Config:   config(""),
				PlanOnly: true,
			},
			// Step 4 - an empty handle stops tagging
			{
				Config: config(`slack_group_handle = ""`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "slack_group_handle", ""),
					sent("PATCH", `"slack_group_handle":""`, true),
				),
			},
		},
	})
}

func TestResourceOnCallCalendarSlackGroupsRead(t *testing.T) {
	server := newResourceServer(t, "/api/v2/on-calls", "123")
	server.ExpectRequest("GET", "/api/v2/on-calls/123", "", http.StatusOK,
		`{"data":{"id":"123","attributes":{"name":"Payments","default_calendar":false,"slack_group_handle":"payments-oncall","use_existing_slack_group":false,`+
			`"slack_groups":[{"slack_workspace_id":"T1","status":"active","error":null},{"slack_workspace_id":"T2","status":"failed","error":"[bad_handle]"}]}}}`)
	server.ExpectRequest("GET", "/api/v2/on-calls/123/rotation", "", http.StatusNotFound, `{"errors":"Not found"}`)
	defer server.Close()

	resource.UnitTest(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_on_call_calendar" "test" {
					name               = "Payments"
					slack_group_handle = "payments-oncall"
				}
				`,
			},
			// Step 2 - the next refresh reads how each workspace took the handle
			{
				Config: `
				provider "betteruptime" {
					api_token = "foo"
				}

				resource "betteruptime_on_call_calendar" "test" {
					name               = "Payments"
					slack_group_handle = "payments-oncall"
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "slack_groups.#", "2"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "slack_groups.0.slack_workspace_id", "T1"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "slack_groups.0.status", "active"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "slack_groups.0.error", ""),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "slack_groups.1.status", "failed"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "slack_groups.1.error", "[bad_handle]"),
				),
			},
		},
	})
}

// A calendar whose Slack user groups are set up in the background: its GET answers pending for the first pendingReads
// reads after the create, then settled. A negative pendingReads fails the first read instead.
func newSlackGroupsServer(t *testing.T, pendingReads int, settled string) (*httptest.Server, *int32) {
	var reads int32
	calendar := func(groups string) string {
		return `{"data":{"id":"123","attributes":{"name":"Payments","default_calendar":false,"slack_group_handle":"payments-oncall",` +
			`"use_existing_slack_group":false,"slack_groups":` + groups + `}}}`
	}
	pending := `[{"slack_workspace_id":"T1","status":"pending","error":null}]`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v2/on-calls/123/rotation":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"errors":"Not found"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/on-calls":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(calendar(pending)))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/on-calls/123":
			read := int(atomic.AddInt32(&reads, 1))
			if pendingReads < 0 && read == 1 {
				w.WriteHeader(http.StatusUnprocessableEntity)
				_, _ = w.Write([]byte(`{"errors":"Try again"}`))
			} else if read <= pendingReads {
				_, _ = w.Write([]byte(calendar(pending)))
			} else {
				_, _ = w.Write([]byte(calendar(settled)))
			}
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v2/on-calls/123":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	return server, &reads
}

func shortenSlackGroupsSetup(t *testing.T, wait time.Duration) {
	savedWait, savedInterval := slackGroupsSetupWait, slackGroupsSetupPollInterval
	slackGroupsSetupWait, slackGroupsSetupPollInterval = wait, 10*time.Millisecond
	t.Cleanup(func() { slackGroupsSetupWait, slackGroupsSetupPollInterval = savedWait, savedInterval })
}

const slackGroupsConfig = `
provider "betteruptime" {
	api_token = "foo"
}

resource "betteruptime_on_call_calendar" "test" {
	name               = "Payments"
	slack_group_handle = "payments-oncall"
}
`

func TestResourceOnCallCalendarAwaitsSlackGroups(t *testing.T) {
	shortenSlackGroupsSetup(t, time.Minute)
	server, reads := newSlackGroupsServer(t, 2,
		`[{"slack_workspace_id":"T1","slack_workspace_name":"Acme","status":"active","error":null,`+
			`"unconnected_users":[{"id":"7","first_name":"Petr","last_name":"Heinz","email":"petr@example.com"}]},`+
			`{"slack_workspace_id":"T2","status":"reinstall_required","error":"Re-install the Slack integration to tag on-call schedules."}]`)
	defer server.Close()

	resource.UnitTest(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config: slackGroupsConfig,
				Check: resource.ComposeTestCheckFunc(
					// The apply's own reads, before the post-apply refresh reads it again.
					func(*terraform.State) error {
						if got := atomic.LoadInt32(reads); got < 3 {
							return fmt.Errorf("expected the apply to read the calendar until nothing was pending, read it %d times", got)
						}
						return nil
					},
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "slack_groups.#", "2"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "slack_groups.0.status", "active"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "slack_groups.0.slack_workspace_name", "Acme"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "slack_groups.0.unconnected_users.#", "1"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "slack_groups.0.unconnected_users.0.email", "petr@example.com"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "slack_groups.1.status", "reinstall_required"),
					resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "slack_groups.1.error", "Re-install the Slack integration to tag on-call schedules."),
				),
			},
		},
	})
}

func TestResourceOnCallCalendarStopsAwaitingSlackGroups(t *testing.T) {
	shortenSlackGroupsSetup(t, 50*time.Millisecond)
	server, _ := newSlackGroupsServer(t, 1000, `[]`)
	defer server.Close()

	resource.UnitTest(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config: slackGroupsConfig,
				Check:  resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "slack_groups.0.status", "pending"),
			},
		},
	})
}

func TestResourceOnCallCalendarKeepsTheCalendarWhenItsSlackGroupsCantBeRead(t *testing.T) {
	shortenSlackGroupsSetup(t, time.Minute)
	server, _ := newSlackGroupsServer(t, -1, `[{"slack_workspace_id":"T1","status":"active","error":null}]`)
	defer server.Close()

	resource.UnitTest(t, resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"betteruptime": func() (*schema.Provider, error) {
				return New(WithURL(server.URL)), nil
			},
		},
		Steps: []resource.TestStep{
			// Tainted, it would plan a replacement and fail the step's empty-plan check.
			{
				Config: slackGroupsConfig,
				Check:  resource.TestCheckResourceAttr("betteruptime_on_call_calendar.test", "id", "123"),
			},
		},
	})
}

func TestSlackGroupWarnings(t *testing.T) {
	groups := []onCallSlackGroup{
		{SlackWorkspaceID: strPtr("T1"), SlackWorkspaceName: strPtr("Acme"), Status: strPtr("active")},
		{SlackWorkspaceID: strPtr("T2"), SlackWorkspaceName: strPtr("Acme EU"), Status: strPtr("name_conflict"), Error: strPtr("This Slack handle is already taken. A Slack user group already uses @payments-oncall.")},
		{SlackWorkspaceID: strPtr("T3"), Status: strPtr("no_slack_users")},
		{SlackWorkspaceID: strPtr("T4"), SlackWorkspaceName: strPtr("Acme US"), Status: strPtr("pending")},
		{SlackWorkspaceID: strPtr("T5"), SlackWorkspaceName: strPtr("Acme"), Status: strPtr("active"), UnconnectedUsers: []onCallSlackUser{
			{FirstName: strPtr("Petr"), LastName: strPtr("Heinz"), Email: strPtr("petr@example.com")},
			{FirstName: strPtr("Kyle"), LastName: strPtr(""), Email: strPtr("kyle@example.com")},
		}},
	}

	got := []string{}
	for _, warning := range slackGroupWarnings("payments-oncall", &groups) {
		if warning.Severity != diag.Warning {
			t.Errorf("expected a warning, got %v", warning.Severity)
		}
		got = append(got, warning.Summary+" | "+warning.Detail)
	}

	want := []string{
		"Slack user group @payments-oncall doesn't work in the Acme EU Slack workspace | This Slack handle is already taken. A Slack user group already uses @payments-oncall.",
		// An API that doesn't name the workspace, nor word the status.
		"Slack user group @payments-oncall doesn't work in Slack workspace T3 | Its status is no_slack_users.",
		"Slack user group @payments-oncall is still being set up in the Acme US Slack workspace | Better Stack hadn't finished setting it up when the apply stopped waiting. The next refresh shows how Slack took it in slack_groups.",
		"Some responders won't be tagged by @payments-oncall in the Acme Slack workspace | They can connect by using the same e-mail in Slack and Better Stack, or in their account settings. Not connected yet: Petr Heinz (petr@example.com), Kyle (kyle@example.com).",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("warnings:\n got %q\nwant %q", got, want)
	}
}

func strPtr(s string) *string { return &s }
