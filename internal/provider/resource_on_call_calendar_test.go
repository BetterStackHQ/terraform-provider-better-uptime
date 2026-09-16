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
